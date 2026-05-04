package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"clawmemory/internal/models"

	"gorm.io/gorm"
)

type AIService struct {
	router *AIRouter
	db     *gorm.DB
}

func NewAIService(router *AIRouter, db *gorm.DB) *AIService {
	return &AIService{
		router: router,
		db:     db,
	}
}

func (s *AIService) chatWithTemplate(ctx context.Context, userID uint, isPro bool, templateID string, templateData map[string]string) (string, error) {
	tmpl, ok := GetPromptTemplate(templateID)
	if !ok {
		return "", fmt.Errorf("prompt template not found: %s", templateID)
	}

	if tmpl.ProOnly && !isPro {
		return "", fmt.Errorf("prompt template %s requires Pro license", templateID)
	}

	systemPrompt := RenderPrompt(tmpl.System, templateData)
	userPrompt := RenderPrompt(tmpl.User, templateData)

	messages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	opts := DefaultChatOptions()
	if isPro {
		opts.Temperature = 0.3
	}

	resp, err := s.router.Chat(ctx, userID, isPro, messages, opts)
	if err != nil {
		return "", err
	}

	s.router.IncrementUsage(userID, resp.TokensIn, resp.TokensOut)

	return resp.Content, nil
}

func (s *AIService) AIExtract(ctx context.Context, userID uint, isPro bool) (map[string]interface{}, error) {
	var memories []models.Memory
	s.db.Where("user_id = ? AND status != ?", userID, "trashed").Find(&memories)

	if len(memories) == 0 {
		return map[string]interface{}{
			"entities": []interface{}{},
			"relations": []interface{}{},
			"total":     0,
			"mode":      "ai",
		}, nil
	}

	memData := make([]map[string]interface{}, len(memories))
	for i, m := range memories {
		memData[i] = map[string]interface{}{
			"id":         m.ID,
			"key":        m.Key,
			"value":      truncateStr(m.Value, 300),
			"layer":      m.Layer,
			"importance": m.Importance,
		}
	}

	content, err := s.chatWithTemplate(ctx, userID, isPro, "extract", map[string]string{
		"Memories": FormatMemoriesForPrompt(memData),
	})
	if err != nil {
		return nil, fmt.Errorf("AI extraction failed: %w", err)
	}

	var result struct {
		Entities []struct {
			Name        string  `json:"name"`
			Type        string  `json:"type"`
			Description string  `json:"description"`
			Confidence  float64 `json:"confidence"`
		} `json:"entities"`
		Relations []struct {
			Source      string  `json:"source"`
			Target      string  `json:"target"`
			Type        string  `json:"type"`
			Description string  `json:"description"`
			Confidence  float64 `json:"confidence"`
		} `json:"relations"`
	}

	if err := ParseAIResponse(content, &result); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	entities := make([]map[string]interface{}, len(result.Entities))
	for i, e := range result.Entities {
		entities[i] = map[string]interface{}{
			"name":        e.Name,
			"entity_type": e.Type,
			"description": e.Description,
			"confidence":  e.Confidence,
			"method":      "ai",
		}
	}

	relations := make([]map[string]interface{}, len(result.Relations))
	for i, r := range result.Relations {
		relations[i] = map[string]interface{}{
			"source":      r.Source,
			"target":      r.Target,
			"type":        r.Type,
			"description": r.Description,
			"confidence":  r.Confidence,
			"method":      "ai",
		}
	}

	return map[string]interface{}{
		"entities":  entities,
		"relations": relations,
		"total":     len(entities) + len(relations),
		"mode":      "ai",
	}, nil
}

