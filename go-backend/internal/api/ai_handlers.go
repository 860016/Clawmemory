package api

import (
	"context"
	"net/http"
	"strings"
	"time"

	"clawmemory/internal/ai"
	"clawmemory/internal/middleware"
	"clawmemory/internal/models"
	"clawmemory/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func handleAIConfig(aiRouter *ai.AIRouter, provider services.ProProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		isPro := provider.IsPro()
		config := aiRouter.GetCurrentUserConfig(userID, isPro)
		c.JSON(http.StatusOK, config)
	}
}

func handleAIConfigUpdate(aiRouter *ai.AIRouter, provider services.ProProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		if err := aiRouter.UpdateProConfig(userID, data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "AI configuration updated",
			"config":  aiRouter.GetCurrentUserConfig(userID, true),
		})
	}
}

func handleAITestConnection(aiRouter *ai.AIRouter, provider services.ProProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		isPro := provider.IsPro()

		result, err := aiRouter.TestConnection(userID, isPro)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func handleAIUsage(aiRouter *ai.AIRouter) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		stats := aiRouter.GetUsageStats(userID)
		c.JSON(http.StatusOK, stats)
	}
}

func handleAIProviders(provider services.ProProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"providers": ai.AllProviders,
			"is_pro":    true,
		})
	}
}

func handleAIExtract(aiSvc *ai.AIService, provider services.ProProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		isPro := provider.IsPro()

		ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
		defer cancel()

		result, err := aiSvc.AIExtract(ctx, userID, isPro)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func handleAIConflictScan(aiSvc *ai.AIService, provider services.ProProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
		defer cancel()

		result, err := aiSvc.ConflictScan(ctx, userID, true)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func handleAIDecayEvaluate(aiSvc *ai.AIService, provider services.ProProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
		defer cancel()

		result, err := aiSvc.DecayEvaluate(ctx, userID, true)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func handleAIDailyReport(aiSvc *ai.AIService, provider services.ProProvider, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		isPro := provider.IsPro()

		date := c.Query("date")
		if date == "" {
			date = time.Now().Format("2006-01-02")
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
		defer cancel()

		result, err := aiSvc.GenerateDailyReport(ctx, userID, isPro, date)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func handleAIWikiGenerate(aiSvc *ai.AIService, provider services.ProProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		var data struct {
			Topic string `json:"topic"`
		}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "topic is required"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
		defer cancel()

		result, err := aiSvc.GenerateWiki(ctx, userID, true, data.Topic)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func handleAICompress(aiSvc *ai.AIService, provider services.ProProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		var data struct {
			MemoryIDs []uint `json:"memory_ids"`
		}
		if err := c.ShouldBindJSON(&data); err != nil || len(data.MemoryIDs) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "memory_ids is required"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
		defer cancel()

		result, err := aiSvc.CompressMemories(ctx, userID, true, data.MemoryIDs)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func handleAIDiscoverRelations(aiSvc *ai.AIService, provider services.ProProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
		defer cancel()

		result, err := aiSvc.DiscoverRelations(ctx, userID, true)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func handleAISmartRoute(aiSvc *ai.AIService, provider services.ProProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		var data struct {
			Text string `json:"text"`
		}
		if err := c.ShouldBindJSON(&data); err != nil || data.Text == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "text is required"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
		defer cancel()

		result, err := aiSvc.SmartRoute(ctx, userID, true, data.Text)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func handleAIExtractFacts(aiSvc *ai.AIService, provider services.ProProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		isPro := provider.IsPro()

		var data struct {
			Messages []map[string]string `json:"messages"`
		}
		if err := c.ShouldBindJSON(&data); err != nil || len(data.Messages) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "messages array is required"})
			return
		}
		if len(data.Messages) > 200 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "too many messages (max 200)"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
		defer cancel()

		result, err := aiSvc.ExtractFacts(ctx, userID, isPro, data.Messages)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func handleAIConsolidate(aiSvc *ai.AIService, provider services.ProProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		isPro := provider.IsPro()

		var data struct {
			Facts []map[string]interface{} `json:"facts"`
		}
		if err := c.ShouldBindJSON(&data); err != nil || len(data.Facts) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "facts array is required"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
		defer cancel()

		result, err := aiSvc.ConsolidateMemories(ctx, userID, isPro, data.Facts)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func handleAIProcessConversation(aiSvc *ai.AIService, provider services.ProProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		isPro := provider.IsPro()

		var data struct {
			Messages []map[string]string `json:"messages"`
		}
		if err := c.ShouldBindJSON(&data); err != nil || len(data.Messages) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "messages array is required"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 180*time.Second)
		defer cancel()

		result, err := aiSvc.ProcessConversation(ctx, userID, isPro, data.Messages)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func handleAIAssembleContext(aiSvc *ai.AIService, provider services.ProProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		var data struct {
			Query       string `json:"query"`
			TokenBudget int    `json:"token_budget"`
		}
		if err := c.ShouldBindJSON(&data); err != nil || data.Query == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "query is required"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
		defer cancel()

		result, err := aiSvc.AssembleContext(ctx, userID, true, data.Query, data.TokenBudget)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func handleAINudgeReflect(aiSvc *ai.AIService, provider services.ProProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		isPro := provider.IsPro()

		ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
		defer cancel()

		result, err := aiSvc.NudgeReflect(ctx, userID, isPro)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func handleAISelfRefine(aiSvc *ai.AIService, provider services.ProProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		var data struct {
			PressureLevel string `json:"pressure_level"`
		}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if data.PressureLevel == "" {
			data.PressureLevel = "medium"
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
		defer cancel()

		result, err := aiSvc.SelfRefine(ctx, userID, true, data.PressureLevel)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func handleAIUserProfile(aiSvc *ai.AIService, provider services.ProProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
		defer cancel()

		result, err := aiSvc.BuildUserProfile(ctx, userID, true)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func handleSkillRecordAction(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		var data struct {
			SessionID  string                   `json:"session_id"`
			AgentName  string                   `json:"agent_name"`
			Platform   string                   `json:"platform"`
			Actions    []map[string]interface{} `json:"actions"`
			ActionType string                   `json:"action_type"`
			ActionName string                   `json:"action_name"`
			Parameters string                   `json:"parameters"`
			Result     string                   `json:"result"`
			Duration   int                      `json:"duration"`
		}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		svc := services.NewSkillLearningService(db)

		if len(data.Actions) > 0 {
			if len(data.Actions) > 100 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "too many actions (max 100)"})
				return
			}
			platform := data.Platform
			if platform == "" {
				platform = middleware.GetPlatform(c)
			}
			created, err := svc.RecordActionBatch(userID, data.SessionID, data.AgentName, platform, data.Actions)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"created": created, "batch": true})
			return
		}

		if data.ActionType == "" || data.ActionName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "action_type and action_name are required"})
			return
		}

		platform := data.Platform
		if platform == "" {
			platform = middleware.GetPlatform(c)
		}

		if err := svc.RecordAction(userID, data.SessionID, data.AgentName, platform, data.ActionType, data.ActionName, data.Parameters, data.Result, data.Duration); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"recorded": true})
	}
}

func handleSkillDetectPatterns(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		svc := services.NewSkillLearningService(db)
		patterns, err := svc.DetectPatterns(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"patterns": patterns,
			"count":    len(patterns),
		})
	}
}

