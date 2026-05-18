package services

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"clawmemory/internal/models"

	"github.com/fsnotify/fsnotify"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	MaxMemoryContentLength = 50000
	SyncInterval           = 60 * time.Second
	DebounceInterval       = 2 * time.Second
	ChunkTargetChars       = 1600
	ChunkOverlapChars      = 320
	MinChunkChars          = 50
)

type memoryCategory string

const (
	categoryLongTerm     memoryCategory = "long_term"
	categoryDailyLog     memoryCategory = "daily_log"
	categorySessionLog   memoryCategory = "session_log"
	categoryConversation memoryCategory = "conversation"
	categoryKnowledge    memoryCategory = "knowledge"
	categoryUnknown      memoryCategory = "unknown"
)

var sensitivePatterns = []string{
	"api_key", "apikey", "api-key",
	"secret", "password", "passwd", "pwd",
	"token", "bearer", "auth",
	"private_key", "privatekey",
	"access_key", "accesskey",
	"credentials", "credential",
}

var sensitiveValuePatterns = []string{
	"sk-", "pk-", "Bearer ", "token_",
	"-----BEGIN", "-----END",
}

type OpenClawSyncService struct {
	db              *gorm.DB
	memService      *MemoryService
	syncInterval    time.Duration
	stopChan        chan struct{}
	running         bool
	mu              sync.Mutex
	keysMu          sync.RWMutex
	lastSyncTime    time.Time
	syncedKeys      map[string]string
	syncedCount     int
	lastError       string
	autoSyncEnabled bool
	skippedCount    int
	mode            string
	watcher         *fsnotify.Watcher
	debounceTimer   *time.Timer
	pendingEvents   map[string]bool
}

type SyncStatus struct {
	Running         bool      `json:"running"`
	Mode            string    `json:"mode"`
	WatchMode       string    `json:"watch_mode"`
	LastSyncTime    time.Time `json:"last_sync_time"`
	SyncedCount     int       `json:"synced_count"`
	SkippedCount    int       `json:"skipped_count"`
	LastError       string    `json:"last_error"`
	AutoSyncEnabled bool      `json:"auto_sync_enabled"`
	OpenClawFound   bool      `json:"openclaw_found"`
	LocalPaths      []string  `json:"local_paths,omitempty"`
	RemoteEndpoint  string    `json:"remote_endpoint,omitempty"`
	WatchedDirs     int       `json:"watched_dirs"`
}

type ConversationMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ConversationPushRequest struct {
	SessionID   string                `json:"session_id"`
	Title       string                `json:"title"`
	ProjectPath string                `json:"project_path"`
	AgentName   string                `json:"agent_name"`
	Messages    []ConversationMessage `json:"messages"`
	Summary     string                `json:"summary"`
	Platform    string                `json:"platform"`
	Visibility  string                `json:"visibility"`
}

var (
	globalSyncService *OpenClawSyncService
	syncServiceMu     sync.Mutex
)

func GetOpenClawSyncService(db *gorm.DB) *OpenClawSyncService {
	syncServiceMu.Lock()
	defer syncServiceMu.Unlock()

	if globalSyncService == nil || globalSyncService.db != db {
		if globalSyncService != nil && globalSyncService.stopChan != nil {
			close(globalSyncService.stopChan)
		}
		globalSyncService = &OpenClawSyncService{
			db:              db,
			memService:      NewMemoryService(db),
			syncInterval:    60 * time.Second,
			stopChan:        make(chan struct{}),
			syncedKeys:      make(map[string]string),
			autoSyncEnabled: true,
			mode:            "local",
		}
		globalSyncService.loadSyncedKeys()
		globalSyncService.detectMode()
	}
	return globalSyncService
}

func (s *OpenClawSyncService) detectMode() {
	if s.detectLocalOpenClaw() {
		s.mode = "local"
	} else {
		s.mode = "remote"
	}
}

func (s *OpenClawSyncService) detectLocalOpenClaw() bool {
	installed := DetectInstalledClients()
	return len(installed) > 0
}

func (s *OpenClawSyncService) loadSyncedKeys() {
	var memories []models.Memory
	_ = s.db.Where("source LIKE ? OR source LIKE ? OR source LIKE ? OR source LIKE ? OR source LIKE ? OR source LIKE ? OR source LIKE ? OR source LIKE ? OR source LIKE ? OR source LIKE ? OR source LIKE ? OR source LIKE ? OR source LIKE ?",
		"openclaw%", "trae%", "codebuddy%", "conversation%", "cursor%", "claude%", "windsurf%", "cline%", "continue%", "hermes%", "chunk:%", "jsonl:%", "ocidx:%").Select("key, value").Find(&memories).Error
	s.keysMu.Lock()
	for _, m := range memories {
		s.syncedKeys[m.Key] = m.Key
		if len(m.Value) > 0 {
			hash := sha256Hash(m.Value)
			s.syncedKeys["__hash__:"+hash] = hash
		}
	}
	s.keysMu.Unlock()
}

func (s *OpenClawSyncService) isKeySynced(key string) bool {
	s.keysMu.RLock()
	defer s.keysMu.RUnlock()
	_, exists := s.syncedKeys[key]
	return exists
}

func (s *OpenClawSyncService) markKeySynced(key string) {
	s.keysMu.Lock()
	s.syncedKeys[key] = key
	s.keysMu.Unlock()
}

func (s *OpenClawSyncService) markKeySkipped(key string) {
	s.keysMu.Lock()
	s.syncedKeys[key] = "__SKIPPED_SENSITIVE__"
	s.keysMu.Unlock()
}

func (s *OpenClawSyncService) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return
	}

	if s.mode == "local" && !s.detectLocalOpenClaw() {
		s.lastError = "No supported IDE directory found (OpenClaw/Trae CN/CodeBuddy CN)"
		return
	}

	s.running = true
	s.stopChan = make(chan struct{})
	s.pendingEvents = make(map[string]bool)

	s.localSync()

	if s.mode == "local" {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			s.lastError = fmt.Sprintf("fsnotify init failed: %v, falling back to polling", err)
			go s.periodicRescan()
			return
		}
		s.watcher = watcher

		s.addWatchDirs()

		go s.watchLoop()
		go s.periodicRescan()
	}
}

func (s *OpenClawSyncService) addWatchDirs() {
	if s.watcher == nil {
		return
	}
	dirs := GetAllSearchDirs()
	for _, dir := range dirs {
		if _, err := os.Stat(dir); err == nil {
			s.watcher.Add(dir)
			filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
				if err != nil || !info.IsDir() {
					return nil
				}
				s.watcher.Add(path)
				return nil
			})
		}
	}
}

