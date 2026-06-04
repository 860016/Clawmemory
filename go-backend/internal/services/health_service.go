package services

import (
	"math"
	"strings"
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
	if err := s.db.Where("user_id = ? AND status != ?", userID, "trashed").Limit(5000).Find(&memories).Error; err != nil {
		return nil, err
	}

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

	expectedLayers := []string{"core", "context", "detail"}
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

type QualityIssue struct {
	MemoryID  uint   `json:"memory_id"`
	Key       string `json:"key"`
	IssueType string `json:"issue_type"`
	Severity  string `json:"severity"`
	Detail    string `json:"detail"`
	AutoFix   bool   `json:"auto_fix"`
	FixAction string `json:"fix_action"`
}

func (s *HealthService) AssessQuality(userID uint) (map[string]interface{}, error) {
	var memories []models.Memory
	if err := s.db.Where("user_id = ? AND status != ?", userID, "trashed").Limit(5000).Find(&memories).Error; err != nil {
		return nil, err
	}

	if len(memories) == 0 {
		return map[string]interface{}{
			"issues":       []QualityIssue{},
			"total":        0,
			"auto_fixable": 0,
		}, nil
	}

	issues := make([]QualityIssue, 0)
	keyCount := make(map[string]int)
	keyFirstID := make(map[string]uint)

	for _, m := range memories {
		keyCount[m.Key]++
		if _, exists := keyFirstID[m.Key]; !exists {
			keyFirstID[m.Key] = m.ID
		}

		if m.Value == "" {
			issues = append(issues, QualityIssue{
				MemoryID:  m.ID,
				Key:       m.Key,
				IssueType: "empty_value",
				Severity:  "high",
				Detail:    "memory value is empty",
				AutoFix:   false,
				FixAction: "delete_or_fill",
			})
		}

		if len(m.Value) > 0 && len(m.Value) < 5 {
			issues = append(issues, QualityIssue{
				MemoryID:  m.ID,
				Key:       m.Key,
				IssueType: "too_short",
				Severity:  "medium",
				Detail:    "memory value is too short to be useful",
				AutoFix:   false,
				FixAction: "enrich_or_merge",
			})
		}

		if m.Importance <= 0.05 && m.DecayStage >= 2 {
			issues = append(issues, QualityIssue{
				MemoryID:  m.ID,
				Key:       m.Key,
				IssueType: "near_death",
				Severity:  "low",
				Detail:    "memory is about to be trashed by decay",
				AutoFix:   true,
				FixAction: "trash",
			})
		}

		if m.Tags == "" && m.Layer == "detail" {
			issues = append(issues, QualityIssue{
				MemoryID:  m.ID,
				Key:       m.Key,
				IssueType: "untagged_detail",
				Severity:  "low",
				Detail:    "detail memory without tags makes search harder",
				AutoFix:   true,
				FixAction: "auto_tag",
			})
		}

		if m.Layer == "" {
			issues = append(issues, QualityIssue{
				MemoryID:  m.ID,
				Key:       m.Key,
				IssueType: "missing_layer",
				Severity:  "medium",
				Detail:    "memory has no layer assigned",
				AutoFix:   true,
				FixAction: "assign_layer",
			})
		}
	}

	for key, count := range keyCount {
		if count > 1 {
			for _, m := range memories {
				if m.Key == key {
					issues = append(issues, QualityIssue{
						MemoryID:  m.ID,
						Key:       key,
						IssueType: "duplicate_key",
						Severity:  "medium",
						Detail:    "duplicate key found, consider merging",
						AutoFix:   true,
						FixAction: "merge_duplicates",
					})
					break
				}
			}
		}
	}

	autoFixable := 0
	for _, issue := range issues {
		if issue.AutoFix {
			autoFixable++
		}
	}

	return map[string]interface{}{
		"issues":       issues,
		"total":        len(issues),
		"auto_fixable": autoFixable,
		"memory_count": len(memories),
	}, nil
}

func (s *HealthService) AutoFix(userID uint, issueTypes []string) (map[string]interface{}, error) {
	qualityResult, err := s.AssessQuality(userID)
	if err != nil {
		return nil, err
	}

	issues, ok := qualityResult["issues"].([]QualityIssue)
	if !ok {
		return map[string]interface{}{
			"fixed": 0,
			"total": 0,
		}, nil
	}

	typeFilter := make(map[string]bool)
	for _, t := range issueTypes {
		typeFilter[t] = true
	}

	fixed := 0
	failed := 0
	skipped := 0

	for _, issue := range issues {
		if !issue.AutoFix {
			skipped++
			continue
		}
		if len(typeFilter) > 0 && !typeFilter[issue.IssueType] {
			skipped++
			continue
		}

		switch issue.IssueType {
		case "near_death":
			result := s.db.Model(&models.Memory{}).Where("id = ? AND user_id = ?", issue.MemoryID, userID).
				Updates(map[string]interface{}{"status": "trashed", "decay_stage": 3})
			if result.Error == nil {
				fixed++
			} else {
				failed++
			}

		case "untagged_detail":
			var mem models.Memory
			if err := s.db.Where("id = ? AND user_id = ?", issue.MemoryID, userID).First(&mem).Error; err == nil {
				autoTags := extractAutoTags(mem.Key, mem.Value)
				if autoTags != "" {
					s.db.Model(&mem).Update("tags", autoTags)
					fixed++
				} else {
					skipped++
				}
			} else {
				failed++
			}

		case "missing_layer":
			var mem models.Memory
			if err := s.db.Where("id = ? AND user_id = ?", issue.MemoryID, userID).First(&mem).Error; err == nil {
				layer := inferLayer(mem)
				s.db.Model(&mem).Update("layer", layer)
				fixed++
			} else {
				failed++
			}

		case "duplicate_key":
			var duplicates []models.Memory
			_ = s.db.Where("user_id = ? AND key = ? AND status != ?", userID, issue.Key, "trashed").
				Order("importance DESC, updated_at DESC").Find(&duplicates).Error

			if len(duplicates) > 1 {
				dedupSvc := NewDedupService(s.db)
				for _, d := range duplicates[1:] {
					_, mergeErr := dedupSvc.Merge(userID, d.ID, duplicates[0].ID)
					if mergeErr != nil {
						failed++
						continue
					}
				}
				fixed++
			} else {
				skipped++
			}

		default:
			skipped++
		}
	}

	return map[string]interface{}{
		"fixed":   fixed,
		"failed":  failed,
		"skipped": skipped,
		"total":   len(issues),
	}, nil
}

func extractAutoTags(key, value string) string {
	tagSet := make(map[string]bool)

	keywords := []string{"api", "config", "database", "docker", "git", "kubernetes",
		"python", "javascript", "golang", "react", "vue", "node", "linux",
		"security", "performance", "testing", "deploy", "monitor", "cache",
		"auth", "jwt", "oauth", "ssl", "http", "grpc", "rest"}

	combined := strings.ToLower(key + " " + value)
	for _, kw := range keywords {
		if strings.Contains(combined, kw) {
			tagSet[kw] = true
		}
	}

	tags := make([]string, 0, len(tagSet))
	for t := range tagSet {
		tags = append(tags, t)
	}

	result := ""
	for i, t := range tags {
		if i > 0 {
			result += ","
		}
		result += t
	}
	return result
}

func inferLayer(mem models.Memory) string {
	key := strings.ToLower(mem.Key)
	value := strings.ToLower(mem.Value)

	coreKeywords := []string{"preference", "name", "role", "identity", "must", "always", "never", "critical"}
	for _, kw := range coreKeywords {
		if strings.Contains(key, kw) || strings.Contains(value, kw) {
			return "core"
		}
	}

	if mem.Importance >= 0.7 {
		return "core"
	}

	contextKeywords := []string{"project", "team", "architecture", "decision", "standard", "workflow"}
	for _, kw := range contextKeywords {
		if strings.Contains(key, kw) || strings.Contains(value, kw) {
			return "context"
		}
	}

	if mem.Importance >= 0.3 {
		return "context"
	}

	return "detail"
}
