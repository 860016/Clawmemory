package services

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"clawmemory/internal/models"

	"gorm.io/gorm"
)

type scoredMemory struct {
	Memory models.Memory
	Score  float64
}

type SmartLoadService struct {
	db *gorm.DB
}

func NewSmartLoadService(db *gorm.DB) *SmartLoadService {
	return &SmartLoadService{db: db}
}

type MemoryLoadLevel string

const (
	LoadLevelSummary  MemoryLoadLevel = "summary"
	LoadLevelStandard MemoryLoadLevel = "standard"
	LoadLevelFull     MemoryLoadLevel = "full"
)

type LoadedMemory struct {
	ID                  uint            `json:"id"`
	Key                 string          `json:"key"`
	Value               string          `json:"value,omitempty"`
	Summary             string          `json:"summary"`
	Layer               string          `json:"layer"`
	MemoryType          string          `json:"memory_type"`
	Importance          float64         `json:"importance"`
	Tags                []string        `json:"tags"`
	Source              string          `json:"source"`
	Score               float64         `json:"score"`
	LoadLevel           MemoryLoadLevel `json:"load_level"`
	RelatedIDs          []uint          `json:"related_ids,omitempty"`
	EstimatedTokens     int             `json:"estimated_tokens"`
	Freshness           string          `json:"freshness"`
	FreshnessWarning    string          `json:"freshness_warning,omitempty"`
	VerifiedAt          string          `json:"verified_at,omitempty"`
	VerificationWarning string          `json:"verification_warning,omitempty"`
	CreatedAt           string          `json:"created_at"`
}

type SmartLoadResult struct {
	Memories    []LoadedMemory `json:"memories"`
	TotalTokens int            `json:"total_tokens"`
	TokenBudget int            `json:"token_budget"`
	LoadLevel   string         `json:"load_level"`
	Engine      string         `json:"engine"`
	Suggestions []string       `json:"suggestions,omitempty"`
}

