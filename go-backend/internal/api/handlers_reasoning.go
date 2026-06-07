package api

import (
	"clawmemory/internal/middleware"
	"clawmemory/internal/models"
	"clawmemory/internal/services"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func handleExternalReason(db *gorm.DB, aiChat services.AIChatProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !middleware.HasAPIKeyPermission(c, "reason:execute") {
			c.JSON(http.StatusForbidden, gin.H{"error": "API key lacks 'reason:execute' permission"})
			return
		}
		userID := middleware.GetUserID(c)

		var req struct {
			Query     string `json:"query" binding:"required"`
			SessionID string `json:"session_id"`
			Depth     int    `json:"depth"`
			Level     string `json:"level"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		svc := services.NewReasoningService(db, aiChat)
		result, err := svc.Reason(userID, req.Query, req.Depth, req.Level)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		auditLog(db, c, "external.reason", req.SessionID, fmt.Sprintf("depth:%d level:%s", req.Depth, req.Level))

		c.JSON(http.StatusOK, gin.H{
			"reasoning": result,
			"depth":     req.Depth,
			"level":     req.Level,
		})
	}
}

func handleGetReasoningConfig(db *gorm.DB, aiChat services.AIChatProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		svc := services.NewReasoningService(db, aiChat)
		config, err := svc.GetConfig(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if config == nil {
			c.JSON(http.StatusOK, gin.H{"configured": false})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"configured":      true,
			"provider":        config.Provider,
			"model":           config.Model,
			"base_url":        config.BaseURL,
			"dialectic_depth": config.DialecticDepth,
			"reasoning_level": config.ReasoningLevel,
			"max_tokens":      config.MaxTokens,
			"enabled":         config.Enabled,
			"has_api_key":     config.APIKey != "",
		})
	}
}

func handleSetReasoningConfig(db *gorm.DB, aiChat services.AIChatProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		svc := services.NewReasoningService(db, aiChat)
		config, err := svc.SetConfig(userID, data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		auditLog(db, c, "reasoning.config_update", "", fmt.Sprintf("provider:%s model:%s", config.Provider, config.Model))

		c.JSON(http.StatusOK, gin.H{
			"provider":        config.Provider,
			"model":           config.Model,
			"base_url":        config.BaseURL,
			"dialectic_depth": config.DialecticDepth,
			"reasoning_level": config.ReasoningLevel,
			"max_tokens":      config.MaxTokens,
			"enabled":         config.Enabled,
		})
	}
}

func handleTestReasoningConnection(db *gorm.DB, aiChat services.AIChatProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		svc := services.NewReasoningService(db, aiChat)
		err := svc.TestConnection(userID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Connection successful"})
	}
}

func handleReason(db *gorm.DB, aiChat services.AIChatProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		var req struct {
			Query string `json:"query" binding:"required"`
			Depth int    `json:"depth"`
			Level string `json:"level"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		svc := services.NewReasoningService(db, aiChat)
		result, err := svc.Reason(userID, req.Query, req.Depth, req.Level)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		auditLog(db, c, "reasoning.execute", "", fmt.Sprintf("depth:%d level:%s", req.Depth, req.Level))

		c.JSON(http.StatusOK, gin.H{"reasoning": result})
	}
}

func handleExternalCreateSessionMemory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !middleware.HasAPIKeyPermission(c, "memories:write") {
			c.JSON(http.StatusForbidden, gin.H{"error": "API key lacks 'memories:write' permission"})
			return
		}
		userID := middleware.GetUserID(c)
		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		sessionID, _ := data["session_id"].(string)
		if sessionID == "" {
			sessionID = "default"
		}

		session := models.SessionMemory{
			UserID:       userID,
			SessionID:    sessionID,
			Title:        getString(data, "title", ""),
			CurrentState: getString(data, "current_state", ""),
			TaskSpec:     getString(data, "task_spec", ""),
			Worklog:      getString(data, "worklog", ""),
			Learnings:    getString(data, "learnings", ""),
			KeyResults:   getString(data, "key_results", ""),
			Docs:         getString(data, "docs", ""),
			Errors:       getString(data, "errors", ""),
			Workflow:     getString(data, "workflow", ""),
			Status:       getString(data, "status", "active"),
		}

		if v, ok := data["token_count"].(float64); ok {
			session.TokenCount = int(v)
		}
		if v, ok := data["compressed_from"].(string); ok {
			session.CompressedFrom = v
		}

		if err := db.Create(&session).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		auditLog(db, c, "external.session_memory_create", sessionID, "")

		c.JSON(http.StatusCreated, session)
	}
}

func handleExternalPushConversation(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !middleware.HasAPIKeyPermission(c, "conversations:write") {
			c.JSON(http.StatusForbidden, gin.H{"error": "API key lacks 'conversations:write' permission"})
			return
		}
		userID := middleware.GetUserID(c)
		platform := middleware.GetPlatform(c)

		var req services.ConversationPushRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if req.AgentName == "" {
			req.AgentName = "openclaw"
		}
		if req.Platform == "" {
			req.Platform = platform
		}

		if len(req.Messages) == 0 && req.Summary == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "messages or summary is required"})
			return
		}

		if len(req.Messages) > 200 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "too many messages (max 200)"})
			return
		}

		syncService := services.GetOpenClawSyncService(db)
		created, err := syncService.PushConversation(userID, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		auditLog(db, c, "external.conversation_push", req.SessionID, fmt.Sprintf("agent:%s platform:%s messages:%d created:%d", req.AgentName, req.Platform, len(req.Messages), created))

		c.JSON(http.StatusOK, gin.H{
			"created":  created,
			"messages": len(req.Messages),
			"agent":    req.AgentName,
			"platform": req.Platform,
		})
	}
}
