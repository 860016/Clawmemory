package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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

	var resp *ChatResponse
	var err error
	maxRetries := 2
	for attempt := 0; attempt <= maxRetries; attempt++ {
		chatCtx, cancel := context.WithTimeout(ctx, 120*time.Second)
		resp, err = s.router.Chat(chatCtx, userID, isPro, messages, opts)
		cancel()

		if err == nil {
			break
		}

		if ctx.Err() != nil {
			return "", ctx.Err()
		}

		if attempt < maxRetries {
			time.Sleep(time.Duration(attempt+1) * 2 * time.Second)
		}
	}

	if err != nil {
		return "", fmt.Errorf("AI call failed after %d retries: %w", maxRetries+1, err)
	}

	s.router.IncrementUsage(userID, resp.TokensIn, resp.TokensOut)

	return resp.Content, nil
}

func (s *AIService) AIExtract(ctx context.Context, userID uint, isPro bool) (map[string]interface{}, error) {
	var memories []models.Memory
	if err := s.db.Where("user_id = ? AND status != ?", userID, "trashed").Limit(5000).Find(&memories).Error; err != nil {
		return nil, fmt.Errorf("failed to load memories: %w", err)
	}

	if len(memories) == 0 {
		return map[string]interface{}{
			"entities":  []interface{}{},
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

func escapeLikeAI(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

func (s *AIService) ConflictScan(ctx context.Context, userID uint, isPro bool) (map[string]interface{}, error) {
	var memories []models.Memory
	if err := s.db.Where("user_id = ? AND status != ?", userID, "trashed").Limit(5000).Find(&memories).Error; err != nil {
		return nil, fmt.Errorf("failed to load memories: %w", err)
	}

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
			MemoryAID   int    `json:"memory_a_id"`
			MemoryBID   int    `json:"memory_b_id"`
			Type        string `json:"type"`
			Description string `json:"description"`
			Severity    string `json:"severity"`
			Suggestion  string `json:"suggestion"`
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
	if err := s.db.Where("user_id = ? AND status != ?", userID, "trashed").Limit(5000).Find(&memories).Error; err != nil {
		return nil, fmt.Errorf("failed to load memories: %w", err)
	}

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
			"id":                m.ID,
			"key":               m.Key,
			"value":             truncateStr(m.Value, 200),
			"importance":        m.Importance,
			"layer":             m.Layer,
			"access_count":      m.AccessCount,
			"reinforce_count":   m.ReinforceCount,
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
			ID            uint    `json:"id"`
			Action        string  `json:"action"`
			Reason        string  `json:"reason"`
			NewImportance float64 `json:"new_importance"`
			MergeWith     []uint  `json:"merge_with"`
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
		"Date":        date,
		"MemoryCount": fmt.Sprintf("%d", memoryCount),
		"Highlights":  highlights,
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
		escaped := escapeLikeAI(topic)
		query = query.Where("key LIKE ? OR value LIKE ?", "%"+escaped+"%", "%"+escaped+"%")
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
		Title        string   `json:"title"`
		Category     string   `json:"category"`
		Content      string   `json:"content"`
		Summary      string   `json:"summary"`
		Tags         []string `json:"tags"`
		KeyDecisions []string `json:"key_decisions"`
		ActionItems  []string `json:"action_items"`
	}

	if err := ParseAIResponse(content, &result); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return map[string]interface{}{
		"title":         result.Title,
		"category":      result.Category,
		"content":       result.Content,
		"summary":       result.Summary,
		"tags":          result.Tags,
		"key_decisions": result.KeyDecisions,
		"action_items":  result.ActionItems,
		"ai_generated":  true,
		"mode":          "ai",
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
		Key         string   `json:"key"`
		Value       string   `json:"value"`
		Importance  float64  `json:"importance"`
		Layer       string   `json:"layer"`
		Tags        []string `json:"tags"`
		MergedCount int      `json:"merged_count"`
		Notes       string   `json:"notes"`
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
		ComplexityScore  int    `json:"complexity_score"`
		Complexity       string `json:"complexity"`
		RecommendedModel string `json:"recommended_model"`
		Reason           string `json:"reason"`
		EstimatedTokens  int    `json:"estimated_tokens"`
		TechnicalTerms   int    `json:"technical_terms"`
		SentenceCount    int    `json:"sentence_count"`
	}

	if err := ParseAIResponse(content, &result); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return map[string]interface{}{
		"model":            result.RecommendedModel,
		"complexity_score": result.ComplexityScore,
		"complexity":       result.Complexity,
		"reason":           result.Reason,
		"estimated_tokens": result.EstimatedTokens,
		"technical_terms":  result.TechnicalTerms,
		"sentence_count":   result.SentenceCount,
		"mode":             "ai",
	}, nil
}

func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

type ExtractedFact struct {
	Content    string  `json:"content"`
	Category   string  `json:"category"`
	Confidence float64 `json:"confidence"`
	Source     string  `json:"source"`
}

type ExtractedPreference struct {
	Topic      string  `json:"topic"`
	Value      string  `json:"value"`
	Strength   string  `json:"strength"`
	Confidence float64 `json:"confidence"`
}

type ExtractedRelation struct {
	Subject    string  `json:"subject"`
	Predicate  string  `json:"predicate"`
	Object     string  `json:"object"`
	Confidence float64 `json:"confidence"`
}

type FactUpdate struct {
	OldFact string `json:"old_fact"`
	NewFact string `json:"new_fact"`
	Reason  string `json:"reason"`
}

type ExtractionResult struct {
	Facts       []ExtractedFact       `json:"facts"`
	Preferences []ExtractedPreference `json:"preferences"`
	Relations   []ExtractedRelation   `json:"relations"`
	Updates     []FactUpdate          `json:"updates"`
}

func (s *AIService) ExtractFacts(ctx context.Context, userID uint, isPro bool, messages []map[string]string) (map[string]interface{}, error) {
	var msgBuilder strings.Builder
	for i, msg := range messages {
		role := msg["role"]
		content := msg["content"]
		if content == "" {
			continue
		}
		msgBuilder.WriteString(fmt.Sprintf("[%s]: %s\n", role, content))
		if i >= 99 {
			msgBuilder.WriteString("... (truncated)\n")
			break
		}
	}

	var existingMemories []models.Memory
	s.db.Where("user_id = ? AND status != ?", userID, "trashed").Order("importance DESC").Limit(50).Find(&existingMemories)

	var memBuilder strings.Builder
	for i, m := range existingMemories {
		memBuilder.WriteString(fmt.Sprintf("- [%d] %s: %s\n", m.ID, m.Key, truncateStr(m.Value, 200)))
		if i >= 29 {
			memBuilder.WriteString("... (showing top 30)\n")
			break
		}
	}

	existingMemStr := "No existing memories"
	if memBuilder.Len() > 0 {
		existingMemStr = memBuilder.String()
	}

	content, err := s.chatWithTemplate(ctx, userID, isPro, "extract_facts", map[string]string{
		"Messages":         msgBuilder.String(),
		"ExistingMemories": existingMemStr,
	})
	if err != nil {
		return nil, fmt.Errorf("fact extraction failed: %w", err)
	}

	var result ExtractionResult
	if err := ParseAIResponse(content, &result); err != nil {
		return nil, fmt.Errorf("failed to parse extraction result: %w", err)
	}

	facts := make([]map[string]interface{}, len(result.Facts))
	for i, f := range result.Facts {
		facts[i] = map[string]interface{}{
			"content":    f.Content,
			"category":   f.Category,
			"confidence": f.Confidence,
			"source":     f.Source,
		}
	}

	prefs := make([]map[string]interface{}, len(result.Preferences))
	for i, p := range result.Preferences {
		prefs[i] = map[string]interface{}{
			"topic":      p.Topic,
			"value":      p.Value,
			"strength":   p.Strength,
			"confidence": p.Confidence,
		}
	}

	rels := make([]map[string]interface{}, len(result.Relations))
	for i, r := range result.Relations {
		rels[i] = map[string]interface{}{
			"subject":    r.Subject,
			"predicate":  r.Predicate,
			"object":     r.Object,
			"confidence": r.Confidence,
		}
	}

	updates := make([]map[string]interface{}, len(result.Updates))
	for i, u := range result.Updates {
		updates[i] = map[string]interface{}{
			"old_fact": u.OldFact,
			"new_fact": u.NewFact,
			"reason":   u.Reason,
		}
	}

	return map[string]interface{}{
		"facts":       facts,
		"preferences": prefs,
		"relations":   rels,
		"updates":     updates,
		"total":       len(facts) + len(prefs) + len(rels) + len(updates),
		"mode":        "ai",
		"algorithm":   "llm_extraction_v1",
	}, nil
}

func (s *AIService) ConsolidateMemories(ctx context.Context, userID uint, isPro bool, newFacts []map[string]interface{}) (map[string]interface{}, error) {
	var factBuilder strings.Builder
	for i, f := range newFacts {
		content, _ := f["content"].(string)
		category, _ := f["category"].(string)
		confidence, _ := f["confidence"].(float64)
		factBuilder.WriteString(fmt.Sprintf("- [%s] %s (confidence: %.2f)\n", category, content, confidence))
		if i >= 49 {
			break
		}
	}

	var existingMemories []models.Memory
	s.db.Where("user_id = ? AND status != ?", userID, "trashed").Order("importance DESC").Limit(50).Find(&existingMemories)

	var memBuilder strings.Builder
	for i, m := range existingMemories {
		memBuilder.WriteString(fmt.Sprintf("- [%d] key=%s layer=%s importance=%.2f\n  %s\n", m.ID, m.Key, m.Layer, m.Importance, truncateStr(m.Value, 200)))
		if i >= 29 {
			break
		}
	}

	existingMemStr := "No existing memories"
	if memBuilder.Len() > 0 {
		existingMemStr = memBuilder.String()
	}

	content, err := s.chatWithTemplate(ctx, userID, isPro, "memory_consolidate", map[string]string{
		"NewFacts":         factBuilder.String(),
		"ExistingMemories": existingMemStr,
	})
	if err != nil {
		return nil, fmt.Errorf("memory consolidation failed: %w", err)
	}

	var result struct {
		Add []struct {
			Key        string  `json:"key"`
			Value      string  `json:"value"`
			Layer      string  `json:"layer"`
			Importance float64 `json:"importance"`
			Category   string  `json:"category"`
		} `json:"add"`
		Update []struct {
			MemoryID int    `json:"memory_id"`
			Field    string `json:"field"`
			OldValue string `json:"old_value"`
			NewValue string `json:"new_value"`
			Reason   string `json:"reason"`
		} `json:"update"`
		Merge []struct {
			SourceIDs   []int  `json:"source_ids"`
			MergedKey   string `json:"merged_key"`
			MergedValue string `json:"merged_value"`
			Layer       string `json:"layer"`
		} `json:"merge"`
		Supersede []struct {
			OldID  int    `json:"old_id"`
			Reason string `json:"reason"`
			NewID  int    `json:"new_id"`
		} `json:"supersede"`
		Skip []struct {
			Fact            string `json:"fact"`
			MatchesMemoryID int    `json:"matches_memory_id"`
		} `json:"skip"`
	}

	if err := ParseAIResponse(content, &result); err != nil {
		return nil, fmt.Errorf("failed to parse consolidation result: %w", err)
	}

	added := 0
	for _, a := range result.Add {
		layer := a.Layer
		if layer == "" {
			layer = "context"
		}
		importance := a.Importance
		if importance == 0 {
			importance = 0.5
		}
		memory := models.Memory{
			UserID:      userID,
			Key:         a.Key,
			Value:       a.Value,
			Layer:       layer,
			Importance:  importance,
			Status:      "active",
			DecayStage:  0,
			AccessCount: 0,
		}
		if err := s.db.Create(&memory).Error; err == nil {
			added++
		}
	}

	updated := 0
	for _, u := range result.Update {
		if u.Field == "value" || u.Field == "importance" || u.Field == "layer" {
			var mem models.Memory
			if err := s.db.Where("id = ? AND user_id = ?", u.MemoryID, userID).First(&mem).Error; err == nil {
				s.db.Model(&mem).Update(u.Field, u.NewValue)
				updated++
			}
		}
	}

	merged := 0
	for _, m := range result.Merge {
		if len(m.SourceIDs) < 2 {
			continue
		}
		var sourceMem models.Memory
		if err := s.db.Where("id = ? AND user_id = ?", m.SourceIDs[0], userID).First(&sourceMem).Error; err != nil {
			continue
		}

		sourceMem.Key = m.MergedKey
		sourceMem.Value = m.MergedValue
		if m.Layer != "" {
			sourceMem.Layer = m.Layer
		}
		s.db.Save(&sourceMem)

		for _, sid := range m.SourceIDs[1:] {
			s.db.Model(&models.Memory{}).Where("id = ? AND user_id = ?", sid, userID).
				Updates(map[string]interface{}{"status": "trashed", "decay_stage": 3})
		}
		merged++
	}

	superseded := 0
	for _, sp := range result.Supersede {
		var oldMem models.Memory
		if err := s.db.Where("id = ? AND user_id = ?", sp.OldID, userID).First(&oldMem).Error; err == nil {
			s.db.Model(&oldMem).Updates(map[string]interface{}{
				"status":     "archived",
				"importance": oldMem.Importance * 0.5,
			})
			superseded++
		}
	}

	skipped := len(result.Skip)

	return map[string]interface{}{
		"added":       added,
		"updated":     updated,
		"merged":      merged,
		"superseded":  superseded,
		"skipped":     skipped,
		"total_input": len(newFacts),
		"mode":        "ai",
		"algorithm":   "llm_consolidation_v1",
	}, nil
}

func (s *AIService) AssembleContext(ctx context.Context, userID uint, isPro bool, query string, tokenBudget int) (map[string]interface{}, error) {
	var memories []models.Memory
	s.db.Where("user_id = ? AND status != ?", userID, "trashed").Order("importance DESC").Limit(100).Find(&memories)

	if len(memories) == 0 {
		return map[string]interface{}{
			"selected_memories": []interface{}{},
			"total_tokens":      0,
			"coverage_score":    0,
			"mode":              "ai",
		}, nil
	}

	var memBuilder strings.Builder
	for i, m := range memories {
		memBuilder.WriteString(fmt.Sprintf("- [%d] key=%s layer=%s importance=%.2f\n  %s\n", m.ID, m.Key, m.Layer, m.Importance, truncateStr(m.Value, 200)))
		if i >= 49 {
			break
		}
	}

	if tokenBudget <= 0 {
		tokenBudget = 4000
	}

	content, err := s.chatWithTemplate(ctx, userID, isPro, "context_assemble", map[string]string{
		"Query":       query,
		"TokenBudget": fmt.Sprintf("%d", tokenBudget),
		"Memories":    memBuilder.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("context assembly failed: %w", err)
	}

	var result struct {
		SelectedMemories []struct {
			MemoryID        int     `json:"memory_id"`
			RelevanceScore  float64 `json:"relevance_score"`
			RelevanceReason string  `json:"relevance_reason"`
			Tokens          int     `json:"tokens"`
		} `json:"selected_memories"`
		TotalTokens       int      `json:"total_tokens"`
		CoverageScore     float64  `json:"coverage_score"`
		MissingContext    []string `json:"missing_context"`
		SuggestedFollowup []string `json:"suggested_followup"`
	}

	if err := ParseAIResponse(content, &result); err != nil {
		return nil, fmt.Errorf("failed to parse context assembly result: %w", err)
	}

	selected := make([]map[string]interface{}, len(result.SelectedMemories))
	for i, sm := range result.SelectedMemories {
		selected[i] = map[string]interface{}{
			"memory_id":        sm.MemoryID,
			"relevance_score":  sm.RelevanceScore,
			"relevance_reason": sm.RelevanceReason,
			"tokens":           sm.Tokens,
		}
	}

	return map[string]interface{}{
		"selected_memories":  selected,
		"total_tokens":       result.TotalTokens,
		"coverage_score":     result.CoverageScore,
		"missing_context":    result.MissingContext,
		"suggested_followup": result.SuggestedFollowup,
		"mode":               "ai",
		"algorithm":          "llm_context_assembly_v1",
	}, nil
}

func (s *AIService) ProcessConversation(ctx context.Context, userID uint, isPro bool, messages []map[string]string) (map[string]interface{}, error) {
	extraction, err := s.ExtractFacts(ctx, userID, isPro, messages)
	if err != nil {
		return nil, fmt.Errorf("extraction phase failed: %w", err)
	}

	facts, _ := extraction["facts"].([]map[string]interface{})
	prefs, _ := extraction["preferences"].([]map[string]interface{})
	updates, _ := extraction["updates"].([]map[string]interface{})

	allNewFacts := make([]map[string]interface{}, 0)
	for _, f := range facts {
		allNewFacts = append(allNewFacts, f)
	}
	for _, p := range prefs {
		allNewFacts = append(allNewFacts, map[string]interface{}{
			"content":    fmt.Sprintf("User prefers %s: %s", p["topic"], p["value"]),
			"category":   "preference",
			"confidence": p["confidence"],
			"source":     "user",
		})
	}
	for _, u := range updates {
		allNewFacts = append(allNewFacts, map[string]interface{}{
			"content":    u["new_fact"],
			"category":   "update",
			"confidence": 0.8,
			"source":     "inferred",
		})
	}

	if len(allNewFacts) == 0 {
		return map[string]interface{}{
			"extraction": extraction,
			"consolidation": map[string]interface{}{
				"added":       0,
				"updated":     0,
				"merged":      0,
				"superseded":  0,
				"skipped":     0,
				"total_input": 0,
				"mode":        "ai",
			},
			"total_processed": 0,
			"mode":            "ai",
		}, nil
	}

	consolidation, err := s.ConsolidateMemories(ctx, userID, isPro, allNewFacts)
	if err != nil {
		return map[string]interface{}{
			"extraction":      extraction,
			"consolidation":   nil,
			"error":           fmt.Sprintf("consolidation failed: %v", err),
			"total_processed": len(allNewFacts),
			"mode":            "ai",
		}, nil
	}

	return map[string]interface{}{
		"extraction":      extraction,
		"consolidation":   consolidation,
		"total_processed": len(allNewFacts),
		"mode":            "ai",
		"algorithm":       "mem0_style_pipeline_v1",
	}, nil
}

func (s *AIService) NudgeReflect(ctx context.Context, userID uint, isPro bool) (map[string]interface{}, error) {
	var recentHistory []models.MemoryHistory
	s.db.Where("user_id = ?", userID).Order("created_at DESC").Limit(50).Find(&recentHistory)

	var recentMemories []models.Memory
	s.db.Where("user_id = ? AND status != ? AND updated_at > ?", userID, "trashed",
		time.Now().AddDate(0, 0, -7)).Order("updated_at DESC").Limit(100).Find(&recentMemories)

	if len(recentHistory) == 0 && len(recentMemories) == 0 {
		return map[string]interface{}{
			"persist":    []interface{}{},
			"compress":   []interface{}{},
			"forget":     []interface{}{},
			"insights":   []interface{}{},
			"nudge_type": "skipped",
			"reason":     "no recent activity to reflect on",
		}, nil
	}

	changesJSON, _ := json.Marshal(recentHistory)

	var profile models.UserProfile
	profileJSON := "{}"
	if err := s.db.Where("user_id = ?", userID).First(&profile).Error; err == nil {
		pj, _ := json.Marshal(map[string]interface{}{
			"identity":    profile.Identity,
			"workflow":    profile.Workflow,
			"preferences": profile.Preferences,
		})
		profileJSON = string(pj)
	}

	var totalMemories int64
	var coreCount, contextCount, detailCount int64
	s.db.Model(&models.Memory{}).Where("user_id = ? AND status != ?", userID, "trashed").Count(&totalMemories)
	s.db.Model(&models.Memory{}).Where("user_id = ? AND layer = ? AND status != ?", userID, "core", "trashed").Count(&coreCount)
	s.db.Model(&models.Memory{}).Where("user_id = ? AND layer = ? AND status != ?", userID, "context", "trashed").Count(&contextCount)
	s.db.Model(&models.Memory{}).Where("user_id = ? AND layer = ? AND status != ?", userID, "detail", "trashed").Count(&detailCount)

	statsJSON, _ := json.Marshal(map[string]interface{}{
		"total":   totalMemories,
		"core":    coreCount,
		"context": contextCount,
		"detail":  detailCount,
	})

	content, err := s.chatWithTemplate(ctx, userID, isPro, "nudge_reflect", map[string]string{
		"RecentChanges": string(changesJSON),
		"UserProfile":   profileJSON,
		"MemoryStats":   string(statsJSON),
	})
	if err != nil {
		return nil, fmt.Errorf("nudge reflection failed: %w", err)
	}

	cleaned := strings.TrimSpace(content)
	if idx := strings.Index(cleaned, "{"); idx > 0 {
		cleaned = cleaned[idx:]
	}
	if idx := strings.LastIndex(cleaned, "}"); idx >= 0 {
		cleaned = cleaned[:idx+1]
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
		return map[string]interface{}{
			"raw_response": content,
			"error":        "failed to parse nudge result",
		}, nil
	}

	persistItems, _ := result["persist"].([]interface{})
	compressItems, _ := result["compress"].([]interface{})
	forgetItems, _ := result["forget"].([]interface{})
	profileUpdates, _ := result["profile_updates"].([]interface{})

	added := 0
	for _, item := range persistItems {
		p, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		contentStr, _ := p["content"].(string)
		category, _ := p["category"].(string)
		if contentStr == "" {
			continue
		}
		layer := "detail"
		if category == "identity" || category == "preference" {
			layer = "core"
		} else if category == "skill" || category == "routine" {
			layer = "context"
		}
		mem := models.Memory{
			UserID:     userID,
			Layer:      layer,
			Key:        category + ":" + truncateString(contentStr, 80),
			Value:      contentStr,
			Importance: 0.6,
			Source:     "nudge",
			Status:     "active",
		}
		if conf, ok := p["confidence"].(float64); ok {
			mem.Importance = math.Max(0.3, math.Min(1.0, conf))
		}
		if err := s.db.Create(&mem).Error; err == nil {
			added++
		}
	}

	merged := 0
	err = s.db.Transaction(func(tx *gorm.DB) error {
		for _, item := range compressItems {
			c, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			ids, _ := c["memory_ids"].([]interface{})
			compressedContent, _ := c["compressed_content"].(string)
			if len(ids) < 2 || compressedContent == "" {
				continue
			}
			var memIDs []uint
			for _, id := range ids {
				if f, ok := id.(float64); ok {
					memIDs = append(memIDs, uint(f))
				}
			}
			if len(memIDs) < 2 {
				continue
			}
			var first models.Memory
			if err := tx.Where("id = ? AND user_id = ?", memIDs[0], userID).First(&first).Error; err != nil {
				continue
			}
			first.Value = compressedContent
			first.Source = "nudge_compress"
			tx.Save(&first)
			tx.Where("id IN ? AND user_id = ? AND id != ?", memIDs[1:], userID, first.ID).
				Updates(map[string]interface{}{"status": "trashed", "decay_stage": 3})
			merged++
		}
		return nil
	})
	if err != nil {
		log.Printf("[NudgeReflect] compress transaction error: %v", err)
	}

	forgotten := 0
	for _, item := range forgetItems {
		f, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		if id, ok := f["memory_id"].(float64); ok {
			s.db.Model(&models.Memory{}).Where("id = ? AND user_id = ?", uint(id), userID).
				Updates(map[string]interface{}{"status": "trashed", "decay_stage": 3})
			forgotten++
		}
	}

	if len(profileUpdates) > 0 {
		s.updateUserProfile(userID, profileUpdates)
	}

	now := time.Now()
	s.db.Model(&models.UserProfile{}).Where("user_id = ?", userID).
		Updates(map[string]interface{}{"last_nudge_at": &now, "nudge_count": gorm.Expr("nudge_count + 1")})

	return map[string]interface{}{
		"persisted":       added,
		"merged":          merged,
		"forgotten":       forgotten,
		"insights":        result["insights"],
		"profile_updates": len(profileUpdates),
		"nudge_type":      "periodic_reflection",
	}, nil
}

func (s *AIService) SelfRefine(ctx context.Context, userID uint, isPro bool, pressureLevel string) (map[string]interface{}, error) {
	var memories []models.Memory
	s.db.Where("user_id = ? AND status != ?", userID, "trashed").Order("importance DESC, access_count DESC").Limit(200).Find(&memories)

	if len(memories) < 10 {
		return map[string]interface{}{
			"kept":       len(memories),
			"merged":     0,
			"demoted":    0,
			"archived":   0,
			"refinement": "skipped",
			"reason":     "too few memories to refine",
		}, nil
	}

	targetCount := len(memories)
	switch pressureLevel {
	case "high":
		targetCount = int(float64(len(memories)) * 0.4)
	case "medium":
		targetCount = int(float64(len(memories)) * 0.65)
	case "low":
		targetCount = int(float64(len(memories)) * 0.85)
	default:
		targetCount = int(float64(len(memories)) * 0.65)
	}
	if targetCount < 5 {
		targetCount = 5
	}

	memoriesJSON, _ := json.Marshal(memories)

	content, err := s.chatWithTemplate(ctx, userID, isPro, "self_refine", map[string]string{
		"Memories":      string(memoriesJSON),
		"TargetCount":   fmt.Sprintf("%d", targetCount),
		"CurrentCount":  fmt.Sprintf("%d", len(memories)),
		"PressureLevel": pressureLevel,
	})
	if err != nil {
		return nil, fmt.Errorf("self-refinement failed: %w", err)
	}

	cleaned := strings.TrimSpace(content)
	if idx := strings.Index(cleaned, "{"); idx > 0 {
		cleaned = cleaned[idx:]
	}
	if idx := strings.LastIndex(cleaned, "}"); idx >= 0 {
		cleaned = cleaned[:idx+1]
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
		return map[string]interface{}{
			"raw_response": content,
			"error":        "failed to parse refinement result",
		}, nil
	}

	kept := 0
	keepItems, _ := result["keep"].([]interface{})
	for _, item := range keepItems {
		k, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		if id, ok := k["memory_id"].(float64); ok {
			var mem models.Memory
			if err := s.db.Where("id = ? AND user_id = ?", uint(id), userID).First(&mem).Error; err == nil {
				mem.ReinforceCount++
				s.db.Save(&mem)
				kept++
			}
		}
	}

	merged := 0
	mergeItems, _ := result["merge"].([]interface{})
	err = s.db.Transaction(func(tx *gorm.DB) error {
		for _, item := range mergeItems {
			m, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			ids, _ := m["source_ids"].([]interface{})
			mergedKey, _ := m["merged_key"].(string)
			mergedValue, _ := m["merged_value"].(string)
			layer, _ := m["layer"].(string)
			if layer == "" {
				layer = "context"
			}
			importance := 0.5
			if imp, ok := m["importance"].(float64); ok {
				importance = imp
			}
			if len(ids) < 2 || mergedValue == "" {
				continue
			}
			var memIDs []uint
			for _, id := range ids {
				if f, ok := id.(float64); ok {
					memIDs = append(memIDs, uint(f))
				}
			}
			if len(memIDs) < 2 {
				continue
			}
			newMem := models.Memory{
				UserID:     userID,
				Layer:      layer,
				Key:        mergedKey,
				Value:      mergedValue,
				Importance: importance,
				Source:     "self_refine",
				Status:     "active",
			}
			if err := tx.Create(&newMem).Error; err == nil {
				tx.Model(&models.Memory{}).Where("id IN ? AND user_id = ?", memIDs, userID).
					Updates(map[string]interface{}{"status": "trashed", "decay_stage": 3})
				merged++
			}
		}
		return nil
	})
	if err != nil {
		log.Printf("[SelfRefine] merge transaction error: %v", err)
	}

	demoted := 0
	demoteItems, _ := result["demote"].([]interface{})
	for _, item := range demoteItems {
		d, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		if id, ok := d["memory_id"].(float64); ok {
			toLayer, _ := d["to_layer"].(string)
			if toLayer == "" {
				toLayer = "detail"
			}
			s.db.Model(&models.Memory{}).Where("id = ? AND user_id = ?", uint(id), userID).
				Update("layer", toLayer)
			demoted++
		}
	}

	archived := 0
	archiveItems, _ := result["archive"].([]interface{})
	for _, item := range archiveItems {
		a, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		if id, ok := a["memory_id"].(float64); ok {
			s.db.Model(&models.Memory{}).Where("id = ? AND user_id = ?", uint(id), userID).
				Updates(map[string]interface{}{"status": "archived", "decay_stage": 2})
			archived++
		}
	}

	s.db.Model(&models.UserProfile{}).Where("user_id = ?", userID).
		Update("total_refinements", gorm.Expr("total_refinements + 1"))

	return map[string]interface{}{
		"kept":           kept,
		"merged":         merged,
		"demoted":        demoted,
		"archived":       archived,
		"stats":          result["stats"],
		"refinement":     "completed",
		"pressure_level": pressureLevel,
	}, nil
}

func (s *AIService) BuildUserProfile(ctx context.Context, userID uint, isPro bool) (map[string]interface{}, error) {
	var memories []models.Memory
	s.db.Where("user_id = ? AND status != ? AND layer IN ?", userID, "trashed",
		[]string{"core", "context"}).Order("importance DESC").Limit(100).Find(&memories)

	var recentHistory []models.MemoryHistory
	s.db.Where("user_id = ?", userID).Order("created_at DESC").Limit(30).Find(&recentHistory)

	memoriesJSON, _ := json.Marshal(memories)
	activityJSON, _ := json.Marshal(recentHistory)

	var existingProfile models.UserProfile
	existingJSON := "{}"
	if err := s.db.Where("user_id = ?", userID).First(&existingProfile).Error; err == nil {
		ej, _ := json.Marshal(map[string]interface{}{
			"identity":      existingProfile.Identity,
			"communication": existingProfile.Communication,
			"workflow":      existingProfile.Workflow,
			"preferences":   existingProfile.Preferences,
			"patterns":      existingProfile.Patterns,
			"version":       existingProfile.ProfileVersion,
		})
		existingJSON = string(ej)
	}

	content, err := s.chatWithTemplate(ctx, userID, isPro, "user_profile_build", map[string]string{
		"Memories":        string(memoriesJSON),
		"RecentActivity":  string(activityJSON),
		"ExistingProfile": existingJSON,
	})
	if err != nil {
		return nil, fmt.Errorf("user profile build failed: %w", err)
	}

	cleaned := strings.TrimSpace(content)
	if idx := strings.Index(cleaned, "{"); idx > 0 {
		cleaned = cleaned[idx:]
	}
	if idx := strings.LastIndex(cleaned, "}"); idx >= 0 {
		cleaned = cleaned[:idx+1]
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
		return map[string]interface{}{
			"raw_response": content,
			"error":        "failed to parse profile result",
		}, nil
	}

	profileData, _ := result["profile"].(map[string]interface{})
	if profileData == nil {
		return result, nil
	}

	identityJSON, _ := json.Marshal(profileData["identity"])
	commJSON, _ := json.Marshal(profileData["communication_style"])
	workflowJSON, _ := json.Marshal(profileData["workflow"])
	prefsJSON, _ := json.Marshal(profileData["preferences"])
	patternsJSON, _ := json.Marshal(profileData["patterns"])
	growthJSON, _ := json.Marshal(profileData["growth_areas"])

	confidence := 0.5
	if conf, ok := result["confidence"].(float64); ok {
		confidence = conf
	}

	profile := models.UserProfile{
		UserID:        userID,
		Identity:      string(identityJSON),
		Communication: string(commJSON),
		Workflow:      string(workflowJSON),
		Preferences:   string(prefsJSON),
		Patterns:      string(patternsJSON),
		GrowthAreas:   string(growthJSON),
		Confidence:    confidence,
	}

	var existing models.UserProfile
	if err := s.db.Where("user_id = ?", userID).First(&existing).Error; err != nil {
		profile.ProfileVersion = 1
		s.db.Create(&profile)
	} else {
		profile.ID = existing.ID
		profile.ProfileVersion = existing.ProfileVersion + 1
		profile.NudgeCount = existing.NudgeCount
		profile.LastNudgeAt = existing.LastNudgeAt
		profile.TotalRefinements = existing.TotalRefinements
		s.db.Save(&profile)
	}

	return map[string]interface{}{
		"profile":         profileData,
		"profile_version": profile.ProfileVersion,
		"confidence":      confidence,
		"changes":         result["changes_from_previous"],
	}, nil
}

func (s *AIService) updateUserProfile(userID uint, updates []interface{}) {
	var profile models.UserProfile
	isNew := false
	if err := s.db.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		profile = models.UserProfile{UserID: userID, ProfileVersion: 1}
		isNew = true
	}

	for _, u := range updates {
		update, ok := u.(map[string]interface{})
		if !ok {
			continue
		}
		field, _ := update["field"].(string)
		newValue, _ := update["new_value"].(string)
		if field == "" || newValue == "" {
			continue
		}
		switch field {
		case "communication_style":
			profile.Communication = newValue
		case "tech_stack", "workflow":
			profile.Workflow = newValue
		case "preference":
			profile.Preferences = newValue
		}
	}

	if isNew {
		s.db.Create(&profile)
	} else {
		s.db.Save(&profile)
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

func (s *AIService) AISkillCreate(ctx context.Context, userID uint, isPro bool, patterns []map[string]interface{}) (map[string]interface{}, error) {
	var traces []models.ActionTrace
	s.db.Where("user_id = ?", userID).Order("created_at DESC").Limit(100).Find(&traces)

	var existingSkills []models.Skill
	s.db.Where("user_id = ? AND status = ?", userID, "active").Find(&existingSkills)

	var profile models.UserProfile
	profileJSON := "{}"
	if err := s.db.Where("user_id = ?", userID).First(&profile).Error; err == nil {
		pj, _ := json.Marshal(map[string]interface{}{
			"identity":    profile.Identity,
			"workflow":    profile.Workflow,
			"preferences": profile.Preferences,
		})
		profileJSON = string(pj)
	}

	patternsJSON, _ := json.Marshal(patterns)
	tracesJSON, _ := json.Marshal(traces[:min(len(traces), 50)])
	skillsJSON, _ := json.Marshal(existingSkills)

	content, err := s.chatWithTemplate(ctx, userID, isPro, "skill_create", map[string]string{
		"Patterns":       string(patternsJSON),
		"Traces":         string(tracesJSON),
		"ExistingSkills": string(skillsJSON),
		"UserProfile":    profileJSON,
	})
	if err != nil {
		return nil, fmt.Errorf("AI skill creation failed: %w", err)
	}

	cleaned := strings.TrimSpace(content)
	if idx := strings.Index(cleaned, "{"); idx > 0 {
		cleaned = cleaned[idx:]
	}
	if idx := strings.LastIndex(cleaned, "}"); idx >= 0 {
		cleaned = cleaned[:idx+1]
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
		return nil, fmt.Errorf("failed to parse AI skill response: %w", err)
	}

	skillData, ok := result["skill"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("AI response missing 'skill' field")
	}

	name, _ := skillData["name"].(string)
	if name == "" {
		name = "unnamed-skill"
	}

	var existing models.Skill
	if err := s.db.Where("user_id = ? AND name = ?", userID, name).First(&existing).Error; err == nil {
		return map[string]interface{}{
			"status":   "duplicate",
			"message":  fmt.Sprintf("Skill '%s' already exists (v%d)", name, existing.Version),
			"existing": existing,
		}, nil
	}

	desc, _ := skillData["description"].(string)
	triggerKW, _ := json.Marshal(skillData["trigger_keywords"])
	steps, _ := json.Marshal(skillData["steps"])
	params, _ := json.Marshal(skillData["parameters"])
	pitfalls, _ := json.Marshal(skillData["known_pitfalls"])
	verification, _ := skillData["verification"].(string)
	category, _ := skillData["category"].(string)
	tags, _ := json.Marshal(skillData["tags"])

	skill := models.Skill{
		UserID:          userID,
		Name:            name,
		Description:     desc,
		TriggerKeywords: string(triggerKW),
		Steps:           string(steps),
		Parameters:      string(params),
		KnownPitfalls:   string(pitfalls),
		Verification:    verification,
		SourceAgent:     "ai_created",
		Category:        category,
		Tags:            string(tags),
		AutoCreated:     true,
		Status:          "active",
	}

	if err := s.db.Create(&skill).Error; err != nil {
		return nil, fmt.Errorf("failed to save skill: %w", err)
	}

	return map[string]interface{}{
		"status":     "created",
		"skill":      skill,
		"confidence": result["confidence"],
		"reasoning":  result["reasoning"],
	}, nil
}

func (s *AIService) AISkillImprove(ctx context.Context, userID uint, isPro bool, skillID uint) (map[string]interface{}, error) {
	var skill models.Skill
	if err := s.db.Where("id = ? AND user_id = ?", skillID, userID).First(&skill).Error; err != nil {
		return nil, fmt.Errorf("skill not found: %w", err)
	}

	var recentTraces []models.ActionTrace
	s.db.Where("user_id = ? AND agent_name = ? AND created_at > ?", userID, skill.SourceAgent,
		time.Now().AddDate(0, 0, -14)).Order("created_at DESC").Limit(50).Find(&recentTraces)

	if len(recentTraces) == 0 && skill.UsageCount == 0 {
		return map[string]interface{}{
			"status":  "kept",
			"message": "No usage data available for improvement",
		}, nil
	}

	currentSkillJSON, _ := json.Marshal(skill)

	usageHistory := map[string]interface{}{
		"usage_count":   skill.UsageCount,
		"success_count": skill.SuccessCount,
		"fail_count":    skill.FailCount,
		"success_rate":  0.0,
	}
	if skill.UsageCount > 0 {
		usageHistory["success_rate"] = float64(skill.SuccessCount) / float64(skill.UsageCount)
	}
	usageJSON, _ := json.Marshal(usageHistory)

	var failTraces, successTraces []models.ActionTrace
	for _, t := range recentTraces {
		if t.Result == "failure" || t.Result == "error" {
			failTraces = append(failTraces, t)
		} else {
			successTraces = append(successTraces, t)
		}
	}
	failJSON, _ := json.Marshal(failTraces[:min(len(failTraces), 10)])
	successJSON, _ := json.Marshal(successTraces[:min(len(successTraces), 10)])
	recentJSON, _ := json.Marshal(recentTraces[:min(len(recentTraces), 30)])

	content, err := s.chatWithTemplate(ctx, userID, isPro, "skill_improve", map[string]string{
		"CurrentSkill":    string(currentSkillJSON),
		"UsageHistory":    string(usageJSON),
		"RecentFailures":  string(failJSON),
		"RecentSuccesses": string(successJSON),
		"RecentTraces":    string(recentJSON),
	})
	if err != nil {
		return nil, fmt.Errorf("AI skill improvement failed: %w", err)
	}

	cleaned := strings.TrimSpace(content)
	if idx := strings.Index(cleaned, "{"); idx > 0 {
		cleaned = cleaned[idx:]
	}
	if idx := strings.LastIndex(cleaned, "}"); idx >= 0 {
		cleaned = cleaned[:idx+1]
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
		return nil, fmt.Errorf("failed to parse AI improvement response: %w", err)
	}

	action, _ := result["action"].(string)
	if action == "keep" {
		return map[string]interface{}{
			"status":    "kept",
			"message":   "Skill is performing well, no changes needed",
			"changelog": result["changelog"],
		}, nil
	}

	improvedSkill, ok := result["improved_skill"].(map[string]interface{})
	if !ok {
		return map[string]interface{}{
			"status":  "error",
			"message": "AI response missing improved_skill data",
			"action":  action,
		}, nil
	}

	if desc, ok := improvedSkill["description"].(string); ok && desc != "" {
		skill.Description = desc
	}
	if triggerKW, ok := improvedSkill["trigger_keywords"].([]interface{}); ok {
		kwJSON, _ := json.Marshal(triggerKW)
		skill.TriggerKeywords = string(kwJSON)
	}
	if steps, ok := improvedSkill["steps"].([]interface{}); ok {
		stepsJSON, _ := json.Marshal(steps)
		skill.Steps = string(stepsJSON)
	}
	if pitfalls, ok := improvedSkill["known_pitfalls"].([]interface{}); ok {
		pitfallsJSON, _ := json.Marshal(pitfalls)
		skill.KnownPitfalls = string(pitfallsJSON)
	}
	if verification, ok := improvedSkill["verification"].(string); ok && verification != "" {
		skill.Verification = verification
	}

	skill.Version++
	skill.AutoCreated = false
	s.db.Save(&skill)

	return map[string]interface{}{
		"status":         "improved",
		"skill":          skill,
		"action":         action,
		"version_change": result["version_change"],
		"changelog":      result["changelog"],
	}, nil
}
