package services

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"clawmemory/internal/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	MaxMemoryContentLength = 50000
	SyncInterval           = 60 * time.Second
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
	lastSyncTime    time.Time
	syncedFiles     map[string]string
	syncedCount     int
	lastError       string
	autoSyncEnabled bool
	skippedCount    int
}

type SyncStatus struct {
	Running         bool      `json:"running"`
	LastSyncTime    time.Time `json:"last_sync_time"`
	SyncedCount     int       `json:"synced_count"`
	SkippedCount    int       `json:"skipped_count"`
	LastError       string    `json:"last_error"`
	AutoSyncEnabled bool      `json:"auto_sync_enabled"`
	OpenClawFound   bool      `json:"openclaw_found"`
}

var (
	globalSyncService *OpenClawSyncService
	syncServiceMu     sync.Mutex
)

func GetOpenClawSyncService(db *gorm.DB) *OpenClawSyncService {
	syncServiceMu.Lock()
	defer syncServiceMu.Unlock()

	if globalSyncService == nil {
		globalSyncService = &OpenClawSyncService{
			db:              db,
			memService:      NewMemoryService(db),
			syncInterval:    60 * time.Second,
			stopChan:        make(chan struct{}),
			syncedFiles:     make(map[string]string),
			autoSyncEnabled: true,
		}
		globalSyncService.loadSyncedFiles()
	}
	return globalSyncService
}

func (s *OpenClawSyncService) loadSyncedFiles() {
	var memories []models.Memory
	s.db.Where("source LIKE ?", "openclaw%").Select("key, source").Find(&memories)
	for _, m := range memories {
		hash := md5Hash(m.Key + m.Source)
		s.syncedFiles[hash] = m.Key
	}
}

func (s *OpenClawSyncService) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return
	}

	homeDir, _ := os.UserHomeDir()
	openclawDir := filepath.Join(homeDir, ".openclaw")
	if _, err := os.Stat(openclawDir); os.IsNotExist(err) {
		s.lastError = "OpenClaw directory not found"
		return
	}

	s.running = true
	s.stopChan = make(chan struct{})

	go s.syncLoop()
}

func (s *OpenClawSyncService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	close(s.stopChan)
	s.running = false
}

func (s *OpenClawSyncService) syncLoop() {
	s.doSync()

	ticker := time.NewTicker(s.syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if s.autoSyncEnabled {
				s.doSync()
			}
		case <-s.stopChan:
			return
		}
	}
}

func (s *OpenClawSyncService) doSync() {
	s.lastSyncTime = time.Now()
	s.lastError = ""

	previews := s.scanOpenClawDirs()
	if len(previews) == 0 {
		return
	}

	recordSensitive := s.getRecordSensitiveSetting()
	var encryptor *Encryptor
	if recordSensitive {
		secretKey := s.getSecretKey()
		var err error
		encryptor, err = NewEncryptor(secretKey)
		if err != nil {
			s.lastError = "failed to init encryptor"
			return
		}
	}

	newCount := 0
	skipped := 0
	for _, p := range previews {
		hash := md5Hash(p.FilePath + p.Key)
		if _, exists := s.syncedFiles[hash]; exists {
			continue
		}

		isSensitive := s.containsSensitiveInfo(p.Key) || s.containsSensitiveInfo(p.Content)
		isSensitiveFile := s.isSensitiveFile(p.FilePath)

		if isSensitive || isSensitiveFile {
			if !recordSensitive {
				skipped++
				s.syncedFiles[hash] = "__SKIPPED_SENSITIVE__"
				continue
			}
		}

		if isSensitiveFile && !recordSensitive {
			skipped++
			s.syncedFiles[hash] = "__SKIPPED_FILE__"
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

		_, err := s.memService.Create(userID, map[string]interface{}{
			"key":          p.Key,
			"value":        value,
			"layer":        p.Layer,
			"source":       source,
			"memory_type":  "knowledge",
			"is_encrypted": isEncrypted,
		})
		if err != nil {
			continue
		}

		s.syncedFiles[hash] = p.Key
		newCount++
	}

	s.syncedCount += newCount
	s.skippedCount += skipped
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
	if key := os.Getenv("SECRET_KEY"); key != "" && key != "clawmemory-default-secret-change-me" {
		return key
	}
	return "clawmemory-encryption-key-" + fmt.Sprintf("%d", time.Now().UnixNano())
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
	if err := s.db.First(&user).Error; err != nil {
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
}

func (s *OpenClawSyncService) scanOpenClawDirs() []memoryPreview {
	searchDirs := s.getOpenClawSearchDirs()
	var allPreviews []memoryPreview

	for _, dir := range searchDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}
		previews := s.extractMemoriesFromDir(dir)
		allPreviews = append(allPreviews, previews...)
	}

	return allPreviews
}

