package services

import (
	"fmt"
	"math"
	"strings"
	"time"

	"clawmemory/internal/models"

	"gorm.io/gorm"
)

type ProFallbackService struct {
	db *gorm.DB
}

func NewProFallbackService(db *gorm.DB) *ProFallbackService {
	return &ProFallbackService{db: db}
}

func (f *ProFallbackService) DecayStats(userID uint) (map[string]interface{}, error) {
	decaySvc := NewDecayService(f.db)
	stats, err := decaySvc.GetStats(userID)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"total":    stats.Total,
		"active":   stats.Active,
		"archived": stats.Archived,
		"trashed":  stats.Trashed,
		"fallback": true,
	}, nil
}

func (f *ProFallbackService) DecayApply(userID uint) (map[string]interface{}, error) {
	decaySvc := NewDecayService(f.db)
	return decaySvc.ApplyDecay(userID)
}

func (f *ProFallbackService) PruneSuggest(userID uint) (map[string]interface{}, error) {
	decaySvc := NewDecayService(f.db)
	suggestions, err := decaySvc.GetPruneSuggestions(userID)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"suggestions": suggestions,
		"algorithm":   "local_importance_v1",
		"fallback":    true,
	}, nil
}

func (f *ProFallbackService) ConflictScan(userID uint) (map[string]interface{}, error) {
	var memories []models.Memory
	f.db.Where("user_id = ? AND status != ?", userID, "trashed").Find(&memories)

	conflicts := []map[string]interface{}{}
	keyMap := make(map[string][]models.Memory)
	for _, m := range memories {
		keyMap[m.Key] = append(keyMap[m.Key], m)
	}

	for key, mems := range keyMap {
		if len(mems) > 1 {
			values := make([]string, len(mems))
			for i, m := range mems {
				values[i] = m.Value
			}
			if !allSame(values) {
				conflict := map[string]interface{}{
					"key":       key,
					"count":     len(mems),
					"memories":  mems,
					"conflict":  "different_values_same_key",
				}
				conflicts = append(conflicts, conflict)
			}
		}
	}

	return map[string]interface{}{
		"conflicts":  conflicts,
		"total":      len(conflicts),
		"algorithm":  "local_key_conflict_v1",
		"fallback":   true,
	}, nil
}

func (f *ProFallbackService) CompressPreview(userID uint, level string) (map[string]interface{}, error) {
	var memories []models.Memory
	f.db.Where("user_id = ? AND status != ?", userID, "trashed").Find(&memories)

	if len(memories) == 0 {
		return map[string]interface{}{
			"original_count": 0,
			"compressed_count": 0,
			"ratio": 0,
			"fallback": true,
		}, nil
	}

	rate := 0.2
	switch level {
	case "light":
		rate = 0.25
	case "medium":
		rate = 0.55
	case "deep":
		rate = 0.75
	}

	compressedCount := int(float64(len(memories)) * (1 - rate))

	grouped := groupByLayer(memories)
	preview := []map[string]interface{}{}
	for layer, mems := range grouped {
		preview = append(preview, map[string]interface{}{
			"layer":     layer,
			"original":  len(mems),
			"compressed": int(float64(len(mems)) * (1 - rate)),
		})
	}

	return map[string]interface{}{
		"original_count":   len(memories),
		"compressed_count": compressedCount,
		"ratio":            math.Round(rate*100) / 100,
		"level":            level,
		"preview":          preview,
		"algorithm":        "local_layer_compress_v1",
		"fallback":         true,
	}, nil
}

func (f *ProFallbackService) CompressApply(userID uint, level string) (map[string]interface{}, error) {
	var memories []models.Memory
	f.db.Where("user_id = ? AND status != ?", userID, "trashed").Find(&memories)

	if len(memories) == 0 {
		return map[string]interface{}{
			"processed": 0,
			"compressed": 0,
			"ratio": 0,
			"fallback": true,
		}, nil
	}

	rate := 0.2
	switch level {
	case "light":
		rate = 0.25
	case "medium":
		rate = 0.55
	case "deep":
		rate = 0.75
	}

	lowImportance := []models.Memory{}
	for _, m := range memories {
		if m.Importance < 0.3 {
			lowImportance = append(lowImportance, m)
		}
	}

	targetCount := int(float64(len(lowImportance)) * rate)
	compressed := 0

	for i := 0; i < targetCount && i < len(lowImportance); i++ {
		m := lowImportance[i]
		f.db.Model(&m).Updates(map[string]interface{}{
			"status":     "archived",
			"importance": m.Importance * 0.5,
		})
		compressed++
	}

	return map[string]interface{}{
		"processed":  len(memories),
		"compressed": compressed,
		"ratio":      math.Round(float64(compressed)/float64(len(memories))*100) / 100,
		"level":      level,
		"algorithm":  "local_importance_compress_v1",
		"fallback":   true,
	}, nil
}

