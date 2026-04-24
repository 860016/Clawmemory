package services

import (
	"math"
	"strings"

	"clawmemory/internal/models"

	"gorm.io/gorm"
)

type DedupService struct {
	db *gorm.DB
}

func NewDedupService(db *gorm.DB) *DedupService {
	return &DedupService{db: db}
}

type DuplicateGroup struct {
	Key       string               `json:"key"`
	Similarity float64             `json:"similarity"`
	Memories  []DuplicateMemory    `json:"memories"`
}

type DuplicateMemory struct {
	ID         uint    `json:"id"`
	Key        string  `json:"key"`
	Value      string  `json:"value"`
	Layer      string  `json:"layer"`
	Importance float64 `json:"importance"`
	Source     string  `json:"source"`
	CreatedAt  string  `json:"created_at"`
}

func (s *DedupService) Scan(userID uint) (map[string]interface{}, error) {
	var memories []models.Memory
	s.db.Where("user_id = ? AND status != ?", userID, "trashed").Find(&memories)

	if len(memories) < 2 {
		return map[string]interface{}{
			"duplicate_groups": []DuplicateGroup{},
			"total_duplicates": 0,
			"potential_savings": 0,
		}, nil
	}

	groups := []DuplicateGroup{}
	processed := make(map[uint]bool)

	for i := 0; i < len(memories); i++ {
		if processed[memories[i].ID] {
			continue
		}

		similarMemories := []DuplicateMemory{}
		for j := i + 1; j < len(memories); j++ {
			if processed[memories[j].ID] {
				continue
			}

			similarity := computeSimilarity(memories[i].Key, memories[i].Value, memories[j].Key, memories[j].Value)
			if similarity > 0.6 {
				if len(similarMemories) == 0 {
					similarMemories = append(similarMemories, DuplicateMemory{
						ID:         memories[i].ID,
						Key:        memories[i].Key,
						Value:      truncateString(memories[i].Value, 100),
						Layer:      memories[i].Layer,
						Importance: memories[i].Importance,
						Source:     memories[i].Source,
						CreatedAt:  memories[i].CreatedAt.Format("2006-01-02 15:04:05"),
					})
				}
				similarMemories = append(similarMemories, DuplicateMemory{
					ID:         memories[j].ID,
					Key:        memories[j].Key,
					Value:      truncateString(memories[j].Value, 100),
					Layer:      memories[j].Layer,
					Importance: memories[j].Importance,
					Source:     memories[j].Source,
					CreatedAt:  memories[j].CreatedAt.Format("2006-01-02 15:04:05"),
				})
				processed[memories[j].ID] = true
			}
		}

		if len(similarMemories) > 1 {
			processed[memories[i].ID] = true
			groups = append(groups, DuplicateGroup{
				Key:        memories[i].Key,
				Similarity: computeSimilarity(memories[i].Key, memories[i].Value, similarMemories[0].Key, similarMemories[0].Value),
				Memories:   similarMemories,
			})
		}
	}

	potentialSavings := 0
	for _, g := range groups {
		potentialSavings += len(g.Memories) - 1
	}

	return map[string]interface{}{
		"duplicate_groups":  groups,
		"total_duplicates":  potentialSavings,
		"potential_savings": potentialSavings,
		"total_memories":    len(memories),
		"dedup_rate":        float64(potentialSavings) / float64(len(memories)),
	}, nil
}

func (s *DedupService) Merge(userID uint, sourceID, targetID uint) (map[string]interface{}, error) {
	var source, target models.Memory
	if err := s.db.Where("id = ? AND user_id = ?", sourceID, userID).First(&source).Error; err != nil {
		return nil, err
	}
	if err := s.db.Where("id = ? AND user_id = ?", targetID, userID).First(&target).Error; err != nil {
		return nil, err
	}

	bestImportance := math.Max(source.Importance, target.Importance)
	bestValue := source.Value
	if len(target.Value) > len(source.Value) {
		bestValue = target.Value
	}

	s.db.Model(&target).Updates(map[string]interface{}{
		"importance": bestImportance,
		"value":      bestValue,
	})

	s.db.Model(&source).Update("status", "trashed")

	return map[string]interface{}{
		"merged_into":       targetID,
		"merged_from":       sourceID,
		"final_importance":  bestImportance,
		"message":           "memories merged successfully",
	}, nil
}

func computeSimilarity(key1, value1, key2, value2 string) float64 {
	keySim := jaccardSimilarity(key1, key2)
	valueSim := jaccardSimilarity(value1, value2)

	if keySim > 0.8 {
		return math.Min(keySim*0.6+valueSim*0.4, 1.0)
	}

	return keySim*0.4 + valueSim*0.6
}

func jaccardSimilarity(a, b string) float64 {
	setA := make(map[string]bool)
	setB := make(map[string]bool)

	for _, w := range tokenize(a) {
		setA[w] = true
	}
	for _, w := range tokenize(b) {
		setB[w] = true
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

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func init() {
	_ = strings.NewReader
	_ = math.Max
}
