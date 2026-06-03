package services

import (
	"fmt"
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
	Key        string            `json:"key"`
	Similarity float64           `json:"similarity"`
	Memories   []DuplicateMemory `json:"memories"`
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
	if err := s.db.Where("user_id = ? AND status != ?", userID, "trashed").
		Order("importance DESC").Limit(1000).Find(&memories).Error; err != nil {
		return nil, fmt.Errorf("failed to load memories: %w", err)
	}

	if len(memories) < 2 {
		return map[string]interface{}{
			"duplicate_groups":  []DuplicateGroup{},
			"total_duplicates":  0,
			"potential_savings": 0,
		}, nil
	}

	groups := []DuplicateGroup{}
	processed := make(map[uint]bool)

	buckets := s.bucketByPrefix(memories)

	for _, bucket := range buckets {
		for i := 0; i < len(bucket); i++ {
			mi := bucket[i]
			if processed[mi.ID] {
				continue
			}

			similarMemories := []DuplicateMemory{}
			for j := i + 1; j < len(bucket); j++ {
				mj := bucket[j]
				if processed[mj.ID] {
					continue
				}

				similarity := computeSimilarity(mi.Key, mi.Value, mj.Key, mj.Value)
				if similarity > 0.6 {
					if len(similarMemories) == 0 {
						similarMemories = append(similarMemories, DuplicateMemory{
							ID:         mi.ID,
							Key:        mi.Key,
							Value:      truncateString(mi.Value, 100),
							Layer:      mi.Layer,
							Importance: mi.Importance,
							Source:     mi.Source,
							CreatedAt:  mi.CreatedAt.Format("2006-01-02 15:04:05"),
						})
					}
					similarMemories = append(similarMemories, DuplicateMemory{
						ID:         mj.ID,
						Key:        mj.Key,
						Value:      truncateString(mj.Value, 100),
						Layer:      mj.Layer,
						Importance: mj.Importance,
						Source:     mj.Source,
						CreatedAt:  mj.CreatedAt.Format("2006-01-02 15:04:05"),
					})
					processed[mj.ID] = true
				}
			}

			if len(similarMemories) > 1 {
				processed[mi.ID] = true
				groups = append(groups, DuplicateGroup{
					Key:        mi.Key,
					Similarity: computeSimilarity(mi.Key, mi.Value, similarMemories[0].Key, similarMemories[0].Value),
					Memories:   similarMemories,
				})
			}
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
	if sourceID == targetID {
		return nil, fmt.Errorf("source and target cannot be the same memory")
	}

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
		"merged_into":      targetID,
		"merged_from":      sourceID,
		"final_importance": bestImportance,
		"message":          "memories merged successfully",
	}, nil
}

func (s *DedupService) AutoMergeSimilar(userID uint, threshold float64) (int, error) {
	var memories []models.Memory
	if err := s.db.Where("user_id = ? AND status = ?", userID, "active").
		Order("importance DESC").Limit(1000).Find(&memories).Error; err != nil {
		return 0, err
	}

	if len(memories) < 2 {
		return 0, nil
	}

	merged := 0
	processed := make(map[uint]bool)

	buckets := s.bucketByPrefix(memories)

	for _, bucket := range buckets {
		for i := 0; i < len(bucket); i++ {
			if processed[bucket[i].ID] {
				continue
			}
			for j := i + 1; j < len(bucket); j++ {
				if processed[bucket[j].ID] {
					continue
				}

				similarity := computeSimilarity(bucket[i].Key, bucket[i].Value, bucket[j].Key, bucket[j].Value)
				if similarity >= threshold {
					target := &bucket[i]
					source := &bucket[j]

					bestImportance := target.Importance
					if source.Importance > bestImportance {
						bestImportance = source.Importance
					}
					bestValue := target.Value
					if len(source.Value) > len(target.Value) {
						bestValue = source.Value
					}

					s.db.Model(target).Updates(map[string]interface{}{
						"importance": bestImportance,
						"value":      bestValue,
					})
					s.db.Model(source).Update("status", "trashed")

					processed[source.ID] = true
					merged++
				}
			}
		}
	}

	return merged, nil
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

func (s *DedupService) bucketByPrefix(memories []models.Memory) [][]models.Memory {
	buckets := make(map[string][]models.Memory)
	for _, m := range memories {
		prefix := keyPrefix(m.Key)
		buckets[prefix] = append(buckets[prefix], m)
	}

	// 也把没有匹配到其他桶的记忆放入一个跨桶组，防止漏检
	// 小桶（<=10个）合并到一起做交叉比较
	var smallBucket []models.Memory
	var result [][]models.Memory
	for _, bucket := range buckets {
		if len(bucket) <= 10 {
			smallBucket = append(smallBucket, bucket...)
		} else {
			result = append(result, bucket)
		}
	}
	if len(smallBucket) > 0 {
		result = append(result, smallBucket)
	}

	return result
}

func keyPrefix(key string) string {
	// 取 key 的第一个单词或前缀作为桶标识
	parts := strings.SplitN(strings.ToLower(strings.TrimSpace(key)), " ", 2)
	if len(parts) == 0 || parts[0] == "" {
		return "_empty"
	}
	// 取前4个字符作为前缀，避免桶太碎
	p := parts[0]
	if len(p) > 4 {
		p = p[:4]
	}
	return p
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
