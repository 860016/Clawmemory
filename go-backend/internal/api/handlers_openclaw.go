package api

import (
	"clawmemory/internal/middleware"
	"clawmemory/internal/models"
	"clawmemory/internal/services"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type memoryPreview struct {
	Key       string `json:"key"`
	Content   string `json:"content"`
	Layer     string `json:"layer"`
	Source    string `json:"source"`
	FilePath  string `json:"file_path"`
	AgentName string `json:"agent_name"`
}

func handleGetConnectedAgents(c *gin.Context) {
	installed := services.DetectInstalledClients()

	var connected []map[string]interface{}
	for _, client := range installed {
		connected = append(connected, map[string]interface{}{
			"name":         client["name"],
			"display_name": client["display_name"],
			"status":       "connected",
			"found_dirs":   client["found_dirs"],
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"agents": connected,
		"total":  len(connected),
	})
}

func handleOpenClawSyncStatus(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		syncService := services.GetOpenClawSyncService(db)
		status := syncService.GetStatus()
		c.JSON(http.StatusOK, status)
	}
}

func handleOpenClawSyncForce(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		syncService := services.GetOpenClawSyncService(db)
		count := syncService.ForceSync()
		c.JSON(http.StatusOK, gin.H{
			"message":      "sync completed",
			"synced_count": count,
		})
	}
}

func handleOpenClawSyncToggle(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Enabled bool `json:"enabled"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "enabled field required"})
			return
		}

		syncService := services.GetOpenClawSyncService(db)
		syncService.SetAutoSync(req.Enabled)

		status := "enabled"
		if !req.Enabled {
			status = "disabled"
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "auto-sync " + status,
			"enabled": req.Enabled,
		})
	}
}

func handleGetAgentsMD(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		scheme := "http"
		if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		host := c.Request.Host
		if host == "" {
			host = "localhost:8765"
		}
		baseURL := scheme + "://" + host

		var keys []models.APIKey
		logDBErr("load api keys", db.Where("revoked_at IS NULL").Order("created_at DESC").Find(&keys).Error)

		hasKey := len(keys) > 0
		apiKeyHint := "cm_your_api_key_here"
		if hasKey {
			apiKeyHint = keys[0].KeyPrefix + "..."
		}

		content := `## 🧠 ClawMemory Auto-Record

每次回复结束时，自动将当次对话写入 ClawMemory。步骤如下：

1. 读取当前 session 的 jsonl 文件：` + "`~/.openclaw/agents/main/sessions/`" + ` 下的 ` + "`.jsonl`" + ` 文件（找最新修改的）
2. 提取用户消息和我的回复
3. 追加写到 ` + "`~/.openclaw/workspace/memory/auto-conversation.yml`" + `（按时间追加，用 ` + "`---`" + ` 分隔）
4. ClawMemory 的 auto-sync 服务每60秒自动扫描并导入

## 🧠 直接读写 ClawMemory（记忆后端）

ClawMemory 是我的长期记忆后端，我可以直接通过 API 读写它。

### 连接信息
- **地址**: ` + baseURL + `
- **API Key**: ` + apiKeyHint + `
- **初始密码**: ` + "`admin123`（首次登录后必须修改）" + `

### 何时读记忆
- 收到需要回忆的任务时，先搜索 ClawMemory：
` + "```" + `
curl -s "` + baseURL + `/api/v1/external/memories/search?q=关键词" -H "X-API-Key: <your-api-key>"
` + "```" + `
- 替代原有 memory_search 工具，因为 ClawMemory 支持语义搜索，更强

### 何时写记忆
- 每次回复末尾，自动将本轮对话（用户消息 + 我的回复）写入 ClawMemory
- 使用 API：` + "`POST /api/v1/external/conversations`" + `
` + "```" + `
curl -X POST "` + baseURL + `/api/v1/external/conversations" \
  -H "X-API-Key: <your-api-key>" \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "<当前session-id>",
    "agent_name": "openclaw",
    "messages": [
      {"role": "user", "content": "<用户消息>"},
      {"role": "assistant", "content": "<我的回复>"}
    ],
    "summary": "<本轮对话摘要>"
  }'
