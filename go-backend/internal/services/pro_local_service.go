package services

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"clawmemory/internal/models"

	"gorm.io/gorm"
)

type ProLocalService struct {
	db *gorm.DB
}

func NewProLocalService(db *gorm.DB) *ProLocalService {
	return &ProLocalService{db: db}
}

func (s *ProLocalService) DecayStats(userID uint) (map[string]interface{}, error) {
	decaySvc := NewDecayService(s.db)
	stats, err := decaySvc.GetStats(userID)
	if err != nil {
		return nil, err
	}

	suggestions := []map[string]interface{}{}
	var lowImportance []models.Memory
	s.db.Where("user_id = ? AND importance < ? AND status != ?", userID, 0.3, "trashed").
		Order("importance ASC").Limit(10).Find(&lowImportance)
	for _, m := range lowImportance {
		suggestions = append(suggestions, map[string]interface{}{
			"memory_id":  m.ID,
			"key":        m.Key,
			"importance": m.Importance,
			"suggestion": "consider archiving or deleting",
			"days_old":   int(time.Since(m.UpdatedAt).Hours() / 24),
		})
	}

	var totalMemories int64
	s.db.Model(&models.Memory{}).Where("user_id = ? AND status != ?", userID, "trashed").Count(&totalMemories)

	avgImportance := 0.0
	if totalMemories > 0 {
		type avgResult struct {
			Avg float64
		}
		var result avgResult
		s.db.Model(&models.Memory{}).
			Where("user_id = ? AND status != ?", userID, "trashed").
			Select("AVG(importance) as avg").
			Scan(&result)
		avgImportance = math.Round(result.Avg*100) / 100
	}

	return map[string]interface{}{
		"total":            stats.Total,
		"active":           stats.Active,
		"archived":         stats.Archived,
		"trashed":          stats.Trashed,
		"prune_candidates": len(lowImportance),
		"avg_importance":   avgImportance,
		"suggestions":      suggestions,
		"mode":             "local_pro",
	}, nil
}

func (s *ProLocalService) DecayApply(userID uint) (map[string]interface{}, error) {
	decaySvc := NewDecayService(s.db)
	return decaySvc.ApplyDecay(userID)
}

func (s *ProLocalService) PruneSuggest(userID uint) (map[string]interface{}, error) {
	decaySvc := NewDecayService(s.db)
	suggestions, err := decaySvc.GetPruneSuggestions(userID)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"suggestions": suggestions,
		"algorithm":   "local_importance_v1",
		"mode":        "local_pro",
	}, nil
}

func (s *ProLocalService) ConflictScan(userID uint) (map[string]interface{}, error) {
	var memories []models.Memory
	s.db.Where("user_id = ? AND status != ?", userID, "trashed").Find(&memories)

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
				severity := "low"
				if len(mems) > 3 {
					severity = "high"
				} else if len(mems) > 2 {
					severity = "medium"
				}
				conflict := map[string]interface{}{
					"key":      key,
					"count":    len(mems),
					"value_a":  values[0],
					"value_b":  values[1],
					"severity": severity,
					"memory_ids": func() []uint {
						ids := make([]uint, len(mems))
						for i, m := range mems {
							ids[i] = m.ID
						}
						return ids
					}(),
					"conflict": "different_values_same_key",
				}
				conflicts = append(conflicts, conflict)
			}
		}
	}

	return map[string]interface{}{
		"conflicts":       conflicts,
		"total":           len(conflicts),
		"auto_resolvable": len(conflicts),
		"needs_review":    0,
		"algorithm":       "local_key_conflict_v1",
		"mode":            "local_pro",
	}, nil
}

