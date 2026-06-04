package services

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"

	"clawmemory/internal/models"

	"gorm.io/gorm"
)

type ToolboxService struct {
	db *gorm.DB
	mu sync.RWMutex
}

func NewToolboxService(db *gorm.DB) *ToolboxService {
	return &ToolboxService{
		db: db,
	}
}

func (s *ToolboxService) ConflictScan(userID uint) (map[string]interface{}, error) {
	var memories []models.Memory
	logDBErr("load memories for conflict scan", s.db.Where("user_id = ? AND status = ?", userID, "active").Limit(5000).Find(&memories).Error)

	keyMap := make(map[string][]models.Memory)
	for _, m := range memories {
		keyMap[m.Key] = append(keyMap[m.Key], m)
	}

	conflicts := []map[string]interface{}{}

	for key, mems := range keyMap {
		if len(mems) > 1 {
			items := make([]map[string]interface{}, 0, len(mems))
			for _, m := range mems {
				items = append(items, map[string]interface{}{
					"id":         m.ID,
					"value":      truncateStr(m.Value, 100),
					"importance": m.Importance,
					"updated_at": m.UpdatedAt,
				})
			}
			conflicts = append(conflicts, map[string]interface{}{
				"key":      key,
				"count":    len(mems),
				"severity": "exact_duplicate",
				"memories": items,
			})
		}
	}

	tagMemories := make(map[string][]models.Memory)
	for _, m := range memories {
		if m.Tags == "" {
			continue
		}
		for _, tag := range parseMemoryTags(m.Tags) {
			if tag != "" {
				tagMemories[tag] = append(tagMemories[tag], m)
			}
		}
	}

	for tag, mems := range tagMemories {
		if len(mems) < 2 || len(mems) > 50 {
			continue
		}
		for i := 0; i < len(mems); i++ {
			for j := i + 1; j < len(mems); j++ {
				if mems[i].Key == mems[j].Key {
					continue
				}
				sim := jaccardSimilarity(mems[i].Value, mems[j].Value)
				if sim >= 0.5 {
					conflicts = append(conflicts, map[string]interface{}{
						"key":        fmt.Sprintf("%s vs %s", mems[i].Key, mems[j].Key),
						"count":      2,
						"severity":   "similar_content",
						"similarity": sim,
						"tag":        tag,
						"memories": []map[string]interface{}{
							{"id": mems[i].ID, "value": truncateStr(mems[i].Value, 100), "importance": mems[i].Importance},
							{"id": mems[j].ID, "value": truncateStr(mems[j].Value, 100), "importance": mems[j].Importance},
						},
					})
				}
			}
		}
	}

	sort.Slice(conflicts, func(i, j int) bool {
		si, _ := conflicts[i]["severity"].(string)
		sj, _ := conflicts[j]["severity"].(string)
		prio := map[string]int{"exact_duplicate": 3, "similar_content": 2, "potential_conflict": 1}
		return prio[si] > prio[sj]
	})

	return map[string]interface{}{
		"mode":      "local",
		"conflicts": conflicts,
		"total":     len(conflicts),
	}, nil
}