func (s *OpenClawSyncService) getOpenClawSearchDirs() []string {
	var dirs []string
	homeDir, _ := os.UserHomeDir()

	addDir := func(path string) {
		if _, err := os.Stat(path); err == nil {
			dirs = append(dirs, path)
		}
	}

	addDir(filepath.Join(homeDir, ".openclaw"))
	addDir(filepath.Join(homeDir, ".openclaw", "workspace"))
	addDir(filepath.Join(homeDir, ".openclaw", "skills"))

	exePath, _ := os.Executable()
	exeDir := filepath.Dir(exePath)
	addDir(filepath.Join(exeDir, "openclaw"))

	wd, _ := os.Getwd()
	addDir(filepath.Join(wd, ".openclaw"))

	return dirs
}

func (s *OpenClawSyncService) extractMemoriesFromDir(dir string) []memoryPreview {
	var previews []memoryPreview

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
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
		}

		return nil
	})

	return previews
}

func (s *OpenClawSyncService) parseMarkdownFile(path string) []memoryPreview {
	data, err := os.ReadFile(path)
	if err != nil || len(data) == 0 {
		return nil
	}

	var previews []memoryPreview
	content := string(data)
	lines := strings.Split(content, "\n")

	currentSection := ""
	currentContent := ""

	for _, line := range lines {
		if strings.HasPrefix(line, "# ") {
			if currentSection != "" && currentContent != "" {
				previews = append(previews, memoryPreview{
					Key:       currentSection,
					Content:   strings.TrimSpace(currentContent),
					Layer:     "knowledge",
					Source:    "openclaw-markdown",
					FilePath:  path,
					AgentName: filepath.Base(filepath.Dir(path)),
				})
			}
			currentSection = strings.TrimSpace(line[2:])
			currentContent = ""
		} else {
			currentContent += line + "\n"
		}
	}

	if currentSection != "" && currentContent != "" {
		previews = append(previews, memoryPreview{
			Key:       currentSection,
			Content:   strings.TrimSpace(currentContent),
			Layer:     "knowledge",
			Source:    "openclaw-markdown",
			FilePath:  path,
			AgentName: filepath.Base(filepath.Dir(path)),
		})
	}

	if len(previews) == 0 && len(content) > 50 {
		previews = append(previews, memoryPreview{
			Key:       filepath.Base(path),
			Content:   content,
			Layer:     "knowledge",
			Source:    "openclaw-markdown",
			FilePath:  path,
			AgentName: filepath.Base(filepath.Dir(path)),
		})
	}

	return previews
}

func (s *OpenClawSyncService) parseJSONFile(path string) []memoryPreview {
	data, err := os.ReadFile(path)
	if err != nil || len(data) == 0 {
		return nil
	}

	var previews []memoryPreview
	agentName := filepath.Base(filepath.Dir(path))

	if strings.Contains(path, "session") || strings.Contains(path, "conversation") {
		previews = append(previews, memoryPreview{
			Key:       filepath.Base(path),
			Content:   string(data),
			Layer:     "episodic",
			Source:    "openclaw-session",
			FilePath:  path,
			AgentName: agentName,
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
	sqlDB, _ := db.DB()
	if sqlDB != nil {
		defer sqlDB.Close()
	}

	var previews []memoryPreview
	agentName := filepath.Base(filepath.Dir(dbPath))
	if agentName == "" || agentName == "." {
		agentName = "sqlite-" + filepath.Base(dbPath)
	}

	type TableName struct {
		Name string
	}
	var tables []TableName
	db.Raw("SELECT name FROM sqlite_master WHERE type='table'").Scan(&tables)

	for _, t := range tables {
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
		db.Raw(fmt.Sprintf("SELECT %s as key, %s as value FROM %s LIMIT 200", keyCol, valueCol, t.Name)).Scan(&kvPairs)

		for _, kv := range kvPairs {
			if kv.Key == "" || kv.Value == "" {
				continue
			}
			previews = append(previews, memoryPreview{
				Key:       kv.Key,
				Content:   kv.Value,
				Layer:     "knowledge",
				Source:    "openclaw-sqlite",
				FilePath:  dbPath,
				AgentName: agentName,
			})
		}
	}

	return previews
}

func (s *OpenClawSyncService) GetStatus() SyncStatus {
	s.mu.Lock()
	defer s.mu.Unlock()

	homeDir, _ := os.UserHomeDir()
	openclawDir := filepath.Join(homeDir, ".openclaw")
	_, found := os.Stat(openclawDir)

	return SyncStatus{
		Running:         s.running,
		LastSyncTime:    s.lastSyncTime,
		SyncedCount:     s.syncedCount,
		SkippedCount:    s.skippedCount,
		LastError:       s.lastError,
		AutoSyncEnabled: s.autoSyncEnabled,
		OpenClawFound:   found == nil,
	}
}

func (s *OpenClawSyncService) SetAutoSync(enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.autoSyncEnabled = enabled
}

func (s *OpenClawSyncService) ForceSync() int {
	s.doSync()
	return s.syncedCount
}

func md5Hash(s string) string {
	h := md5.Sum([]byte(s))
	return hex.EncodeToString(h[:])
}
