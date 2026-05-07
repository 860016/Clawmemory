package services

import (
	"encoding/json"
	"time"

	"clawmemory/internal/models"

	"gorm.io/gorm"
)

type MemoryService struct {
	db *gorm.DB
}

// MemoryModel 用于返回的记忆模型
type MemoryModel struct {
	ID             uint       `json:"id"`
	UserID         uint       `json:"user_id"`
	Layer          string     `json:"layer"`
	Key            string     `json:"key"`
	Value          string     `json:"value"`
	Summary        string     `json:"summary"`
	Importance     float64    `json:"importance"`
	AccessCount    int        `json:"access_count"`
	LastAccessedAt *time.Time `json:"last_accessed_at"`
	IsEncrypted    bool       `json:"is_encrypted"`
	Tags           []string   `json:"tags"`
	Source         string     `json:"source"`
	Platform       string     `json:"platform"`
	Status         string     `json:"status"`
	TrashedAt      *time.Time `json:"trashed_at"`
	DecayStage     int        `json:"decay_stage"`
	ReinforceCount int        `json:"reinforce_count"`
	MemoryType     string     `json:"memory_type"`
	VerifiedAt     *time.Time `json:"verified_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func NewMemoryService(db *gorm.DB) *MemoryService {
	return &MemoryService{db: db}
}

func (s *MemoryService) Create(userID uint, data map[string]interface{}) (*MemoryModel, error) {
	tags := "[]"
	if t, ok := data["tags"].([]interface{}); ok {
		tagStrings := make([]string, len(t))
		for i, v := range t {
			tagStrings[i] = v.(string)
		}
		b, _ := json.Marshal(tagStrings)
		tags = string(b)
	}

	memory := &models.Memory{
		UserID:      userID,
		Layer:       getString(data, "layer", "short"),
		Key:         getString(data, "key", ""),
		Value:       getString(data, "value", ""),
		Importance:  getFloat(data, "importance", 0.5),
		Tags:        tags,
		Summary:     generateSummary(getString(data, "key", ""), getString(data, "value", "")),
		Source:      getString(data, "source", "manual"),
		Platform:    getString(data, "platform", "openclaw"),
		MemoryType:  getString(data, "memory_type", "knowledge"),
		IsEncrypted: getBool(data, "is_encrypted", false),
		Status:      "active",
		DecayStage:  0,
	}

	if err := s.db.Create(memory).Error; err != nil {
		return nil, err
	}
	return ToMemoryModel(memory), nil
}

func (s *MemoryService) Get(userID, id uint) (*MemoryModel, error) {
	var memory models.Memory
	if err := s.db.Where("user_id = ? AND id = ?", userID, id).First(&memory).Error; err != nil {
		return nil, err
	}
	return ToMemoryModel(&memory), nil
}

func (s *MemoryService) List(userID uint, layer string, page, size int, status string, memoryType ...string) ([]*MemoryModel, int64, error) {
	var memories []models.Memory
	var total int64

	query := s.db.Model(&models.Memory{}).Where("user_id = ?", userID)

	if layer != "" {
		query = query.Where("layer = ?", layer)
	}

	if len(memoryType) > 0 && memoryType[0] != "" {
		query = query.Where("memory_type = ?", memoryType[0])
	}

	if status != "" {
		query = query.Where("status = ?", status)
	} else {
		query = query.Where("status != ?", "trashed")
	}

	query.Count(&total)
	err := query.Order("updated_at DESC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&memories).Error

	result := make([]*MemoryModel, len(memories))
	for i, m := range memories {
		result[i] = ToMemoryModel(&m)
	}
	return result, total, err
}

func (s *MemoryService) Update(userID, id uint, data map[string]interface{}) (*MemoryModel, error) {
	existing, err := s.Get(userID, id)
	if err != nil {
		return nil, err
	}

	updates := map[string]interface{}{}
	if v, ok := data["value"]; ok {
		updates["value"] = v
		key := existing.Key
		if k, ok2 := data["key"]; ok2 {
			key = k.(string)
		}
		updates["summary"] = generateSummary(key, v.(string))
	}
	if v, ok := data["key"]; ok {
		updates["key"] = v
		value := existing.Value
		if val, ok2 := data["value"]; ok2 {
			value = val.(string)
		}
		updates["summary"] = generateSummary(v.(string), value)
	}
	if v, ok := data["importance"]; ok {
		updates["importance"] = v
	}
	if v, ok := data["memory_type"]; ok {
		updates["memory_type"] = v
	}
	if v, ok := data["tags"]; ok {
		if tags, ok := v.([]interface{}); ok {
			tagStrings := make([]string, len(tags))
			for i, t := range tags {
				tagStrings[i] = t.(string)
			}
			b, _ := json.Marshal(tagStrings)
			updates["tags"] = string(b)
		}
	}

	if err := s.db.Model(&models.Memory{}).Where("user_id = ? AND id = ?", userID, id).Updates(updates).Error; err != nil {
		return nil, err
	}
	return s.Get(userID, id)
}

func (s *MemoryService) Delete(userID, id uint) error {
	return s.db.Model(&models.Memory{}).Where("user_id = ? AND id = ?", userID, id).Update("status", "trashed").Error
}

func (s *MemoryService) Restore(userID, id uint) error {
	return s.db.Model(&models.Memory{}).Where("user_id = ? AND id = ?", userID, id).Update("status", "active").Error
}

func (s *MemoryService) SearchKeyword(userID uint, q string, limit int) ([]*MemoryModel, error) {
	var memories []models.Memory
	err := s.db.Where("user_id = ? AND (key LIKE ? OR value LIKE ?)", userID, "%"+q+"%", "%"+q+"%").
		Limit(limit).
		Find(&memories).Error

	result := make([]*MemoryModel, len(memories))
	for i, m := range memories {
		result[i] = ToMemoryModel(&m)
	}
	return result, err
}

func (s *MemoryService) IncrementAccess(id uint) error {
	return s.db.Model(&models.Memory{}).Where("id = ?", id).Updates(map[string]interface{}{
		"access_count":     gorm.Expr("access_count + 1"),
		"last_accessed_at": time.Now(),
	}).Error
}

func ToMemoryModel(m *models.Memory) *MemoryModel {
	var tags []string
	if m.Tags != "" {
		json.Unmarshal([]byte(m.Tags), &tags)
	}
	if tags == nil {
		tags = []string{}
	}
	return &MemoryModel{
		ID:             m.ID,
		UserID:         m.UserID,
		Layer:          m.Layer,
		Key:            m.Key,
		Value:          m.Value,
		Summary:        m.Summary,
		Importance:     m.Importance,
		AccessCount:    m.AccessCount,
		LastAccessedAt: m.LastAccessedAt,
		IsEncrypted:    m.IsEncrypted,
		Tags:           tags,
		Source:         m.Source,
		Platform:       m.Platform,
		Status:         m.Status,
		TrashedAt:      m.TrashedAt,
		DecayStage:     m.DecayStage,
		ReinforceCount: m.ReinforceCount,
		MemoryType:     m.MemoryType,
		VerifiedAt:     m.VerifiedAt,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}