func (s *OpenClawSyncService) watchLoop() {
	for {
		select {
		case <-s.stopChan:
			return
		case event, ok := <-s.watcher.Events:
			if !ok {
				return
			}
			if !s.autoSyncEnabled {
				continue
			}
			if event.Op&fsnotify.Create == fsnotify.Create || event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Rename == fsnotify.Rename {
				ext := strings.ToLower(filepath.Ext(event.Name))
				if ext == ".md" || ext == ".json" || ext == ".jsonl" || ext == ".db" || ext == ".sqlite" || ext == ".sqlite3" || ext == ".vscdb" {
					if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
						s.watcher.Add(event.Name)
						continue
					}
					s.mu.Lock()
					s.pendingEvents[event.Name] = true
					if s.debounceTimer != nil {
						s.debounceTimer.Stop()
					}
					s.debounceTimer = time.AfterFunc(DebounceInterval, func() {
						s.mu.Lock()
						events := s.pendingEvents
						s.pendingEvents = make(map[string]bool)
						s.mu.Unlock()
						s.processFileEvents(events)
					})
					s.mu.Unlock()
				}
			}
			if event.Op&fsnotify.Create == fsnotify.Create {
				if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
					s.watcher.Add(event.Name)
				}
			}
		case err, ok := <-s.watcher.Errors:
			if !ok {
				return
			}
			s.mu.Lock()
			s.lastError = err.Error()
			s.mu.Unlock()
		}
	}
}

func (s *OpenClawSyncService) processFileEvents(events map[string]bool) {
	for filePath := range events {
		previews := s.extractFromFile(filePath)
		if len(previews) == 0 {
			continue
		}

		recordSensitive := s.getRecordSensitiveSetting()
		var encryptor *Encryptor
		if recordSensitive {
			secretKey := s.getSecretKey()
			if secretKey == "" {
				recordSensitive = false
			} else {
				var err error
				encryptor, err = NewEncryptor(secretKey)
				if err != nil {
					recordSensitive = false
				}
			}
		}

		userID := s.getDefaultUserID()
		if userID == 0 {
			continue
		}

		chunkCount := 0
		for _, p := range previews {
			if s.isChunkHashSynced(p.ChunkHash) {
				continue
			}
			if s.isKeySynced(p.Key) {
				continue
			}
			if s.isLowQualityContent(p.Content) {
				continue
			}
			if !s.isSignificant(p) {
				continue
			}

			isSensitive := s.containsSensitiveInfo(p.Key) || s.containsSensitiveInfo(p.Content)
			isSensitiveFile := s.isSensitiveFile(p.FilePath)
			if (isSensitive || isSensitiveFile) && !recordSensitive {
				s.skippedCount++
				s.markKeySkipped(p.Key)
				continue
			}

			if len(p.Content) > MaxMemoryContentLength {
				p.Content = p.Content[:MaxMemoryContentLength]
			}

			value := p.Content
			source := p.Source
			isEncrypted := false
			if isSensitive && recordSensitive && encryptor != nil {
				encrypted, err := EncryptValue(encryptor, p.Content)
				if err != nil {
					continue
				}
				value = encrypted
				source = p.Source + ":encrypted"
				isEncrypted = true
			}

			platform := s.inferPlatformFromPath(p.FilePath)
			if p.Category != "" {
				_, _, catSource := s.categoryToLayerAndType(p.Category)
				if catSource != "" {
					source = catSource
					if isEncrypted {
						source = catSource + ":encrypted"
					}
				}
			}

			_, err := s.memService.Create(userID, map[string]interface{}{
				"key":          p.Key,
				"value":        value,
				"layer":        p.Layer,
				"source":       source,
				"memory_type":  s.inferMemoryType(p),
				"is_encrypted": isEncrypted,
				"platform":     platform,
				"source_agent": p.AgentName,
			})
			if err == nil {
				s.markKeySynced(p.Key)
				s.markChunkHashSynced(p.ChunkHash)
				s.syncedCount++
				chunkCount++
			}
		}

		if chunkCount > 0 {
			hash := s.computeFileHash(filePath)
			info, err := os.Stat(filePath)
			var size int64
			if err == nil {
				size = info.Size()
			}
			platform := s.inferPlatformFromPath(filePath)
			src := "unknown"
			if len(previews) > 0 {
				src = previews[0].Source
			}
			s.updateFileIndex(filePath, hash, size, chunkCount, src, platform)
		}
	}

	s.mu.Lock()
	s.lastSyncTime = time.Now()
	s.mu.Unlock()
}

func (s *OpenClawSyncService) extractFromFile(filePath string) []memoryPreview {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".md":
		return s.parseMarkdownFile(filePath)
	case ".db", ".sqlite", ".sqlite3":
		return s.extractSqliteMemories(filePath)
	case ".json":
		return s.parseJSONFile(filePath)
	case ".jsonl":
		return s.parseJSONLFile(filePath)
	case ".vscdb":
		return s.extractVscdbMemories(filePath)
	}
	return nil
}

func (s *OpenClawSyncService) periodicRescan() {
	ticker := time.NewTicker(SyncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !s.autoSyncEnabled {
				continue
			}
			s.addWatchDirs()
			s.localSync()
		case <-s.stopChan:
			return
		}
	}
}

func (s *OpenClawSyncService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	close(s.stopChan)
	if s.watcher != nil {
		s.watcher.Close()
		s.watcher = nil
	}
	if s.debounceTimer != nil {
		s.debounceTimer.Stop()
	}
	s.running = false
}

