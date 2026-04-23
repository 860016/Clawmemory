package services

import (
	"encoding/json"

	"clawmemory/internal/models"

	"gorm.io/gorm"
)

type KnowledgeService struct {
	db *gorm.DB
}

func NewKnowledgeService(db *gorm.DB) *KnowledgeService {
	return &KnowledgeService{db: db}
}

func (s *KnowledgeService) CreateEntity(data map[string]interface{}) (*models.Entity, error) {
	properties := "{}"
	if p, ok := data["properties"].(map[string]interface{}); ok {
		b, _ := json.Marshal(p)
		properties = string(b)
	}

	aliases := "[]"
	if a, ok := data["aliases"].([]string); ok {
		b, _ := json.Marshal(a)
		aliases = string(b)
	}

	entity := &models.Entity{
		UserID:        1,
		Name:          getString(data, "name", ""),
		EntityType:    getString(data, "entity_type", ""),
		Description:   getString(data, "description", ""),
		Properties:    properties,
		Confidence:    getFloat(data, "confidence", 1.0),
		ExtractMethod: getString(data, "extract_method", "manual"),
		CanonicalName: getString(data, "canonical_name", ""),
		Aliases:       aliases,
	}

	if err := s.db.Create(entity).Error; err != nil {
		return nil, err
	}
	return entity, nil
}

func (s *KnowledgeService) ListEntities(entityType string, page, size int) ([]models.Entity, int64, error) {
	var entities []models.Entity
	var total int64

	query := s.db.Model(&models.Entity{}).Where("user_id = ?", 1)
	if entityType != "" {
		query = query.Where("entity_type = ?", entityType)
	}

	query.Count(&total)
	err := query.Order("updated_at DESC").Offset((page - 1) * size).Limit(size).Find(&entities).Error
	return entities, total, err
}

func (s *KnowledgeService) CreateRelation(data map[string]interface{}) (*models.Relation, error) {
	relation := &models.Relation{
		UserID:         1,
		SourceID:       uint(getFloat(data, "source_id", 0)),
		TargetID:       uint(getFloat(data, "target_id", 0)),
		RelationType:   getString(data, "relation_type", ""),
		Description:    getString(data, "description", ""),
		Confidence:     getFloat(data, "confidence", 1.0),
		DiscoverMethod: getString(data, "discover_method", "manual"),
		Weight:         getFloat(data, "weight", 1.0),
	}

	if err := s.db.Create(relation).Error; err != nil {
		return nil, err
	}
	return relation, nil
}

func (s *KnowledgeService) GetGraph() ([]models.Entity, []models.Relation, error) {
	var entities []models.Entity
	var relations []models.Relation

	if err := s.db.Where("user_id = ?", 1).Find(&entities).Error; err != nil {
		return nil, nil, err
	}
	if err := s.db.Where("user_id = ?", 1).Find(&relations).Error; err != nil {
		return nil, nil, err
	}

	return entities, relations, nil
}
