// +build pro

package router

import (
	"regexp"
	"strings"
)

// ComplexityResult 复杂度分析结果
type ComplexityResult struct {
	Tokens     int    `json:"tokens"`
	Complexity string `json:"complexity"`
	Score      int    `json:"score"`
}

// RouteResult 路由结果
type RouteResult struct {
	Model  string `json:"model"`
	Reason string `json:"reason"`
}

// EstimateComplexity 估计文本复杂度
func EstimateComplexity(text string) *ComplexityResult {
	if text == "" {
		return &ComplexityResult{Tokens: 0, Complexity: "low", Score: 1}
	}

	// 统计中文字符和英文单词
	chineseChars := len(regexp.MustCompile(`[\u4e00-\u9fff]`).FindAllString(text, -1))
	englishWords := len(regexp.MustCompile(`[a-zA-Z]+`).FindAllString(text, -1))
	tokens := chineseChars + englishWords

	var complexity string
	var score int

	switch {
	case tokens < 100:
		complexity = "low"
		score = 1
	case tokens < 500:
		complexity = "medium"
		score = 2
	case tokens < 2000:
		complexity = "high"
		score = 3
	default:
		complexity = "very_high"
		score = 4
	}

	return &ComplexityResult{
		Tokens:     tokens,
		Complexity: complexity,
		Score:      score,
	}
}

// RouteModel 根据文本复杂度路由到合适的模型
func RouteModel(text string, availableModels []string) *RouteResult {
	comp := EstimateComplexity(text)

	if availableModels == nil {
		availableModels = []string{"gpt-3.5-turbo", "gpt-4", "gpt-4-turbo"}
	}

	var model string
	switch comp.Score {
	case 1:
		model = availableModels[0]
	case 2:
		if len(availableModels) > 1 {
			model = availableModels[1]
		} else {
			model = availableModels[0]
		}
	case 3, 4:
		if len(availableModels) > 0 {
			model = availableModels[len(availableModels)-1]
		}
	}

	return &RouteResult{
		Model:  model,
		Reason: "Complexity: " + comp.Complexity + " (score: " + string(rune('0'+comp.Score)) + ")",
	}
}

// BatchRoute 批量路由
func BatchRoute(texts []string, availableModels []string) []*RouteResult {
	results := make([]*RouteResult, len(texts))
	for i, text := range texts {
		results[i] = RouteModel(text, availableModels)
	}
	return results
}

// GetTokenStats 获取 Token 统计
func GetTokenStats(texts []string) map[string]interface{} {
	totalTokens := 0
	complexityDistribution := map[string]int{
		"low":       0,
		"medium":    0,
		"high":      0,
		"very_high": 0,
	}

	for _, text := range texts {
		comp := EstimateComplexity(text)
		totalTokens += comp.Tokens
		complexityDistribution[comp.Complexity]++
	}

	avgTokens := 0
	if len(texts) > 0 {
		avgTokens = totalTokens / len(texts)
	}

	return map[string]interface{}{
		"total_texts":           len(texts),
		"total_tokens":          totalTokens,
		"average_tokens":        avgTokens,
		"complexity_distribution": complexityDistribution,
	}
}
