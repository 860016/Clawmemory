// +build pro

package conflict

import (
	"strings"
)

// ConflictResult 冲突检测结果
type ConflictResult struct {
	HasConflict bool     `json:"has_conflict"`
	Similarity  float64  `json:"similarity"`
	CommonWords []string `json:"common_words"`
}

// DetectConflict 检测两个记忆是否冲突
func DetectConflict(content1, content2 string) *ConflictResult {
	if content1 == "" || content2 == "" {
		return &ConflictResult{HasConflict: false, Similarity: 0}
	}

	// 分词（简化版）
	words1 := tokenize(content1)
	words2 := tokenize(content2)

	if len(words1) == 0 || len(words2) == 0 {
		return &ConflictResult{HasConflict: false, Similarity: 0}
	}

	// 计算相似度
	intersection := intersect(words1, words2)
	union := union(words1, words2)

	similarity := float64(len(intersection)) / float64(len(union))

	// 相似度高但内容不同可能是冲突
	hasConflict := similarity > 0.5 && content1 != content2

	return &ConflictResult{
		HasConflict: hasConflict,
		Similarity:  similarity,
		CommonWords: intersection,
	}
}

// ResolveConflict 尝试合并冲突的记忆
func ResolveConflict(memory1, memory2 map[string]interface{}) map[string]interface{} {
	content1 := getString(memory1, "content", "")
	content2 := getString(memory2, "content", "")

	conflict := DetectConflict(content1, content2)
	if !conflict.HasConflict {
		return map[string]interface{}{
			"resolved": true,
			"merged":   memory1,
		}
	}

	// 保留更重要的记忆
	importance1 := getFloat(memory1, "importance", 0.5)
	importance2 := getFloat(memory2, "importance", 0.5)

	var merged map[string]interface{}
	var strategy string

	if importance1 >= importance2 {
		merged = copyMap(memory1)
		merged["conflict_resolved"] = true
		merged["merged_with"] = memory2["id"]
		strategy = "keep_higher_importance"
	} else {
		merged = copyMap(memory2)
		merged["conflict_resolved"] = true
		merged["merged_with"] = memory1["id"]
		strategy = "keep_higher_importance"
	}

	return map[string]interface{}{
		"resolved": true,
		"merged":   merged,
		"strategy": strategy,
	}
}

// ScanConflicts 扫描所有冲突
func ScanConflicts(memories []map[string]interface{}) []map[string]interface{} {
	conflicts := []map[string]interface{}{}

	for i := 0; i < len(memories); i++ {
		for j := i + 1; j < len(memories); j++ {
			content1 := getString(memories[i], "content", "")
			content2 := getString(memories[j], "content", "")

			result := DetectConflict(content1, content2)
			if result.HasConflict {
				conflicts = append(conflicts, map[string]interface{}{
					"memory1_id": memories[i]["id"],
					"memory2_id": memories[j]["id"],
					"similarity": result.Similarity,
				})
			}
		}
	}

	return conflicts
}

// Helper functions
func tokenize(text string) []string {
	words := strings.Fields(strings.ToLower(text))
	// 去重
	seen := make(map[string]bool)
	result := []string{}
	for _, w := range words {
		if !seen[w] {
			seen[w] = true
			result = append(result, w)
		}
	}
	return result
}

func intersect(a, b []string) []string {
	set := make(map[string]bool)
	for _, v := range b {
		set[v] = true
	}
	result := []string{}
	for _, v := range a {
		if set[v] {
			result = append(result, v)
		}
	}
	return result
}

func union(a, b []string) []string {
	set := make(map[string]bool)
	for _, v := range a {
		set[v] = true
	}
	for _, v := range b {
		set[v] = true
	}
	result := []string{}
	for k := range set {
		result = append(result, k)
	}
	return result
}

func getString(m map[string]interface{}, key, defaultValue string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return defaultValue
}

func getFloat(m map[string]interface{}, key string, defaultValue float64) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	return defaultValue
}

func copyMap(m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		result[k] = v
	}
	return result
}
