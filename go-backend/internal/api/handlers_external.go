package api

import (
	"clawmemory/internal/ai"
	"clawmemory/internal/middleware"
	"clawmemory/internal/models"
	"clawmemory/internal/services"
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// normalizeLayer maps legacy MCP layer names (episodic/semantic/procedural)
// to the backend's canonical layer names (core/context/detail).
// Canonical names pass through unchanged.
func normalizeLayer(layer string) string {
	switch layer {
	case "episodic":
		return "detail"
	case "semantic":
		return "context"
	case "procedural":
		return "core"
	default:
		return layer
	}
}

func handleExternalBatchPushConversations(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !middleware.HasAPIKeyPermission(c, "conversations:write") {
			c.JSON(http.StatusForbidden, gin.H{"error": "API key lacks 'conversations:write' permission"})
			return
		}
		userID := middleware.GetUserID(c)
		platform := middleware.GetPlatform(c)

		var req struct {
			Turns []services.ConversationPushRequest `json:"turns" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if len(req.Turns) > 50 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "too many turns (max 50)"})
			return
		}

		syncService := services.GetOpenClawSyncService(db)
		totalCreated := 0
		totalMessages := 0
		var errorDetails []string

		for _, turn := range req.Turns {
			if turn.AgentName == "" {
				turn.AgentName = "openclaw"
			}
			if turn.Platform == "" {
				turn.Platform = platform
			}
			created, err := syncService.PushConversation(userID, turn)
			if err != nil {
				errorDetails = append(errorDetails, fmt.Sprintf("session=%s: %v", turn.SessionID, err))
				continue
			}
			totalCreated += created
			totalMessages += len(turn.Messages)
		}

		auditLog(db, c, "external.conversation_batch_push", "", fmt.Sprintf("turns:%d messages:%d created:%d platform:%s", len(req.Turns), totalMessages, totalCreated, platform))

		c.JSON(http.StatusOK, gin.H{
			"turns":    len(req.Turns),
			"created":  totalCreated,
			"messages": totalMessages,
			"platform": platform,
			"errors":   len(errorDetails),
		})
	}
}

func handleExternalMemoryContext(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !middleware.HasAPIKeyPermission(c, "memories:read") {
			c.JSON(http.StatusForbidden, gin.H{"error": "API key lacks 'memories:read' permission"})
			return
		}
		userID := middleware.GetUserID(c)
		platform := middleware.GetPlatform(c)
		q := c.Query("q")
		if q == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "q is required"})
			return
		}
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "5"))
		if limit <= 0 {
			limit = 5
		}
		if limit > 50 {
			limit = 50
		}
		platformFilter := c.Query("platform")
		if platformFilter == "" {
			platformFilter = c.DefaultQuery("source", "")
		}

		svc := services.NewMemoryService(db)
		memories, err := svc.SearchKeywordWithPlatform(userID, q, platform, platformFilter, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var contextParts []string
		for _, m := range memories {
			if m.Key != "" && m.Value != "" {
				contextParts = append(contextParts, fmt.Sprintf("- %s: %s", m.Key, m.Value))
			}
		}

		systemPromptAddition := ""
		if len(contextParts) > 0 {
			systemPromptAddition = "\n\n[ClawMemory Relevant Memories]\n" + strings.Join(contextParts, "\n")
		}

		c.JSON(http.StatusOK, gin.H{
			"memories":               memories,
			"count":                  len(memories),
			"system_prompt_addition": systemPromptAddition,
		})
	}
}

func handleExternalSessionTrack(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !middleware.HasAPIKeyPermission(c, "sessions:write") {
			c.JSON(http.StatusForbidden, gin.H{"error": "API key lacks 'sessions:write' permission"})
			return
		}
		userID := middleware.GetUserID(c)

		var req struct {
			SessionID string `json:"session_id" binding:"required"`
			Metadata  string `json:"metadata"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var session models.SessionMemory
		result := db.Where("user_id = ? AND session_id = ?", userID, req.SessionID).First(&session)
		if result.Error == gorm.ErrRecordNotFound {
			session = models.SessionMemory{
				UserID:    userID,
				SessionID: req.SessionID,
				Title:     "OpenClaw Session",
				Status:    "active",
			}
			if req.Metadata != "" {
				session.CurrentState = req.Metadata
			}
			if err := db.Create(&session).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
				return
			}
		} else {
			if req.Metadata != "" {
				if err := db.Model(&session).Update("current_state", req.Metadata).Error; err != nil {
					log.Printf("[SessionTrack] failed to update metadata for session %s: %v", session.SessionID, err)
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"session_id": session.SessionID,
			"status":     session.Status,
			"tracked":    true,
		})
	}
}

func handleExternalAINudgeReflect(aiSvc *ai.AIService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !middleware.HasAPIKeyPermission(c, "ai:execute") {
			c.JSON(http.StatusForbidden, gin.H{"error": "API key lacks 'ai:execute' permission"})
			return
		}
		userID := middleware.GetUserID(c)

		ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
		defer cancel()

		result, err := aiSvc.NudgeReflect(ctx, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func handleExternalAIProcessConversation(aiSvc *ai.AIService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !middleware.HasAPIKeyPermission(c, "ai:execute") {
			c.JSON(http.StatusForbidden, gin.H{"error": "API key lacks 'ai:execute' permission"})
			return
		}
		userID := middleware.GetUserID(c)

		var req struct {
			Messages []map[string]string `json:"messages"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if len(req.Messages) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "messages are required"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
		defer cancel()

		result, err := aiSvc.ProcessConversation(ctx, userID, req.Messages)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func handleExternalSkillRecordAction(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !middleware.HasAPIKeyPermission(c, "ai:execute") {
			c.JSON(http.StatusForbidden, gin.H{"error": "API key lacks 'ai:execute' permission"})
			return
		}
		userID := middleware.GetUserID(c)

		var req struct {
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
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		svc := services.NewSkillLearningService(db)
		platform := req.Platform
		if platform == "" {
			platform = "openclaw"
		}

		if len(req.Actions) > 0 {
			created, err := svc.RecordActionBatch(userID, req.SessionID, req.AgentName, platform, req.Actions)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"created": created, "batch": true})
			return
		}

		if req.ActionType == "" || req.ActionName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "action_type and action_name are required"})
			return
		}

		if err := svc.RecordAction(userID, req.SessionID, req.AgentName, platform, req.ActionType, req.ActionName, req.Parameters, req.Result, req.Duration); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"recorded": true})
	}
}

func handleExternalCreateMemory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !middleware.HasAPIKeyPermission(c, "memories:write") {
			c.JSON(http.StatusForbidden, gin.H{"error": "API key lacks 'memories:write' permission"})
			return
		}
		userID := middleware.GetUserID(c)
		platform := middleware.GetPlatform(c)
		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		key, _ := data["key"].(string)
		value, _ := data["value"].(string)
		if key == "" || value == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "key and value are required"})
			return
		}

		if len(key) > 500 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "key too long (max 500 characters)"})
			return
		}
		if len(value) > 50000 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "value too long (max 50000 characters)"})
			return
		}

		source, _ := data["source"].(string)
		if source == "" {
			source = platform
		}

		svc := services.NewMemoryService(db)
		agentName := middleware.GetAgentName(c)
		reqSourceAgent, _ := data["source_agent"].(string)
		if reqSourceAgent == "" {
			reqSourceAgent = agentName
		}
		reqVisibility, _ := data["visibility"].(string)
		if reqVisibility == "" {
			reqVisibility = "private"
		}
		memory, err := svc.Create(userID, map[string]interface{}{
			"key":          key,
			"value":        value,
			"layer":        normalizeLayer(getString(data, "layer", "detail")),
			"importance":   data["importance"],
			"source":       source,
			"memory_type":  getString(data, "memory_type", "knowledge"),
			"platform":     platform,
			"source_agent": reqSourceAgent,
			"visibility":   reqVisibility,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		auditLog(db, c, "external.memory_create", key, fmt.Sprintf("source:%s platform:%s", source, platform))

		c.JSON(http.StatusCreated, memory)
	}
}

func handleExternalBatchCreateMemories(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !middleware.HasAPIKeyPermission(c, "memories:write") {
			c.JSON(http.StatusForbidden, gin.H{"error": "API key lacks 'memories:write' permission"})
			return
		}
		userID := middleware.GetUserID(c)
		platform := middleware.GetPlatform(c)
		var req struct {
			Memories []map[string]interface{} `json:"memories" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if len(req.Memories) > 100 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "batch size exceeds limit (max 100)"})
			return
		}

		var created, errorsCount int
		var errorDetails []string

		tx := db.Begin()
		if tx.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to begin transaction"})
			return
		}
		svcTx := services.NewMemoryService(tx)

		for _, m := range req.Memories {
			key, _ := m["key"].(string)
			value, _ := m["value"].(string)
			if key == "" || value == "" {
				errorsCount++
				continue
			}

			if len(key) > 500 || len(value) > 50000 {
				errorsCount++
				continue
			}

			source, _ := m["source"].(string)
			if source == "" {
				source = platform
			}

			agentName := middleware.GetAgentName(c)
			reqSourceAgent, _ := m["source_agent"].(string)
			if reqSourceAgent == "" {
				reqSourceAgent = agentName
			}
			reqVisibility, _ := m["visibility"].(string)
			if reqVisibility == "" {
				reqVisibility = "private"
			}

			_, err := svcTx.Create(userID, map[string]interface{}{
				"key":          key,
				"value":        value,
				"layer":        normalizeLayer(getString(m, "layer", "detail")),
				"importance":   m["importance"],
				"source":       source,
				"memory_type":  getString(m, "memory_type", "knowledge"),
				"platform":     platform,
				"source_agent": reqSourceAgent,
				"visibility":   reqVisibility,
			})
			if err != nil {
				errorsCount++
				errorDetails = append(errorDetails, fmt.Sprintf("key=%s: %v", key, err))
				continue
			}
			created++
		}

		if errorsCount > 0 && created == 0 {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "all items failed", "details": errorDetails})
			return
		}

		if err := tx.Commit().Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "transaction commit failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"created": created,
			"errors":  errorsCount,
			"total":   len(req.Memories),
		})
	}
}

func handleExternalSearchMemories(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !middleware.HasAPIKeyPermission(c, "memories:read") {
			c.JSON(http.StatusForbidden, gin.H{"error": "API key lacks 'memories:read' permission"})
			return
		}
		userID := middleware.GetUserID(c)
		platform := middleware.GetPlatform(c)
		q := c.Query("q")
		if q == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "q is required"})
			return
		}
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
		if limit <= 0 {
			limit = 10
		}
		if limit > 100 {
			limit = 100
		}
		platformFilter := c.Query("platform")
		if platformFilter == "" {
			platformFilter = c.DefaultQuery("source", "")
		}

		svc := services.NewMemoryService(db)
		memories, err := svc.SearchKeywordWithPlatform(userID, q, platform, platformFilter, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": memories})
	}
}