func (s *AIService) ConflictScan(ctx context.Context, userID uint, isPro bool) (map[string]interface{}, error) {
	var memories []models.Memory
	s.db.Where("user_id = ? AND status != ?", userID, "trashed").Find(&memories)

	if len(memories) < 2 {
		return map[string]interface{}{
			"conflicts": []interface{}{},
			"total":     0,
			"mode":      "ai",
		}, nil
	}

	memData := make([]map[string]interface{}, len(memories))
	for i, m := range memories {
		memData[i] = map[string]interface{}{
			"id":    m.ID,
			"key":   m.Key,
			"value": truncateStr(m.Value, 300),
			"layer": m.Layer,
		}
	}

	content, err := s.chatWithTemplate(ctx, userID, isPro, "conflict_scan", map[string]string{
		"Memories": FormatMemoriesForPrompt(memData),
	})
	if err != nil {
		return nil, fmt.Errorf("AI conflict scan failed: %w", err)
	}

	var result struct {
		Conflicts []struct {
			MemoryAID   int     `json:"memory_a_id"`
			MemoryBID   int     `json:"memory_b_id"`
			Type        string  `json:"type"`
			Description string  `json:"description"`
			Severity    string  `json:"severity"`
			Suggestion  string  `json:"suggestion"`
		} `json:"conflicts"`
	}

	if err := ParseAIResponse(content, &result); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	conflicts := make([]map[string]interface{}, len(result.Conflicts))
	for i, c := range result.Conflicts {
		conflicts[i] = map[string]interface{}{
			"memory_a_id": c.MemoryAID,
			"memory_b_id": c.MemoryBID,
			"type":        c.Type,
			"description": c.Description,
			"severity":    c.Severity,
			"suggestion":  c.Suggestion,
		}
	}

	return map[string]interface{}{
		"conflicts":       conflicts,
		"total":           len(conflicts),
		"auto_resolvable": len(conflicts),
		"needs_review":    0,
		"algorithm":       "ai_semantic_v1",
		"mode":            "ai",
	}, nil
}

func (s *AIService) DecayEvaluate(ctx context.Context, userID uint, isPro bool) (map[string]interface{}, error) {
	var memories []models.Memory
	s.db.Where("user_id = ? AND status != ?", userID, "trashed").Find(&memories)

	if len(memories) == 0 {
		return map[string]interface{}{
			"evaluations": []interface{}{},
			"total":       0,
			"mode":        "ai",
		}, nil
	}

	memData := make([]map[string]interface{}, len(memories))
	for i, m := range memories {
		daysSinceUpdate := time.Since(m.UpdatedAt).Hours() / 24
		memData[i] = map[string]interface{}{
			"id":            m.ID,
			"key":           m.Key,
			"value":         truncateStr(m.Value, 200),
			"importance":    m.Importance,
			"layer":         m.Layer,
			"access_count":  m.AccessCount,
			"reinforce_count": m.ReinforceCount,
			"days_since_update": fmt.Sprintf("%.0f", daysSinceUpdate),
		}
	}

	content, err := s.chatWithTemplate(ctx, userID, isPro, "decay_evaluate", map[string]string{
		"Memories": FormatMemoriesForPrompt(memData),
	})
	if err != nil {
		return nil, fmt.Errorf("AI decay evaluation failed: %w", err)
	}

	var result struct {
		Evaluations []struct {
			ID             uint    `json:"id"`
			Action         string  `json:"action"`
			Reason         string  `json:"reason"`
			NewImportance  float64 `json:"new_importance"`
			MergeWith      []uint  `json:"merge_with"`
		} `json:"evaluations"`
	}

	if err := ParseAIResponse(content, &result); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	evaluations := make([]map[string]interface{}, len(result.Evaluations))
	for i, e := range result.Evaluations {
		evaluations[i] = map[string]interface{}{
			"id":             e.ID,
			"action":         e.Action,
			"reason":         e.Reason,
			"new_importance": math.Round(e.NewImportance*100) / 100,
			"merge_with":     e.MergeWith,
		}
	}

	return map[string]interface{}{
		"evaluations": evaluations,
		"total":       len(evaluations),
		"algorithm":   "ai_decay_v1",
		"mode":        "ai",
	}, nil
}