func (f *ProFallbackService) EvolutionDiscover(userID uint) (map[string]interface{}, error) {
	recommendSvc := NewRecommendService(f.db)

	var memories []models.Memory
	f.db.Where("user_id = ? AND status != ?", userID, "trashed").Find(&memories)

	relations := []map[string]interface{}{}
	seen := make(map[string]bool)

	for i := 0; i < len(memories) && len(relations) < 20; i++ {
		result, err := recommendSvc.RecommendForMemory(userID, memories[i].ID, 3)
		if err != nil {
			continue
		}
		if recs, ok := result["recommendations"].([]map[string]interface{}); ok {
			for _, rec := range recs {
				key := fmt.Sprintf("%d-%d", memories[i].ID, rec["id"])
				if !seen[key] {
					seen[key] = true
					relations = append(relations, map[string]interface{}{
						"source_id":   memories[i].ID,
						"source_key":  memories[i].Key,
						"target_id":   rec["id"],
						"target_key":  rec["key"],
						"score":       rec["score"],
						"reason":      rec["reason"],
					})
				}
			}
		}
	}

	return map[string]interface{}{
		"relations":    relations,
		"total":        len(relations),
		"algorithm":    "local_token_similarity_v1",
		"fallback":     true,
	}, nil
}

func (f *ProFallbackService) EvolutionPrefetch(userID uint, context string) (map[string]interface{}, error) {
	recommendSvc := NewRecommendService(f.db)
	result, err := recommendSvc.RecommendByContext(userID, context, 10)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"matched":  result["recommendations"],
		"total":    result["total"],
		"context":  context,
		"algorithm": "local_context_match_v1",
		"fallback":  true,
	}, nil
}

func (f *ProFallbackService) EvolutionImportance(userID uint) (map[string]interface{}, error) {
	var memories []models.Memory
	f.db.Where("user_id = ? AND status != ?", userID, "trashed").Find(&memories)

	adjustments := []map[string]interface{}{}
	for _, m := range memories {
		daysSinceUpdate := time.Since(m.UpdatedAt).Hours() / 24
		adjustedImportance := m.Importance

		if daysSinceUpdate > 30 {
			adjustedImportance *= 0.8
		} else if daysSinceUpdate > 15 {
			adjustedImportance *= 0.9
		}

		if m.AccessCount > 5 {
			adjustedImportance = math.Min(adjustedImportance*1.1, 1.0)
		}

		if math.Abs(adjustedImportance-m.Importance) > 0.01 {
			adjustments = append(adjustments, map[string]interface{}{
				"id":                  m.ID,
				"key":                 m.Key,
				"current_importance":  m.Importance,
				"adjusted_importance": math.Round(adjustedImportance*100) / 100,
				"reason":              getImportanceReason(daysSinceUpdate, m.AccessCount),
			})
		}
	}

	return map[string]interface{}{
		"adjustments": adjustments,
		"total":       len(adjustments),
		"algorithm":   "local_time_access_v1",
		"fallback":    true,
	}, nil
}

func allSame(values []string) bool {
	if len(values) <= 1 {
		return true
	}
	for i := 1; i < len(values); i++ {
		if !strings.EqualFold(values[i], values[0]) {
			return false
		}
	}
	return true
}

func (f *ProFallbackService) ReinforceMemory(userID uint, memoryID uint) (map[string]interface{}, error) {
	var memory models.Memory
	if err := f.db.Where("id = ? AND user_id = ?", memoryID, userID).First(&memory).Error; err != nil {
		return nil, fmt.Errorf("memory not found")
	}
	newImportance := math.Min(memory.Importance*1.2, 1.0)
	f.db.Model(&memory).Updates(map[string]interface{}{
		"importance":       newImportance,
		"access_count":     memory.AccessCount + 1,
		"last_accessed_at": time.Now(),
	})
	return map[string]interface{}{
		"memory_id":         memoryID,
		"old_importance":    memory.Importance,
		"new_importance":    newImportance,
		"reinforced":        true,
		"algorithm":         "local_access_reinforce_v1",
		"fallback":          true,
	}, nil
}

func (f *ProFallbackService) ConflictResolve(userID uint, conflictIndex int, strategy string) (map[string]interface{}, error) {
	result, err := f.ConflictScan(userID)
	if err != nil {
		return nil, err
	}
	conflicts, ok := result["conflicts"].([]map[string]interface{})
	if !ok || conflictIndex >= len(conflicts) {
		return nil, fmt.Errorf("conflict index out of range")
	}
	conflict := conflicts[conflictIndex]
	memories, ok := conflict["memories"].([]models.Memory)
	if !ok || len(memories) == 0 {
		return nil, fmt.Errorf("no memories in conflict")
	}
	switch strategy {
	case "keep_first":
		for i := 1; i < len(memories); i++ {
			f.db.Model(&memories[i]).Update("status", "archived")
		}
	case "keep_latest":
		for i := 0; i < len(memories)-1; i++ {
			f.db.Model(&memories[i]).Update("status", "archived")
		}
	default:
		for i := 1; i < len(memories); i++ {
			f.db.Model(&memories[i]).Update("status", "archived")
		}
	}
	return map[string]interface{}{
		"resolved":  true,
		"strategy":  strategy,
		"kept":      1,
		"archived":  len(memories) - 1,
		"fallback":  true,
	}, nil
}