func handleSkillAutoCreate(db *gorm.DB, aiSvc *ai.AIService, provider services.ProProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		isPro := provider.IsPro()

		svc := services.NewSkillLearningService(db)

		var data struct {
			UseAI    bool                     `json:"use_ai"`
			Patterns []map[string]interface{} `json:"patterns"`
		}
		_ = c.ShouldBindJSON(&data)

		if data.UseAI && isPro {
			patterns := data.Patterns
			if len(patterns) == 0 {
				detected, err := svc.DetectPatterns(userID)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				patterns = detected
			}

			ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
			defer cancel()

			result, err := aiSvc.AISkillCreate(ctx, userID, isPro, patterns)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, result)
			return
		}

		result, err := svc.DetectAndCreateSkills(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func handleSkillList(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		svc := services.NewSkillLearningService(db)
		stats := svc.GetSkillStats(userID)

		var skills []models.Skill
		statusFilter := c.Query("status")
		query := db.Where("user_id = ?", userID)
		if statusFilter != "" {
			query = query.Where("status = ?", statusFilter)
		} else {
			query = query.Where("status = ?", "active")
		}
		if err := query.Order("usage_count DESC, updated_at DESC").Find(&skills).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"skills": skills,
			"stats":  stats,
		})
	}
}

func handleSkillMatch(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		query := c.Query("q")
		if query == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'q' is required"})
			return
		}

		svc := services.NewSkillLearningService(db)
		skills, err := svc.MatchSkill(userID, query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"skills": skills,
			"query":  query,
			"count":  len(skills),
		})
	}
}

func handleSkillPatch(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		skillID := c.Param("id")

		var data struct {
			Field    string `json:"field"`
			OldValue string `json:"old_value"`
			NewValue string `json:"new_value"`
		}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		svc := services.NewSkillLearningService(db)
		var skill models.Skill
		if err := db.Where("id = ? AND user_id = ?", skillID, userID).First(&skill).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "skill not found"})
			return
		}

		if err := svc.PatchSkill(userID, skill.ID, data.Field, data.OldValue, data.NewValue); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"patched": true, "skill_id": skill.ID})
	}
}

func handleSkillImprove(db *gorm.DB, aiSvc *ai.AIService, provider services.ProProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		skillIDStr := c.Param("id")

		var skill models.Skill
		if err := db.Where("id = ? AND user_id = ?", skillIDStr, userID).First(&skill).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "skill not found"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
		defer cancel()

		result, err := aiSvc.AISkillImprove(ctx, userID, true, skill.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func handleSkillRecordUsage(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		skillIDStr := c.Param("id")

		var data struct {
			Success bool `json:"success"`
		}
		_ = c.ShouldBindJSON(&data)

		svc := services.NewSkillLearningService(db)
		var skill models.Skill
		if err := db.Where("id = ? AND user_id = ?", skillIDStr, userID).First(&skill).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "skill not found"})
			return
		}

		if err := svc.RecordSkillUsage(userID, skill.ID, data.Success); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"recorded": true})
	}
}

func handleSkillSuggestions(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		svc := services.NewSkillLearningService(db)

		path := c.Request.URL.Path
		if strings.Contains(path, "/dismiss") {
			sugIDStr := c.Param("id")
			var sug models.SkillSuggestion
			if err := db.Where("id = ? AND user_id = ?", sugIDStr, userID).First(&sug).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "suggestion not found"})
				return
			}
			if err := svc.DismissSuggestion(userID, sug.ID); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"dismissed": true})
			return
		}

		if strings.Contains(path, "/generate") {
			suggestions, err := svc.GenerateAgentSuggestions(userID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			for _, sug := range suggestions {
				svc.SaveSuggestion(userID, sug)
			}
			c.JSON(http.StatusOK, gin.H{
				"suggestions": suggestions,
				"count":       len(suggestions),
			})
			return
		}

		suggestions, err := svc.GetPendingSuggestions(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"suggestions": suggestions,
			"count":       len(suggestions),
		})
	}
}