func (s *AIService) GenerateDailyReport(ctx context.Context, userID uint, isPro bool, date string) (map[string]interface{}, error) {
	var memoryCount int64
	s.db.Model(&models.Memory{}).Where("user_id = ? AND DATE(created_at) = ? AND status != ?", userID, date, "trashed").Count(&memoryCount)

	var memories []models.Memory
	s.db.Where("user_id = ? AND DATE(created_at) = ? AND status != ?", userID, date, "trashed").
		Order("importance DESC").Limit(10).Find(&memories)

	highlights := ""
	for i, m := range memories {
		highlights += fmt.Sprintf("- %s: %s\n", m.Key, truncateStr(m.Value, 100))
		if i >= 9 {
			break
		}
	}

	if highlights == "" {
		highlights = "No significant activity today"
	}

	content, err := s.chatWithTemplate(ctx, userID, isPro, "daily_report", map[string]string{
		"Date":         date,
		"MemoryCount":  fmt.Sprintf("%d", memoryCount),
		"Highlights":   highlights,
	})
	if err != nil {
		return nil, fmt.Errorf("AI daily report generation failed: %w", err)
	}

	var result struct {
		Summary             string   `json:"summary"`
		Highlights          []string `json:"highlights"`
		KnowledgeGained     []string `json:"knowledge_gained"`
		PendingTasks        []string `json:"pending_tasks"`
		TomorrowSuggestions []string `json:"tomorrow_suggestions"`
		Mood                string   `json:"mood"`
	}

	if err := ParseAIResponse(content, &result); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return map[string]interface{}{
		"summary":              result.Summary,
		"highlights":           result.Highlights,
		"knowledge_gained":     result.KnowledgeGained,
		"pending_tasks":        result.PendingTasks,
		"tomorrow_suggestions": result.TomorrowSuggestions,
		"mood":                 result.Mood,
		"mode":                 "ai",
	}, nil
}

func (s *AIService) GenerateWiki(ctx context.Context, userID uint, isPro bool, topic string) (map[string]interface{}, error) {
	var memories []models.Memory
	query := s.db.Where("user_id = ? AND status != ?", userID, "trashed")
	if topic != "" {
		query = query.Where("key LIKE ? OR value LIKE ?", "%"+topic+"%", "%"+topic+"%")
	}
	query.Find(&memories)

	if len(memories) == 0 {
		return nil, fmt.Errorf("no memories found for topic: %s", topic)
	}

	memData := make([]map[string]interface{}, len(memories))
	for i, m := range memories {
		memData[i] = map[string]interface{}{
			"id":    m.ID,
			"key":   m.Key,
			"value": truncateStr(m.Value, 300),
			"layer": m.Layer,
		}
	}

	content, err := s.chatWithTemplate(ctx, userID, isPro, "wiki_generate", map[string]string{
		"Topic":    topic,
		"Memories": FormatMemoriesForPrompt(memData),
	})
	if err != nil {
		return nil, fmt.Errorf("AI wiki generation failed: %w", err)
	}

	var result struct {
		Title         string   `json:"title"`
		Category      string   `json:"category"`
		Content       string   `json:"content"`
		Summary       string   `json:"summary"`
		Tags          []string `json:"tags"`
		KeyDecisions  []string `json:"key_decisions"`
		ActionItems   []string `json:"action_items"`
	}

	if err := ParseAIResponse(content, &result); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return map[string]interface{}{
		"title":          result.Title,
		"category":       result.Category,
		"content":        result.Content,
		"summary":        result.Summary,
		"tags":           result.Tags,
		"key_decisions":  result.KeyDecisions,
		"action_items":   result.ActionItems,
		"ai_generated":   true,
		"mode":           "ai",
	}, nil
}

