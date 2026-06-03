package services

import (
	"fmt"
	"math"
	"sort"
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

type labelPropagationResult struct {
	Labels      map[uint]uint
	Communities map[uint][]models.Entity
	EntityMap   map[uint]models.Entity
}

func (s *EvolutionService) runLabelPropagation(entities []models.Entity, relations []models.Relation) *labelPropagationResult {
	adj := make(map[uint][]uint)
	for _, r := range relations {
		adj[r.SourceID] = append(adj[r.SourceID], r.TargetID)
		adj[r.TargetID] = append(adj[r.TargetID], r.SourceID)
	}

	entityIDSet := make(map[uint]bool)
	entityMap := make(map[uint]models.Entity)
	for _, e := range entities {
		entityIDSet[e.ID] = true
		entityMap[e.ID] = e
	}

	labels := make(map[uint]uint)
	for i, e := range entities {
		labels[e.ID] = uint(i)
	}

	for iter := 0; iter < 20; iter++ {
		changed := false
		for _, e := range entities {
			neighbors := adj[e.ID]
			if len(neighbors) == 0 {
				continue
			}
			labelCount := make(map[uint]int)
			for _, nb := range neighbors {
				if entityIDSet[nb] {
					labelCount[labels[nb]]++
				}
			}
			if len(labelCount) == 0 {
				continue
			}
			maxCount := 0
			var bestLabel uint
			for lbl, cnt := range labelCount {
				if cnt > maxCount {
					maxCount = cnt
					bestLabel = lbl
				}
			}
			if labels[e.ID] != bestLabel {
				labels[e.ID] = bestLabel
				changed = true
			}
		}
		if !changed {
			break
		}
	}

	communities := make(map[uint][]models.Entity)
	for _, e := range entities {
		communities[labels[e.ID]] = append(communities[labels[e.ID]], e)
	}

	return &labelPropagationResult{
		Labels:      labels,
		Communities: communities,
		EntityMap:   entityMap,
	}
}

func (s *EvolutionService) CommunityDiscovery(userID uint) (map[string]interface{}, error) {
	var entities []models.Entity
	logDBErr("load entities for community discovery", s.db.Where("user_id = ?", userID).Limit(5000).Find(&entities).Error)

	var relations []models.Relation
	logDBErr("load relations for community discovery", s.db.Where("user_id = ?", userID).Limit(5000).Find(&relations).Error)

	if len(entities) == 0 {
		return map[string]interface{}{
			"mode":        "local",
			"communities": []map[string]interface{}{},
			"total":       0,
		}, nil
	}

	lpResult := s.runLabelPropagation(entities, relations)

	type Community struct {
		ID        uint                     `json:"id"`
		Size      int                      `json:"size"`
		Entities  []map[string]interface{} `json:"entities"`
		TopEntity string                   `json:"top_entity"`
	}

	var result []Community
	for label, ents := range lpResult.Communities {
		if len(ents) < 2 {
			continue
		}
		topEntity := ents[0].Name
		if len(ents) > 1 {
			topEntity = ents[0].Name
		}
		entityList := make([]map[string]interface{}, 0, len(ents))
		for _, e := range ents {
			entityList = append(entityList, map[string]interface{}{
				"id":          e.ID,
				"name":        e.Name,
				"entity_type": e.EntityType,
				"confidence":  e.Confidence,
			})
		}
		if len(entityList) > 30 {
			entityList = entityList[:30]
		}
		result = append(result, Community{
			ID:        label,
			Size:      len(ents),
			Entities:  entityList,
			TopEntity: topEntity,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Size > result[j].Size
	})

	return map[string]interface{}{
		"mode":        "local",
		"communities": result,
		"total":       len(result),
	}, nil
}

func (s *EvolutionService) CommunitiesToWiki(userID uint) (map[string]interface{}, error) {
	var entities []models.Entity
	logDBErr("load entities for communities to wiki", s.db.Where("user_id = ?", userID).Limit(5000).Find(&entities).Error)

	var relations []models.Relation
	logDBErr("load relations for communities to wiki", s.db.Where("user_id = ?", userID).Limit(5000).Find(&relations).Error)

	lpResult := s.runLabelPropagation(entities, relations)
	communityEntities := lpResult.Communities

	wikiSvc := NewWikiService(s.db)
	existingCategories, _ := wikiSvc.GetCategories(userID)
	existingCatSet := make(map[string]bool)
	for _, cat := range existingCategories {
		existingCatSet[strings.ToLower(cat)] = true
	}

	createdPages := 0
	createdCategories := []string{}
	skippedCommunities := 0

	for _, ents := range communityEntities {
		if len(ents) < 2 {
			continue
		}

		topEntity := ents[0]
		for _, e := range ents[1:] {
			if e.Confidence > topEntity.Confidence {
				topEntity = e
			}
		}

		category := topEntity.EntityType
		if category == "" || category == "concept" {
			category = "knowledge_domain"
		}
		category = "domain_" + category

		title := topEntity.Name + " Domain"
		contentParts := []string{}
		for _, e := range ents {
			desc := e.Name
			if e.Description != "" {
				desc += ": " + e.Description
			}
			if e.Confidence > 0 {
				desc += fmt.Sprintf(" (%.0f%%)", e.Confidence*100)
			}
			contentParts = append(contentParts, desc)
		}
		content := "Entities in this domain:\n" + strings.Join(contentParts, "\n")

		if existingCatSet[strings.ToLower(category)] {
			var existingPages []models.WikiPage
			s.db.Where("user_id = ? AND category = ? AND title = ?", userID, category, title).Limit(1).Find(&existingPages)
			if len(existingPages) > 0 {
				s.db.Model(&existingPages[0]).Updates(map[string]interface{}{
					"content": content,
				})
				createdPages++
				continue
			}
		}

		_, err := wikiSvc.Create(userID, map[string]interface{}{
			"title":    title,
			"content":  content,
			"category": category,
			"status":   "in_progress",
		})
		if err == nil {
			createdPages++
			if !existingCatSet[strings.ToLower(category)] {
				createdCategories = append(createdCategories, category)
				existingCatSet[strings.ToLower(category)] = true
			}
		}
	}

	return map[string]interface{}{
		"mode":                "local",
		"communities_found":   len(communityEntities),
		"wiki_pages_created":  createdPages,
		"categories_created":  createdCategories,
		"skipped_communities": skippedCommunities,
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
	lowQuality := 0
	createdRelations := []map[string]interface{}{}

	minRelationSimilarity := 0.3

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
		candidates := mems
		if len(candidates) > 20 {
			candidates = candidates[:20]
		}
		for i := 0; i < len(candidates); i++ {
			for j := i + 1; j < len(candidates); j++ {
				sim := computeSimilarity(candidates[i].Key, candidates[i].Value, candidates[j].Key, candidates[j].Value)
				if sim < minRelationSimilarity {
					lowQuality++
					continue
				}
				pairKey := fmt.Sprintf("%d-same_topic-%d", candidates[i].ID, candidates[j].ID)
				if existingPairs[pairKey] && !overwrite {
					skipped++
					continue
				}
				weight := 0.4 + sim*0.4
				rel := models.Relation{
					SourceID:     candidates[i].ID,
					TargetID:     candidates[j].ID,
					RelationType: "same_topic",
					UserID:       userID,
					Weight:       weight,
				}
				if err := s.db.Create(&rel).Error; err == nil {
					created++
					createdRelations = append(createdRelations, map[string]interface{}{
						"source":        candidates[i].Key,
						"target":        candidates[j].Key,
						"relation_type": "same_topic",
						"weight":        weight,
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
		for _, tag := range parseMemoryTags(m.Tags) {
			if tag != "" {
				tagMemories[tag] = append(tagMemories[tag], m)
			}
		}
	}

	for tag, mems := range tagMemories {
		if len(mems) < 2 || len(mems) > 100 {
			continue
		}
		candidates := mems
		if len(candidates) > 20 {
			candidates = candidates[:20]
		}
		for i := 0; i < len(candidates); i++ {
			for j := i + 1; j < len(candidates); j++ {
				if candidates[i].Key == candidates[j].Key {
					continue
				}
				sim := computeSimilarity(candidates[i].Key, candidates[i].Value, candidates[j].Key, candidates[j].Value)
				if sim < minRelationSimilarity {
					lowQuality++
					continue
				}
				pairKey := fmt.Sprintf("%d-shared_tags-%d", candidates[i].ID, candidates[j].ID)
				if existingPairs[pairKey] && !overwrite {
					skipped++
					continue
				}
				weight := 0.2 + sim*0.3
				rel := models.Relation{
					SourceID:     candidates[i].ID,
					TargetID:     candidates[j].ID,
					RelationType: "shared_tags",
					UserID:       userID,
					Weight:       weight,
				}
				if err := s.db.Create(&rel).Error; err == nil {
					created++
					createdRelations = append(createdRelations, map[string]interface{}{
						"source":        candidates[i].Key,
						"target":        candidates[j].Key,
						"relation_type": "shared_tags",
						"weight":        weight,
						"tag":           tag,
					})
				}
			}
		}
	}

	var entities []models.Entity
	logDBErr("load entities for auto graph", s.db.Where("user_id = ?", userID).Limit(1000).Find(&entities).Error)

	entityMemories := make(map[string][]uint)
	for _, e := range entities {
		if e.SourceMemoryID != nil {
			name := e.Name
			if e.CanonicalName != "" {
				name = e.CanonicalName
			}
			entityMemories[name] = append(entityMemories[name], *e.SourceMemoryID)
		}
	}

	for ename, memIDs := range entityMemories {
		if len(memIDs) < 2 {
			continue
		}
		// deduplicate memory IDs
		seen := make(map[uint]bool)
		var uniqueIDs []uint
		for _, id := range memIDs {
			if !seen[id] {
				seen[id] = true
				uniqueIDs = append(uniqueIDs, id)
			}
		}
		memIDs = uniqueIDs
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
						"entity_name":   ename,
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
		"low_quality": lowQuality,
		"total_pairs": len(existingPairs),
		"relations":   createdRelations,
	}, nil
}

func (s *EvolutionService) GraphReasoning(userID uint) (map[string]interface{}, error) {
	var entities []models.Entity
	logDBErr("load entities for graph reasoning", s.db.Where("user_id = ?", userID).Limit(5000).Find(&entities).Error)

	var relations []models.Relation
	logDBErr("load relations for graph reasoning", s.db.Where("user_id = ?", userID).Limit(5000).Find(&relations).Error)

	adj := make(map[uint][]struct {
		target  uint
		relType string
		weight  float64
	})
	for _, r := range relations {
		adj[r.SourceID] = append(adj[r.SourceID], struct {
			target  uint
			relType string
			weight  float64
		}{target: r.TargetID, relType: r.RelationType, weight: r.Weight})
		adj[r.TargetID] = append(adj[r.TargetID], struct {
			target  uint
			relType string
			weight  float64
		}{target: r.SourceID, relType: r.RelationType, weight: r.Weight})
	}

	inferredRelations := []map[string]interface{}{}
	existingPairs := make(map[string]bool)
	for _, r := range relations {
		key := fmt.Sprintf("%d-%s-%d", r.SourceID, r.RelationType, r.TargetID)
		existingPairs[key] = true
	}

	entityName := make(map[uint]string)
	for _, e := range entities {
		entityName[e.ID] = e.Name
	}

	for _, e := range entities {
		neighbors := adj[e.ID]
		for i := 0; i < len(neighbors); i++ {
			for j := i + 1; j < len(neighbors); j++ {
				ni := neighbors[i]
				nj := neighbors[j]
				if ni.target == nj.target {
					continue
				}
				inferredType := "co_occurrence"
				if ni.relType == nj.relType {
					inferredType = "transitive_" + ni.relType
				}
				pairKey := fmt.Sprintf("%d-%s-%d", ni.target, inferredType, nj.target)
				if existingPairs[pairKey] {
					continue
				}
				inferredScore := ni.weight * nj.weight * 0.5
				if inferredScore < 0.2 {
					continue
				}
				inferredRelations = append(inferredRelations, map[string]interface{}{
					"source":        entityName[ni.target],
					"target":        entityName[nj.target],
					"via":           entityName[e.ID],
					"relation_type": inferredType,
					"confidence":    inferredScore,
				})
				existingPairs[pairKey] = true
			}
		}
	}

	if len(inferredRelations) > 50 {
		inferredRelations = inferredRelations[:50]
	}

	return map[string]interface{}{
		"mode":               "local",
		"inferred_relations": inferredRelations,
		"total_inferred":     len(inferredRelations),
		"entity_count":       len(entities),
		"relation_count":     len(relations),
	}, nil
}

func (s *EvolutionService) CentralityAnalysis(userID uint) (map[string]interface{}, error) {
	var entities []models.Entity
	logDBErr("load entities for centrality", s.db.Where("user_id = ?", userID).Limit(5000).Find(&entities).Error)

	var relations []models.Relation
	logDBErr("load relations for centrality", s.db.Where("user_id = ?", userID).Limit(5000).Find(&relations).Error)

	degree := make(map[uint]int)
	for _, r := range relations {
		degree[r.SourceID]++
		degree[r.TargetID]++
	}

	type entityCentrality struct {
		ID         uint    `json:"id"`
		Name       string  `json:"name"`
		EntityType string  `json:"entity_type"`
		Degree     int     `json:"degree"`
		Score      float64 `json:"score"`
	}

	entityMap := make(map[uint]models.Entity)
	for _, e := range entities {
		entityMap[e.ID] = e
	}

	maxDegree := 0
	for _, d := range degree {
		if d > maxDegree {
			maxDegree = d
		}
	}

	var centralities []entityCentrality
	for _, e := range entities {
		d := degree[e.ID]
		score := 0.0
		if maxDegree > 0 {
			score = float64(d) / float64(maxDegree)
		}
		score = score*0.6 + e.Confidence*0.4
		centralities = append(centralities, entityCentrality{
			ID:         e.ID,
			Name:       e.Name,
			EntityType: e.EntityType,
			Degree:     d,
			Score:      math.Round(score*1000) / 1000,
		})
	}

	sort.Slice(centralities, func(i, j int) bool {
		return centralities[i].Score > centralities[j].Score
	})

	top := centralities
	if len(top) > 20 {
		top = top[:20]
	}

	hubCount := 0
	for _, c := range centralities {
		if c.Degree >= 3 {
			hubCount++
		}
	}
	isolatedCount := 0
	for _, c := range centralities {
		if c.Degree == 0 {
			isolatedCount++
		}
	}

	return map[string]interface{}{
		"mode":            "local",
		"top_entities":    top,
		"hub_count":       hubCount,
		"isolated_count":  isolatedCount,
		"total_entities":  len(entities),
		"total_relations": len(relations),
		"avg_degree":      float64(len(relations)*2) / float64(max(len(entities), 1)),
	}, nil
}
