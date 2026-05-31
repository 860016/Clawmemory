package services

import (
	"fmt"
	"math"
	"strings"

	"clawmemory/internal/models"

	"gorm.io/gorm"
)

type EvolutionService struct {
	db *gorm.DB
}

func NewEvolutionService(db *gorm.DB) *EvolutionService {
	return &EvolutionService{db: db}
}

func (s *EvolutionService) Insights(userID uint) (map[string]interface{}, error) {
	var memories []models.Memory
	logDBErr("load memories for evolution insights", s.db.Where("user_id = ? AND status = ?", userID, "active").Limit(5000).Find(&memories).Error)

	layerCount := make(map[string]int)
	sourceCount := make(map[string]int)
	for _, m := range memories {
		layerCount[m.Layer]++
		if m.Source != "" {
			sourceCount[m.Source]++
		}
	}

	var relationsCount int64
	logDBErr("count relations for evolution insights", s.db.Model(&models.Relation{}).Where("user_id = ?", userID).Count(&relationsCount).Error)

	var entitiesCount int64
	logDBErr("count entities for evolution insights", s.db.Model(&models.Entity{}).Where("user_id = ?", userID).Count(&entitiesCount).Error)

	return map[string]interface{}{
		"mode":                 "local",
		"total":                len(memories),
		"layer_stats":          layerCount,
		"source_stats":         sourceCount,
		"relations_count":      relationsCount,
		"discovered_relations": relationsCount,
		"inferred_chains":      entitiesCount,
	}, nil
}

func (s *EvolutionService) Discover(userID uint) (map[string]interface{}, error) {
	var relations []models.Relation
	logDBErr("load relations for evolution discover", s.db.Where("user_id = ?", userID).Limit(5000).Find(&relations).Error)

	typeCount := make(map[string]int)
	for _, r := range relations {
		typeCount[r.RelationType]++
	}

	discoveries := []map[string]interface{}{}
	for rtype, count := range typeCount {
		if count >= 2 {
			discoveries = append(discoveries, map[string]interface{}{
				"relation_type": rtype,
				"count":         count,
				"confidence":    math.Min(float64(count)/10.0, 1.0),
			})
		}
	}

	return map[string]interface{}{
		"mode":        "local",
		"discoveries": discoveries,
		"total":       len(discoveries),
	}, nil
}

func (s *EvolutionService) HighConfidenceEntities(userID uint) (map[string]interface{}, error) {
	var entities []models.Entity
	logDBErr("load entities for high confidence", s.db.Where("user_id = ?", userID).Limit(5000).Find(&entities).Error)

	inferences := []map[string]interface{}{}
	for _, e := range entities {
		if e.Confidence >= 0.7 {
			inferences = append(inferences, map[string]interface{}{
				"entity_id":   e.ID,
				"entity_name": e.Name,
				"confidence":  e.Confidence,
				"reason":      "high_importance_entity",
			})
		}
	}

	return map[string]interface{}{
		"mode":       "local",
		"inferences": inferences,
		"total":      len(inferences),
	}, nil
}

func (s *EvolutionService) ImportanceBuckets(userID uint) (map[string]interface{}, error) {
	var memories []models.Memory
	logDBErr("load memories for evolution importance", s.db.Where("user_id = ? AND status = ?", userID, "active").Limit(5000).Find(&memories).Error)

	importanceBuckets := map[string]int{
		"critical": 0,
		"high":     0,
		"medium":   0,
		"low":      0,
	}
	for _, m := range memories {
		switch {
		case m.Importance >= 0.8:
			importanceBuckets["critical"]++
		case m.Importance >= 0.5:
			importanceBuckets["high"]++
		case m.Importance >= 0.3:
			importanceBuckets["medium"]++
		default:
			importanceBuckets["low"]++
		}
	}

	return map[string]interface{}{
		"mode":    "local",
		"buckets": importanceBuckets,
		"total":   len(memories),
	}, nil
}