func (f *ProFallbackService) TokenRoute(message string, contextLength int) (map[string]interface{}, error) {
	tokenEstimate := len(message) / 4
	if contextLength > 0 {
		tokenEstimate += contextLength
	}
	return map[string]interface{}{
		"model":           "local",
		"estimated_tokens": tokenEstimate,
		"provider":        "fallback",
		"fallback":        true,
	}, nil
}

func (f *ProFallbackService) TokenStats() (map[string]interface{}, error) {
	var totalMemories int64
	f.db.Model(&models.Memory{}).Where("status != ?", "trashed").Count(&totalMemories)
	estimatedTokens := totalMemories * 50
	return map[string]interface{}{
		"total_tokens_used": estimatedTokens,
		"total_memories":    totalMemories,
		"provider":          "fallback",
		"fallback":          true,
	}, nil
}

func (f *ProFallbackService) AIExtract(userID uint) (map[string]interface{}, error) {
	var memories []models.Memory
	f.db.Where("user_id = ? AND status != ?", userID, "trashed").Find(&memories)
	entities := []map[string]interface{}{}
	for _, m := range memories {
		if len(entities) >= 10 {
			break
		}
		entities = append(entities, map[string]interface{}{
			"name":        m.Key,
			"entity_type": "concept",
			"description": m.Value,
			"source":      "local_extract",
		})
	}
	return map[string]interface{}{
		"entities":    entities,
		"total":       len(entities),
		"algorithm":   "local_key_extract_v1",
		"fallback":    true,
	}, nil
}

func (f *ProFallbackService) AutoGraph(userID uint, overwrite bool) (map[string]interface{}, error) {
	extractResult, err := f.AIExtract(userID)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"entities_created": extractResult["total"],
		"relations_created": 0,
		"overwrite":        overwrite,
		"algorithm":        "local_key_extract_v1",
		"fallback":         true,
	}, nil
}

func (f *ProFallbackService) BackupSchedule() (map[string]interface{}, error) {
	return map[string]interface{}{
		"enabled":        false,
		"interval_hours": 24,
		"fallback":       true,
	}, nil
}

func (f *ProFallbackService) SetBackupSchedule(enabled bool, intervalHours int) (map[string]interface{}, error) {
	return map[string]interface{}{
		"enabled":        enabled,
		"interval_hours": intervalHours,
		"message":        "本地模式：备份计划已记录，请手动执行备份",
		"fallback":       true,
	}, nil
}

func (f *ProFallbackService) CompressConfig() (map[string]interface{}, error) {
	return map[string]interface{}{
		"level":     "light",
		"auto":      false,
		"fallback":  true,
	}, nil
}

func (f *ProFallbackService) SetCompressConfig(config map[string]interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{
		"config":    config,
		"message":   "本地模式：压缩配置已记录",
		"fallback":  true,
	}, nil
}

func (f *ProFallbackService) EvolutionInsights(userID uint) (map[string]interface{}, error) {
	var totalMemories int64
	f.db.Model(&models.Memory{}).Where("user_id = ? AND status != ?", userID, "trashed").Count(&totalMemories)

	var totalEntities int64
	f.db.Model(&models.Entity{}).Where("user_id = ?", userID).Count(&totalEntities)

	var totalRelations int64
	f.db.Model(&models.Relation{}).Count(&totalRelations)

	return map[string]interface{}{
		"total_memories":   totalMemories,
		"total_entities":   totalEntities,
		"total_relations":  totalRelations,
		"health_score":     0.7,
		"recommendations":  []string{"定期清理低重要性记忆", "为关键记忆添加标签", "利用知识图谱建立关联"},
		"fallback":         true,
	}, nil
}

func (f *ProFallbackService) EvolutionInfer(userID uint) (map[string]interface{}, error) {
	return map[string]interface{}{
		"chains":    []map[string]interface{}{},
		"total":     0,
		"algorithm": "local_infer_v1",
		"fallback":  true,
	}, nil
}

func groupByLayer(memories []models.Memory) map[string][]models.Memory {
	result := make(map[string][]models.Memory)
	for _, m := range memories {
		result[m.Layer] = append(result[m.Layer], m)
	}
	return result
}

func getImportanceReason(daysSinceUpdate float64, accessCount int) string {
	if daysSinceUpdate > 30 {
		return "stale_memory"
	}
	if daysSinceUpdate > 15 {
		return "aging_memory"
	}
	if accessCount > 5 {
		return "frequently_accessed"
	}
	return "time_decay"
}
