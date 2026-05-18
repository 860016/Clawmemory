package services

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"clawmemory/internal/models"

	"gorm.io/gorm"
)

type ToolboxService struct {
	db *gorm.DB

	compressCfg map[string]interface{}
	backupCfg   struct {
		Enabled       bool
		IntervalHours int
	}
	mu sync.RWMutex
}

func NewToolboxService(db *gorm.DB) *ToolboxService {
	return &ToolboxService{
		db: db,
		compressCfg: map[string]interface{}{
			"auto_compress":      false,
			"threshold":          5000,
			"level":              "light",
			"preserve_important": true,
		},
		backupCfg: struct {
			Enabled       bool
			IntervalHours int
		}{
			Enabled:       false,
			IntervalHours: 24,
		},
	}
}

func (s *ToolboxService) DecayStats(userID uint) (map[string]interface{}, error) {
	var stats struct {
		Total    int64
		Active   int64
		Archived int64
		Trashed  int64
	}
	_ = s.db.Model(&models.Memory{}).Where("user_id = ?", userID).Count(&stats.Total).Error
	_ = s.db.Model(&models.Memory{}).Where("user_id = ? AND status = ?", userID, "active").Count(&stats.Active).Error
	_ = s.db.Model(&models.Memory{}).Where("user_id = ? AND status = ?", userID, "archived").Count(&stats.Archived).Error
	_ = s.db.Model(&models.Memory{}).Where("user_id = ? AND status = ?", userID, "trashed").Count(&stats.Trashed).Error

	var avgImportance float64
	s.db.Model(&models.Memory{}).Where("user_id = ? AND status = ?", userID, "active").
		Select("COALESCE(AVG(importance), 0)").Row().Scan(&avgImportance)

	return map[string]interface{}{
		"total":          stats.Total,
		"active":         stats.Active,
		"archived":       stats.Archived,
		"trashed":        stats.Trashed,
		"avg_importance": math.Round(avgImportance*1000) / 1000,
	}, nil
}

func (s *ToolboxService) DecayApply(userID uint) (map[string]interface{}, error) {
	var memories []models.Memory
	_ = s.db.Where("user_id = ? AND status = ?", userID, "active").Limit(5000).Find(&memories).Error

	archived := 0
	trashed := 0
	kept := 0

	for _, m := range memories {
		switch {
		case m.Importance < 0.1 && time.Since(m.UpdatedAt) > 30*24*time.Hour:
			_ = s.db.Model(&models.Memory{}).Where("id = ?", m.ID).Update("status", "trashed").Error
			trashed++
		case m.Importance < 0.3 && time.Since(m.UpdatedAt) > 14*24*time.Hour:
			_ = s.db.Model(&models.Memory{}).Where("id = ?", m.ID).Update("status", "archived").Error
			archived++
		default:
			kept++
		}
	}

	return map[string]interface{}{
		"mode":      "local",
		"processed": len(memories),
		"archived":  archived,
		"trashed":   trashed,
		"adjusted":  kept,
		"algorithm": "local_decay_v1",
	}, nil
}

func (s *ToolboxService) PruneSuggest(userID uint) (map[string]interface{}, error) {
	suggestions, err := NewDecayService(s.db).GetPruneSuggestions(userID)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"mode":        "local",
		"suggestions": suggestions,
		"total":       len(suggestions),
	}, nil
}