` + "```" + `
- 不需要等文件同步，直接写，即刻生效
`

		c.JSON(http.StatusOK, gin.H{
			"content":  content,
			"base_url": baseURL,
			"has_key":  hasKey,
		})
	}
}

func getOpenClawSearchDirs() []string {
	return services.GetAllSearchDirs()
}

func extractMemoriesFromDir(dir string) ([]memoryPreview, map[string]int) {
	var previews []memoryPreview
	agentCountMap := make(map[string]int)

	memFile := filepath.Join(dir, "MEMORY.md")
	if data, err := os.ReadFile(memFile); err == nil && len(data) > 0 {
		previews, agentCountMap = parseMarkdownMemory(string(data), memFile, "workspace", previews, agentCountMap)
	}

	claudeMdFile := filepath.Join(dir, "CLAUDE.md")
	if data, err := os.ReadFile(claudeMdFile); err == nil && len(data) > 0 {
		previews, agentCountMap = parseMarkdownMemory(string(data), claudeMdFile, "claude", previews, agentCountMap)
	}

	memoryDir := filepath.Join(dir, "memory")
	if info, err := os.Stat(memoryDir); err == nil && info.IsDir() {
		previews, agentCountMap = extractWorkspaceMemory(memoryDir, dir, previews, agentCountMap)
	}

	memoriesDir := filepath.Join(dir, "memories")
	if info, err := os.Stat(memoriesDir); err == nil && info.IsDir() {
		previews, agentCountMap = extractWorkspaceMemory(memoriesDir, dir, previews, agentCountMap)
	}

	sessionsDir := filepath.Join(dir, "agents")
	if info, err := os.Stat(sessionsDir); err == nil && info.IsDir() {
		previews, agentCountMap = extractSessionMemories(sessionsDir, previews, agentCountMap)
	}

	projectsDir := filepath.Join(dir, "projects")
	if info, err := os.Stat(projectsDir); err == nil && info.IsDir() {
		previews, agentCountMap = extractClaudeProjects(projectsDir, previews, agentCountMap)
	}

	sessionsDir2 := filepath.Join(dir, "sessions")
	if info, err := os.Stat(sessionsDir2); err == nil && info.IsDir() {
		previews, agentCountMap = extractSessionDir(sessionsDir2, previews, agentCountMap)
	}

	dataDir := filepath.Join(dir, "data")
	if info, err := os.Stat(dataDir); err == nil && info.IsDir() {
		previews, agentCountMap = extractWorkspaceMemory(dataDir, dir, previews, agentCountMap)
	}

	sqliteFiles, _ := filepath.Glob(filepath.Join(dir, "*.db"))
	sqliteFiles2, _ := filepath.Glob(filepath.Join(dir, "*.sqlite"))
	sqliteFiles3, _ := filepath.Glob(filepath.Join(dir, "*.sqlite3"))
	allSqlite := append(append(sqliteFiles, sqliteFiles2...), sqliteFiles3...)
	for _, dbFile := range allSqlite {
		previews, agentCountMap = extractSqliteMemories(dbFile, previews, agentCountMap)
	}

	vscdbFiles, _ := filepath.Glob(filepath.Join(dir, "*.vscdb"))
	for _, vscdbFile := range vscdbFiles {
		previews, agentCountMap = extractVscdbMemories(vscdbFile, previews, agentCountMap)
	}

	wsStorageDir := filepath.Join(dir, "User", "workspaceStorage")
	if info, err := os.Stat(wsStorageDir); err == nil && info.IsDir() {
		previews, agentCountMap = extractWorkspaceStorage(wsStorageDir, previews, agentCountMap)
	}

	return previews, agentCountMap
}

func extractSessionMemories(sessionsDir string, previews []memoryPreview, agentCountMap map[string]int) ([]memoryPreview, map[string]int) {
	agentDirs, _ := os.ReadDir(sessionsDir)
	for _, ad := range agentDirs {
		if !ad.IsDir() {
			continue
		}
		agentName := ad.Name()
		sessDir := filepath.Join(sessionsDir, agentName, "sessions")
		if info, err := os.Stat(sessDir); err != nil || !info.IsDir() {
			continue
		}
		files, _ := os.ReadDir(sessDir)
		for _, f := range files {
			if f.IsDir() {
				continue
			}
			ext := strings.ToLower(filepath.Ext(f.Name()))
			if ext != ".jsonl" {
				continue
			}
			path := filepath.Join(sessDir, f.Name())
			data, err := os.ReadFile(path)
			if err != nil || len(data) == 0 {
				continue
			}
			pvs, cnt := parseJSONLSession(string(data), path, agentName)
			previews = append(previews, pvs...)
			agentCountMap[agentName] += cnt
		}
	}
	return previews, agentCountMap
}

func parseJSONLSession(content string, filePath string, agentName string) ([]memoryPreview, int) {
	var previews []memoryPreview
	count := 0
	lines := strings.Split(content, "\n")
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
		text, _ := msg["text"].(string)
		if text == "" {
			text, _ = msg["content"].(string)
		}
		if text == "" {
			continue
		}

		isUserMsg := (msgType == "user" || msgType == "human")
		role := msgType
		if role == "" {
			role, _ = msg["role"].(string)
			if role == "" {
				role = "unknown"
			}
		}

		key := role + ": "
		if len(text) > 40 {
			key += text[:40] + "..."
		} else {
			key += text
		}

		preview := text
		if len(preview) > 200 {
			preview = preview[:200] + "..."
		}

		layer := "episodic"
		if isUserMsg {
			layer = "episodic"
		} else {
			layer = "semantic"
		}

		source := "session"
		if strings.Contains(filePath, "sessions") {
			source = "openclaw-session"
		}

		previews = append(previews, memoryPreview{
			Key: key, Content: preview, Layer: layer,
			Source: source, FilePath: filePath, AgentName: agentName,
		})
		count++
	}
	return previews, count
}

func extractWorkspaceMemory(memoryDir string, baseDir string, previews []memoryPreview, agentCountMap map[string]int) ([]memoryPreview, map[string]int) {
	files, _ := os.ReadDir(memoryDir)
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(f.Name()))
		if ext != ".md" {
			continue
		}
		path := filepath.Join(memoryDir, f.Name())
		data, err := os.ReadFile(path)
		if err != nil || len(data) == 0 {
			continue
		}
		previews, agentCountMap = parseMarkdownMemory(string(data), path, "workspace", previews, agentCountMap)
	}
	return previews, agentCountMap
}

func extractWorkspaceStorage(wsDir string, previews []memoryPreview, agentCountMap map[string]int) ([]memoryPreview, map[string]int) {
	entries, err := os.ReadDir(wsDir)
	if err != nil {
		return previews, agentCountMap
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		subDir := filepath.Join(wsDir, entry.Name())
		vscdbFiles, _ := filepath.Glob(filepath.Join(subDir, "*.vscdb"))
		for _, vscdbFile := range vscdbFiles {
			previews, agentCountMap = extractVscdbMemories(vscdbFile, previews, agentCountMap)
		}
		dbFiles, _ := filepath.Glob(filepath.Join(subDir, "*.db"))
		for _, dbFile := range dbFiles {
			previews, agentCountMap = extractSqliteMemories(dbFile, previews, agentCountMap)
		}
	}
	return previews, agentCountMap
}

func extractClaudeProjects(projectsDir string, previews []memoryPreview, agentCountMap map[string]int) ([]memoryPreview, map[string]int) {
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return previews, agentCountMap
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		projectDir := filepath.Join(projectsDir, entry.Name())
		files, err := os.ReadDir(projectDir)
		if err != nil {
			continue
		}
		for _, f := range files {
			if f.IsDir() {
				continue
			}
			ext := strings.ToLower(filepath.Ext(f.Name()))
			if ext != ".jsonl" && ext != ".md" {
				continue
			}
			path := filepath.Join(projectDir, f.Name())
			data, err := os.ReadFile(path)
			if err != nil || len(data) == 0 {
				continue
			}
			if ext == ".jsonl" {
				pvs, cnt := parseJSONLSession(string(data), path, "claude")
				previews = append(previews, pvs...)
				agentCountMap["claude"] += cnt
			} else if ext == ".md" {
				previews, agentCountMap = parseMarkdownMemory(string(data), path, "claude", previews, agentCountMap)
			}
		}
	}
	return previews, agentCountMap
}

func extractSessionDir(sessionsDir string, previews []memoryPreview, agentCountMap map[string]int) ([]memoryPreview, map[string]int) {
	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		return previews, agentCountMap
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".json" && ext != ".jsonl" {
			continue
		}
		path := filepath.Join(sessionsDir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil || len(data) == 0 {
			continue
		}
		agentName := "session"
		if ext == ".jsonl" {
			pvs, cnt := parseJSONLSession(string(data), path, agentName)
			previews = append(previews, pvs...)
			agentCountMap[agentName] += cnt
		} else if ext == ".json" {
			var jsonData map[string]interface{}
			if json.Unmarshal(data, &jsonData) == nil {
				if title, ok := jsonData["title"].(string); ok && title != "" {
					key := title
					content := string(data)
					if len(content) > 2000 {
						content = content[:2000]
					}
					previews = append(previews, memoryPreview{
						Key: key, Content: content, Layer: "episodic",
						Source: "session-json", FilePath: path, AgentName: agentName,
					})
					agentCountMap[agentName]++
				}
			}
		}
	}
	return previews, agentCountMap
}

func extractVscdbMemories(dbPath string, previews []memoryPreview, agentCountMap map[string]int) ([]memoryPreview, map[string]int) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return previews, agentCountMap
	}
	sqlDB, _ := db.DB()
	if sqlDB != nil {
		defer sqlDB.Close()
	}

	agentName := "vscdb"
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
		return previews, agentCountMap
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
		var entries []struct {
			InputText string `json:"inputText"`
		}
		if json.Unmarshal([]byte(row.Value), &entries) != nil {
			continue
		}
		for _, e := range entries {
			if len(e.InputText) < 10 {
				continue
			}
			key := agentName + "-chat-" + fmt.Sprintf("%x", md5.Sum([]byte(e.InputText)))[:12]
			previews = append(previews, memoryPreview{
				Key: key, Content: e.InputText, Layer: "episodic",
				Source: agentName + "-chat", FilePath: dbPath, AgentName: agentName,
			})
			agentCountMap[agentName]++
		}
	}

	var sessionRows []KV
	db.Raw("SELECT key, value FROM ItemTable WHERE key LIKE 'session:%'").Scan(&sessionRows)
	for _, row := range sessionRows {
		if row.Value == "" || len(row.Value) < 20 {
			continue
		}
		type SessionData struct {
			Title string `json:"title"`
			Cwd   string `json:"cwd"`
		}
		var session SessionData
		if json.Unmarshal([]byte(row.Value), &session) != nil || session.Title == "" {
			continue
		}
		content := session.Title
		if session.Cwd != "" {
			content += " | Project: " + session.Cwd
		}
		key := agentName + "-session-" + fmt.Sprintf("%x", md5.Sum([]byte(row.Value)))[:12]
		previews = append(previews, memoryPreview{
			Key: key, Content: content, Layer: "episodic",
			Source: agentName + "-session", FilePath: dbPath, AgentName: agentName,
		})
		agentCountMap[agentName]++
	}

	var cursorChatRows []KV
	db.Raw("SELECT key, value FROM ItemTable WHERE key LIKE 'cursor%chat%' OR key LIKE 'composerData%'").Scan(&cursorChatRows)
	for _, row := range cursorChatRows {
		if row.Value == "" || len(row.Value) < 50 {
			continue
		}
		key := agentName + "-chat-" + fmt.Sprintf("%x", md5.Sum([]byte(row.Key)))[:12]
		content := row.Value
		if len(content) > 2000 {
			content = content[:2000]
		}
		previews = append(previews, memoryPreview{
			Key: key, Content: content, Layer: "episodic",
			Source: agentName + "-chat", FilePath: dbPath, AgentName: agentName,
		})
		agentCountMap[agentName]++
	}

	return previews, agentCountMap
}

func parseMarkdownMemory(content string, filePath string, agentName string, previews []memoryPreview, agentCountMap map[string]int) ([]memoryPreview, map[string]int) {
	lines := strings.Split(content, "\n")
	currentSection := ""
	currentContent := ""
	for _, line := range lines {
		if strings.HasPrefix(line, "# ") {
			if currentSection != "" || currentContent != "" {
				key := currentSection
				if key == "" {
					key = filepath.Base(filePath)
				}
				preview := strings.TrimSpace(currentContent)
				if len(preview) > 100 {
					preview = preview[:100] + "..."
				}
				previews = append(previews, memoryPreview{Key: key, Content: preview, Layer: "knowledge", Source: "markdown", FilePath: filePath, AgentName: agentName})
				agentCountMap[agentName]++
			}
			currentSection = strings.TrimSpace(line[2:])
			currentContent = ""
		} else {
			currentContent += line + "\n"
		}
	}
	if currentSection != "" || currentContent != "" {
		key := currentSection
		if key == "" {
			key = filepath.Base(filePath)
		}
		preview := strings.TrimSpace(currentContent)
		if len(preview) > 100 {
			preview = preview[:100] + "..."
		}
		previews = append(previews, memoryPreview{Key: key, Content: preview, Layer: "knowledge", Source: "markdown", FilePath: filePath, AgentName: agentName})
		agentCountMap[agentName]++
	}
	return previews, agentCountMap
}

func extractSqliteMemories(dbPath string, previews []memoryPreview, agentCountMap map[string]int) ([]memoryPreview, map[string]int) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return previews, agentCountMap
	}
	sqlDB, err := db.DB()
	if err != nil {
		return previews, agentCountMap
	}
	defer sqlDB.Close()

	type TableName struct {
		Name string
	}
	var tables []TableName
	db.Raw("SELECT name FROM sqlite_master WHERE type='table'").Scan(&tables)

	agentName := filepath.Base(filepath.Dir(dbPath))
	if agentName == "" || agentName == "." {
		agentName = "sqlite-" + filepath.Base(dbPath)
	}

	for _, t := range tables {
		type ColInfo struct {
			CID       int
			Name      string
			Type      string
			NotNull   int
			DfltValue interface{}
			PK        int
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
			if valueCol == "" && (cl == "value" || cl == "content" || cl == "text" || cl == "description" || cl == "body" || cl == "message") {
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
			preview := kv.Value
			if len(preview) > 200 {
				preview = preview[:200] + "..."
			}
			previews = append(previews, memoryPreview{
				Key: kv.Key, Content: preview, Layer: "knowledge",
				Source: "sqlite", FilePath: dbPath, AgentName: agentName,
			})
			agentCountMap[agentName]++
		}
	}

	return previews, agentCountMap
}

func handleScanOpenClawMemories(c *gin.Context) {
	searchDirs := getOpenClawSearchDirs()

	var allPreviews []memoryPreview
	agentCountMap := make(map[string]int)
	var foundDirs []string

	for _, dir := range searchDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}

		previews, acm := extractMemoriesFromDir(dir)
		if len(previews) > 0 {
			allPreviews = append(allPreviews, previews...)
			for name, count := range acm {
				agentCountMap[name] += count
			}
			foundDirs = append(foundDirs, dir)
		}
	}

	if len(allPreviews) > 0 {
		agents := make([]map[string]interface{}, 0)
		for name, count := range agentCountMap {
			agentPreviews := make([]map[string]interface{}, 0)
			for _, p := range allPreviews {
				if p.AgentName == name {
					agentPreviews = append(agentPreviews, map[string]interface{}{
						"key":    p.Key,
						"value":  p.Content,
						"layer":  p.Layer,
						"source": p.Source,
					})
				}
			}
			agents = append(agents, map[string]interface{}{
				"agent_name":   name,
				"layout":       "v2",
				"files":        count,
				"memory_count": count,
				"previews":     agentPreviews,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"found":          true,
			"scanned_dirs":   strings.Join(foundDirs, ", "),
			"openclaw_dir":   strings.Join(foundDirs, ", "),
			"clients":        services.DetectInstalledClients(),
			"agents":         agents,
			"total_memories": len(allPreviews),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"found": false,
	})
}

func handleScanOpenClawAgent(c *gin.Context) {
	agentName := c.Param("agentName")
	searchDirs := getOpenClawSearchDirs()

	var allFiltered []map[string]interface{}

	for _, dir := range searchDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}

		previews, _ := extractMemoriesFromDir(dir)

		for _, p := range previews {
			if p.AgentName == agentName {
				allFiltered = append(allFiltered, map[string]interface{}{
					"key":    p.Key,
					"value":  p.Content,
					"layer":  p.Layer,
					"source": p.Source,
				})
			}
		}
	}

	if len(allFiltered) > 0 {
		c.JSON(http.StatusOK, gin.H{
			"agent_name": agentName,
			"preview":    allFiltered,
			"total":      len(allFiltered),
		})
		return
	}

	c.JSON(http.StatusNotFound, gin.H{
		"error": "agent not found",
	})
}

func handleImportOpenClawMemories(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		var req struct {
			AgentName    string `json:"agent_name"`
			Layer        string `json:"layer"`
			SkipExisting bool   `json:"skip_existing"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if req.Layer == "" {
			req.Layer = "knowledge"
		}

		searchDirs := getOpenClawSearchDirs()

		var imported, skipped, errorsCount int

		seenKeys := make(map[string]bool)
		var existingKeys []string
		db.Table("memories").Where("user_id = ? AND status != ?", userID, "trashed").Pluck("key", &existingKeys)
		for _, k := range existingKeys {
			seenKeys[k] = true
		}

		for _, dir := range searchDirs {
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				continue
			}

			previews, _ := extractMemoriesFromDir(dir)

			for _, p := range previews {
				if req.AgentName != "" && p.AgentName != req.AgentName {
					continue
				}

				if p.Key == "" {
					errorsCount++
					continue
				}

				if seenKeys[p.Key] {
					skipped++
					continue
				}

				if req.SkipExisting {
					var count int64
					logDBErr("count memories by key for import", db.Table("memories").Where("user_id = ? AND key = ?", userID, p.Key).Count(&count).Error)
					if count > 0 {
						skipped++
						seenKeys[p.Key] = true
						continue
					}
				}

				fullContent := p.Content
				if strings.HasSuffix(fullContent, "...") {
					data, err := os.ReadFile(p.FilePath)
					if err == nil {
						ext := strings.ToLower(filepath.Ext(p.FilePath))
						if ext == ".jsonl" {
							lines := strings.Split(string(data), "\n")
							for _, line := range lines {
								line = strings.TrimSpace(line)
								if line == "" {
									continue
								}
								var msg map[string]interface{}
								if json.Unmarshal([]byte(line), &msg) != nil {
									continue
								}
								text, _ := msg["text"].(string)
								if text == "" {
									text, _ = msg["content"].(string)
								}
								key := ""
								msgType, _ := msg["type"].(string)
								if text != "" {
									key = msgType + ": "
									if len(text) > 40 {
										key += text[:40] + "..."
									} else {
										key += text
									}
								}
								if key == p.Key && text != "" {
									fullContent = text
									break
								}
							}
						} else if ext == ".json" {
							var jsonItems []map[string]interface{}
							if json.Unmarshal(data, &jsonItems) == nil {
								for _, m := range jsonItems {
									key, _ := m["key"].(string)
									if key == "" {
										key, _ = m["name"].(string)
									}
									if key == p.Key {
										if v, ok := m["content"].(string); ok && v != "" {
											fullContent = v
										} else if v, ok := m["value"].(string); ok && v != "" {
											fullContent = v
										} else if v, ok := m["text"].(string); ok && v != "" {
											fullContent = v
										} else if v, ok := m["description"].(string); ok && v != "" {
											fullContent = v
										}
										break
									}
								}
							} else {
								var single map[string]interface{}
								if json.Unmarshal(data, &single) == nil {
									if v, ok := single["content"].(string); ok && v != "" {
										fullContent = v
									} else if v, ok := single["value"].(string); ok && v != "" {
										fullContent = v
									} else if v, ok := single["text"].(string); ok && v != "" {
										fullContent = v
									}
								}
							}
						} else if ext == ".md" {
							mdLines := strings.Split(string(data), "\n")
							curSection := ""
							curContent := ""
							for _, l := range mdLines {
								if strings.HasPrefix(l, "# ") {
									if curSection == p.Key && curContent != "" {
										fullContent = strings.TrimSpace(curContent)
										break
									}
									curSection = strings.TrimSpace(l[2:])
									curContent = ""
								} else {
									curContent += l + "\n"
								}
							}
							if curSection == p.Key && fullContent == p.Content {
								fullContent = strings.TrimSpace(curContent)
							}
							if fullContent == p.Content {
								fullContent = strings.TrimSpace(string(data))
							}
						} else {
							fullContent = string(data)
						}
					}
				}

				if fullContent == "" {
					errorsCount++
					continue
				}

				layer := req.Layer
				if p.Layer != "" {
					layer = p.Layer
				}

				memSvc := services.NewMemoryService(db)
				_, err := memSvc.Create(userID, map[string]interface{}{
					"key":        p.Key,
					"value":      fullContent,
					"layer":      layer,
					"importance": 0.5,
					"source":     p.Source,
				})
				if err != nil {
					errorsCount++
				} else {
					imported++
					seenKeys[p.Key] = true
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"imported": imported,
			"skipped":  skipped,
			"errors":   errorsCount,
		})
	}
}

func handleAutoImportMemories(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		svc := services.NewImportService(db)
		result := svc.AutoImport(userID)
		c.JSON(http.StatusOK, gin.H{
			"imported":         result.Imported,
			"skipped":          result.Skipped,
			"entities_created": result.EntitiesCreated,
			"files_found":      result.FoundFiles,
			"message":          fmt.Sprintf("imported %d memories, created %d entities, skipped %d", result.Imported, result.EntitiesCreated, result.Skipped),
		})
	}
}
