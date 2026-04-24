package services

import (
	"math"
	"strings"

	"clawmemory/internal/models"

	"gorm.io/gorm"
)

type RecommendService struct {
	db *gorm.DB
}

func NewRecommendService(db *gorm.DB) *RecommendService {
	return &RecommendService{db: db}
}

func (s *RecommendService) RecommendForMemory(userID uint, memoryID uint, limit int) (map[string]interface{}, error) {
	var target models.Memory
	if err := s.db.Where("id = ? AND user_id = ?", memoryID, userID).First(&target).Error; err != nil {
		return nil, err
	}

	var memories []models.Memory
	s.db.Where("user_id = ? AND status != ? AND id != ?", userID, "trashed", memoryID).Find(&memories)

	targetTokens := tokenize(target.Key + " " + target.Value)
	if len(targetTokens) == 0 {
		return map[string]interface{}{"recommendations": []map[string]interface{}{}}, nil
	}

	type scored struct {
		Memory models.Memory
		Score  float64
		Reason string
	}

	scored_memories := make([]scored, 0, len(memories))

	for _, m := range memories {
		score := 0.0
		reason := ""

		if m.Layer == target.Layer {
			score += 0.3
			reason = "same_layer"
		}

		mTokens := tokenize(m.Key + " " + m.Value)
		similarity := tokenSimilarity(targetTokens, mTokens)
		score += similarity * 0.5

		if similarity > 0.3 && reason == "" {
			reason = "content_similar"
		}

		tagOverlap := tagOverlapScore(target.Tags, m.Tags)
		score += tagOverlap * 0.2

		if tagOverlap > 0 && reason == "" {
			reason = "shared_tags"
		}

		if reason == "" && score > 0.2 {
			reason = "related"
		}

		if score > 0.15 {
			scored_memories = append(scored_memories, scored{Memory: m, Score: score, Reason: reason})
		}
	}

	for i := 0; i < len(scored_memories)-1; i++ {
		for j := i + 1; j < len(scored_memories); j++ {
			if scored_memories[j].Score > scored_memories[i].Score {
				scored_memories[i], scored_memories[j] = scored_memories[j], scored_memories[i]
			}
		}
	}

	if limit > len(scored_memories) {
		limit = len(scored_memories)
	}

	recommendations := make([]map[string]interface{}, 0, limit)
	for i := 0; i < limit; i++ {
		m := scored_memories[i].Memory
		recommendations = append(recommendations, map[string]interface{}{
			"id":         m.ID,
			"key":        m.Key,
			"value":      truncateStr(m.Value, 150),
			"layer":      m.Layer,
			"importance": m.Importance,
			"score":      math.Round(scored_memories[i].Score*100) / 100,
			"reason":     scored_memories[i].Reason,
		})
	}

	return map[string]interface{}{
		"target_id":       memoryID,
		"recommendations": recommendations,
		"total":           len(scored_memories),
	}, nil
}

func (s *RecommendService) RecommendByContext(userID uint, context string, limit int) (map[string]interface{}, error) {
	var memories []models.Memory
	s.db.Where("user_id = ? AND status != ?", userID, "trashed").Find(&memories)

	if len(memories) == 0 || context == "" {
		return map[string]interface{}{"recommendations": []map[string]interface{}{}}, nil
	}

	contextTokens := tokenize(context)
	if len(contextTokens) == 0 {
		return map[string]interface{}{"recommendations": []map[string]interface{}{}}, nil
	}

	type scored struct {
		Memory models.Memory
		Score  float64
	}

	scored_memories := make([]scored, 0, len(memories))

	for _, m := range memories {
		mTokens := tokenize(m.Key + " " + m.Value)
		similarity := tokenSimilarity(contextTokens, mTokens)
		score := similarity*0.7 + m.Importance*0.3

		if score > 0.1 {
			scored_memories = append(scored_memories, scored{Memory: m, Score: score})
		}
	}

	for i := 0; i < len(scored_memories)-1; i++ {
		for j := i + 1; j < len(scored_memories); j++ {
			if scored_memories[j].Score > scored_memories[i].Score {
				scored_memories[i], scored_memories[j] = scored_memories[j], scored_memories[i]
			}
		}
	}

	if limit > len(scored_memories) {
		limit = len(scored_memories)
	}

	recommendations := make([]map[string]interface{}, 0, limit)
	for i := 0; i < limit; i++ {
		m := scored_memories[i].Memory
		recommendations = append(recommendations, map[string]interface{}{
			"id":         m.ID,
			"key":        m.Key,
			"value":      truncateStr(m.Value, 150),
			"layer":      m.Layer,
			"importance": m.Importance,
			"score":      math.Round(scored_memories[i].Score*100) / 100,
		})
	}

	return map[string]interface{}{
		"context":         context,
		"recommendations": recommendations,
		"total":           len(scored_memories),
	}, nil
}

func tokenSimilarity(a, b []string) float64 {
	setA := make(map[string]bool)
	setB := make(map[string]bool)
	for _, t := range a {
		setA[t] = true
	}
	for _, t := range b {
		setB[t] = true
	}

	if len(setA) == 0 && len(setB) == 0 {
		return 1.0
	}

	intersection := 0
	for w := range setA {
		if setB[w] {
			intersection++
		}
	}

	union := len(setA) + len(setB) - intersection
	if union == 0 {
		return 0
	}

	return float64(intersection) / float64(union)
}

func tagOverlapScore(tagsA, tagsB string) float64 {
	if tagsA == "" || tagsB == "" || tagsA == "[]" || tagsB == "[]" {
		return 0
	}

	setA := parseTags(tagsA)
	setB := parseTags(tagsB)

	if len(setA) == 0 || len(setB) == 0 {
		return 0
	}

	intersection := 0
	for t := range setA {
		if setB[t] {
			intersection++
		}
	}

	return float64(intersection) / math.Max(float64(len(setA)), float64(len(setB)))
}

func parseTags(tags string) map[string]bool {
	result := make(map[string]bool)
	cleaned := strings.Trim(tags, "[]\"")
	for _, t := range strings.Split(cleaned, ",") {
		t = strings.TrimSpace(strings.Trim(t, "\""))
		if t != "" {
			result[t] = true
		}
	}
	return result
}

func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