func (s *ToolboxService) ConflictScan(userID uint) (map[string]interface{}, error) {
	var memories []models.Memory
	_ = s.db.Where("user_id = ? AND status = ?", userID, "active").Limit(5000).Find(&memories).Error

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
		for _, tag := range strings.Split(m.Tags, ",") {
			tag = strings.TrimSpace(tag)
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
	_ = s.db.Where("user_id = ? AND status = ?", userID, "active").Limit(5000).Find(&memories).Error

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

func (s *ToolboxService) CompressPreview(userID uint, level string) (map[string]interface{}, error) {
	var memories []models.Memory
	_ = s.db.Where("user_id = ? AND status != ?", userID, "trashed").Limit(5000).Find(&memories).Error

	threshold := 0.3
	switch level {
	case "light":
		threshold = 0.2
	case "medium":
		threshold = 0.35
	case "heavy", "deep":
		threshold = 0.5
	default:
		threshold = 0.2
	}

	preview := []map[string]interface{}{}
	for _, m := range memories {
		if m.Importance < threshold {
			preview = append(preview, map[string]interface{}{
				"memory_id":  m.ID,
				"key":        m.Key,
				"value_len":  len(m.Value),
				"importance": m.Importance,
				"action":     "archive",
			})
		}
	}

	return map[string]interface{}{
		"mode":      "local",
		"level":     level,
		"threshold": threshold,
		"preview":   preview,
		"total":     len(preview),
	}, nil
}

func (s *ToolboxService) CompressApply(userID uint, level string) (map[string]interface{}, error) {
	preview, err := s.CompressPreview(userID, level)
	if err != nil {
		return nil, err
	}

	previewItems, _ := preview["preview"].([]map[string]interface{})
	archived := 0
	for _, item := range previewItems {
		if id, ok := item["memory_id"]; ok {
			var memoryID uint
			switch v := id.(type) {
			case uint:
				memoryID = v
			case float64:
				memoryID = uint(v)
			case int:
				memoryID = uint(v)
			}
			if err := s.db.Model(&models.Memory{}).Where("id = ? AND user_id = ?", memoryID, userID).
				Update("status", "archived").Error; err == nil {
				archived++
			}
		}
	}

	return map[string]interface{}{
		"mode":     "local",
		"level":    level,
		"archived": archived,
		"total":    len(previewItems),
	}, nil
}

func (s *ToolboxService) CompressConfig(userID uint) (map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return map[string]interface{}{
		"config": s.compressCfg,
	}, nil
}

func (s *ToolboxService) SetCompressConfig(userID uint, cfg map[string]interface{}) (map[string]interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if v, ok := cfg["auto_compress"]; ok {
		s.compressCfg["auto_compress"] = v
	}
	if v, ok := cfg["threshold"]; ok {
		s.compressCfg["threshold"] = v
	}
	if v, ok := cfg["level"]; ok {
		s.compressCfg["level"] = v
	}
	if v, ok := cfg["preserve_important"]; ok {
		s.compressCfg["preserve_important"] = v
	}
	return map[string]interface{}{
		"updated": true,
		"config":  s.compressCfg,
	}, nil
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
	layer := "core"
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
		layer = "semantic"
		strategy = "semantic_priority"
	case 4:
		complexityLabel = "extreme"
		layer = "summary"
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
	_ = s.db.Model(&models.Memory{}).Where("user_id = ? AND status != ?", userID, "trashed").Count(&totalMemories).Error

	var totalEntities int64
	_ = s.db.Model(&models.Entity{}).Where("user_id = ?", userID).Count(&totalEntities).Error

	var totalRelations int64
	_ = s.db.Model(&models.Relation{}).Where("user_id = ?", userID).Count(&totalRelations).Error

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

func (s *ToolboxService) EvolutionInsights(userID uint) (map[string]interface{}, error) {
	var memories []models.Memory
	_ = s.db.Where("user_id = ? AND status = ?", userID, "active").Limit(5000).Find(&memories).Error

	layerCount := make(map[string]int)
	sourceCount := make(map[string]int)
	for _, m := range memories {
		layerCount[m.Layer]++
		if m.Source != "" {
			sourceCount[m.Source]++
		}
	}

	return map[string]interface{}{
		"mode":         "local",
		"total":        len(memories),
		"layer_stats":  layerCount,
		"source_stats": sourceCount,
	}, nil
}

func (s *ToolboxService) EvolutionDiscover(userID uint) (map[string]interface{}, error) {
	var relations []models.Relation
	_ = s.db.Where("user_id = ?", userID).Limit(5000).Find(&relations).Error

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

func (s *ToolboxService) EvolutionInfer(userID uint) (map[string]interface{}, error) {
	var entities []models.Entity
	_ = s.db.Where("user_id = ?", userID).Limit(5000).Find(&entities).Error

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

func (s *ToolboxService) EvolutionImportance(userID uint) (map[string]interface{}, error) {
	var memories []models.Memory
	_ = s.db.Where("user_id = ? AND status = ?", userID, "active").Limit(5000).Find(&memories).Error

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

func (s *ToolboxService) EvolutionPrefetch(userID uint, context string) (map[string]interface{}, error) {
	var memories []models.Memory
	escaped := escapeLike(context)
	_ = s.db.Where("user_id = ? AND status != ?", userID, "trashed").
		Where("key LIKE ? OR value LIKE ?", "%"+escaped+"%", "%"+escaped+"%").
		Limit(200).Find(&memories).Error

	results := make([]map[string]interface{}, 0, len(memories))
	for _, m := range memories {
		results = append(results, map[string]interface{}{
			"id":         m.ID,
			"key":        m.Key,
			"importance": m.Importance,
			"layer":      m.Layer,
		})
	}

	return map[string]interface{}{
		"mode":    "local",
		"context": context,
		"results": results,
		"total":   len(results),
	}, nil
}

func (s *ToolboxService) ReinforceMemory(userID uint, memoryID uint) (map[string]interface{}, error) {
	var memory models.Memory
	if err := s.db.Where("id = ? AND user_id = ?", memoryID, userID).First(&memory).Error; err != nil {
		return nil, fmt.Errorf("memory not found: %w", err)
	}

	newImportance := math.Min(memory.Importance+0.1, 1.0)
	if err := s.db.Model(&memory).Updates(map[string]interface{}{
		"importance":   newImportance,
		"access_count": gorm.Expr("access_count + 1"),
	}).Error; err != nil {
		return nil, fmt.Errorf("failed to reinforce memory: %w", err)
	}

	return map[string]interface{}{
		"mode":           "local",
		"memory_id":      memoryID,
		"old_importance": memory.Importance,
		"new_importance": newImportance,
		"reinforced":     true,
	}, nil
}

func (s *ToolboxService) AutoGraph(userID uint, overwrite bool) (map[string]interface{}, error) {
	var memories []models.Memory
	_ = s.db.Where("user_id = ? AND status = ?", userID, "active").Limit(5000).Find(&memories).Error

	var relations []models.Relation
	_ = s.db.Where("user_id = ?", userID).Limit(5000).Find(&relations).Error

	existingPairs := make(map[string]bool)
	for _, r := range relations {
		key := fmt.Sprintf("%d-%s-%d", r.SourceID, r.RelationType, r.TargetID)
		existingPairs[key] = true
	}

	created := 0

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

	for _, mems := range tagMemories {
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
				}
			}
		}
	}

	var entities []models.Entity
	_ = s.db.Where("user_id = ?", userID).Limit(1000).Find(&entities).Error

	entityMemories := make(map[uint][]uint)
	for _, e := range entities {
		if e.SourceMemoryID != nil {
			entityMemories[e.ID] = append(entityMemories[e.ID], *e.SourceMemoryID)
		}
	}

	for _, memIDs := range entityMemories {
		if len(memIDs) < 2 {
			continue
		}
		for i := 0; i < len(memIDs); i++ {
			for j := i + 1; j < len(memIDs); j++ {
				pairKey := fmt.Sprintf("%d-shared_entity-%d", memIDs[i], memIDs[j])
				if existingPairs[pairKey] && !overwrite {
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
				}
			}
		}
	}

	return map[string]interface{}{
		"mode":        "local",
		"created":     created,
		"total_pairs": len(existingPairs),
	}, nil
}

func (s *ToolboxService) AIExtract(userID uint) (map[string]interface{}, error) {
	var memories []models.Memory
	_ = s.db.Where("user_id = ? AND status != ?", userID, "trashed").Select("value").Limit(5000).Find(&memories).Error

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

func (s *ToolboxService) BackupSchedule(userID uint) (map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return map[string]interface{}{
		"enabled":        s.backupCfg.Enabled,
		"interval_hours": s.backupCfg.IntervalHours,
	}, nil
}

func (s *ToolboxService) SetBackupSchedule(userID uint, enabled bool, intervalHours int) (map[string]interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if intervalHours < 1 {
		intervalHours = 24
	}
	if intervalHours > 720 {
		intervalHours = 720
	}
	s.backupCfg.Enabled = enabled
	s.backupCfg.IntervalHours = intervalHours
	return map[string]interface{}{
		"enabled":        enabled,
		"interval_hours": intervalHours,
		"updated":        true,
	}, nil
}

func (s *ToolboxService) SmartLoad(userID uint, context string) (map[string]interface{}, error) {
	var memories []models.Memory
	escaped := escapeLike(context)
	_ = s.db.Where("user_id = ? AND status != ?", userID, "trashed").
		Where("key LIKE ? OR value LIKE ?", "%"+escaped+"%", "%"+escaped+"%").
		Limit(200).Find(&memories).Error

	sort.Slice(memories, func(i, j int) bool {
		return memories[i].Importance > memories[j].Importance
	})

	if len(memories) > 50 {
		memories = memories[:50]
	}

	results := make([]map[string]interface{}, 0, len(memories))
	for _, m := range memories {
		results = append(results, map[string]interface{}{
			"id":         m.ID,
			"key":        m.Key,
			"value":      truncateStr(m.Value, 200),
			"importance": m.Importance,
			"layer":      m.Layer,
		})
	}

	return map[string]interface{}{
		"mode":    "local",
		"context": context,
		"results": results,
		"total":   len(results),
	}, nil
}