func (s *SmartLoadService) SmartLoad(userID uint, query string, tokenBudget int, loadLevel string) (*SmartLoadResult, error) {
	if tokenBudget <= 0 {
		tokenBudget = 2000
	}

	var memories []models.Memory
	_ = s.db.Where("user_id = ? AND status = ?", userID, "active").Limit(5000).Find(&memories).Error

	if len(memories) == 0 {
		return &SmartLoadResult{
			Memories:    []LoadedMemory{},
			TotalTokens: 0,
			TokenBudget: tokenBudget,
			LoadLevel:   loadLevel,
			Engine:      "smart_v1",
		}, nil
	}

	scored := make([]scoredMemory, 0, len(memories))
	for _, m := range memories {
		score := s.computeRelevanceScore(m, query)
		if score > 0.01 {
			scored = append(scored, scoredMemory{Memory: m, Score: score})
		}
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	loaded := s.allocateByBudget(scored, tokenBudget, loadLevel)

	totalTokens := 0
	for i := range loaded {
		totalTokens += loaded[i].EstimatedTokens
	}

	suggestions := s.generateSuggestions(loaded, query, tokenBudget)

	return &SmartLoadResult{
		Memories:    loaded,
		TotalTokens: totalTokens,
		TokenBudget: tokenBudget,
		LoadLevel:   loadLevel,
		Engine:      "smart_v1",
		Suggestions: suggestions,
	}, nil
}

func (s *SmartLoadService) computeRelevanceScore(m models.Memory, query string) float64 {
	score := 0.0

	score += m.Importance * 0.25

	accessBoost := math.Min(float64(m.AccessCount)/50.0, 1.0) * 0.15
	score += accessBoost

	reinforceBoost := math.Min(float64(m.ReinforceCount)/10.0, 1.0) * 0.1
	score += reinforceBoost

	layerWeights := map[string]float64{
		"preference": 0.15,
		"knowledge":  0.12,
		"short_term": 0.05,
		"private":    0.10,
	}
	if w, ok := layerWeights[m.Layer]; ok {
		score += w
	}

	if query != "" {
		queryTokens := tokenize(query)
		memTokens := tokenize(m.Key + " " + m.Value)
		similarity := tokenSimilarity(queryTokens, memTokens)
		score += similarity * 0.35

		keyTokens := tokenize(m.Key)
		keyMatch := tokenSimilarity(queryTokens, keyTokens)
		score += keyMatch * 0.15
	}

	if m.LastAccessedAt != nil {
		daysSince := time.Since(*m.LastAccessedAt).Hours() / 24
		recencyBoost := math.Max(0, 1.0-daysSince/30.0) * 0.1
		score += recencyBoost
	}

	if m.DecayStage > 0 {
		score *= (1.0 - float64(m.DecayStage)*0.15)
	}

	return score
}

func (s *SmartLoadService) allocateByBudget(scored []scoredMemory, tokenBudget int, loadLevel string) []LoadedMemory {
	result := make([]LoadedMemory, 0)
	remainingBudget := tokenBudget

	for _, sm := range scored {
		m := sm.Memory
		summary := s.getOrGenerateSummary(m)

		var level MemoryLoadLevel
		switch loadLevel {
		case "summary":
			level = LoadLevelSummary
		case "full":
			level = LoadLevelFull
		default:
			if sm.Score >= 0.6 {
				level = LoadLevelFull
			} else if sm.Score >= 0.3 {
				level = LoadLevelStandard
			} else {
				level = LoadLevelSummary
			}
		}

		var value string
		estTokens := estimateTokenCount(m.Key)
		if level == LoadLevelFull || level == LoadLevelStandard {
			value = m.Value
			estTokens += estimateTokenCount(m.Value)
		} else {
			estTokens += estimateTokenCount(summary)
		}

		freshness, freshnessWarning := computeFreshness(m.UpdatedAt)
		verifiedAt, verificationWarning := computeVerification(m)

		mem := LoadedMemory{
			ID:                  m.ID,
			Key:                 m.Key,
			Value:               value,
			Summary:             summary,
			Layer:               m.Layer,
			MemoryType:          m.MemoryType,
			Importance:          m.Importance,
			Tags:                parseTagsSlice(m.Tags),
			Source:              m.Source,
			Score:               math.Round(sm.Score*1000) / 1000,
			LoadLevel:           level,
			EstimatedTokens:     estTokens,
			Freshness:           freshness,
			FreshnessWarning:    freshnessWarning,
			VerifiedAt:          verifiedAt,
			VerificationWarning: verificationWarning,
			CreatedAt:           m.CreatedAt.Format("2006-01-02 15:04:05"),
		}

		if remainingBudget-mem.EstimatedTokens < 0 && len(result) > 0 {
			break
		}
		remainingBudget -= mem.EstimatedTokens
		result = append(result, mem)
	}

	return result
}

func computeFreshness(updatedAt time.Time) (string, string) {
	days := time.Since(updatedAt).Hours() / 24

	switch {
	case days <= 1:
		return "fresh", ""
	case days <= 7:
		return "recent", ""
	case days <= 30:
		return "aging", fmt.Sprintf("This memory is %d days old. Verify before relying on it.", int(days))
	default:
		return "stale", fmt.Sprintf("This memory is %d days old and may be outdated. Memories are point-in-time observations, not live state. Verify against current context before asserting as fact.", int(days))
	}
}

func computeVerification(m models.Memory) (string, string) {
	if m.VerifiedAt == nil {
		if time.Since(m.CreatedAt).Hours()/24 > 7 {
			return "", "This memory has never been verified. Consider verifying before use."
		}
		return "", ""
	}

	daysSinceVerify := time.Since(*m.VerifiedAt).Hours() / 24
	if daysSinceVerify > 30 {
		return m.VerifiedAt.Format("2006-01-02"), fmt.Sprintf("Last verified %d days ago. Content may have changed since verification.", int(daysSinceVerify))
	}
	return m.VerifiedAt.Format("2006-01-02"), ""
}

func (s *SmartLoadService) getOrGenerateSummary(m models.Memory) string {
	if m.Summary != "" {
		return m.Summary
	}
	return generateSummary(m.Key, m.Value)
}

func generateSummary(key, value string) string {
	if len(value) <= 80 {
		return value
	}

	sentences := splitSentences(value)
	if len(sentences) <= 1 {
		if len(value) > 80 {
			return value[:77] + "..."
		}
		return value
	}

	var sb strings.Builder
	sb.WriteString(sentences[0])
	for i := 1; i < len(sentences) && sb.Len() < 70; i++ {
		sb.WriteString(" ")
		sb.WriteString(sentences[i])
	}

	result := sb.String()
	if len(result) > 80 {
		result = result[:77] + "..."
	}
	return result
}

func splitSentences(text string) []string {
	var sentences []string
	var current strings.Builder

	for _, r := range text {
		current.WriteRune(r)
		if r == '.' || r == '。' || r == '!' || r == '！' || r == '?' || r == '？' || r == '\n' {
			s := strings.TrimSpace(current.String())
			if s != "" {
				sentences = append(sentences, s)
			}
			current.Reset()
		}
	}

	if current.Len() > 0 {
		s := strings.TrimSpace(current.String())
		if s != "" {
			sentences = append(sentences, s)
		}
	}

	if len(sentences) == 0 {
		return []string{text}
	}
	return sentences
}

func estimateTokenCount(text string) int {
	cjk := 0
	ascii := 0
	for _, r := range text {
		if r >= 0x4E00 && r <= 0x9FFF || r >= 0x3040 && r <= 0x30FF || r >= 0xAC00 && r <= 0xD7AF {
			cjk++
		} else if r <= 127 {
			ascii++
		} else {
			cjk++
		}
	}
	return cjk*2 + ascii/4 + 1
}

func parseTagsSlice(tags string) []string {
	if tags == "" || tags == "[]" {
		return []string{}
	}
	var result []string
	cleaned := strings.Trim(tags, "[]\"")
	for _, t := range strings.Split(cleaned, ",") {
		t = strings.TrimSpace(strings.Trim(t, "\""))
		if t != "" {
			result = append(result, t)
		}
	}
	if result == nil {
		result = []string{}
	}
	return result
}

func (s *SmartLoadService) generateSuggestions(loaded []LoadedMemory, query string, budget int) []string {
	var suggestions []string

	summaryCount := 0
	standardCount := 0
	fullCount := 0
	for _, m := range loaded {
		switch m.LoadLevel {
		case LoadLevelSummary:
			summaryCount++
		case LoadLevelStandard:
			standardCount++
		case LoadLevelFull:
			fullCount++
		}
	}

	if summaryCount > 0 {
		suggestions = append(suggestions, fmt.Sprintf("%d memories shown as summaries (save ~%d tokens). Use load_level=standard or full to see more detail.", summaryCount, summaryCount*50))
	}
	if fullCount > 0 {
		suggestions = append(suggestions, fmt.Sprintf("%d high-relevance memories loaded in full detail.", fullCount))
	}
	if len(loaded) > 0 && loaded[len(loaded)-1].Score > 0.1 {
		suggestions = append(suggestions, "More memories available. Increase token_budget to load additional results.")
	}
	if query == "" {
		suggestions = append(suggestions, "Provide a query for more targeted memory selection.")
	}

	return suggestions
}

func (s *SmartLoadService) ReinforceMemory(userID uint, memoryID uint) error {
	var memory models.Memory
	if err := s.db.Where("user_id = ? AND id = ?", userID, memoryID).First(&memory).Error; err != nil {
		return err
	}

	newImportance := math.Min(memory.Importance+0.05, 1.0)
	newReinforceCount := memory.ReinforceCount + 1
	now := time.Now()

	return s.db.Model(&memory).Updates(map[string]interface{}{
		"importance":       newImportance,
		"reinforce_count":  newReinforceCount,
		"access_count":     memory.AccessCount + 1,
		"last_accessed_at": now,
		"decay_stage":      0,
		"status":           "active",
	}).Error
}

func (s *SmartLoadService) GenerateAndSaveSummary(userID uint, memoryID uint) (string, error) {
	var memory models.Memory
	if err := s.db.Where("user_id = ? AND id = ?", userID, memoryID).First(&memory).Error; err != nil {
		return "", err
	}

	summary := generateSummary(memory.Key, memory.Value)
	s.db.Model(&memory).Update("summary", summary)
	return summary, nil
}

func (s *SmartLoadService) BatchGenerateSummaries(userID uint) (int, error) {
	var memories []models.Memory
	if err := s.db.Where("user_id = ? AND summary = '' AND status = ?", userID, "active").Find(&memories).Error; err != nil {
		return 0, err
	}

	count := 0
	for _, m := range memories {
		summary := generateSummary(m.Key, m.Value)
		s.db.Model(&m).Update("summary", summary)
		count++
	}
	return count, nil
}
