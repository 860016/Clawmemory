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

func (s *KnowledgeService) CreateEntity(userID uint, data map[string]interface{}) (*models.Entity, error) {
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
		UserID:        userID,
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

func (s *KnowledgeService) ListEntities(userID uint, entityType string, page, size int) ([]models.Entity, int64, error) {
	var entities []models.Entity
	var total int64

	query := s.db.Model(&models.Entity{}).Where("user_id = ?", userID)
	if entityType != "" {
		query = query.Where("entity_type = ?", entityType)
	}

	query.Count(&total)
	err := query.Order("updated_at DESC").Offset((page - 1) * size).Limit(size).Find(&entities).Error
	return entities, total, err
}

func (s *KnowledgeService) CreateRelation(userID uint, data map[string]interface{}) (*models.Relation, error) {
	relation := &models.Relation{
		UserID:         userID,
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

func (s *KnowledgeService) GetGraph(userID uint) ([]models.Entity, []models.Relation, error) {
	var entities []models.Entity
	var relations []models.Relation

	if err := s.db.Where("user_id = ?", userID).Find(&entities).Error; err != nil {
		return nil, nil, err
	}
	if err := s.db.Where("user_id = ?", userID).Find(&relations).Error; err != nil {
		return nil, nil, err
	}

	return entities, relations, nil
}

func (s *KnowledgeService) GetEntity(userID uint, id uint) (*models.Entity, error) {
	var entity models.Entity
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&entity).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

func (s *KnowledgeService) UpdateEntity(userID uint, id uint, data map[string]interface{}) (*models.Entity, error) {
	var entity models.Entity
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&entity).Error; err != nil {
		return nil, err
	}

	updates := map[string]interface{}{}
	if name, ok := data["name"].(string); ok && name != "" {
		updates["name"] = name
	}
	if entityType, ok := data["entity_type"].(string); ok && entityType != "" {
		updates["entity_type"] = entityType
	}
	if desc, ok := data["description"].(string); ok {
		updates["description"] = desc
	}
	if props, ok := data["properties"].(map[string]interface{}); ok {
		b, _ := json.Marshal(props)
		updates["properties"] = string(b)
	}
	if conf, ok := data["confidence"].(float64); ok {
		updates["confidence"] = conf
	}

	if len(updates) > 0 {
		if err := s.db.Model(&entity).Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	s.db.Where("id = ?", id).First(&entity)
	return &entity, nil
}

func (s *KnowledgeService) DeleteEntity(userID uint, id uint) error {
	s.db.Where("source_id = ? OR target_id = ?", id, id).Delete(&models.Relation{})
	return s.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Entity{}).Error
}

func (s *KnowledgeService) DeleteRelation(userID uint, id uint) error {
	return s.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Relation{}).Error
}