func (s *OpenClawSyncService) localSync() {
	s.lastSyncTime = time.Now()
	s.lastError = ""

	previews := s.readLocalDatabases()
	if len(previews) == 0 {
		return
	}

	recordSensitive := s.getRecordSensitiveSetting()
	var encryptor *Encryptor
	if recordSensitive {
		secretKey := s.getSecretKey()
		if secretKey == "" {
			s.lastError = "SECRET_KEY not configured, cannot encrypt sensitive content"
			recordSensitive = false
		} else {
			var err error
			encryptor, err = NewEncryptor(secretKey)
			if err != nil {
				s.lastError = "failed to init encryptor"
				return
			}
		}
	}

	newCount := 0
	skipped := 0
	fileChunkCounts := make(map[string]int)

	for _, p := range previews {
		if s.isChunkHashSynced(p.ChunkHash) {
			skipped++
			continue
		}

		if s.isKeySynced(p.Key) {
			skipped++
			continue
		}

		if s.isLowQualityContent(p.Content) {
			skipped++
			continue
		}

		if !s.isSignificant(p) {
			skipped++
			continue
		}

		isSensitive := s.containsSensitiveInfo(p.Key) || s.containsSensitiveInfo(p.Content)
		isSensitiveFile := s.isSensitiveFile(p.FilePath)

		if (isSensitive || isSensitiveFile) && !recordSensitive {
			skipped++
			s.markKeySkipped(p.Key)
			continue
		}

		if len(p.Content) > MaxMemoryContentLength {
			p.Content = p.Content[:MaxMemoryContentLength]
		}

		userID := s.getDefaultUserID()
		if userID == 0 {
			s.lastError = "no user found"
			continue
		}

		value := p.Content
		source := p.Source
		isEncrypted := false
		if isSensitive && recordSensitive && encryptor != nil {
			encrypted, err := EncryptValue(encryptor, p.Content)
			if err != nil {
				skipped++
				continue
			}
			value = encrypted
			source = p.Source + ":encrypted"
			isEncrypted = true
		}

		platform := s.inferPlatformFromPath(p.FilePath)
		if p.Category != "" {
			_, _, catSource := s.categoryToLayerAndType(p.Category)
			if catSource != "" {
				source = catSource
				if isEncrypted {
					source = catSource + ":encrypted"
				}
			}
		}

		_, err := s.memService.Create(userID, map[string]interface{}{
			"key":          p.Key,
			"value":        value,
			"layer":        p.Layer,
			"source":       source,
			"memory_type":  s.inferMemoryType(p),
			"is_encrypted": isEncrypted,
			"platform":     platform,
			"source_agent": p.AgentName,
		})
		if err != nil {
			continue
		}

		s.markKeySynced(p.Key)
		s.markChunkHashSynced(p.ChunkHash)
		fileChunkCounts[p.FilePath]++
		newCount++
	}

	for filePath, count := range fileChunkCounts {
		hash := s.computeFileHash(filePath)
		info, err := os.Stat(filePath)
		var size int64
		if err == nil {
			size = info.Size()
		}
		platform := s.inferPlatformFromPath(filePath)
		source := "unknown"
		previews := s.extractFromFile(filePath)
		if len(previews) > 0 {
			source = previews[0].Source
		}
		s.updateFileIndex(filePath, hash, size, count, source, platform)
	}

	s.syncedCount += newCount
	s.skippedCount += skipped
}

func (s *OpenClawSyncService) isChunkHashSynced(hash string) bool {
	if hash == "" {
		return false
	}
	s.keysMu.RLock()
	defer s.keysMu.RUnlock()
	_, exists := s.syncedKeys["__hash__:"+hash]
	return exists
}

func (s *OpenClawSyncService) markChunkHashSynced(hash string) {
	if hash == "" {
		return
	}
	s.keysMu.Lock()
	s.syncedKeys["__hash__:"+hash] = hash
	s.keysMu.Unlock()
}

func (s *OpenClawSyncService) PushConversation(userID uint, req ConversationPushRequest) (int, error) {
	created := 0
	platform := req.Platform
	if platform == "" {
		platform = req.AgentName
	}
	visibility := req.Visibility
	if visibility == "" {
		visibility = "private"
	}

	if req.SessionID != "" && req.Summary != "" {
		key := "conversation-" + req.AgentName + "-" + req.SessionID
		if !s.isKeySynced(key) {
			value := req.Summary
			if req.Title != "" {
				value = req.Title + "\n\n" + value
			}
			if req.ProjectPath != "" {
				value = "Project: " + req.ProjectPath + "\n" + value
			}

			if !s.isLowQualityContent(value) {
				_, err := s.memService.Create(userID, map[string]interface{}{
					"key":          key,
					"value":        value,
					"layer":        "episodic",
					"source":       "conversation-" + req.AgentName,
					"memory_type":  "episodic",
					"platform":     platform,
					"source_agent": req.AgentName,
					"visibility":   visibility,
				})
				if err == nil {
					s.markKeySynced(key)
					created++
				}
			}
		}
	}

	for _, msg := range req.Messages {
		if msg.Content == "" || len(msg.Content) < 10 {
			continue
		}
		if s.isLowQualityContent(msg.Content) {
			continue
		}

		key := "conversation-" + req.AgentName + "-" + msg.Role + "-" + md5Hash(msg.Content)[:12]
		if s.isKeySynced(key) {
			continue
		}

		value := msg.Content
		if len(value) > MaxMemoryContentLength {
			value = value[:MaxMemoryContentLength]
		}

		layer := "episodic"
		memoryType := "episodic"
		if msg.Role == "assistant" {
			layer = "semantic"
			memoryType = "knowledge"
		}

		_, err := s.memService.Create(userID, map[string]interface{}{
			"key":          key,
			"value":        value,
			"layer":        layer,
			"source":       "conversation-" + req.AgentName + "-" + msg.Role,
			"memory_type":  memoryType,
			"platform":     platform,
			"source_agent": req.AgentName,
			"visibility":   visibility,
		})
		if err == nil {
			s.markKeySynced(key)
			created++
		}
	}

	s.syncedCount += created
	return created, nil
}

func (s *OpenClawSyncService) getRecordSensitiveSetting() bool {
	userID := s.getDefaultUserID()
	if userID == 0 {
		return false
	}
	svc := NewSettingsService(s.db)
	val, err := svc.GetByKey(userID, "record_sensitive_content")
	if err != nil {
		return false
	}
	if b, ok := val.(bool); ok {
		return b
	}
	return false
}

func (s *OpenClawSyncService) getSecretKey() string {
	return GetEncryptionKey()
}

func (s *OpenClawSyncService) containsSensitiveInfo(content string) bool {
	contentLower := strings.ToLower(content)

	for _, pattern := range sensitivePatterns {
		if strings.Contains(contentLower, pattern) {
			for _, vp := range sensitiveValuePatterns {
				if strings.Contains(content, vp) || strings.Contains(contentLower, strings.ToLower(vp)) {
					return true
				}
			}
		}
	}

	if strings.Contains(contentLower, "api_key") || strings.Contains(contentLower, "apikey") {
		if len(content) > 20 {
			potentialKey := content
			for _, vp := range sensitiveValuePatterns {
				if strings.Contains(potentialKey, vp) {
					return true
				}
			}
		}
	}

	return false
}

func (s *OpenClawSyncService) isSensitiveFile(filePath string) bool {
	fileName := strings.ToLower(filepath.Base(filePath))
	sensitiveFiles := []string{
		".env", "credentials", "secrets",
		"config.json", "settings.json",
		"api_key", "apikey", "private",
	}

	for _, sf := range sensitiveFiles {
		if strings.Contains(fileName, sf) {
			return true
		}
	}

	return false
}

func (s *OpenClawSyncService) getDefaultUserID() uint {
	var user models.User
	if err := s.db.Where("role = ?", "admin").First(&user).Error; err == nil {
		return user.ID
	}
	if err := s.db.Order("id ASC").First(&user).Error; err != nil {
		return 0
	}
	return user.ID
}