func (s *EvolutionService) AutoGraph(userID uint, overwrite bool) (map[string]interface{}, error) {
	var memories []models.Memory
	logDBErr("load memories for auto graph", s.db.Where("user_id = ? AND status = ?", userID, "active").Limit(5000).Find(&memories).Error)

	var relations []models.Relation
	logDBErr("load relations for auto graph", s.db.Where("user_id = ?", userID).Limit(5000).Find(&relations).Error)

	existingPairs := make(map[string]bool)
	for _, r := range relations {
		key := fmt.Sprintf("%d-%s-%d", r.SourceID, r.RelationType, r.TargetID)
		existingPairs[key] = true
	}

	created := 0
	skipped := 0
	createdRelations := []map[string]interface{}{}

	memoryMap := make(map[uint]models.Memory)
	for _, m := range memories {
		memoryMap[m.ID] = m
	}

	keyMemories := make(map[string][]models.Memory)
	for _, m := range memories {
		parts := strings.SplitN(m.Key, ":", 2)
		if len(parts) == 2 {
			keyMemories[parts[0]] = append(keyMemories[parts[0]], m)
		}
	}

	for _, mems := range keyMemories {
		if len(mems) < 2 {
			continue
		}
		for i := 0; i < len(mems); i++ {
			for j := i + 1; j < len(mems); j++ {
				pairKey := fmt.Sprintf("%d-same_topic-%d", mems[i].ID, mems[j].ID)
				if existingPairs[pairKey] && !overwrite {
					skipped++
					continue
				}
				rel := models.Relation{
					SourceID:     mems[i].ID,
					TargetID:     mems[j].ID,
					RelationType: "same_topic",
					UserID:       userID,
					Weight:       0.6,
				}
				if err := s.db.Create(&rel).Error; err == nil {
					created++
					createdRelations = append(createdRelations, map[string]interface{}{
						"source":        mems[i].Key,
						"target":        mems[j].Key,
						"relation_type": "same_topic",
						"weight":        0.6,
					})
				}
			}
		}
	}

	tagMemories := make(map[string][]models.Memory)
	for _, m := range memories {
		if m.Tags == "" {
			continue
		}
		for _, tag := range strings.Split(m.Tags, ",") {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				tagMemories[tag] = append(tagMemories[tag], m)
			}
		}
	}

	for tag, mems := range tagMemories {
		if len(mems) < 2 || len(mems) > 100 {
			continue
		}
		for i := 0; i < len(mems); i++ {
			for j := i + 1; j < len(mems); j++ {
				if mems[i].Key == mems[j].Key {
					continue
				}
				pairKey := fmt.Sprintf("%d-shared_tags-%d", mems[i].ID, mems[j].ID)
				if existingPairs[pairKey] && !overwrite {
					skipped++
					continue
				}
				rel := models.Relation{
					SourceID:     mems[i].ID,
					TargetID:     mems[j].ID,
					RelationType: "shared_tags",
					UserID:       userID,
					Weight:       0.4,
				}
				if err := s.db.Create(&rel).Error; err == nil {
					created++
					createdRelations = append(createdRelations, map[string]interface{}{
						"source":        mems[i].Key,
						"target":        mems[j].Key,
						"relation_type": "shared_tags",
						"weight":        0.4,
						"tag":           tag,
					})
				}
			}
		}
	}

	var entities []models.Entity
	logDBErr("load entities for auto graph", s.db.Where("user_id = ?", userID).Limit(1000).Find(&entities).Error)

	entityMemories := make(map[uint][]uint)
	for _, e := range entities {
		if e.SourceMemoryID != nil {
			entityMemories[e.ID] = append(entityMemories[e.ID], *e.SourceMemoryID)
		}
	}

	for eid, memIDs := range entityMemories {
		if len(memIDs) < 2 {
			continue
		}
		for i := 0; i < len(memIDs); i++ {
			for j := i + 1; j < len(memIDs); j++ {
				pairKey := fmt.Sprintf("%d-shared_entity-%d", memIDs[i], memIDs[j])
				if existingPairs[pairKey] && !overwrite {
					skipped++
					continue
				}
				rel := models.Relation{
					SourceID:     memIDs[i],
					TargetID:     memIDs[j],
					RelationType: "shared_entity",
					UserID:       userID,
					Weight:       0.5,
				}
				if err := s.db.Create(&rel).Error; err == nil {
					created++
					srcKey := fmt.Sprintf("memory_%d", memIDs[i])
					tgtKey := fmt.Sprintf("memory_%d", memIDs[j])
					if m, ok := memoryMap[memIDs[i]]; ok {
						srcKey = m.Key
					}
					if m, ok := memoryMap[memIDs[j]]; ok {
						tgtKey = m.Key
					}
					createdRelations = append(createdRelations, map[string]interface{}{
						"source":        srcKey,
						"target":        tgtKey,
						"relation_type": "shared_entity",
						"weight":        0.5,
						"entity_id":     eid,
					})
				}
			}
		}
	}

	if len(createdRelations) > 30 {
		createdRelations = createdRelations[:30]
	}

	return map[string]interface{}{
		"mode":        "local",
		"created":     created,
		"skipped":     skipped,
		"total_pairs": len(existingPairs),
		"relations":   createdRelations,
	}, nil
}
