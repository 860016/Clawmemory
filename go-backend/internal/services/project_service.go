package services

import (
	"encoding/json"
	"fmt"
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