type memoryPreview struct {
	Key       string
	Content   string
	Layer     string
	Source    string
	FilePath  string
	AgentName string
	StartLine int
	EndLine   int
	ChunkHash string
	Category  memoryCategory
}

func (s *OpenClawSyncService) readLocalDatabases() []memoryPreview {
	searchDirs := s.getLocalSearchDirs()
	var allPreviews []memoryPreview

	for _, dir := range searchDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}
		previews := s.extractFromDir(dir)
		allPreviews = append(allPreviews, previews...)
	}

	return allPreviews
}

func (s *OpenClawSyncService) getLocalSearchDirs() []string {
	allDirs := GetAllSearchDirs()
	var existing []string
	for _, d := range allDirs {
		if _, err := os.Stat(d); err == nil {
			existing = append(existing, d)
		}
	}
	return existing
}

func (s *OpenClawSyncService) extractFromDir(dir string) []memoryPreview {
	var previews []memoryPreview

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		if s.isFileIndexUpToDate(path, info) {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".md":
			previews = append(previews, s.parseMarkdownFile(path)...)
		case ".db", ".sqlite", ".sqlite3":
			previews = append(previews, s.extractSqliteMemories(path)...)
		case ".json":
			previews = append(previews, s.parseJSONFile(path)...)
		case ".jsonl":
			previews = append(previews, s.parseJSONLFile(path)...)
		case ".vscdb":
			previews = append(previews, s.extractVscdbMemories(path)...)
		}

		return nil
	})

	return previews
}

func (s *OpenClawSyncService) isFileIndexUpToDate(path string, info os.FileInfo) bool {
	var idx models.FileSyncIndex
	err := s.db.Where("file_path = ?", path).First(&idx).Error
	if err != nil {
		return false
	}
	if idx.FileHash == "" {
		return false
	}
	hash := s.computeFileHash(path)
	if hash == idx.FileHash && info.Size() == idx.FileSize {
		return true
	}
	s.removeStaleChunks(path, idx.FileHash)
	return false
}

func (s *OpenClawSyncService) computeFileHash(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func (s *OpenClawSyncService) updateFileIndex(path string, hash string, size int64, chunkCount int, source string, platform string) {
	var idx models.FileSyncIndex
	err := s.db.Where("file_path = ?", path).First(&idx).Error
	now := time.Now()
	if err != nil {
		idx = models.FileSyncIndex{
			FilePath:   path,
			FileHash:   hash,
			FileSize:   size,
			ModTime:    now.Unix(),
			Source:     source,
			ChunkCount: chunkCount,
			Platform:   platform,
			SyncedAt:   now,
		}
		s.db.Create(&idx)
	} else {
		s.db.Model(&idx).Updates(map[string]interface{}{
			"file_hash":   hash,
			"file_size":   size,
			"mod_time":    now.Unix(),
			"source":      source,
			"chunk_count": chunkCount,
			"platform":    platform,
			"synced_at":   now,
		})
	}
}

func (s *OpenClawSyncService) removeStaleChunks(path string, oldHash string) {
	pathHash := sha256Hash(path)[:8]
	s.db.Where("key LIKE ?", "chunk:"+pathHash+":%").Delete(&models.Memory{})
	s.db.Where("key LIKE ?", "ocidx:%").Where("source_agent = ?", path).Delete(&models.Memory{})
}

func (s *OpenClawSyncService) classifyFilePath(path string) memoryCategory {
	base := strings.ToLower(filepath.Base(path))
	dir := strings.ToLower(filepath.Dir(path))

	if base == "memory.md" {
		return categoryLongTerm
	}

	datePattern := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}\.md$`)
	if datePattern.MatchString(base) && strings.Contains(dir, "memory") {
		return categoryDailyLog
	}

	slugPattern := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}-.+\.md$`)
	if slugPattern.MatchString(base) {
		return categorySessionLog
	}

	if strings.Contains(dir, "session") || strings.Contains(dir, "conversation") {
		return categoryConversation
	}

	if strings.Contains(dir, "memory") || strings.Contains(dir, "workspace") {
		return categoryKnowledge
	}

	return categoryUnknown
}

func (s *OpenClawSyncService) categoryToLayerAndType(cat memoryCategory) (layer string, memType string, source string) {
	switch cat {
	case categoryLongTerm:
		return "semantic", "knowledge", "openclaw-memory-md"
	case categoryDailyLog:
		return "episodic", "episodic", "openclaw-daily-log"
	case categorySessionLog:
		return "episodic", "episodic", "openclaw-session-archive"
	case categoryConversation:
		return "episodic", "episodic", "openclaw-conversation"
	case categoryKnowledge:
		return "semantic", "knowledge", "openclaw-knowledge"
	default:
		return "semantic", "knowledge", "openclaw-markdown"
	}
}

