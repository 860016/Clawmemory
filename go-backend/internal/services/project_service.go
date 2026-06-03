package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"clawmemory/internal/models"

	"gorm.io/gorm"
)

type ProjectService struct {
	db *gorm.DB
}

func NewProjectService(db *gorm.DB) *ProjectService {
	return &ProjectService{db: db}
}

func (s *ProjectService) List(userID uint, page, size int, status, category string) ([]models.Project, int64, error) {
	var projects []models.Project
	var total int64
	query := s.db.Model(&models.Project{}).Where("user_id = ?", userID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if category != "" {
		query = query.Where("category = ?", category)
	}
	query.Count(&total)
	err := query.Order("is_pinned DESC, updated_at DESC").Offset((page - 1) * size).Limit(size).Find(&projects).Error
	return projects, total, err
}

func (s *ProjectService) Get(userID uint, id uint) (*models.Project, error) {
	var project models.Project
	if err := s.db.Where("user_id = ? AND id = ?", userID, id).First(&project).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

func (s *ProjectService) Create(userID uint, data map[string]interface{}) (*models.Project, error) {
	project := &models.Project{
		UserID:          userID,
		Name:            getString(data, "name", "Untitled Project"),
		Description:     getString(data, "description", ""),
		Status:          getString(data, "status", "active"),
		Category:        getString(data, "category", ""),
		Tags:            toJSONStr(data["tags"]),
		KeyDecisions:    toJSONStr(data["key_decisions"]),
		ActionItems:     toJSONStr(data["action_items"]),
		SourceAgent:     getString(data, "source_agent", ""),
		SourceSessionID: getString(data, "source_session_id", ""),
	}
	if p, ok := data["progress"].(float64); ok {
		project.Progress = int(p)
	}
	if pinned, ok := data["is_pinned"].(bool); ok {
		project.IsPinned = pinned
	}
	if err := s.db.Create(project).Error; err != nil {
		return nil, err
	}
	return project, nil
}

func (s *ProjectService) Update(userID uint, id uint, data map[string]interface{}) (*models.Project, error) {
	var project models.Project
	if err := s.db.Where("user_id = ? AND id = ?", userID, id).First(&project).Error; err != nil {
		return nil, err
	}
	updates := map[string]interface{}{}
	if v, ok := data["name"].(string); ok {
		updates["name"] = v
	}
	if v, ok := data["description"].(string); ok {
		updates["description"] = v
	}
	if v, ok := data["status"].(string); ok {
		updates["status"] = v
	}
	if v, ok := data["category"].(string); ok {
		updates["category"] = v
	}
	if v, ok := data["tags"]; ok {
		updates["tags"] = toJSONStr(v)
	}
	if v, ok := data["key_decisions"]; ok {
		updates["key_decisions"] = toJSONStr(v)
	}
	if v, ok := data["action_items"]; ok {
		updates["action_items"] = toJSONStr(v)
	}
	if v, ok := data["progress"].(float64); ok {
		updates["progress"] = int(v)
	}
	if v, ok := data["is_pinned"].(bool); ok {
		updates["is_pinned"] = v
	}
	if len(updates) > 0 {
		if err := s.db.Model(&project).Updates(updates).Error; err != nil {
			return nil, err
		}
	}
	if err := s.db.Where("user_id = ? AND id = ?", userID, id).First(&project).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

func (s *ProjectService) Delete(userID uint, id uint) error {
	return s.db.Where("user_id = ? AND id = ?", userID, id).Delete(&models.Project{}).Error
}

func (s *ProjectService) GetNotes(userID uint, projectID uint) ([]models.ProjectNote, error) {
	var notes []models.ProjectNote
	err := s.db.Where("user_id = ? AND project_id = ?", userID, projectID).
		Order("is_key_point DESC, created_at DESC").Find(&notes).Error
	return notes, err
}

func (s *ProjectService) AddNote(userID uint, projectID uint, data map[string]interface{}) (*models.ProjectNote, error) {
	var project models.Project
	if err := s.db.Where("user_id = ? AND id = ?", userID, projectID).First(&project).Error; err != nil {
		return nil, fmt.Errorf("project not found")
	}
	note := &models.ProjectNote{
		UserID:    userID,
		ProjectID: projectID,
		Content:   getString(data, "content", ""),
		NoteType:  getString(data, "note_type", "note"),
		Source:    getString(data, "source", "manual"),
	}
	if kp, ok := data["is_key_point"].(bool); ok {
		note.IsKeyPoint = kp
	}
	if err := s.db.Create(note).Error; err != nil {
		return nil, err
	}
	now := time.Now()
	s.db.Model(&project).Updates(map[string]interface{}{"updated_at": now})
	return note, nil
}

func (s *ProjectService) UpdateNote(userID uint, noteID uint, data map[string]interface{}) (*models.ProjectNote, error) {
	var note models.ProjectNote
	if err := s.db.Where("user_id = ? AND id = ?", userID, noteID).First(&note).Error; err != nil {
		return nil, err
	}
	updates := map[string]interface{}{}
	if v, ok := data["content"].(string); ok {
		updates["content"] = v
	}
	if v, ok := data["note_type"].(string); ok {
		updates["note_type"] = v
	}
	if v, ok := data["is_key_point"].(bool); ok {
		updates["is_key_point"] = v
	}
	if len(updates) > 0 {
		s.db.Model(&note).Updates(updates)
	}
	s.db.Where("user_id = ? AND id = ?", userID, noteID).First(&note)
	return &note, nil
}

func (s *ProjectService) DeleteNote(userID uint, noteID uint) error {
	return s.db.Where("user_id = ? AND id = ?", userID, noteID).Delete(&models.ProjectNote{}).Error
}

func (s *ProjectService) GetCategories(userID uint) ([]string, error) {
	var categories []string
	err := s.db.Model(&models.Project{}).Where("user_id = ? AND category != ''", userID).
		Distinct("category").Pluck("category", &categories).Error
	return categories, err
}

func (s *ProjectService) ExtractFromMemories(userID uint, projectID uint) (int, error) {
	var project models.Project
	if err := s.db.Where("user_id = ? AND id = ?", userID, projectID).First(&project).Error; err != nil {
		return 0, fmt.Errorf("project not found")
	}

	var memories []models.Memory
	escapedName := EscapeLikeQuery(project.Name)
	_ = s.db.Where("user_id = ? AND status != ?", userID, "trashed").
		Where("key LIKE ? OR value LIKE ?", "%"+escapedName+"%", "%"+escapedName+"%").
		Order("importance DESC").Limit(50).Find(&memories).Error

	extracted := 0
	for _, m := range memories {
		var existingCount int64
		s.db.Model(&models.ProjectNote{}).Where("user_id = ? AND project_id = ? AND content = ?", userID, projectID, m.Value).Count(&existingCount)
		if existingCount > 0 {
			continue
		}
		note := &models.ProjectNote{
			UserID:     userID,
			ProjectID:  projectID,
			Content:    m.Value,
			NoteType:   "memory_extract",
			Source:     "memory:" + m.Key,
			IsKeyPoint: m.Importance >= 0.7,
		}
		if err := s.db.Create(note).Error; err == nil {
			extracted++
		}
	}
	return extracted, nil
}

func (s *ProjectService) GetContextForOpenClaw(userID uint, projectName string) (string, error) {
	var project models.Project
	if err := s.db.Where("user_id = ? AND name LIKE ?", userID, "%"+EscapeLikeQuery(projectName)+"%").First(&project).Error; err != nil {
		return "", fmt.Errorf("project not found")
	}

	var notes []models.ProjectNote
	_ = s.db.Where("user_id = ? AND project_id = ?", userID, project.ID).
		Order("is_key_point DESC, created_at DESC").Limit(20).Find(&notes).Error

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Project: %s\n", project.Name))
	if project.Description != "" {
		sb.WriteString(fmt.Sprintf("Description: %s\n", project.Description))
	}
	sb.WriteString(fmt.Sprintf("Status: %s | Progress: %d%%\n\n", project.Status, project.Progress))

	if project.KeyDecisions != "" && project.KeyDecisions != "[]" {
		var decisions []string
		if json.Unmarshal([]byte(project.KeyDecisions), &decisions) == nil && len(decisions) > 0 {
			sb.WriteString("## Key Decisions\n")
			for _, d := range decisions {
				sb.WriteString(fmt.Sprintf("- %s\n", d))
			}
			sb.WriteString("\n")
		}
	}

	if project.ActionItems != "" && project.ActionItems != "[]" {
		var items []string
		if json.Unmarshal([]byte(project.ActionItems), &items) == nil && len(items) > 0 {
			sb.WriteString("## Action Items\n")
			for _, item := range items {
				sb.WriteString(fmt.Sprintf("- [ ] %s\n", item))
			}
			sb.WriteString("\n")
		}
	}

	if len(notes) > 0 {
		sb.WriteString("## Notes\n")
		for _, n := range notes {
			prefix := "  -"
			if n.IsKeyPoint {
				prefix = "  ★"
			}
			sb.WriteString(fmt.Sprintf("%s [%s] %s\n", prefix, n.NoteType, n.Content))
		}
	}

	return sb.String(), nil
}

func (s *ProjectService) Search(userID uint, query string, limit int) ([]models.Project, error) {
	escaped := EscapeLikeQuery(query)
	var projects []models.Project
	err := s.db.Where("user_id = ? AND (name LIKE ? OR description LIKE ? OR category LIKE ?)",
		userID, "%"+escaped+"%", "%"+escaped+"%", "%"+escaped+"%").
		Order("is_pinned DESC, updated_at DESC").Limit(limit).Find(&projects).Error
	return projects, err
}

type DiscoveredProject struct {
	Name              string   `json:"name"`
	Path              string   `json:"path"`
	Description       string   `json:"description"`
	Category          string   `json:"category"`
	Status            string   `json:"status"`
	Progress          int      `json:"progress"`
	KeyDecisions      []string `json:"key_decisions"`
	ActionItems       []string `json:"action_items"`
	Confidence        float64  `json:"confidence"`
	Source            string   `json:"source"`
	MemoryCount       int      `json:"memory_count"`
	FileCount         int      `json:"file_count"`
	ConversationCount int      `json:"conversation_count"`
	Agents            []string `json:"agents"`
}

func (s *ProjectService) DiscoverFromMemories(userID uint) ([]DiscoveredProject, error) {
	var existingProjects []models.Project
	s.db.Where("user_id = ?", userID).Find(&existingProjects)
	existingNames := map[string]bool{}
	existingPaths := map[string]bool{}
	for _, p := range existingProjects {
		existingNames[strings.ToLower(p.Name)] = true
		if p.Description != "" {
			existingPaths[strings.ToLower(p.Description)] = true
		}
	}

	projectGroups := map[string]*projectGroup{}

	var fileIndexes []models.FileSyncIndex
	s.db.Find(&fileIndexes)

	for _, fi := range fileIndexes {
		projectRoot := extractProjectRoot(fi.FilePath)
		if projectRoot == "" {
			continue
		}
		key := strings.ToLower(projectRoot)
		if projectGroups[key] == nil {
			projectGroups[key] = &projectGroup{
				Root:     projectRoot,
				Files:    []string{},
				Agents:   map[string]bool{},
				Platform: map[string]bool{},
			}
		}
		pg := projectGroups[key]
		pg.Files = append(pg.Files, fi.FilePath)
		pg.ChunkCount += fi.ChunkCount
		if fi.Platform != "" {
			pg.Platform[fi.Platform] = true
		}
		if fi.Source != "" {
			pg.Agents[fi.Source] = true
		}
	}

	var conversations []models.Memory
	s.db.Where("user_id = ? AND status != ? AND key LIKE ?", userID, "trashed", "conversation-%").
		Order("importance DESC").Limit(500).Find(&conversations)

	for _, conv := range conversations {
		projectPath := extractProjectPathFromConversation(conv.Value)
		if projectPath == "" {
			continue
		}
		projectRoot := extractProjectRoot(projectPath)
		if projectRoot == "" {
			continue
		}
		key := strings.ToLower(projectRoot)
		if projectGroups[key] == nil {
			projectGroups[key] = &projectGroup{
				Root:          projectRoot,
				Files:         []string{},
				Agents:        map[string]bool{},
				Platform:      map[string]bool{},
				Conversations: []models.Memory{},
			}
		}
		pg := projectGroups[key]
		pg.Conversations = append(pg.Conversations, conv)
		if conv.SourceAgent != "" {
			pg.Agents[conv.SourceAgent] = true
		}
	}

	var results []DiscoveredProject
	for _, pg := range projectGroups {
		totalItems := len(pg.Files) + len(pg.Conversations)
		if totalItems < 2 {
			continue
		}

		name := filepath.Base(pg.Root)
		if name == "" || len(name) < 2 {
			name = filepath.Base(filepath.Dir(pg.Root))
		}
		if existingNames[strings.ToLower(name)] {
			name = name + " (" + filepath.Base(filepath.Dir(pg.Root)) + ")"
		}

		agentList := []string{}
		for a := range pg.Agents {
			agentList = append(agentList, a)
		}

		memoryCount := pg.ChunkCount
		if memoryCount == 0 {
			memoryCount = len(pg.Files) * 3
		}

		confidence := 0.5
		if len(pg.Files) >= 5 {
			confidence += 0.2
		}
		if len(pg.Conversations) >= 3 {
			confidence += 0.15
		}
		if len(pg.Agents) >= 2 {
			confidence += 0.1
		}
		if confidence > 0.95 {
			confidence = 0.95
		}

		desc := fmt.Sprintf("从 %d 个文件和 %d 条对话中发现", len(pg.Files), len(pg.Conversations))

		results = append(results, DiscoveredProject{
			Name:              name,
			Path:              pg.Root,
			Description:       desc,
			Category:          inferCategoryFromFiles(pg.Files),
			Status:            "active",
			Progress:          30,
			Confidence:        confidence,
			Source:            "heuristic:filepath",
			MemoryCount:       memoryCount,
			FileCount:         len(pg.Files),
			ConversationCount: len(pg.Conversations),
			Agents:            agentList,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Confidence > results[j].Confidence
	})

	return results, nil
}

type projectGroup struct {
	Root          string
	Files         []string
	Conversations []models.Memory
	Agents        map[string]bool
	Platform      map[string]bool
	ChunkCount    int
}

func extractProjectRoot(filePath string) string {
	sep := string(filepath.Separator)
	parts := strings.Split(filepath.Clean(filePath), sep)

	projectMarkers := map[string]bool{
		".git": true, "go.mod": true, "package.json": true, "Cargo.toml": true,
		"pom.xml": true, "build.gradle": true, "Gemfile": true, "pyproject.toml": true,
		"requirements.txt": true, ".csproj": true, "composer.json": true,
	}

	for i := len(parts) - 1; i >= 1; i-- {
		dir := strings.Join(parts[:i+1], sep)
		for marker := range projectMarkers {
			candidate := filepath.Join(dir, marker)
			if _, err := os.Stat(candidate); err == nil {
				return dir
			}
		}
	}

	commonRoots := map[string]bool{
		"src": true, "projects": true, "code": true, "workspace": true,
		"repos": true, "dev": true, "work": true, "home": true,
		"Users": true, "Developer": true,
	}

	for i := 1; i < len(parts)-1; i++ {
		if commonRoots[strings.ToLower(parts[i])] || commonRoots[parts[i]] {
			if i+2 < len(parts) {
				return strings.Join(parts[:i+2], sep)
			}
		}
	}

	if len(parts) >= 4 {
		for i := len(parts) - 2; i >= 2; i-- {
			candidate := strings.Join(parts[:i+1], sep)
			base := filepath.Base(candidate)
			if len(base) > 2 && !commonRoots[strings.ToLower(base)] {
				return candidate
			}
		}
	}

	if len(parts) >= 3 {
		return strings.Join(parts[:3], sep)
	}

	return ""
}

func extractProjectPathFromConversation(value string) string {
	lines := strings.Split(value, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Project: ") {
			return strings.TrimPrefix(line, "Project: ")
		}
	}

	pathPatterns := []string{
		`c:\Users\`, `C:\Users\`, `D:\`,
		`/home/`, `/Users/`, `/root/`,
	}
	lower := strings.ToLower(value)
	for _, pat := range pathPatterns {
		lowerPat := strings.ToLower(pat)
		idx := strings.Index(lower, lowerPat)
		if idx >= 0 {
			rest := value[idx:]
			parts := strings.SplitN(rest, " ", 2)
			path := strings.TrimRight(parts[0], `/\,.`)
			if len(path) > len(pat)+3 {
				return path
			}
		}
	}

	return ""
}

func inferCategoryFromFiles(files []string) string {
	extCounts := map[string]int{}
	for _, f := range files {
		ext := strings.ToLower(filepath.Ext(f))
		extCounts[ext]++
	}

	frontendExts := map[string]bool{".vue": true, ".jsx": true, ".tsx": true, ".svelte": true, ".css": true, ".scss": true, ".html": true}
	backendExts := map[string]bool{".go": true, ".py": true, ".rs": true, ".java": true, ".rb": true, ".php": true}

	feCount, beCount := 0, 0
	for ext, count := range extCounts {
		if frontendExts[ext] {
			feCount += count
		}
		if backendExts[ext] {
			beCount += count
		}
	}

	if feCount > 0 && beCount > 0 {
		return "全栈"
	}
	if feCount > beCount {
		return "前端"
	}
	if beCount > feCount {
		return "后端"
	}
	return "开发"
}

func (s *ProjectService) GenerateWikiFromMemories(userID uint, projectID uint) (int, error) {
	var project models.Project
	if err := s.db.Where("user_id = ? AND id = ?", userID, projectID).First(&project).Error; err != nil {
		return 0, fmt.Errorf("project not found")
	}

	projectPath := project.Description
	if projectPath == "" {
		projectPath = project.Name
	}

	var fileIndexes []models.FileSyncIndex
	escapedPath := EscapeLikeQuery(projectPath)
	s.db.Where("file_path LIKE ?", "%"+escapedPath+"%").Find(&fileIndexes)

	var conversations []models.Memory
	s.db.Where("user_id = ? AND status != ? AND key LIKE ?", userID, "trashed", "conversation-%").
		Where("value LIKE ?", "%"+escapedPath+"%").
		Order("importance DESC").Limit(100).Find(&conversations)

	var relatedMemories []models.Memory
	s.db.Where("user_id = ? AND status != ?", userID, "trashed").
		Where("source_agent IN (?) OR value LIKE ?",
			s.db.Model(&models.Memory{}).Select("DISTINCT source_agent").Where("user_id = ? AND source_agent != ''", userID),
			"%"+escapedPath+"%").
		Order("importance DESC").Limit(200).Find(&relatedMemories)

	if len(fileIndexes) == 0 && len(conversations) == 0 && len(relatedMemories) == 0 {
		return 0, nil
	}

	wikiSvc := NewWikiService(s.db)
	created := 0

	overviewContent := buildRealWikiOverview(&project, fileIndexes, conversations, relatedMemories)
	_, err := wikiSvc.Create(userID, map[string]interface{}{
		"title":    project.Name + " - 项目概览",
		"content":  overviewContent,
		"category": project.Name,
		"status":   "in_progress",
		"tags":     project.Name,
	})
	if err == nil {
		created++
	}

	if len(fileIndexes) > 0 {
		dirStructure := buildFileStructureWiki(&project, fileIndexes)
		_, err := wikiSvc.Create(userID, map[string]interface{}{
			"title":    project.Name + " - 文件结构",
			"content":  dirStructure,
			"category": project.Name,
			"status":   "draft",
			"tags":     project.Name + ",文件结构",
		})
		if err == nil {
			created++
		}
	}

	if len(conversations) > 0 {
		convContent := buildConversationWiki(&project, conversations)
		_, err := wikiSvc.Create(userID, map[string]interface{}{
			"title":    project.Name + " - 对话记录",
			"content":  convContent,
			"category": project.Name,
			"status":   "draft",
			"tags":     project.Name + ",对话",
		})
		if err == nil {
			created++
		}
	}

	if len(relatedMemories) > 0 {
		keyContent := buildKeyMemoriesWiki(&project, relatedMemories)
		_, err := wikiSvc.Create(userID, map[string]interface{}{
			"title":    project.Name + " - 关键记忆",
			"content":  keyContent,
			"category": project.Name,
			"status":   "draft",
			"tags":     project.Name + ",关键记忆",
		})
		if err == nil {
			created++
		}
	}

	return created, nil
}

func buildRealWikiOverview(project *models.Project, files []models.FileSyncIndex, convs []models.Memory, mems []models.Memory) string {
	var sb strings.Builder
	sb.WriteString("# " + project.Name + "\n\n")

	if project.Description != "" {
		sb.WriteString("**项目路径**: " + project.Description + "\n\n")
	}

	sb.WriteString(fmt.Sprintf("- 状态: %s\n", project.Status))
	sb.WriteString(fmt.Sprintf("- 进度: %d%%\n", project.Progress))
	if project.Category != "" {
		sb.WriteString(fmt.Sprintf("- 分类: %s\n", project.Category))
	}
	sb.WriteString(fmt.Sprintf("- 同步文件数: %d\n", len(files)))
	sb.WriteString(fmt.Sprintf("- 相关对话数: %d\n", len(convs)))
	sb.WriteString(fmt.Sprintf("- 关联记忆数: %d\n\n", len(mems)))

	agentSet := map[string]bool{}
	for _, c := range convs {
		if c.SourceAgent != "" {
			agentSet[c.SourceAgent] = true
		}
	}
	if len(agentSet) > 0 {
		sb.WriteString("## 涉及的 Agent\n\n")
		for a := range agentSet {
			sb.WriteString("- " + a + "\n")
		}
		sb.WriteString("\n")
	}

	if len(files) > 0 {
		sb.WriteString("## 文件类型分布\n\n")
		extCounts := map[string]int{}
		for _, f := range files {
			ext := strings.ToLower(filepath.Ext(f.FilePath))
			if ext == "" {
				ext = "(无扩展名)"
			}
			extCounts[ext]++
		}
		for ext, count := range extCounts {
			sb.WriteString(fmt.Sprintf("- `%s`: %d 个文件\n", ext, count))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func buildFileStructureWiki(project *models.Project, files []models.FileSyncIndex) string {
	var sb strings.Builder
	sb.WriteString("# " + project.Name + " - 文件结构\n\n")

	dirFiles := map[string][]string{}
	for _, f := range files {
		dir := filepath.Dir(f.FilePath)
		dirFiles[dir] = append(dirFiles[dir], filepath.Base(f.FilePath))
	}

	for dir, filenames := range dirFiles {
		sb.WriteString("## " + dir + "\n\n")
		for _, name := range filenames {
			sb.WriteString("- " + name + "\n")
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func buildConversationWiki(project *models.Project, convs []models.Memory) string {
	var sb strings.Builder
	sb.WriteString("# " + project.Name + " - 对话记录\n\n")

	agentConvs := map[string][]models.Memory{}
	for _, c := range convs {
		agent := c.SourceAgent
		if agent == "" {
			agent = "unknown"
		}
		agentConvs[agent] = append(agentConvs[agent], c)
	}

	for agent, mems := range agentConvs {
		sb.WriteString("## " + agent + "\n\n")
		for i, m := range mems {
			if i >= 20 {
				sb.WriteString(fmt.Sprintf("\n... 还有 %d 条对话\n", len(mems)-20))
				break
			}
			val := m.Value
			if len(val) > 500 {
				val = val[:500] + "..."
			}
			sb.WriteString("### " + m.Key + "\n\n")
			sb.WriteString(val + "\n\n")
		}
	}

	return sb.String()
}

func buildKeyMemoriesWiki(project *models.Project, mems []models.Memory) string {
	var sb strings.Builder
	sb.WriteString("# " + project.Name + " - 关键记忆\n\n")

	for i, m := range mems {
		if i >= 50 {
			sb.WriteString(fmt.Sprintf("\n... 还有 %d 条记忆\n", len(mems)-50))
			break
		}
		val := m.Value
		if len(val) > 300 {
			val = val[:300] + "..."
		}
		sb.WriteString(fmt.Sprintf("### %s (重要度: %.0f%%)\n\n", m.Key, m.Importance*100))
		sb.WriteString(val + "\n\n")
	}

	return sb.String()
}

func buildWikiOverview(project *models.Project, memories []models.Memory) string {
	var sb strings.Builder
	sb.WriteString("# " + project.Name + "\n\n")
	if project.Description != "" {
		sb.WriteString(project.Description + "\n\n")
	}
	sb.WriteString(fmt.Sprintf("- 状态: %s\n", project.Status))
	sb.WriteString(fmt.Sprintf("- 进度: %d%%\n", project.Progress))
	if project.Category != "" {
		sb.WriteString(fmt.Sprintf("- 分类: %s\n", project.Category))
	}
	sb.WriteString(fmt.Sprintf("- 关联记忆数: %d\n\n", len(memories)))

	if project.KeyDecisions != "" && project.KeyDecisions != "[]" {
		var decisions []string
		if json.Unmarshal([]byte(project.KeyDecisions), &decisions) == nil && len(decisions) > 0 {
			sb.WriteString("## 关键决策\n\n")
			for _, d := range decisions {
				sb.WriteString("- " + d + "\n")
			}
			sb.WriteString("\n")
		}
	}

	if project.ActionItems != "" && project.ActionItems != "[]" {
		var items []string
		if json.Unmarshal([]byte(project.ActionItems), &items) == nil && len(items) > 0 {
			sb.WriteString("## 待办事项\n\n")
			for _, item := range items {
				sb.WriteString("- [ ] " + item + "\n")
			}
			sb.WriteString("\n")
		}
	}

	sb.WriteString("## 重要记忆摘要\n\n")
	for i, m := range memories {
		if i >= 10 {
			break
		}
		if m.Importance >= 0.6 {
			val := m.Value
			if len(val) > 200 {
				val = val[:200] + "..."
			}
			sb.WriteString(fmt.Sprintf("- **%s**: %s\n", m.Key, val))
		}
	}
	return sb.String()
}

func buildWikiSection(section string, mems []models.Memory) string {
	var sb strings.Builder
	sb.WriteString("# " + layerLabel(section) + "\n\n")
	for i, m := range mems {
		if i >= 30 {
			sb.WriteString(fmt.Sprintf("\n... 还有 %d 条记忆\n", len(mems)-30))
			break
		}
		val := m.Value
		if len(val) > 300 {
			val = val[:300] + "..."
		}
		sb.WriteString(fmt.Sprintf("### %s\n\n%s\n\n", m.Key, val))
	}
	return sb.String()
}

func layerLabel(layer string) string {
	labels := map[string]string{
		"context":    "核心上下文",
		"detail":     "详细信息",
		"episodic":   "事件记录",
		"semantic":   "语义知识",
		"procedural": "过程知识",
	}
	if label, ok := labels[layer]; ok {
		return label
	}
	if strings.HasPrefix(layer, "agent:") {
		return strings.TrimPrefix(layer, "agent:")
	}
	return layer
}

func extractProjectDir(key string) string {
	if !strings.HasPrefix(key, "chunk:") {
		return ""
	}
	parts := strings.SplitN(key, ":", 4)
	if len(parts) < 3 {
		return ""
	}
	filename := parts[2]
	filename = strings.TrimSuffix(filename, ".md")
	if strings.Contains(filename, "-") {
		segments := strings.Split(filename, "-")
		for _, seg := range segments {
			if len(seg) > 2 && seg[0] >= 'A' && seg[0] <= 'Z' {
				return seg
			}
		}
	}
	if len(filename) > 2 && filename[0] >= 'A' && filename[0] <= 'Z' {
		return filename
	}
	return ""
}

func buildProjectDescription(mems []models.Memory) string {
	var parts []string
	for i, m := range mems {
		if i >= 3 {
			break
		}
		val := m.Value
		if len(val) > 150 {
			val = val[:150] + "..."
		}
		parts = append(parts, val)
	}
	return strings.Join(parts, " | ")
}

func inferCategory(mems []models.Memory) string {
	techKeywords := map[string]string{
		"frontend": "前端", "react": "前端", "vue": "前端", "angular": "前端", "css": "前端", "html": "前端",
		"backend": "后端", "api": "后端", "server": "后端", "database": "后端", "go": "后端", "python": "后端", "node": "后端",
		"mobile": "移动端", "ios": "移动端", "android": "移动端", "flutter": "移动端", "react-native": "移动端",
		"devops": "运维", "docker": "运维", "kubernetes": "运维", "ci/cd": "运维", "deploy": "运维",
		"test": "测试", "testing": "测试", "e2e": "测试", "unit test": "测试",
		"design": "设计", "ui": "设计", "ux": "设计", "figma": "设计",
		"data": "数据", "analytics": "数据", "ml": "数据", "ai": "数据",
	}

	counts := map[string]int{}
	for _, m := range mems {
		lower := strings.ToLower(m.Value + " " + m.Key + " " + m.Tags)
		for kw, cat := range techKeywords {
			if strings.Contains(lower, kw) {
				counts[cat]++
			}
		}
	}

	maxCat := ""
	maxCount := 0
	for cat, count := range counts {
		if count > maxCount {
			maxCount = count
			maxCat = cat
		}
	}
	if maxCat == "" {
		maxCat = "其他"
	}
	return maxCat
}

func estimateProgress(mems []models.Memory) int {
	done := 0
	total := 0
	for _, m := range mems {
		lower := strings.ToLower(m.Value)
		if strings.Contains(lower, "done") || strings.Contains(lower, "completed") ||
			strings.Contains(lower, "完成") || strings.Contains(lower, "已实现") ||
			strings.Contains(lower, "fixed") || strings.Contains(lower, "resolved") {
			done++
		}
		if strings.Contains(lower, "todo") || strings.Contains(lower, "pending") ||
			strings.Contains(lower, "待办") || strings.Contains(lower, "未完成") ||
			strings.Contains(lower, "need to") || strings.Contains(lower, "should") {
			total++
		}
		total++
	}
	if total == 0 {
		return 30
	}
	pct := done * 100 / total
	if pct > 95 {
		pct = 95
	}
	if pct < 10 {
		pct = 10
	}
	return pct
}

func extractKeyDecisions(mems []models.Memory) []string {
	var decisions []string
	keywords := []string{"decided", "decision", "决定", "选择", "use ", "using ", "采用", "switch to", "migrate"}
	for _, m := range mems {
		lower := strings.ToLower(m.Value)
		for _, kw := range keywords {
			if strings.Contains(lower, kw) {
				line := m.Value
				if len(line) > 100 {
					line = line[:100] + "..."
				}
				decisions = append(decisions, line)
				break
			}
		}
		if len(decisions) >= 5 {
			break
		}
	}
	return decisions
}

func extractActionItems(mems []models.Memory) []string {
	var items []string
	keywords := []string{"todo:", "need to", "should", "must", "待办", "需要", "计划", "will ", "going to"}
	for _, m := range mems {
		lower := strings.ToLower(m.Value)
		for _, kw := range keywords {
			if strings.Contains(lower, kw) {
				line := m.Value
				if len(line) > 100 {
					line = line[:100] + "..."
				}
				items = append(items, line)
				break
			}
		}
		if len(items) >= 5 {
			break
		}
	}
	return items
}

func calcConfidence(mems []models.Memory, count int) float64 {
	conf := 0.3
	if count >= 10 {
		conf += 0.2
	}
	if count >= 30 {
		conf += 0.2
	}
	hasHighImportance := false
	for _, m := range mems {
		if m.Importance >= 0.7 {
			hasHighImportance = true
			break
		}
	}
	if hasHighImportance {
		conf += 0.15
	}
	if conf > 0.95 {
		conf = 0.95
	}
	return conf
}

func groupByTag(mems []models.Memory) map[string][]models.Memory {
	groups := map[string][]models.Memory{}
	for _, m := range mems {
		if m.Tags == "" {
			continue
		}
		for _, tag := range parseMemoryTags(m.Tags) {
			if tag == "" || len(tag) < 2 {
				continue
			}
			groups[tag] = append(groups[tag], m)
		}
	}
	return groups
}
