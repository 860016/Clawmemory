package services

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"clawmemory/internal/models"

	"gorm.io/gorm"
)

type MemoryService struct {
	db          *gorm.DB
	searchCache *SearchCache
}

// SetSearchCache injects the shared search cache for invalidation on writes.
func (s *MemoryService) SetSearchCache(c *SearchCache) {
	s.searchCache = c
}

func (s *MemoryService) invalidateSearchCache(userID uint) {
	if s.searchCache != nil {
		s.searchCache.Invalidate(userID)
	}
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
	SourceAgent    string     `json:"source_agent"`
	Visibility     string     `json:"visibility"`
	OriginChain    string     `json:"origin_chain"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func NewMemoryService(db *gorm.DB) *MemoryService {
	return &MemoryService{db: db}
}

func (s *MemoryService) Create(userID uint, data map[string]interface{}) (*MemoryModel, error) {
	if err := ValidateMemoryCreate(data); err != nil {
		return nil, fmt.Errorf("validation failed: %s", err.Error())
	}

	layer := getString(data, "layer", "context")

	if layer == "core" {
		var coreCount int64
		s.db.Model(&models.Memory{}).Where(
			"user_id = ? AND layer = ? AND status != ?", userID, "core", "trashed",
		).Count(&coreCount)
		if coreCount >= 50 {
			return nil, fmt.Errorf("core memory capacity reached (%d/50). Consolidate or archive existing core memories first", coreCount)
		}
	}

	tags := "[]"
	switch t := data["tags"].(type) {
	case []interface{}:
		tagStrings := make([]string, 0, len(t))
		for _, v := range t {
			if s, ok := v.(string); ok {
				tagStrings = append(tagStrings, s)
			}
		}
		b, _ := json.Marshal(tagStrings)
		tags = string(b)
	case string:
		if t != "" {
			tagStrings := strings.Split(t, ",")
			filtered := make([]string, 0, len(tagStrings))
			for _, s := range tagStrings {
				s = strings.TrimSpace(s)
				if s != "" {
					filtered = append(filtered, s)
				}
			}
			b, _ := json.Marshal(filtered)
			tags = string(b)
		}
	}

	memory := &models.Memory{
		UserID:      userID,
		Layer:       layer,
		Key:         getString(data, "key", ""),
		Value:       getString(data, "value", ""),
		Importance:  getFloat(data, "importance", 0.5),
		Tags:        tags,
		Summary:     generateSummary(getString(data, "key", ""), getString(data, "value", "")),
		Source:      getString(data, "source", "manual"),
		Platform:    getString(data, "platform", "clawmemory"),
		MemoryType:  getString(data, "memory_type", "knowledge"),
		IsEncrypted: getBool(data, "is_encrypted", false),
		SourceAgent: getString(data, "source_agent", ""),
		Visibility:  getString(data, "visibility", "private"),
		OriginChain: getString(data, "origin_chain", ""),
		Status:      "active",
		DecayStage:  0,
	}

	if err := s.db.Create(memory).Error; err != nil {
		return nil, err
	}

	go s.postCreate(memory)

	s.invalidateSearchCache(userID)

	return ToMemoryModel(memory), nil
}

func (s *MemoryService) postCreate(memory *models.Memory) {
	shareSvc := NewMemoryShareService(s.db)
	if err := shareSvc.ProcessAutoShareRules(memory.UserID, memory); err != nil {
		log.Printf("[MemoryService] postCreate: auto share rules error for memory %d: %v", memory.ID, err)
	}

	chromaSvc := NewChromaDBService(s.db)
	if chromaSvc.IsAvailable() {
		if err := chromaSvc.IndexSingleMemory(memory); err != nil {
			log.Printf("[MemoryService] postCreate: chromadb index error for memory %d: %v", memory.ID, err)
		}
	}
}

func (s *MemoryService) Get(userID, id uint) (*MemoryModel, error) {
	var memory models.Memory
	if err := s.db.Where("user_id = ? AND id = ? AND status != ?", userID, id, "trashed").First(&memory).Error; err != nil {
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

	var sourceAgent, visibility, source string
	if len(memoryType) > 1 && memoryType[1] != "" {
		sourceAgent = memoryType[1]
	}
	if len(memoryType) > 2 && memoryType[2] != "" {
		visibility = memoryType[2]
	}
	if len(memoryType) > 3 && memoryType[3] != "" {
		source = memoryType[3]
	}

	if sourceAgent != "" {
		query = query.Where("source_agent = ?", sourceAgent)
	}
	if source != "" {
		query = query.Where("source = ?", source)
	}
	if visibility != "" {
		query = query.Where("visibility = ?", visibility)
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
	keyStr := existing.Key
	valStr := existing.Value
	if v, ok := data["key"].(string); ok {
		keyStr = v
		updates["key"] = v
	}
	if v, ok := data["value"].(string); ok {
		valStr = v
		updates["value"] = v
	}
	if _, ok := data["key"]; ok || data["value"] != nil {
		updates["summary"] = generateSummary(keyStr, valStr)
	}
	if v, ok := data["importance"]; ok {
		updates["importance"] = v
	}
	if v, ok := data["memory_type"]; ok {
		updates["memory_type"] = v
	}
	if v, ok := data["tags"]; ok {
		switch t := v.(type) {
		case []interface{}:
			tagStrings := make([]string, 0, len(t))
			for _, item := range t {
				if s, ok := item.(string); ok {
					tagStrings = append(tagStrings, s)
				}
			}
			b, _ := json.Marshal(tagStrings)
			updates["tags"] = string(b)
		case string:
			if t != "" {
				tagStrings := strings.Split(t, ",")
				filtered := make([]string, 0, len(tagStrings))
				for _, s := range tagStrings {
					s = strings.TrimSpace(s)
					if s != "" {
						filtered = append(filtered, s)
					}
				}
				b, _ := json.Marshal(filtered)
				updates["tags"] = string(b)
			}
		}
	}
	if v, ok := data["layer"]; ok {
		updates["layer"] = v
	}
	if v, ok := data["visibility"]; ok {
		updates["visibility"] = v
	}
	if v, ok := data["source_agent"]; ok {
		updates["source_agent"] = v
	}
	if v, ok := data["source"]; ok {
		updates["source"] = v
	}

	if err := s.db.Model(&models.Memory{}).Where("user_id = ? AND id = ?", userID, id).Updates(updates).Error; err != nil {
		return nil, err
	}
	s.invalidateSearchCache(userID)
	return s.Get(userID, id)
}

func (s *MemoryService) Delete(userID, id uint) error {
	err := s.db.Model(&models.Memory{}).Where("user_id = ? AND id = ?", userID, id).Update("status", "trashed").Error
	if err == nil {
		s.invalidateSearchCache(userID)
	}
	return err
}

func (s *MemoryService) Restore(userID, id uint) error {
	err := s.db.Model(&models.Memory{}).Where("user_id = ? AND id = ?", userID, id).Update("status", "active").Error
	if err == nil {
		s.invalidateSearchCache(userID)
	}
	return err
}

func (s *MemoryService) SearchKeyword(userID uint, q string, limit int) ([]*MemoryModel, error) {
	escaped := EscapeLikeQuery(q)
	var memories []models.Memory
	err := s.db.Where("user_id = ? AND status != ? AND (key LIKE ? OR value LIKE ?)", userID, "trashed", "%"+escaped+"%", "%"+escaped+"%").
		Limit(limit).
		Find(&memories).Error

	result := make([]*MemoryModel, len(memories))
	for i, m := range memories {
		result[i] = ToMemoryModel(&m)
	}
	return result, err
}

func (s *MemoryService) SearchKeywordWithPlatform(userID uint, q string, platform string, platformFilter string, limit int) ([]*MemoryModel, error) {
	escaped := EscapeLikeQuery(q)
	var memories []models.Memory
	query := s.db.Where("user_id = ? AND status != ? AND (key LIKE ? OR value LIKE ?)", userID, "trashed", "%"+escaped+"%", "%"+escaped+"%")

	if platformFilter != "" {
		query = query.Where("platform = ?", platformFilter)
	} else if platform != "" && platform != "clawmemory" {
		query = query.Where("platform = ? OR platform = ? OR platform = ?", platform, "clawmemory", "")
	}

	err := query.Limit(limit).Find(&memories).Error

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
		SourceAgent:    m.SourceAgent,
		Visibility:     m.Visibility,
		OriginChain:    m.OriginChain,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}

func EscapeLikeQuery(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}