func (s *OpenClawSyncService) parseMarkdownFile(path string) []memoryPreview {
	data, err := os.ReadFile(path)
	if err != nil || len(data) == 0 {
		return nil
	}

	content := string(data)
	cat := s.classifyFilePath(path)
	layer, _, sourcePrefix := s.categoryToLayerAndType(cat)
	agentName := s.inferAgentFromPath(path)

	var previews []memoryPreview

	type section struct {
		title     string
		level     int
		startLine int
		content   string
	}

	var sections []section
	lines := strings.Split(content, "\n")
	currentSection := section{level: 0, startLine: 1}

	for i, line := range lines {
		headingMatch := regexp.MustCompile(`^(#{1,6})\s+(.+)$`)
		if m := headingMatch.FindStringSubmatch(line); m != nil {
			if currentSection.content != "" || currentSection.title != "" {
				sections = append(sections, currentSection)
			}
			level := len(m[1])
			title := strings.TrimSpace(m[2])
			currentSection = section{
				title:     title,
				level:     level,
				startLine: i + 1,
				content:   "",
			}
		} else {
			currentSection.content += line + "\n"
		}
	}
	if currentSection.content != "" || currentSection.title != "" {
		sections = append(sections, currentSection)
	}

	if len(sections) == 0 {
		if len(content) > MinChunkChars {
			chunks := s.slidingWindowChunk(content, 1, len(lines))
			for _, ch := range chunks {
				chunkHash := sha256Hash(ch.text)
				key := s.buildChunkKey(path, ch.startLine, ch.endLine, chunkHash)
				previews = append(previews, memoryPreview{
					Key:       key,
					Content:   strings.TrimSpace(ch.text),
					Layer:     layer,
					Source:    sourcePrefix,
					FilePath:  path,
					AgentName: agentName,
					StartLine: ch.startLine,
					EndLine:   ch.endLine,
					ChunkHash: chunkHash,
					Category:  cat,
				})
			}
		}
		return previews
	}

	for _, sec := range sections {
		trimmedContent := strings.TrimSpace(sec.content)
		if trimmedContent == "" {
			continue
		}

		if len(trimmedContent) > ChunkTargetChars*2 {
			chunks := s.slidingWindowChunk(sec.content, sec.startLine, sec.startLine+strings.Count(sec.content, "\n"))
			for _, ch := range chunks {
				chunkHash := sha256Hash(ch.text)
				sectionPrefix := ""
				if sec.title != "" {
					sectionPrefix = sec.title + ": "
				}
				key := s.buildChunkKey(path, ch.startLine, ch.endLine, chunkHash)
				previews = append(previews, memoryPreview{
					Key:       key,
					Content:   sectionPrefix + strings.TrimSpace(ch.text),
					Layer:     layer,
					Source:    sourcePrefix,
					FilePath:  path,
					AgentName: agentName,
					StartLine: ch.startLine,
					EndLine:   ch.endLine,
					ChunkHash: chunkHash,
					Category:  cat,
				})
			}
		} else {
			endLine := sec.startLine + strings.Count(sec.content, "\n")
			chunkHash := sha256Hash(sec.content)
			key := s.buildChunkKey(path, sec.startLine, endLine, chunkHash)
			fullContent := sec.content
			if sec.title != "" {
				fullContent = sec.title + "\n" + sec.content
			}
			previews = append(previews, memoryPreview{
				Key:       key,
				Content:   strings.TrimSpace(fullContent),
				Layer:     layer,
				Source:    sourcePrefix,
				FilePath:  path,
				AgentName: agentName,
				StartLine: sec.startLine,
				EndLine:   endLine,
				ChunkHash: chunkHash,
				Category:  cat,
			})
		}
	}

	return previews
}

type chunkResult struct {
	text      string
	startLine int
	endLine   int
}

func (s *OpenClawSyncService) slidingWindowChunk(content string, baseStartLine int, baseEndLine int) []chunkResult {
	lines := strings.Split(content, "\n")
	var chunks []chunkResult

	currentLines := []string{}
	currentLen := 0
	chunkStartLine := baseStartLine
	lineIdx := 0

	for i, line := range lines {
		lineLen := len(line) + 1
		if currentLen+lineLen > ChunkTargetChars && currentLen > 0 {
			text := strings.Join(currentLines, "\n")
			if len(strings.TrimSpace(text)) >= MinChunkChars {
				chunks = append(chunks, chunkResult{
					text:      text,
					startLine: chunkStartLine,
					endLine:   baseStartLine + lineIdx - 1,
				})
			}

			overlapLen := 0
			overlapStart := len(currentLines) - 1
			for overlapStart > 0 && overlapLen < ChunkOverlapChars {
				overlapStart--
				overlapLen += len(currentLines[overlapStart]) + 1
			}
			currentLines = currentLines[overlapStart:]
			currentLen = overlapLen
			chunkStartLine = baseStartLine + i - len(currentLines)
		}
		currentLines = append(currentLines, line)
		currentLen += lineLen
		lineIdx = i + 1
	}

	if len(currentLines) > 0 {
		text := strings.Join(currentLines, "\n")
		if len(strings.TrimSpace(text)) >= MinChunkChars {
			chunks = append(chunks, chunkResult{
				text:      text,
				startLine: chunkStartLine,
				endLine:   baseEndLine,
			})
		}
	}

	return chunks
}

func (s *OpenClawSyncService) buildChunkKey(path string, startLine int, endLine int, chunkHash string) string {
	pathHash := sha256Hash(path)[:8]
	shortPath := filepath.Base(path)
	return fmt.Sprintf("chunk:%s:%s:%d-%d:%s", pathHash, shortPath, startLine, endLine, chunkHash[:12])
}

func (s *OpenClawSyncService) inferAgentFromPath(path string) string {
	lower := strings.ToLower(path)
	agentPatterns := map[string]string{
		"openclaw":  "openclaw",
		"trae":      "trae",
		"codebuddy": "codebuddy",
		"cursor":    "cursor",
		"claude":    "claude",
		"windsurf":  "windsurf",
		"cline":     "cline",
		"continue":  "continue",
		"hermes":    "hermes",
		"aider":     "aider",
		"augment":   "augment",
	}
	for pat, agent := range agentPatterns {
		if strings.Contains(lower, pat) {
			return agent
		}
	}
	return filepath.Base(filepath.Dir(path))
}

func (s *OpenClawSyncService) inferPlatformFromPath(path string) string {
	lower := strings.ToLower(path)
	platformPatterns := map[string][]string{
		"openclaw":  {"openclaw", "claude-code"},
		"hermes":    {"hermes"},
		"cursor":    {"cursor"},
		"trae":      {"trae"},
		"codebuddy": {"codebuddy"},
		"windsurf":  {"windsurf", "codeium"},
		"cline":     {"cline"},
		"continue":  {"continue"},
		"aider":     {"aider"},
		"augment":   {"augment"},
	}
	for platform, patterns := range platformPatterns {
		for _, pat := range patterns {
			if strings.Contains(lower, pat) {
				return platform
			}
		}
	}
	return "clawmemory"
}

func (s *OpenClawSyncService) parseJSONFile(path string) []memoryPreview {
	data, err := os.ReadFile(path)
	if err != nil || len(data) == 0 {
		return nil
	}

	var previews []memoryPreview
	agentName := s.inferAgentFromPath(path)
	platform := s.inferPlatformFromPath(path)
	cat := s.classifyFilePath(path)
	layer, _, sourcePrefix := s.categoryToLayerAndType(cat)

	var parsed interface{}
	if json.Unmarshal(data, &parsed) != nil {
		return nil
	}

	switch v := parsed.(type) {
	case map[string]interface{}:
		if title, ok := v["title"].(string); ok && title != "" {
			content := s.extractMeaningfulJSONContent(v)
			if content != "" && len(content) >= MinChunkChars {
				chunkHash := sha256Hash(content)
				key := s.buildChunkKey(path, 1, 1, chunkHash)
				previews = append(previews, memoryPreview{
					Key:       key,
					Content:   title + "\n" + content,
					Layer:     layer,
					Source:    sourcePrefix,
					FilePath:  path,
					AgentName: agentName,
					StartLine: 1,
					EndLine:   1,
					ChunkHash: chunkHash,
					Category:  cat,
				})
			}
		}
		if messages, ok := v["messages"].([]interface{}); ok {
			previews = append(previews, s.extractMessagesFromJSON(messages, path, agentName, platform)...)
		}
	case []interface{}:
		previews = append(previews, s.extractMessagesFromJSON(v, path, agentName, platform)...)
	}

	return previews
}

