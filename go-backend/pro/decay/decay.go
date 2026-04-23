// +build pro

package decay

import (
	"math"
	"time"
)

// CalculateDecay 计算记忆衰减值（Pro 功能）
func CalculateDecay(ageHours float64, importance float64, accessCount int) float64 {
	if importance <= 0 {
		return 1.0
	}

	// Pro 版使用更精细的算法
	timeFactor := math.Log1p(ageHours) / 10.0
	accessFactor := math.Log1p(float64(accessCount)) / 5.0
	importanceFactor := 1.1 - math.Min(importance, 1.0)

	// 加入时间衰减的非线性调整
	if ageHours > 168 { // 超过一周
		timeFactor *= 1.5
	}
	if ageHours > 720 { // 超过一个月
		timeFactor *= 2.0
	}

	decay := timeFactor*importanceFactor - accessFactor
	return math.Max(0.0, math.Min(1.0, decay))
}

// DecayMemory 对记忆进行衰减计算
type Memory struct {
	ID            uint      `json:"id"`
	Importance    float64   `json:"importance"`
	AccessCount   int       `json:"access_count"`
	CreatedAt     time.Time `json:"created_at"`
	DecayValue    float64   `json:"decay_value"`
	ShouldPrune   bool      `json:"should_prune"`
}

func DecayMemory(memory *Memory) *Memory {
	ageHours := time.Since(memory.CreatedAt).Hours()
	decayValue := CalculateDecay(ageHours, memory.Importance, memory.AccessCount)
	memory.DecayValue = decayValue
	memory.ShouldPrune = decayValue > 0.8
	return memory
}

// BatchDecay 批量衰减计算
func BatchDecay(memories []*Memory) []*Memory {
	results := make([]*Memory, len(memories))
	for i, m := range memories {
		results[i] = DecayMemory(m)
	}
	return results
}

// GetStageInfo 获取衰减阶段信息
func GetStageInfo(decayValue float64) map[string]interface{} {
	stage := 0
	label := "fresh"
	description := "记忆新鲜"

	switch {
	case decayValue < 0.2:
		stage = 0
		label = "fresh"
		description = "记忆新鲜"
	case decayValue < 0.4:
		stage = 1
		label = "stable"
		description = "记忆稳定"
	case decayValue < 0.6:
		stage = 2
		label = "fading"
		description = "开始模糊"
	case decayValue < 0.8:
		stage = 3
		label = "weak"
		description = "记忆薄弱"
	default:
		stage = 4
		label = "prune"
		description = "建议清理"
	}

	return map[string]interface{}{
		"stage":       stage,
		"label":       label,
		"description": description,
		"decay_value": decayValue,
	}
}
