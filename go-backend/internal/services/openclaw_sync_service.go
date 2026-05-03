package services

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
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
	syncedKeys      map[string]string
	syncedCount     int
	lastError       string
	autoSyncEnabled bool
	skippedCount    int
	mode            string
}

type SyncStatus struct {
	Running         bool      `json:"running"`
	Mode            string    `json:"mode"`
	LastSyncTime    time.Time `json:"last_sync_time"`
	SyncedCount     int       `json:"synced_count"`
	SkippedCount    int       `json:"skipped_count"`
	LastError       string    `json:"last_error"`
	AutoSyncEnabled bool      `json:"auto_sync_enabled"`
	OpenClawFound   bool      `json:"openclaw_found"`
	LocalPaths      []string  `json:"local_paths,omitempty"`
	RemoteEndpoint  string    `json:"remote_endpoint,omitempty"`
}

type ConversationMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ConversationPushRequest struct {
	SessionID   string               `json:"session_id"`
	Title       string               `json:"title"`
	ProjectPath string               `json:"project_path"`
	AgentName   string               `json:"agent_name"`
	Messages    []ConversationMessage `json:"messages"`
	Summary     string               `json:"summary"`
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
	homeDir, _ := os.UserHomeDir()
	openclawDir := filepath.Join(homeDir, ".openclaw")
	if _, err := os.Stat(openclawDir); err == nil {
		return true
	}

	appData := os.Getenv("APPDATA")
	if appData != "" {
		if _, err := os.Stat(filepath.Join(appData, "Trae CN")); err == nil {
			return true
		}
		if _, err := os.Stat(filepath.Join(appData, "CodeBuddy CN")); err == nil {
			return true
		}
		if _, err := os.Stat(filepath.Join(appData, "CodeBuddy")); err == nil {
			return true
		}
	}

	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData != "" {
		if _, err := os.Stat(filepath.Join(localAppData, "Trae CN")); err == nil {
			return true
		}
		if _, err := os.Stat(filepath.Join(localAppData, "CodeBuddy CN")); err == nil {
			return true
		}
	}

	return false
}

func (s *OpenClawSyncService) loadSyncedKeys() {
	var memories []models.Memory
	s.db.Where("source LIKE ? OR source LIKE ? OR source LIKE ? OR source LIKE ?",
		"openclaw%", "trae%", "codebuddy%", "conversation%").Select("key").Find(&memories)
	for _, m := range memories {
		s.syncedKeys[m.Key] = m.Key
	}
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

	if s.mode == "local" {
		go s.localSyncLoop()
	}
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

func (s *OpenClawSyncService) localSyncLoop() {
	s.localSync()

	ticker := time.NewTicker(s.syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if s.autoSyncEnabled {
				s.localSync()
			}
		case <-s.stopChan:
			return
		}
	}
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
	for _, p := range previews {
		if _, exists := s.syncedKeys[p.Key]; exists {
			continue
		}

		if s.isLowQualityContent(p.Content) {
			skipped++
			continue
		}

		isSensitive := s.containsSensitiveInfo(p.Key) || s.containsSensitiveInfo(p.Content)
		isSensitiveFile := s.isSensitiveFile(p.FilePath)

		if (isSensitive || isSensitiveFile) && !recordSensitive {
			skipped++
			s.syncedKeys[p.Key] = "__SKIPPED_SENSITIVE__"
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
			"memory_type":  s.inferMemoryType(p),
			"is_encrypted": isEncrypted,
		})
		if err != nil {
			continue
		}

		s.syncedKeys[p.Key] = p.Key
		newCount++
	}

	s.syncedCount += newCount
	s.skippedCount += skipped
}

func (s *OpenClawSyncService) PushConversation(userID uint, req ConversationPushRequest) (int, error) {
	created := 0

	if req.SessionID != "" && req.Summary != "" {
		key := "conversation-" + req.AgentName + "-" + req.SessionID
		if _, exists := s.syncedKeys[key]; !exists {
			value := req.Summary
			if req.Title != "" {
				value = req.Title + "\n\n" + value
			}
			if req.ProjectPath != "" {
				value = "Project: " + req.ProjectPath + "\n" + value
			}

			if !s.isLowQualityContent(value) {
				_, err := s.memService.Create(userID, map[string]interface{}{
					"key":         key,
					"value":       value,
					"layer":       "episodic",
					"source":      "conversation-" + req.AgentName,
					"memory_type": "episodic",
				})
				if err == nil {
					s.syncedKeys[key] = key
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
		if _, exists := s.syncedKeys[key]; exists {
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
			"key":         key,
			"value":       value,
			"layer":       layer,
			"source":      "conversation-" + req.AgentName + "-" + msg.Role,
			"memory_type": memoryType,
		})
		if err == nil {
			s.syncedKeys[key] = key
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

	appData := os.Getenv("APPDATA")
	if appData != "" {
		addDir(filepath.Join(appData, "Trae CN", "User", "workspaceStorage"))
		addDir(filepath.Join(appData, "Trae CN", "ModularData", "ai-agent"))
		addDir(filepath.Join(appData, "CodeBuddy CN", "User", "workspaceStorage"))
		addDir(filepath.Join(appData, "CodeBuddy", "User", "workspaceStorage"))
	}

	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData != "" {
		addDir(filepath.Join(localAppData, "Trae CN"))
		addDir(filepath.Join(localAppData, "CodeBuddy CN"))
	}

	return dirs
}

func (s *OpenClawSyncService) extractFromDir(dir string) []memoryPreview {
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
		case ".vscdb":
			previews = append(previews, s.extractVscdbMemories(path)...)
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

	agentName := "trae"
	if strings.Contains(strings.ToLower(dbPath), "codebuddy") {
		agentName = "codebuddy"
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

func (s *OpenClawSyncService) inferMemoryType(p memoryPreview) string {
	if p.Layer == "episodic" {
		return "episodic"
	}
	chatSources := []string{"-chat", "-session", "openclaw-session"}
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

	return SyncStatus{
		Running:         s.running,
		Mode:            s.mode,
		LastSyncTime:    s.lastSyncTime,
		SyncedCount:     s.syncedCount,
		SkippedCount:    s.skippedCount,
		LastError:       s.lastError,
		AutoSyncEnabled: s.autoSyncEnabled,
		OpenClawFound:   s.detectLocalOpenClaw(),
		LocalPaths:      existingPaths,
		RemoteEndpoint:  "/api/v1/external/conversations",
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

func md5Hash(s string) string {
	h := md5.Sum([]byte(s))
	return hex.EncodeToString(h[:])
}