func (s *OpenClawSyncService) parseJSONLFile(path string) []memoryPreview {
	data, err := os.ReadFile(path)
	if err != nil || len(data) == 0 {
		return nil
	}

	var previews []memoryPreview
	agentName := s.inferAgentFromPath(path)

	lines := strings.Split(string(data), "\n")
	var userMessages []string
	var assistantMessages []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var msg map[string]interface{}
		if json.Unmarshal([]byte(line), &msg) != nil {
			continue
		}

		msgType, _ := msg["type"].(string)
		role, _ := msg["role"].(string)
		if role == "" {
			role = msgType
		}

		text := s.extractMessageText(msg)
		if text == "" || len(text) < 10 {
			continue
		}
		if s.isLowQualityContent(text) {
			continue
		}

		if role == "user" || role == "human" {
			userMessages = append(userMessages, text)
		} else if role == "assistant" || role == "ai" {
			assistantMessages = append(assistantMessages, text)
		}
	}

	for _, text := range userMessages {
		if len(text) > MaxMemoryContentLength {
			text = text[:MaxMemoryContentLength]
		}
		chunkHash := sha256Hash(text)
		key := "jsonl:" + agentName + ":user:" + chunkHash[:12]
		previews = append(previews, memoryPreview{
			Key:       key,
			Content:   text,
			Layer:     "episodic",
			Source:    "openclaw-session-jsonl",
			FilePath:  path,
			AgentName: agentName,
			StartLine: 0,
			EndLine:   0,
			ChunkHash: chunkHash,
			Category:  categoryConversation,
		})
	}

	for _, text := range assistantMessages {
		if len(text) > MaxMemoryContentLength {
			text = text[:MaxMemoryContentLength]
		}
		chunkHash := sha256Hash(text)
		key := "jsonl:" + agentName + ":assistant:" + chunkHash[:12]
		previews = append(previews, memoryPreview{
			Key:       key,
			Content:   text,
			Layer:     "semantic",
			Source:    "openclaw-session-jsonl",
			FilePath:  path,
			AgentName: agentName,
			StartLine: 0,
			EndLine:   0,
			ChunkHash: chunkHash,
			Category:  categoryConversation,
		})
	}

	return previews
}

func (s *OpenClawSyncService) extractMessageText(msg map[string]interface{}) string {
	if text, ok := msg["text"].(string); ok && text != "" {
		return text
	}
	if content, ok := msg["content"].(string); ok && content != "" {
		return content
	}
	if parts, ok := msg["content"].([]interface{}); ok {
		var texts []string
		for _, part := range parts {
			if p, ok := part.(map[string]interface{}); ok {
				if t, ok := p["text"].(string); ok && t != "" {
					texts = append(texts, t)
				}
			}
		}
		return strings.Join(texts, "\n")
	}
	return ""
}

func (s *OpenClawSyncService) extractMeaningfulJSONContent(v map[string]interface{}) string {
	var parts []string
	skipKeys := map[string]bool{
		"id": true, "_id": true, "timestamp": true, "created_at": true,
		"updated_at": true, "version": true, "type": true, "messages": true,
	}
	for key, val := range v {
		if skipKeys[key] {
			continue
		}
		if str, ok := val.(string); ok && str != "" && len(str) >= 10 {
			parts = append(parts, str)
		}
	}
	return strings.Join(parts, "\n")
}

func (s *OpenClawSyncService) extractMessagesFromJSON(messages []interface{}, path string, agentName string, _ string) []memoryPreview {
	var previews []memoryPreview
	for _, msg := range messages {
		m, ok := msg.(map[string]interface{})
		if !ok {
			continue
		}
		role, _ := m["role"].(string)
		text := s.extractMessageText(m)
		if text == "" || len(text) < 10 {
			continue
		}
		if s.isLowQualityContent(text) {
			continue
		}
		if len(text) > MaxMemoryContentLength {
			text = text[:MaxMemoryContentLength]
		}

		chunkHash := sha256Hash(text)
		key := "json:" + agentName + ":" + role + ":" + chunkHash[:12]
		layer := "episodic"
		if role == "assistant" {
			layer = "semantic"
		}
		previews = append(previews, memoryPreview{
			Key:       key,
			Content:   text,
			Layer:     layer,
			Source:    "openclaw-session-json",
			FilePath:  path,
			AgentName: agentName,
			StartLine: 0,
			EndLine:   0,
			ChunkHash: chunkHash,
			Category:  categoryConversation,
		})
	}
	return previews
}

func (s *OpenClawSyncService) extractSqliteMemories(dbPath string) []memoryPreview {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil
	}
	defer sqlDB.Close()

	var previews []memoryPreview
	agentName := s.inferAgentFromPath(dbPath)
	platform := s.inferPlatformFromPath(dbPath)

	if s.isOpenClawMemoryDB(db) {
		previews = append(previews, s.extractOpenClawMemoryChunks(db, dbPath, agentName, platform)...)
		if len(previews) > 0 {
			return previews
		}
	}

	skipTables := map[string]bool{
		"sqlite_master": true, "sqlite_sequence": true,
		"sqlite_stat1": true, "sqlite_stat4": true,
		"meta": true, "schema_migrations": true,
		"embedding_cache": true, "chunks_vec": true, "chunks_fts": true,
		"files": true, "ItemTable": true,
	}

	type TableName struct {
		Name string
	}
	var tables []TableName
	db.Raw("SELECT name FROM sqlite_master WHERE type='table'").Scan(&tables)

	for _, t := range tables {
		if skipTables[strings.ToLower(t.Name)] {
			continue
		}

		type ColInfo struct {
			Name string
		}
		var cols []ColInfo
		db.Raw(fmt.Sprintf("PRAGMA table_info(%s)", t.Name)).Scan(&cols)

		keyCol := ""
		valueCol := ""
		for _, c := range cols {
			cl := strings.ToLower(c.Name)
			if keyCol == "" && (cl == "key" || cl == "name" || cl == "title" || cl == "id") {
				keyCol = c.Name
			}
			if valueCol == "" && (cl == "value" || cl == "content" || cl == "text" || cl == "message") {
				valueCol = c.Name
			}
		}

		if keyCol == "" || valueCol == "" {
			continue
		}

		type KV struct {
			Key   string
			Value string
		}
		var kvPairs []KV
		db.Raw(fmt.Sprintf("SELECT %s as key, %s as value FROM %s LIMIT 500", keyCol, valueCol, t.Name)).Scan(&kvPairs)

		for _, kv := range kvPairs {
			if kv.Key == "" || kv.Value == "" {
				continue
			}
			if s.isLowQualityContent(kv.Value) {
				continue
			}
			chunkHash := sha256Hash(kv.Value)
			previews = append(previews, memoryPreview{
				Key:       "sqlite:" + t.Name + ":" + chunkHash[:12],
				Content:   kv.Value,
				Layer:     "knowledge",
				Source:    "openclaw-sqlite",
				FilePath:  dbPath,
				AgentName: agentName,
				StartLine: 0,
				EndLine:   0,
				ChunkHash: chunkHash,
				Category:  categoryKnowledge,
			})
		}
	}

	return previews
}

