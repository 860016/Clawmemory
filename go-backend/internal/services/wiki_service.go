package services

import (
	"encoding/json"

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
	}

	page := &models.WikiPage{
		UserID:    userID,
		Title:     getString(data, "title", ""),
		Content:   getString(data, "content", ""),
		Category:  getString(data, "category", ""),
		Tags:      tags,
		IsPublic:  getBool(data, "is_public", false),
	}

	if err := s.db.Create(page).Error; err != nil {
		return nil, err
	}
	return page, nil
}

func (s *WikiService) List(userID uint, category string, page, size int) ([]models.WikiPage, int64, error) {
	var pages []models.WikiPage
	var total int64

	query := s.db.Model(&models.WikiPage{}).Where("user_id = ?", userID)
	if category != "" {
		query = query.Where("category = ?", category)
	}

	query.Count(&total)
	err := query.Order("updated_at DESC").Offset((page - 1) * size).Limit(size).Find(&pages).Error
	return pages, total, err
}

func (s *WikiService) Get(id uint) (*models.WikiPage, error) {
	var page models.WikiPage
	if err := s.db.First(&page, id).Error; err != nil {
		return nil, err
	}
	return &page, nil
}

func (s *WikiService) Update(id uint, data map[string]interface{}) (*models.WikiPage, error) {
	page, err := s.Get(id)
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

	if err := s.db.Model(page).Updates(updates).Error; err != nil {
		return nil, err
	}
	return page, nil
}

func (s *WikiService) Delete(id uint) error {
	return s.db.Delete(&models.WikiPage{}, id).Error
}