func (s *ToolboxService) ConflictResolve(userID uint, conflictIndex int, strategy string) (map[string]interface{}, error) {
	var memories []models.Memory
	logDBErr("load memories for conflict resolve", s.db.Where("user_id = ? AND status = ?", userID, "active").Limit(5000).Find(&memories).Error)

	keyMap := make(map[string][]models.Memory)
	for _, m := range memories {
		keyMap[m.Key] = append(keyMap[m.Key], m)
	}

	keys := make([]string, 0, len(keyMap))
	for k := range keyMap {
		if len(keyMap[k]) > 1 {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	if conflictIndex < 0 || conflictIndex >= len(keys) {
		return nil, fmt.Errorf("conflict index %d out of range (total: %d)", conflictIndex, len(keys))
	}

	conflictKey := keys[conflictIndex]
	conflictMems := keyMap[conflictKey]

	switch strategy {
	case "keep_newest":
		for i := 0; i < len(conflictMems)-1; i++ {
			s.db.Model(&models.Memory{}).Where("id = ?", conflictMems[i].ID).Update("status", "archived")
		}
		return map[string]interface{}{
			"strategy": strategy, "key": conflictKey,
			"kept": conflictMems[len(conflictMems)-1].ID, "archived": len(conflictMems) - 1,
		}, nil
	case "keep_important":
		sort.Slice(conflictMems, func(i, j int) bool {
			return conflictMems[i].Importance > conflictMems[j].Importance
		})
		for i := 1; i < len(conflictMems); i++ {
			s.db.Model(&models.Memory{}).Where("id = ?", conflictMems[i].ID).Update("status", "archived")
		}
		return map[string]interface{}{
			"strategy": strategy, "key": conflictKey,
			"kept": conflictMems[0].ID, "archived": len(conflictMems) - 1,
		}, nil
	case "merge":
		var merged strings.Builder
		for i, m := range conflictMems {
			if i > 0 {
				merged.WriteString("\n---\n")
			}
			merged.WriteString(m.Value)
		}
		s.db.Model(&models.Memory{}).Where("id = ?", conflictMems[0].ID).Update("value", merged.String())
		for i := 1; i < len(conflictMems); i++ {
			s.db.Model(&models.Memory{}).Where("id = ?", conflictMems[i].ID).Update("status", "archived")
		}
		return map[string]interface{}{
			"strategy": strategy, "key": conflictKey,
			"kept": conflictMems[0].ID, "archived": len(conflictMems) - 1,
		}, nil
	default:
		return nil, fmt.Errorf("unknown strategy: %s", strategy)
	}
}

func (s *ToolboxService) TokenRoute(message string, contextLength int) (map[string]interface{}, error) {
	length := len([]rune(message))
	sentenceCount := 0
	for _, ch := range message {
		if ch == '。' || ch == '.' || ch == '？' || ch == '?' || ch == '！' || ch == '!' {
			sentenceCount++
		}
	}
	if sentenceCount == 0 {
		sentenceCount = 1
	}
	avgSentenceLen := float64(length) / float64(sentenceCount)

	technicalTerms := []string{"算法", "架构", "API", "数据库", "函数", "模型", "优化", "部署", "系统", "框架", "协议", "接口"}
	techCount := 0
	for _, term := range technicalTerms {
		if strings.Contains(message, term) {
			techCount++
		}
	}

	score := 1
	if length > 100 {
		score++
	}
	if avgSentenceLen > 30 {
		score++
	}
	if techCount >= 2 {
		score++
	}

	complexityLabel := "simple"
	layer := "context"
	strategy := "direct"
	switch score {
	case 1:
		complexityLabel = "simple"
		layer = "core"
		strategy = "direct"
	case 2:
		complexityLabel = "medium"
		layer = "core"
		strategy = "keyword_priority"
	case 3:
		complexityLabel = "complex"
		layer = "context"
		strategy = "semantic_priority"
	case 4:
		complexityLabel = "extreme"
		layer = "detail"
		strategy = "full_context"
	}

	tokenEstimate := length / 4
	if tokenEstimate == 0 {
		tokenEstimate = 1
	}

	return map[string]interface{}{
		"mode":              "local",
		"token_estimate":    tokenEstimate,
		"recommended_layer": layer,
		"strategy":          strategy,
		"complexity":        complexityLabel,
		"complexity_score":  score,
		"technical_terms":   techCount,
		"avg_sentence_len":  math.Round(avgSentenceLen*10) / 10,
	}, nil
}

func (s *ToolboxService) TokenStats(userID uint) (map[string]interface{}, error) {
	var totalMemories int64
	logDBErr("count total memories for token stats", s.db.Model(&models.Memory{}).Where("user_id = ? AND status != ?", userID, "trashed").Count(&totalMemories).Error)

	var totalEntities int64
	logDBErr("count entities for token stats", s.db.Model(&models.Entity{}).Where("user_id = ?", userID).Count(&totalEntities).Error)

	var totalRelations int64
	logDBErr("count relations for token stats", s.db.Model(&models.Relation{}).Where("user_id = ?", userID).Count(&totalRelations).Error)

	estimatedTokens := totalMemories*200 + totalEntities*50 + totalRelations*30

	return map[string]interface{}{
		"mode":            "local",
		"memory_tokens":   totalMemories * 200,
		"entity_tokens":   totalEntities * 50,
		"relation_tokens": totalRelations * 30,
		"total_tokens":    estimatedTokens,
		"memory_count":    totalMemories,
		"entity_count":    totalEntities,
		"relation_count":  totalRelations,
	}, nil
}

func (s *ToolboxService) ExtractEntities(userID uint) (map[string]interface{}, error) {
	var memories []models.Memory
	logDBErr("load memories for extract entities", s.db.Where("user_id = ? AND status != ?", userID, "trashed").Select("value").Limit(5000).Find(&memories).Error)

	extracted := 0
	for _, m := range memories {
		if m.Value == "" {
			continue
		}
		var existingCount int64
		s.db.Model(&models.Entity{}).Where("user_id = ? AND name = ?", userID, truncateStr(m.Key, 200)).Count(&existingCount)
		if existingCount > 0 {
			continue
		}
		entity := models.Entity{
			Name:       truncateStr(m.Key, 200),
			EntityType: "concept",
			Confidence: m.Importance,
			UserID:     userID,
		}
		if err := s.db.Create(&entity).Error; err == nil {
			extracted++
		}
	}

	return map[string]interface{}{
		"mode":      "local",
		"extracted": extracted,
		"scanned":   len(memories),
	}, nil
}