func (s *OpenClawSyncService) isOpenClawMemoryDB(db *gorm.DB) bool {
	type TableName struct {
		Name string
	}
	var tables []TableName
	db.Raw("SELECT name FROM sqlite_master WHERE type='table'").Scan(&tables)
	hasChunks := false
	hasFiles := false
	for _, t := range tables {
		if t.Name == "chunks" {
			hasChunks = true
		}
		if t.Name == "files" {
			hasFiles = true
		}
	}
	return hasChunks && hasFiles
}

func (s *OpenClawSyncService) extractOpenClawMemoryChunks(db *gorm.DB, dbPath string, agentName string, _ string) []memoryPreview {
	type Chunk struct {
		ID        string `gorm:"column:id"`
		Path      string `gorm:"column:path"`
		Source    string `gorm:"column:source"`
		StartLine int    `gorm:"column:start_line"`
		EndLine   int    `gorm:"column:end_line"`
		Hash      string `gorm:"column:hash"`
		Text      string `gorm:"column:text"`
		Model     string `gorm:"column:model"`
	}

	var chunks []Chunk
	if err := db.Table("chunks").Select("id, path, source, start_line, end_line, hash, text, model").
		Where("length(text) >= ?", MinChunkChars).
		Limit(1000).Find(&chunks).Error; err != nil {
		return nil
	}

	var previews []memoryPreview
	for _, ch := range chunks {
		if s.isLowQualityContent(ch.Text) {
			continue
		}
		text := ch.Text
		if len(text) > MaxMemoryContentLength {
			text = text[:MaxMemoryContentLength]
		}

		cat := categoryKnowledge
		layer := "semantic"
		memSource := "openclaw-index-chunk"
		if ch.Source == "sessions" {
			cat = categoryConversation
			layer = "episodic"
			memSource = "openclaw-index-session"
		} else if strings.Contains(strings.ToLower(ch.Path), "memory/") && !strings.Contains(strings.ToLower(ch.Path), "memory.md") {
			cat = categoryDailyLog
			layer = "episodic"
			memSource = "openclaw-index-daily"
		}

		previews = append(previews, memoryPreview{
			Key:       "ocidx:" + ch.ID[:min(20, len(ch.ID))],
			Content:   text,
			Layer:     layer,
			Source:    memSource,
			FilePath:  dbPath,
			AgentName: agentName,
			StartLine: ch.StartLine,
			EndLine:   ch.EndLine,
			ChunkHash: ch.Hash,
			Category:  cat,
		})
	}

	return previews
}

func (s *OpenClawSyncService) extractVscdbMemories(dbPath string) []memoryPreview {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil
	}
	sqlDB, _ := db.DB()
	if sqlDB != nil {
		defer sqlDB.Close()
	}

	var previews []memoryPreview

	type TableName struct {
		Name string
	}
	var tables []TableName
	db.Raw("SELECT name FROM sqlite_master WHERE type='table'").Scan(&tables)
	hasItemTable := false
	for _, t := range tables {
		if t.Name == "ItemTable" {
			hasItemTable = true
			break
		}
	}
	if !hasItemTable {
		return nil
	}

	agentName := "unknown-vscdb"
	lowerPath := strings.ToLower(dbPath)
	if strings.Contains(lowerPath, "trae") {
		agentName = "trae"
	} else if strings.Contains(lowerPath, "codebuddy") {
		agentName = "codebuddy"
	} else if strings.Contains(lowerPath, "cursor") {
		agentName = "cursor"
	} else if strings.Contains(lowerPath, "windsurf") {
		agentName = "windsurf"
	}

	type KV struct {
		Key   string
		Value string
	}

	var chatRows []KV
	db.Raw("SELECT key, value FROM ItemTable WHERE key = 'icube-ai-agent-storage-input-history'").Scan(&chatRows)
	for _, row := range chatRows {
		if row.Value == "" {
			continue
		}
		inputs := s.parseInputHistory(row.Value)
		for _, input := range inputs {
			if len(input) < 10 {
				continue
			}
			if s.isLowQualityContent(input) {
				continue
			}
			key := agentName + "-chat-" + md5Hash(input)[:12]
			previews = append(previews, memoryPreview{
				Key:       key,
				Content:   input,
				Layer:     "episodic",
				Source:    agentName + "-chat",
				FilePath:  dbPath,
				AgentName: agentName,
			})
		}
	}

	var sessionRows []KV
	db.Raw("SELECT key, value FROM ItemTable WHERE key LIKE 'session:%'").Scan(&sessionRows)
	for _, row := range sessionRows {
		if row.Value == "" || len(row.Value) < 20 {
			continue
		}
		sessionInfo := s.parseSessionInfo(row.Value)
		if sessionInfo == "" {
			continue
		}
		key := agentName + "-session-" + md5Hash(row.Value)[:12]
		previews = append(previews, memoryPreview{
			Key:       key,
			Content:   sessionInfo,
			Layer:     "episodic",
			Source:    agentName + "-session",
			FilePath:  dbPath,
			AgentName: agentName,
		})
	}

	return previews
}

func (s *OpenClawSyncService) parseSessionInfo(jsonStr string) string {
	type SessionData struct {
		ConversationID string `json:"conversationId"`
		Cwd            string `json:"cwd"`
		Title          string `json:"title"`
		Status         string `json:"status"`
		CreatedAt      int64  `json:"createdAt"`
		UpdatedAt      int64  `json:"updatedAt"`
	}
	var session SessionData
	if err := json.Unmarshal([]byte(jsonStr), &session); err != nil {
		return jsonStr
	}
	if session.Title == "" {
		return ""
	}
	result := session.Title
	if session.Cwd != "" {
		result += " | Project: " + session.Cwd
	}
	if session.Status != "" {
		result += " | Status: " + session.Status
	}
	return result
}

func (s *OpenClawSyncService) parseInputHistory(jsonStr string) []string {
	type InputEntry struct {
		InputText string `json:"inputText"`
	}
	var entries []InputEntry
	if err := json.Unmarshal([]byte(jsonStr), &entries); err != nil {
		return nil
	}
	var result []string
	for _, e := range entries {
		if e.InputText != "" {
			result = append(result, e.InputText)
		}
	}
	return result
}

