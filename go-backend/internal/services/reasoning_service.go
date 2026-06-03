package services

import (
	"context"
	"fmt"
	"log"

	"clawmemory/internal/models"

	"gorm.io/gorm"
)

type AIChatProvider interface {
	Chat(ctx context.Context, userID uint, messages []AIChatMessage, opts AIChatOptions) (*AIChatResponse, error)
}

type AIChatMessage struct {
	Role    string
	Content string
}

type AIChatOptions struct {
	Temperature float64
	MaxTokens   int
}

type AIChatResponse struct {
	Content  string
	TokensIn int
	Model    string
}

var coldPrompt = `Who is this person based on their conversation history? What are their preferences, goals, and working style? Focus on facts that would help an AI assistant be immediately useful. Be concise and structured.`

var warmPrompt = `Given what has been discussed in this session so far, what context about this user is most relevant to the current conversation? Prioritize active context over biographical facts. Be concise and structured.`

var auditPrompt = `Review your previous assessment. Identify any gaps, contradictions, or missing insights. Synthesize a more complete picture.`

var reconcilePrompt = `Check for contradictions between your assessments. Produce a final, reconciled synthesis that resolves any conflicts.`

type ReasoningResult struct {
	Content  string `json:"content"`
	Pass     int    `json:"pass"`
	Model    string `json:"model"`
	TokensIn int    `json:"tokens_in"`
}

type ReasoningService struct {
	db     *gorm.DB
	aiChat AIChatProvider
}

func NewReasoningService(db *gorm.DB, aiChat AIChatProvider) *ReasoningService {
	return &ReasoningService{db: db, aiChat: aiChat}
}

func (s *ReasoningService) GetConfig(userID uint) (*models.ReasoningConfig, error) {
	var config models.ReasoningConfig
	if err := s.db.Where("user_id = ?", userID).First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &config, nil
}

