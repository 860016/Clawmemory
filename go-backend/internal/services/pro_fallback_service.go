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
