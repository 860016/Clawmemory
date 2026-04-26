package services

import (
	"encoding/json"
	"strings"

	"clawmemory/internal/models"

	"gorm.io/gorm"
)

type WikiService struct {
	db *gorm.DB
}

func NewWikiService(db *gorm.DB) *WikiService {
	return &WikiService{db: db}
}

func (s *WikiService) Create(userID uint, data map[string]interface{}) (*models.WikiPage, error) {
	tags := "[]"
	if t, ok := data["tags"].([]string); ok {
		b, _ := json.Marshal(t)
		tags = string(b)
	} else if t, ok := data["tags"].([]interface{}); ok {
		tagStrs := make([]string, 0, len(t))
		for _, v := range t {
			if s, ok := v.(string); ok {
				tagStrs = append(tagStrs, s)
			}
		}
		b, _ := json.Marshal(tagStrs)
		tags = string(b)
	}

	keyDecisions := "[]"
	if kd, ok := data["key_decisions"]; ok {
		switch v := kd.(type) {
		case []string:
			b, _ := json.Marshal(v)
			keyDecisions = string(b)
		case []interface{}:
			items := make([]string, 0, len(v))
			for _, item := range v {
				if s, ok := item.(string); ok {
					items = append(items, s)
				}
			}
			b, _ := json.Marshal(items)
			keyDecisions = string(b)
		case string:
			keyDecisions = v
		}
	}

	actionItems := "[]"
	if ai, ok := data["action_items"]; ok {
		switch v := ai.(type) {
		case []string:
			b, _ := json.Marshal(v)
			actionItems = string(b)
		case []interface{}:
			items := make([]string, 0, len(v))
			for _, item := range v {
				if s, ok := item.(string); ok {
					items = append(items, s)
				}
			}
			b, _ := json.Marshal(items)
			actionItems = string(b)
		case string:
			actionItems = v
		}
	}

	page := &models.WikiPage{
		UserID:       userID,
		Title:        getString(data, "title", ""),
		Content:      getString(data, "content", ""),
		Category:     getString(data, "category", ""),
		Tags:         tags,
		Status:       getString(data, "status", "draft"),
		Summary:      getString(data, "summary", ""),
		IsPublic:     getBool(data, "is_public", false),
		IsPinned:     getBool(data, "is_pinned", false),
		AIGenerated:  getBool(data, "ai_generated", false),
		AIConfidence: getFloat(data, "ai_confidence", 0),
		KeyDecisions: keyDecisions,
		ActionItems:  actionItems,
	}

	if parentID, ok := data["parent_id"]; ok {
		if pid, ok := parentID.(float64); ok && pid > 0 {
			uid := uint(pid)
			page.ParentID = &uid
		}
	}

	if err := s.db.Create(page).Error; err != nil {
		return nil, err
	}
	return page, nil
}

func (s *WikiService) List(userID uint, category, status string, page, size int) ([]models.WikiPage, int64, error) {
	var pages []models.WikiPage
	var total int64

	query := s.db.Model(&models.WikiPage{}).Where("user_id = ?", userID)
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)
	err := query.Order("updated_at DESC").Offset((page - 1) * size).Limit(size).Find(&pages).Error
	return pages, total, err
}

func (s *WikiService) Get(userID, id uint) (*models.WikiPage, error) {
	var page models.WikiPage
	if err := s.db.Where("user_id = ? AND id = ?", userID, id).First(&page).Error; err != nil {
		return nil, err
	}
	return &page, nil
}