func (s *ReasoningService) SetConfig(userID uint, data map[string]interface{}) (*models.ReasoningConfig, error) {
	var config models.ReasoningConfig
	result := s.db.Where("user_id = ?", userID).First(&config)

	if result.Error == gorm.ErrRecordNotFound {
		config = models.ReasoningConfig{
			UserID:         userID,
			Provider:       getString(data, "provider", "openai"),
			Model:          getString(data, "model", "gpt-4o-mini"),
			APIKey:         getString(data, "api_key", ""),
			BaseURL:        getString(data, "base_url", ""),
			DialecticDepth: getInt(data, "dialectic_depth", 1),
			ReasoningLevel: getString(data, "reasoning_level", "medium"),
			MaxTokens:      getInt(data, "max_tokens", 1000),
			Enabled:        getBool(data, "enabled", false),
		}
		if err := s.db.Create(&config).Error; err != nil {
			return nil, err
		}
		return &config, nil
	}

	updates := make(map[string]interface{})
	if v, ok := data["provider"].(string); ok && v != "" {
		updates["provider"] = v
	}
	if v, ok := data["model"].(string); ok && v != "" {
		updates["model"] = v
	}
	if v, ok := data["api_key"].(string); ok && v != "" {
		updates["api_key"] = v
	}
	if v, ok := data["base_url"].(string); ok {
		updates["base_url"] = v
	}
	if v, ok := data["dialectic_depth"]; ok {
		updates["dialectic_depth"] = v
	}
	if v, ok := data["reasoning_level"].(string); ok && v != "" {
		updates["reasoning_level"] = v
	}
	if v, ok := data["max_tokens"]; ok {
		updates["max_tokens"] = v
	}
	if v, ok := data["enabled"]; ok {
		updates["enabled"] = v
	}

	if len(updates) > 0 {
		if err := s.db.Model(&config).Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	s.db.Where("user_id = ?", userID).First(&config)
	return &config, nil
}

func (s *ReasoningService) TestConnection(userID uint) error {
	config, err := s.GetConfig(userID)
	if err != nil {
		return err
	}
	if config == nil {
		return fmt.Errorf("no reasoning config found")
	}
	if !config.Enabled {
		return fmt.Errorf("reasoning is not enabled")
	}

	_, err = s.callLLM(context.Background(), userID, "Reply with OK", "minimal", config.MaxTokens)
	return err
}

func (s *ReasoningService) Reason(userID uint, query string, depth int, level string) (string, error) {
	config, err := s.GetConfig(userID)
	if err != nil {
		return "", fmt.Errorf("failed to get reasoning config: %w", err)
	}
	if config == nil {
		return "", fmt.Errorf("no reasoning model configured. Please configure one in Settings > AI Config")
	}
	if !config.Enabled {
		return "", fmt.Errorf("reasoning is not enabled. Please enable it in Settings > AI Config")
	}

	if depth < 1 {
		depth = config.DialecticDepth
	}
	if depth < 1 {
		depth = 1
	}
	if depth > 3 {
		depth = 3
	}
	if level == "" {
		level = config.ReasoningLevel
	}
	if level == "" {
		level = "medium"
	}

	memSvc := NewMemoryService(s.db)
	existing, _ := memSvc.SearchKeyword(userID, query, 3)
	hasExistingContext := len(existing) > 0

	pass0Prompt := coldPrompt
	if hasExistingContext {
		pass0Prompt = warmPrompt
	}

	fullQuery := pass0Prompt + "\n\nConversation context:\n" + truncate(query, 10000)

	pass0Result, err := s.callLLM(context.Background(), userID, fullQuery, level, config.MaxTokens)
	if err != nil {
		return "", fmt.Errorf("pass 0 failed: %w", err)
	}
	results := []ReasoningResult{pass0Result}

	if depth >= 2 && len(pass0Result.Content) > 300 {
		auditQuery := auditPrompt + "\n\nPrevious assessment:\n" + pass0Result.Content
		pass1Result, err := s.callLLM(context.Background(), userID, auditQuery, "low", config.MaxTokens)
		if err == nil {
			results = append(results, pass1Result)
		}
	}

	if depth >= 3 && len(results) >= 2 {
		reconcileQuery := reconcilePrompt + "\n\nAssessment 1:\n" + results[0].Content + "\n\nAssessment 2:\n" + results[1].Content
		pass2Result, err := s.callLLM(context.Background(), userID, reconcileQuery, "low", config.MaxTokens)
		if err == nil {
			results = append(results, pass2Result)
		}
	}

	finalResult := results[len(results)-1].Content

	_, err = memSvc.Create(userID, map[string]interface{}{
		"key":         fmt.Sprintf("reasoning-%s-%d", level, userID),
		"value":       finalResult,
		"layer":       "context",
		"source":      "dialectic-reasoning",
		"memory_type": "knowledge",
	})
	if err != nil {
		log.Printf("[Reasoning] failed to save reasoning result: %v", err)
	}

	return finalResult, nil
}

func (s *ReasoningService) callLLM(ctx context.Context, userID uint, prompt string, level string, maxTokenLimit int) (ReasoningResult, error) {
	maxTokens := 600
	switch level {
	case "minimal":
		maxTokens = 200
	case "low":
		maxTokens = 400
	case "medium":
		maxTokens = 600
	case "high":
		maxTokens = 1000
	case "max":
		maxTokens = 2000
	}

	if maxTokenLimit > 0 && maxTokens > maxTokenLimit {
		maxTokens = maxTokenLimit
	}

	resp, err := s.aiChat.Chat(ctx, userID, []AIChatMessage{
		{Role: "user", Content: prompt},
	}, AIChatOptions{
		Temperature: 0.7,
		MaxTokens:   maxTokens,
	})
	if err != nil {
		return ReasoningResult{}, err
	}

	return ReasoningResult{
		Content:  resp.Content,
		Pass:     0,
		Model:    resp.Model,
		TokensIn: resp.TokensIn,
	}, nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}
