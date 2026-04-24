package services

import (
	"math"
	"time"

	"clawmemory/internal/models"

	"gorm.io/gorm"
)

type HealthService struct {
	db *gorm.DB
}

func NewHealthService(db *gorm.DB) *HealthService {
	return &HealthService{db: db}
}

func (s *HealthService) GetHealthScore(userID uint) (map[string]interface{}, error) {
	var memories []models.Memory
	s.db.Where("user_id = ? AND status != ?", userID, "trashed").Find(&memories)

	if len(memories) == 0 {
		return map[string]interface{}{
			"overall_score": 100,
			"grade":         "A+",
			"dimensions":    map[string]interface{}{},
			"suggestions":   []string{},
		}, nil
	}

	dedupScore := s.computeDedupScore(memories)
	freshnessScore := s.computeFreshnessScore(memories)
	coverageScore := s.computeCoverageScore(memories)
	qualityScore := s.computeQualityScore(memories)
	balanceScore := s.computeBalanceScore(memories)

	overallScore := dedupScore*0.25 + freshnessScore*0.25 + coverageScore*0.2 + qualityScore*0.15 + balanceScore*0.15
	overallScore = math.Round(overallScore*100) / 100

	grade := "A+"
	switch {
	case overallScore >= 90:
		grade = "A+"
	case overallScore >= 80:
		grade = "A"
	case overallScore >= 70:
		grade = "B+"
	case overallScore >= 60:
		grade = "B"
	case overallScore >= 50:
		grade = "C"
	default:
		grade = "D"
	}

	suggestions := s.generateSuggestions(dedupScore, freshnessScore, coverageScore, qualityScore, balanceScore, memories)

	return map[string]interface{}{
		"overall_score": overallScore,
		"grade":         grade,
		"dimensions": map[string]interface{}{
			"dedup":     map[string]interface{}{"score": math.Round(dedupScore*100) / 100, "label": "去重率"},
			"freshness": map[string]interface{}{"score": math.Round(freshnessScore*100) / 100, "label": "时效性"},
			"coverage":  map[string]interface{}{"score": math.Round(coverageScore*100) / 100, "label": "覆盖率"},
			"quality":   map[string]interface{}{"score": math.Round(qualityScore*100) / 100, "label": "质量"},
			"balance":   map[string]interface{}{"score": math.Round(balanceScore*100) / 100, "label": "均衡性"},
		},
		"total_memories": len(memories),
		"suggestions":    suggestions,
	}, nil
}

func (s *HealthService) computeDedupScore(memories []models.Memory) float64 {
	keyCount := make(map[string]int)
	for _, m := range memories {
		keyCount[m.Key]++
	}

	duplicates := 0
	for _, count := range keyCount {
		if count > 1 {
			duplicates += count - 1
		}
	}

	if len(memories) == 0 {
		return 100
	}

	dupRate := float64(duplicates) / float64(len(memories))
	return math.Max(0, 100-dupRate*200)
}

func (s *HealthService) computeFreshnessScore(memories []models.Memory) float64 {
	now := time.Now()
	recentCount := 0
	for _, m := range memories {
		daysSinceUpdate := now.Sub(m.UpdatedAt).Hours() / 24
		if daysSinceUpdate < 30 {
			recentCount++
		}
	}

	if len(memories) == 0 {
		return 100
	}

	freshRate := float64(recentCount) / float64(len(memories))
	return freshRate * 100
}

func (s *HealthService) computeCoverageScore(memories []models.Memory) float64 {
	layerCount := make(map[string]int)
	for _, m := range memories {
		layerCount[m.Layer]++
	}

	expectedLayers := []string{"preference", "knowledge", "short_term"}
	covered := 0
	for _, layer := range expectedLayers {
		if layerCount[layer] > 0 {
			covered++
		}
	}

	return float64(covered) / float64(len(expectedLayers)) * 100
}

func (s *HealthService) computeQualityScore(memories []models.Memory) float64 {
	if len(memories) == 0 {
		return 100
	}

	totalImportance := 0.0
	emptyValueCount := 0
	for _, m := range memories {
		totalImportance += m.Importance
		if m.Value == "" {
			emptyValueCount++
		}
	}

	avgImportance := totalImportance / float64(len(memories))
	emptyRate := float64(emptyValueCount) / float64(len(memories))

	importanceScore := avgImportance * 80
	emptyPenalty := emptyRate * 50

	return math.Max(0, math.Min(100, importanceScore+20-emptyPenalty))
}

func (s *HealthService) computeBalanceScore(memories []models.Memory) float64 {
	layerCount := make(map[string]int)
	for _, m := range memories {
		layerCount[m.Layer]++
	}

	if len(layerCount) <= 1 {
		return 30
	}

	counts := make([]float64, 0, len(layerCount))
	for _, count := range layerCount {
		counts = append(counts, float64(count))
	}

	mean := 0.0
	for _, c := range counts {
		mean += c
	}
	mean /= float64(len(counts))

	variance := 0.0
	for _, c := range counts {
		variance += (c - mean) * (c - mean)
	}
	variance /= float64(len(counts))

	stdDev := math.Sqrt(variance)
	cv := stdDev / mean

	balanceScore := math.Max(0, 100-cv*100)
	return math.Min(100, balanceScore)
}

func (s *HealthService) generateSuggestions(dedup, freshness, coverage, quality, balance float64, memories []models.Memory) []string {
	suggestions := []string{}

	if dedup < 70 {
		suggestions = append(suggestions, "发现较多重复记忆，建议使用去重功能合并相似记忆")
	}
	if freshness < 50 {
		suggestions = append(suggestions, "记忆更新频率较低，建议定期更新和清理过时记忆")
	}
	if coverage < 60 {
		suggestions = append(suggestions, "记忆层级覆盖不完整，建议补充各层级记忆数据")
	}
	if quality < 60 {
		suggestions = append(suggestions, "部分记忆质量较低，建议补充内容或调整重要性")
	}
	if balance < 50 {
		suggestions = append(suggestions, "记忆分布不均衡，某些层级过多或过少")
	}

	lowImportance := 0
	for _, m := range memories {
		if m.Importance < 0.2 {
			lowImportance++
		}
	}
	if lowImportance > len(memories)/5 {
		suggestions = append(suggestions, "低重要性记忆较多，建议清理或提升重要性")
	}

	if len(suggestions) == 0 {
		suggestions = append(suggestions, "记忆库状态良好，继续保持！")
	}

	return suggestions
}