func (s *ProLocalService) CompressPreview(userID uint, level string) (map[string]interface{}, error) {
	var memories []models.Memory
	s.db.Where("user_id = ? AND status != ?", userID, "trashed").Find(&memories)

	if len(memories) == 0 {
		return map[string]interface{}{
			"original_count":   0,
			"compressed_count": 0,
			"ratio":            0,
			"mode":             "local_pro",
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
			"layer":      layer,
			"original":   len(mems),
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
		"mode":             "local_pro",
	}, nil
}

func (s *ProLocalService) CompressApply(userID uint, level string) (map[string]interface{}, error) {
	var memories []models.Memory
	s.db.Where("user_id = ? AND status != ?", userID, "trashed").Find(&memories)

	if len(memories) == 0 {
		return map[string]interface{}{
			"processed":  0,
			"compressed": 0,
			"ratio":      0,
			"mode":       "local_pro",
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

	for i := 0; i < len(lowImportance)-1; i++ {
		for j := i + 1; j < len(lowImportance); j++ {
			if lowImportance[j].Importance < lowImportance[i].Importance {
				lowImportance[i], lowImportance[j] = lowImportance[j], lowImportance[i]
			}
		}
	}

	targetCount := int(float64(len(lowImportance)) * rate)
	compressed := 0

	for i := 0; i < targetCount && i < len(lowImportance); i++ {
		m := lowImportance[i]
		s.db.Model(&m).Updates(map[string]interface{}{
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
		"mode":       "local_pro",
	}, nil
}

func (s *ProLocalService) EvolutionDiscover(userID uint) (map[string]interface{}, error) {
	recommendSvc := NewRecommendService(s.db)

	var memories []models.Memory
	s.db.Where("user_id = ? AND status != ?", userID, "trashed").Find(&memories)

	relations := []map[string]interface{}{}
	seen := make(map[string]bool)

	for i := 0; i < len(memories) && len(relations) < 20; i++ {
		result, err := recommendSvc.RecommendForMemory(userID, memories[i].ID, 3)
		if err != nil {
			continue
		}
		if recs, ok := result["recommendations"].([]map[string]interface{}); ok {
			for _, rec := range recs {
				var targetID uint
				switch v := rec["id"].(type) {
				case float64:
					targetID = uint(v)
				case uint:
					targetID = v
				case int:
					targetID = uint(v)
				default:
					continue
				}
				key := fmt.Sprintf("%d-%d", memories[i].ID, targetID)
				if !seen[key] {
					seen[key] = true
					score := 0.5
					if s, ok := rec["score"].(float64); ok {
						score = s
					}
					relationType := "related_to"
					if reason, ok := rec["reason"].(string); ok {
						relationType = reason
					}
					relations = append(relations, map[string]interface{}{
						"source":     memories[i].Key,
						"target":     rec["key"],
						"type":       relationType,
						"confidence": math.Round(score*100) / 100,
						"source_id":  memories[i].ID,
						"target_id":  targetID,
						"reason":     rec["reason"],
					})
				}
			}
		}
	}

	return map[string]interface{}{
		"relations": relations,
		"total":     len(relations),
		"algorithm": "local_token_similarity_v1",
		"mode":      "local_pro",
	}, nil
}

func (s *ProLocalService) EvolutionPrefetch(userID uint, context string) (map[string]interface{}, error) {
	recommendSvc := NewRecommendService(s.db)
	result, err := recommendSvc.RecommendByContext(userID, context, 10)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"matched":   result["recommendations"],
		"total":     result["total"],
		"context":   context,
		"algorithm": "local_context_match_v1",
		"mode":      "local_pro",
	}, nil
}

func (s *ProLocalService) EvolutionImportance(userID uint) (map[string]interface{}, error) {
	var memories []models.Memory
	s.db.Where("user_id = ? AND status != ?", userID, "trashed").Find(&memories)

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
			adjustedImportance = math.Round(adjustedImportance*100) / 100
			s.db.Model(&m).Update("importance", adjustedImportance)
			adjustments = append(adjustments, map[string]interface{}{
				"id":                  m.ID,
				"key":                 m.Key,
				"current_importance":  m.Importance,
				"adjusted_importance": adjustedImportance,
				"reason":              getImportanceReason(daysSinceUpdate, m.AccessCount),
			})
		}
	}

	return map[string]interface{}{
		"adjustments": adjustments,
		"total":       len(adjustments),
		"algorithm":   "local_time_access_v1",
		"mode":        "local_pro",
	}, nil
}

func (s *ProLocalService) ReinforceMemory(userID uint, memoryID uint) (map[string]interface{}, error) {
	var memory models.Memory
	if err := s.db.Where("id = ? AND user_id = ?", memoryID, userID).First(&memory).Error; err != nil {
		return nil, fmt.Errorf("memory not found")
	}
	newImportance := math.Min(memory.Importance*1.2, 1.0)
	s.db.Model(&memory).Updates(map[string]interface{}{
		"importance":       newImportance,
		"access_count":     memory.AccessCount + 1,
		"reinforce_count":  memory.ReinforceCount + 1,
		"last_accessed_at": time.Now(),
	})
	return map[string]interface{}{
		"memory_id":       memoryID,
		"old_importance":  memory.Importance,
		"new_importance":  newImportance,
		"reinforced":      true,
		"reinforce_count": memory.ReinforceCount + 1,
		"algorithm":       "local_access_reinforce_v1",
		"mode":            "local_pro",
	}, nil
}

func (s *ProLocalService) ConflictResolve(userID uint, conflictIndex int, strategy string) (map[string]interface{}, error) {
	result, err := s.ConflictScan(userID)
	if err != nil {
		return nil, err
	}
	conflicts, ok := result["conflicts"].([]map[string]interface{})
	if !ok || conflictIndex >= len(conflicts) {
		return nil, fmt.Errorf("conflict index out of range")
	}
	conflict := conflicts[conflictIndex]
	memoryIDs, ok := conflict["memory_ids"].([]uint)
	if !ok || len(memoryIDs) == 0 {
		return nil, fmt.Errorf("no memories in conflict")
	}

	var memories []models.Memory
	s.db.Where("id IN ? AND user_id = ?", memoryIDs, userID).Order("id ASC").Find(&memories)
	if len(memories) == 0 {
		return nil, fmt.Errorf("no memories found for conflict")
	}

	switch strategy {
	case "keep_first":
		for i := 1; i < len(memories); i++ {
			s.db.Model(&memories[i]).Update("status", "archived")
		}
	case "keep_latest":
		latest := 0
		for i := 1; i < len(memories); i++ {
			if memories[i].UpdatedAt.After(memories[latest].UpdatedAt) {
				latest = i
			}
		}
		for i := 0; i < len(memories); i++ {
			if i != latest {
				s.db.Model(&memories[i]).Update("status", "archived")
			}
		}
	default:
		for i := 1; i < len(memories); i++ {
			s.db.Model(&memories[i]).Update("status", "archived")
		}
	}
	return map[string]interface{}{
		"resolved": true,
		"strategy": strategy,
		"kept":     1,
		"archived": len(memories) - 1,
		"mode":     "local_pro",
	}, nil
}

func (s *ProLocalService) TokenRoute(message string, contextLength int) (map[string]interface{}, error) {
	tokenEstimate := len(message) / 4
	if contextLength > 0 {
		tokenEstimate += contextLength
	}
	return map[string]interface{}{
		"model":            "local",
		"estimated_tokens": tokenEstimate,
		"provider":         "local_pro",
		"mode":             "local_pro",
	}, nil
}

func (s *ProLocalService) TokenStats(userID uint) (map[string]interface{}, error) {
	var totalMemories int64
	s.db.Model(&models.Memory{}).Where("user_id = ? AND status != ?", userID, "trashed").Count(&totalMemories)
	estimatedTokens := totalMemories * 50
	return map[string]interface{}{
		"total_tokens_used": estimatedTokens,
		"total_memories":    totalMemories,
		"provider":          "local_pro",
		"mode":              "local_pro",
	}, nil
}

func inferEntityType(key, value string) string {
	k := strings.ToLower(key)
	v := strings.ToLower(value)

	personPatterns := []string{"user", "name", "author", "developer", "person", "people", "member", "colleague", "friend", "manager", "boss", "同事", "用户", "开发者", "经理"}
	for _, p := range personPatterns {
		if strings.Contains(k, p) {
			return "person"
		}
	}

	orgPatterns := []string{"company", "organization", "team", "org", "department", "group", "公司", "团队", "组织", "部门"}
	for _, p := range orgPatterns {
		if strings.Contains(k, p) {
			return "organization"
		}
	}

	locationPatterns := []string{"location", "address", "city", "country", "place", "server", "host", "ip", "url", "endpoint", "地址", "位置", "城市"}
	for _, p := range locationPatterns {
		if strings.Contains(k, p) {
			return "location"
		}
	}

	techPatterns := []string{"tech", "tool", "framework", "library", "language", "database", "api", "sdk", "package", "plugin", "version", "config", "技术", "工具", "框架", "库"}
	for _, p := range techPatterns {
		if strings.Contains(k, p) {
			return "technology"
		}
	}

	eventPatterns := []string{"event", "meeting", "deadline", "schedule", "task", "todo", "plan", "milestone", "release", "deploy", "事件", "会议", "任务", "计划"}
	for _, p := range eventPatterns {
		if strings.Contains(k, p) {
			return "event"
		}
	}

	if strings.Contains(v, "http://") || strings.Contains(v, "https://") || strings.Contains(v, "github.com") {
		return "location"
	}

	if strings.Contains(v, "v1.") || strings.Contains(v, "v2.") || regexpCheck(`^\d+\.\d+`, v) {
		return "technology"
	}

	return "concept"
}

func regexpCheck(pattern, s string) bool {
	matched, _ := regexp.MatchString(pattern, s)
	return matched
}

func (s *ProLocalService) AIExtract(userID uint) (map[string]interface{}, error) {
	var memories []models.Memory
	s.db.Where("user_id = ? AND status != ?", userID, "trashed").Find(&memories)

	entities := []map[string]interface{}{}
	seen := make(map[string]bool)

	for _, m := range memories {
		name := strings.TrimSpace(m.Key)
		if name == "" || len(name) < 2 || len(name) > 100 {
			continue
		}
		if seen[strings.ToLower(name)] {
			continue
		}
		seen[strings.ToLower(name)] = true

		entityType := inferEntityType(name, m.Value)
		description := m.Value
		if len(description) > 500 {
			description = description[:500]
		}

		entities = append(entities, map[string]interface{}{
			"name":             name,
			"entity_type":      entityType,
			"description":      description,
			"extract_method":   "auto",
			"confidence":       m.Importance,
			"source_memory_id": float64(m.ID),
		})
	}

	if len(entities) > 200 {
		entities = entities[:200]
	}

	return map[string]interface{}{
		"entities":  entities,
		"total":     len(entities),
		"algorithm": "local_key_extract_v2",
		"mode":      "local_pro",
	}, nil
}

func (s *ProLocalService) AutoGraph(userID uint, overwrite bool) (map[string]interface{}, error) {
	if overwrite {
		s.db.Where("user_id = ?", userID).Delete(&models.Entity{})
		s.db.Where("user_id = ?", userID).Delete(&models.Relation{})
	}

	extractResult, err := s.AIExtract(userID)
	if err != nil {
		return nil, err
	}

	entitiesRaw, _ := extractResult["entities"].([]map[string]interface{})
	knowledgeSvc := NewKnowledgeService(s.db)

	entitiesCreated := 0
	entityMap := make(map[string]uint)

	for _, eData := range entitiesRaw {
		name, _ := eData["name"].(string)
		if name == "" {
			continue
		}

		if existingID, found := entityMap[name]; found && !overwrite {
			entityMap[name] = existingID
			continue
		}

		var existing models.Entity
		if !overwrite {
			if s.db.Where("user_id = ? AND name = ?", userID, name).First(&existing).Error == nil {
				entityMap[name] = existing.ID
				continue
			}
		}

		entity, err := knowledgeSvc.CreateEntity(userID, eData)
		if err != nil {
			continue
		}
		entityMap[name] = entity.ID
		entitiesCreated++
	}

	relationsCreated := 0
	var allEntities []models.Entity
	s.db.Where("user_id = ?", userID).Find(&allEntities)

	for i := 0; i < len(allEntities) && relationsCreated < 50; i++ {
		for j := i + 1; j < len(allEntities) && relationsCreated < 50; j++ {
			e1 := allEntities[i]
			e2 := allEntities[j]

			relType, desc, conf := inferRelation(e1, e2)
			if relType == "" {
				continue
			}

			var existing models.Relation
			err := s.db.Where("user_id = ? AND source_id = ? AND target_id = ? AND relation_type = ?",
				userID, e1.ID, e2.ID, relType).First(&existing).Error
			if err == nil {
				continue
			}

			if err := s.db.Create(&models.Relation{
				UserID:         userID,
				SourceID:       e1.ID,
				TargetID:       e2.ID,
				RelationType:   relType,
				Description:    desc,
				Confidence:     conf,
				DiscoverMethod: "auto_graph",
				Weight:         conf,
			}).Error; err == nil {
				relationsCreated++
			}
		}
	}

	return map[string]interface{}{
		"entities_created":  entitiesCreated,
		"relations_created": relationsCreated,
		"overwrite":         overwrite,
		"algorithm":         "local_key_extract_v1",
		"mode":              "local_pro",
	}, nil
}

func (s *ProLocalService) BackupSchedule(userID uint) (map[string]interface{}, error) {
	settingsSvc := NewSettingsService(s.db)
	enabled, _ := settingsSvc.GetByKey(userID, "pro_backup_enabled")
	if enabled == nil {
		enabled = false
	}
	intervalHours := 24
	if ih, _ := settingsSvc.GetByKey(userID, "pro_backup_interval_hours"); ih != nil {
		if v, ok := ih.(float64); ok {
			intervalHours = int(v)
		}
	}
	return map[string]interface{}{
		"enabled":        enabled,
		"interval_hours": intervalHours,
		"mode":           "local_pro",
	}, nil
}

func (s *ProLocalService) SetBackupSchedule(userID uint, enabled bool, intervalHours int) (map[string]interface{}, error) {
	settingsSvc := NewSettingsService(s.db)
	settingsSvc.SetByKey(userID, "pro_backup_enabled", enabled)
	settingsSvc.SetByKey(userID, "pro_backup_interval_hours", intervalHours)
	return map[string]interface{}{
		"enabled":        enabled,
		"interval_hours": intervalHours,
		"message":        "backup schedule saved",
		"mode":           "local_pro",
	}, nil
}

func (s *ProLocalService) CompressConfig(userID uint) (map[string]interface{}, error) {
	settingsSvc := NewSettingsService(s.db)
	level := "light"
	if l, _ := settingsSvc.GetByKey(userID, "pro_compress_level"); l != nil {
		if v, ok := l.(string); ok {
			level = v
		}
	}
	autoEnabled := false
	if ae, _ := settingsSvc.GetByKey(userID, "pro_compress_auto_enabled"); ae != nil {
		if v, ok := ae.(bool); ok {
			autoEnabled = v
		}
	}
	threshold := 1000
	if t, _ := settingsSvc.GetByKey(userID, "pro_compress_threshold"); t != nil {
		if v, ok := t.(float64); ok {
			threshold = int(v)
		}
	}
	return map[string]interface{}{
		"level":        level,
		"auto_enabled": autoEnabled,
		"threshold":    threshold,
		"mode":         "local_pro",
	}, nil
}

func (s *ProLocalService) SetCompressConfig(userID uint, config map[string]interface{}) (map[string]interface{}, error) {
	settingsSvc := NewSettingsService(s.db)
	if level, ok := config["level"].(string); ok {
		settingsSvc.SetByKey(userID, "pro_compress_level", level)
	}
	if autoEnabled, ok := config["auto_enabled"].(bool); ok {
		settingsSvc.SetByKey(userID, "pro_compress_auto_enabled", autoEnabled)
	}
	if threshold, ok := config["threshold"].(float64); ok {
		settingsSvc.SetByKey(userID, "pro_compress_threshold", int(threshold))
	}
	return map[string]interface{}{
		"config":  config,
		"message": "compress config saved",
		"mode":    "local_pro",
	}, nil
}

func (s *ProLocalService) EvolutionInsights(userID uint) (map[string]interface{}, error) {
	var totalMemories int64
	s.db.Model(&models.Memory{}).Where("user_id = ? AND status != ?", userID, "trashed").Count(&totalMemories)

	var totalEntities int64
	s.db.Model(&models.Entity{}).Where("user_id = ?", userID).Count(&totalEntities)

	var totalRelations int64
	s.db.Model(&models.Relation{}).Where("user_id = ?", userID).Count(&totalRelations)

	var discoveredRelations int64
	s.db.Model(&models.Relation{}).Where("user_id = ? AND discover_method != ?", userID, "manual").Count(&discoveredRelations)

	healthScore := 0.5
	if totalMemories > 0 && totalEntities > 0 {
		ratio := float64(totalRelations) / float64(totalEntities)
		if ratio > 1.0 {
			healthScore = 0.9
		} else if ratio > 0.5 {
			healthScore = 0.7
		}
	}

	return map[string]interface{}{
		"total_memories":       totalMemories,
		"total_entities":       totalEntities,
		"total_relations":      totalRelations,
		"discovered_relations": discoveredRelations,
		"inferred_chains":      0,
		"health_score":         math.Round(healthScore*100) / 100,
		"recommendations":      []string{"regularly clean low-importance memories", "add tags to key memories", "use knowledge graph to build associations"},
		"mode":                 "local_pro",
	}, nil
}

func (s *ProLocalService) EvolutionInfer(userID uint) (map[string]interface{}, error) {
	var entities []models.Entity
	s.db.Where("user_id = ?", userID).Find(&entities)

	var relations []models.Relation
	s.db.Where("user_id = ?", userID).Find(&relations)

	entityMap := make(map[uint]models.Entity)
	for _, e := range entities {
		entityMap[e.ID] = e
	}

	adj := make(map[uint][]models.Relation)
	for _, r := range relations {
		adj[r.SourceID] = append(adj[r.SourceID], r)
	}

	chains := []map[string]interface{}{}
	visited := make(map[string]bool)

	for startID := range adj {
		paths := [][]uint{}
		var dfs func(current uint, path []uint, depth int)
		dfs = func(current uint, path []uint, depth int) {
			if depth >= 3 {
				return
			}
			for _, r := range adj[current] {
				next := r.TargetID
				inPath := false
				for _, p := range path {
					if p == next {
						inPath = true
						break
					}
				}
				if inPath {
					continue
				}
				newPath := append([]uint{}, path...)
				newPath = append(newPath, next)
				if len(newPath) >= 3 {
					paths = append(paths, newPath)
				}
				dfs(next, newPath, depth+1)
			}
		}
		dfs(startID, []uint{startID}, 0)

		for _, path := range paths {
			if len(chains) >= 10 {
				break
			}
			key := fmt.Sprintf("%v", path)
			if visited[key] {
				continue
			}
			visited[key] = true

			nodes := []string{}
			for _, id := range path {
				if e, ok := entityMap[id]; ok {
					nodes = append(nodes, e.Name)
				} else {
					nodes = append(nodes, fmt.Sprintf("entity_%d", id))
				}
			}

			lastID := path[len(path)-1]
			conclusion := "related entities form a cluster"
			if lastE, ok := entityMap[lastID]; ok {
				conclusion = fmt.Sprintf("%s is connected through a chain of relationships", lastE.Name)
			}

			chains = append(chains, map[string]interface{}{
				"nodes":      nodes,
				"conclusion": conclusion,
				"length":     len(nodes),
			})
		}
		if len(chains) >= 10 {
			break
		}
	}

	return map[string]interface{}{
		"chains":    chains,
		"total":     len(chains),
		"algorithm": "local_graph_traversal_v1",
		"mode":      "local_pro",
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

func inferRelation(e1, e2 models.Entity) (string, string, float64) {
	name1 := strings.ToLower(e1.Name)
	name2 := strings.ToLower(e2.Name)
	desc1 := strings.ToLower(e1.Description)
	desc2 := strings.ToLower(e2.Description)

	if strings.Contains(desc2, name1) || strings.Contains(desc1, name2) {
		return "references", fmt.Sprintf("%s references %s", e1.Name, e2.Name), 0.7
	}

	typeRelationMap := map[string]map[string]struct {
		relType string
		desc    string
		conf    float64
	}{
		"person": {
			"organization": {"member_of", "is a member of", 0.7},
			"technology":   {"uses", "uses technology", 0.6},
			"concept":      {"understands", "understands concept", 0.5},
			"event":        {"participated_in", "participated in", 0.7},
			"location":     {"located_at", "is located at", 0.6},
			"person":       {"collaborates_with", "collaborates with", 0.5},
		},
		"technology": {
			"concept":      {"implements", "implements concept", 0.6},
			"technology":   {"depends_on", "depends on", 0.6},
			"organization": {"used_by", "is used by", 0.5},
			"person":       {"used_by", "is used by", 0.6},
			"event":        {"released_at", "released at", 0.5},
		},
		"organization": {
			"location": {"headquartered_at", "headquartered at", 0.6},
			"event":    {"organized", "organized", 0.7},
			"person":   {"employs", "employs", 0.6},
		},
		"event": {
			"location":   {"held_at", "held at", 0.7},
			"concept":    {"about", "is about", 0.5},
			"person":     {"attended_by", "attended by", 0.6},
			"technology": {"features", "features", 0.5},
		},
		"concept": {
			"concept": {"related_to", "is related to", 0.4},
		},
		"location": {
			"location": {"near", "is near", 0.4},
		},
	}

	if typeRels, ok := typeRelationMap[e1.EntityType]; ok {
		if rel, ok := typeRels[e2.EntityType]; ok {
			return rel.relType, fmt.Sprintf("%s %s %s", e1.Name, rel.desc, e2.Name), rel.conf
		}
	}

	if e1.EntityType == e2.EntityType {
		return "same_type", fmt.Sprintf("Both are %s", e1.EntityType), 0.3
	}

	commonWords := 0
	words1 := strings.Fields(desc1)
	wordSet := make(map[string]bool)
	for _, w := range words1 {
		if len(w) > 3 {
			wordSet[w] = true
		}
	}
	for _, w := range strings.Fields(desc2) {
		if len(w) > 3 && wordSet[w] {
			commonWords++
		}
	}
	if commonWords >= 2 {
		return "related_to", fmt.Sprintf("%s and %s share common context", e1.Name, e2.Name), 0.4
	}

	return "", "", 0
}