func (s *WikiService) Update(userID, id uint, data map[string]interface{}) (*models.WikiPage, error) {
	page, err := s.Get(userID, id)
	if err != nil {
		return nil, err
	}

	updates := map[string]interface{}{}
	if v, ok := data["title"]; ok {
		updates["title"] = v
	}
	if v, ok := data["content"]; ok {
		updates["content"] = v
	}
	if v, ok := data["category"]; ok {
		updates["category"] = v
	}
	if v, ok := data["status"]; ok {
		updates["status"] = v
	}
	if v, ok := data["summary"]; ok {
		updates["summary"] = v
	}
	if v, ok := data["is_pinned"]; ok {
		updates["is_pinned"] = v
	}
	if v, ok := data["is_public"]; ok {
		updates["is_public"] = v
	}
	if v, ok := data["parent_id"]; ok {
		if v == nil {
			updates["parent_id"] = nil
		} else {
			updates["parent_id"] = v
		}
	}
	if t, ok := data["tags"]; ok {
		switch v := t.(type) {
		case []string:
			b, _ := json.Marshal(v)
			updates["tags"] = string(b)
		case []interface{}:
			tagStrs := make([]string, 0, len(v))
			for _, item := range v {
				if s, ok := item.(string); ok {
					tagStrs = append(tagStrs, s)
				}
			}
			b, _ := json.Marshal(tagStrs)
			updates["tags"] = string(b)
		}
	}

	if err := s.db.Model(page).Updates(updates).Error; err != nil {
		return nil, err
	}
	return s.Get(userID, id)
}

func (s *WikiService) Delete(userID, id uint) error {
	return s.db.Where("user_id = ? AND id = ?", userID, id).Delete(&models.WikiPage{}).Error
}

func (s *WikiService) GetCategories(userID uint) ([]string, error) {
	var categories []string
	err := s.db.Model(&models.WikiPage{}).
		Where("user_id = ?", userID).
		Distinct("category").
		Where("category != ''").
		Pluck("category", &categories).Error
	return categories, err
}

func (s *WikiService) GetStats(userID uint) (map[string]interface{}, error) {
	var total, completed, inProgress, draft, aiGenerated int64
	s.db.Model(&models.WikiPage{}).Where("user_id = ?", userID).Count(&total)
	s.db.Model(&models.WikiPage{}).Where("user_id = ? AND status = ?", userID, "completed").Count(&completed)
	s.db.Model(&models.WikiPage{}).Where("user_id = ? AND status = ?", userID, "in_progress").Count(&inProgress)
	s.db.Model(&models.WikiPage{}).Where("user_id = ? AND status = ?", userID, "draft").Count(&draft)
	s.db.Model(&models.WikiPage{}).Where("user_id = ? AND ai_generated = ?", userID, true).Count(&aiGenerated)

	return map[string]interface{}{
		"total":        total,
		"completed":    completed,
		"in_progress":  inProgress,
		"draft":        draft,
		"ai_generated": aiGenerated,
	}, nil
}

func (s *WikiService) Search(userID uint, query string, limit int) ([]models.WikiPage, error) {
	var pages []models.WikiPage
	err := s.db.Where("user_id = ? AND (title LIKE ? OR content LIKE ?)", userID, "%"+query+"%", "%"+query+"%").
		Order("updated_at DESC").
		Limit(limit).
		Find(&pages).Error
	return pages, err
}

func (s *WikiService) MarkComplete(userID, id uint) error {
	return s.db.Model(&models.WikiPage{}).Where("user_id = ? AND id = ?", userID, id).Update("status", "completed").Error
}

func (s *WikiService) MarkInProgress(userID, id uint) error {
	return s.db.Model(&models.WikiPage{}).Where("user_id = ? AND id = ?", userID, id).Update("status", "in_progress").Error
}

func (s *WikiService) GetTree(userID uint) ([]models.WikiPage, error) {
	var pages []models.WikiPage
	err := s.db.Where("user_id = ?", userID).Order("category, title").Find(&pages).Error
	return pages, err
}

func (s *WikiService) parseTags(tags string) []string {
	if tags == "" || tags == "[]" {
		return []string{}
	}
	var result []string
	if err := json.Unmarshal([]byte(tags), &result); err != nil {
		return strings.Split(tags, ",")
	}
	return result
}