func (s *OpenClawSyncService) isLowQualityContent(content string) bool {
	if len(content) < 10 {
		return true
	}

	lower := strings.ToLower(content)

	errorPatterns := []string{
		"stack trace:", "#0 ", "#1 ", "#2 ",
		"uncaught typeerror", "fatal error:",
		"at module._compile", "at object.module",
		"at require (internal", "at process._tickCallback",
		"at new ", "at class ", "at function ",
		"thrown at", "exception in thread",
	}
	errorCount := 0
	for _, p := range errorPatterns {
		if strings.Contains(lower, p) {
			errorCount++
		}
	}
	if errorCount >= 3 {
		return true
	}

	lines := strings.Split(content, "\n")
	nonEmptyLines := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") && !strings.HasPrefix(trimmed, "at ") {
			nonEmptyLines++
		}
	}
	if nonEmptyLines == 0 && len(lines) > 5 {
		return true
	}

	uiPatterns := []string{
		"loading project", "please wait",
		"click here", "press enter",
		"yes/no", "y/n",
	}
	for _, p := range uiPatterns {
		if strings.Contains(lower, p) && len(content) < 100 {
			return true
		}
	}

	codeOnlyPatterns := []struct {
		pattern string
		minOcc  int
	}{
		{"import ", 3},
		{"from '", 3},
		{"require('", 3},
		{"func ", 3},
		{"var ", 3},
		{"let ", 3},
		{"const ", 3},
	}
	for _, cop := range codeOnlyPatterns {
		if strings.Count(lower, cop.pattern) >= cop.minOcc && nonEmptyLines < 3 {
			return true
		}
	}

	if len(content) > 5000 {
		codeLineCount := 0
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "//") ||
				strings.HasPrefix(trimmed, "/*") ||
				strings.HasPrefix(trimmed, "*") ||
				strings.HasPrefix(trimmed, "*/") ||
				strings.HasPrefix(trimmed, "<!--") ||
				strings.HasPrefix(trimmed, "-->") ||
				strings.HasPrefix(trimmed, "#") ||
				(len(trimmed) > 0 && (trimmed[len(trimmed)-1] == ';' || trimmed[len(trimmed)-1] == '{' || trimmed[len(trimmed)-1] == '}')) {
				codeLineCount++
			}
		}
		if float64(codeLineCount)/float64(len(lines)) > 0.8 {
			return true
		}
	}

	return false
}

func (s *OpenClawSyncService) computeSignificance(p memoryPreview) float64 {
	score := 0.3
	content := p.Content
	lower := strings.ToLower(content)

	switch p.Category {
	case categoryLongTerm:
		score += 0.4
	case categoryKnowledge:
		score += 0.2
	case categoryDailyLog:
		score += 0.1
	case categorySessionLog:
		score += 0.05
	case categoryConversation:
		score += 0.0
	}

	preferenceSignals := []string{"prefer", "like", "hate", "want", "always", "never", "should", "must", "need to"}
	for _, sig := range preferenceSignals {
		if strings.Contains(lower, sig) {
			score += 0.1
			break
		}
	}

	decisionSignals := []string{"decided", "chose", "will use", "agreed", "concluded", "resolved", "determined"}
	for _, sig := range decisionSignals {
		if strings.Contains(lower, sig) {
			score += 0.15
			break
		}
	}

	factSignals := []string{"is called", "lives in", "works at", "born on", "birthday", "email", "phone", "address"}
	for _, sig := range factSignals {
		if strings.Contains(lower, sig) {
			score += 0.1
			break
		}
	}

	infoDensity := 0.0
	if len(content) > 0 {
		words := strings.Fields(content)
		uniqueWords := make(map[string]bool)
		for _, w := range words {
			uniqueWords[strings.ToLower(w)] = true
		}
		if len(words) > 0 {
			infoDensity = float64(len(uniqueWords)) / float64(len(words))
		}
	}
	score += infoDensity * 0.1

	if len(content) > 50 && len(content) < 5000 {
		score += 0.05
	}

	if score > 1.0 {
		score = 1.0
	}
	return score
}

func (s *OpenClawSyncService) isSignificant(p memoryPreview) bool {
	score := s.computeSignificance(p)
	threshold := 0.3
	switch p.Category {
	case categoryLongTerm:
		threshold = 0.2
	case categoryKnowledge:
		threshold = 0.25
	case categoryDailyLog:
		threshold = 0.3
	case categoryConversation:
		threshold = 0.4
	}
	return score >= threshold
}

func (s *OpenClawSyncService) inferMemoryType(p memoryPreview) string {
	if p.Category != "" {
		switch p.Category {
		case categoryLongTerm, categoryKnowledge:
			return "knowledge"
		case categoryDailyLog, categorySessionLog, categoryConversation:
			return "episodic"
		}
	}
	if p.Layer == "episodic" {
		return "episodic"
	}
	chatSources := []string{"-chat", "-session", "openclaw-session", "openclaw-session-jsonl", "openclaw-session-json", "openclaw-session-archive", "openclaw-daily-log"}
	for _, cs := range chatSources {
		if strings.Contains(p.Source, cs) {
			return "episodic"
		}
	}
	return "knowledge"
}

func (s *OpenClawSyncService) GetStatus() SyncStatus {
	s.mu.Lock()
	defer s.mu.Unlock()

	localPaths := s.getLocalSearchDirs()
	var existingPaths []string
	for _, p := range localPaths {
		if _, err := os.Stat(p); err == nil {
			existingPaths = append(existingPaths, p)
		}
	}

	watchMode := "polling"
	watchedDirs := 0
	if s.watcher != nil {
		watchMode = "fsnotify"
		watchedDirs = len(s.watcher.WatchList())
	}

	return SyncStatus{
		Running:         s.running,
		Mode:            s.mode,
		WatchMode:       watchMode,
		LastSyncTime:    s.lastSyncTime,
		SyncedCount:     s.syncedCount,
		SkippedCount:    s.skippedCount,
		LastError:       s.lastError,
		AutoSyncEnabled: s.autoSyncEnabled,
		OpenClawFound:   s.detectLocalOpenClaw(),
		LocalPaths:      existingPaths,
		RemoteEndpoint:  "/api/v1/external/conversations",
		WatchedDirs:     watchedDirs,
	}
}

func (s *OpenClawSyncService) SetAutoSync(enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.autoSyncEnabled = enabled
}

func (s *OpenClawSyncService) ForceSync() int {
	if s.mode == "local" {
		s.localSync()
	}
	return s.syncedCount
}

func sha256Hash(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func md5Hash(s string) string {
	h := md5.Sum([]byte(s))
	return hex.EncodeToString(h[:])
}