func (s *AIService) CompressMemories(ctx context.Context, userID uint, isPro bool, memoryIDs []uint) (map[string]interface{}, error) {
	var memories []models.Memory
	s.db.Where("id IN ? AND user_id = ?", memoryIDs, userID).Find(&memories)

	if len(memories) == 0 {
		return nil, fmt.Errorf("no memories found")
	}

	memData := make([]map[string]interface{}, len(memories))
	for i, m := range memories {
		memData[i] = map[string]interface{}{
			"id":    m.ID,
			"key":   m.Key,
			"value": m.Value,
			"layer": m.Layer,
		}
	}

	content, err := s.chatWithTemplate(ctx, userID, isPro, "compress", map[string]string{
		"Memories": FormatMemoriesForPrompt(memData),
	})
	if err != nil {
		return nil, fmt.Errorf("AI compression failed: %w", err)
	}

	var result struct {
		Key          string   `json:"key"`
		Value        string   `json:"value"`
		Importance   float64  `json:"importance"`
		Layer        string   `json:"layer"`
		Tags         []string `json:"tags"`
		MergedCount  int      `json:"merged_count"`
		Notes        string   `json:"notes"`
	}

	if err := ParseAIResponse(content, &result); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return map[string]interface{}{
		"key":          result.Key,
		"value":        result.Value,
		"importance":   math.Round(result.Importance*100) / 100,
		"layer":        result.Layer,
		"tags":         result.Tags,
		"merged_count": result.MergedCount,
		"notes":        result.Notes,
		"source_ids":   memoryIDs,
		"mode":         "ai",
	}, nil
}

func (s *AIService) DiscoverRelations(ctx context.Context, userID uint, isPro bool) (map[string]interface{}, error) {
	var memories []models.Memory
	s.db.Where("user_id = ? AND status != ?", userID, "trashed").Limit(30).Find(&memories)

	if len(memories) < 2 {
		return map[string]interface{}{
			"relations": []interface{}{},
			"total":     0,
			"mode":      "ai",
		}, nil
	}

	memData := make([]map[string]interface{}, len(memories))
	for i, m := range memories {
		memData[i] = map[string]interface{}{
			"id":    m.ID,
			"key":   m.Key,
			"value": truncateStr(m.Value, 200),
			"layer": m.Layer,
		}
	}

	content, err := s.chatWithTemplate(ctx, userID, isPro, "evolution_discover", map[string]string{
		"Memories": FormatMemoriesForPrompt(memData),
	})
	if err != nil {
		return nil, fmt.Errorf("AI relation discovery failed: %w", err)
	}

	var result struct {
		Relations []struct {
			SourceID    uint    `json:"source_id"`
			TargetID    uint    `json:"target_id"`
			Type        string  `json:"type"`
			Description string  `json:"description"`
			Confidence  float64 `json:"confidence"`
		} `json:"relations"`
	}

	if err := ParseAIResponse(content, &result); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	relations := make([]map[string]interface{}, len(result.Relations))
	for i, r := range result.Relations {
		relations[i] = map[string]interface{}{
			"source_id":   r.SourceID,
			"target_id":   r.TargetID,
			"type":        r.Type,
			"description": r.Description,
			"confidence":  math.Round(r.Confidence*100) / 100,
		}
	}

	return map[string]interface{}{
		"relations": relations,
		"total":     len(relations),
		"algorithm": "ai_discovery_v1",
		"mode":      "ai",
	}, nil
}

func (s *AIService) SmartRoute(ctx context.Context, userID uint, isPro bool, text string) (map[string]interface{}, error) {
	content, err := s.chatWithTemplate(ctx, userID, isPro, "smart_route", map[string]string{
		"Text": text,
	})
	if err != nil {
		return nil, fmt.Errorf("AI smart route failed: %w", err)
	}

	var result struct {
		ComplexityScore int    `json:"complexity_score"`
		Complexity      string `json:"complexity"`
		RecommendedModel string `json:"recommended_model"`
		Reason          string `json:"reason"`
		EstimatedTokens int    `json:"estimated_tokens"`
		TechnicalTerms  int    `json:"technical_terms"`
		SentenceCount   int    `json:"sentence_count"`
	}

	if err := ParseAIResponse(content, &result); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return map[string]interface{}{
		"model":             result.RecommendedModel,
		"complexity_score":  result.ComplexityScore,
		"complexity":        result.Complexity,
		"reason":            result.Reason,
		"estimated_tokens":  result.EstimatedTokens,
		"technical_terms":   result.TechnicalTerms,
		"sentence_count":    result.SentenceCount,
		"mode":              "ai",
	}, nil
}

func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func init() {
	_ = strings.ReplaceAll
	_ = json.Marshal
}
