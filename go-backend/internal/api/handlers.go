package api

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"clawmemory/internal/ai"
	"clawmemory/internal/config"
	"clawmemory/internal/middleware"
	"clawmemory/internal/models"
	"clawmemory/internal/services"

	"golang.org/x/crypto/bcrypt"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func logDBErr(context string, err error) {
	if err != nil {
		log.Printf("[DB] %s: %v", context, err)
	}
}

func truncateStr(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}

func parseIDParam(c *gin.Context, name string) (int, bool) {
	id, err := strconv.Atoi(c.Param(name))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid " + name})
		return 0, false
	}
	return id, true
}

func handleInitStatus(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		passwordSet, err := authService.CheckInitStatus()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check init status"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"password_set": passwordSet,
		})
	}
}

func handleSetPassword(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if len(req.Password) < 6 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "password must be at least 6 characters"})
			return
		}

		username := req.Username
		if username == "" {
			username = "admin"
		}

		result, err := authService.SetPasswordWithUsername(username, req.Password)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		resp := gin.H{
			"access_token":  result.AccessToken,
			"refresh_token": result.RefreshToken,
		}
		if result.APIKey != "" {
			resp["api_key"] = result.APIKey
		}

		c.JSON(http.StatusOK, resp)
	}
}

func handleLogin(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password" binding:"required"`
			Captcha  string `json:"captcha"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if req.Username == "" {
			req.Username = "admin"
		}

		result, err := authService.Login(req.Username, req.Password, req.Captcha)
		if err != nil {
			status := http.StatusUnauthorized
			resp := gin.H{"error": err.Error()}
			if result != nil {
				resp["requires_captcha"] = result.RequiresCaptcha
				if result.LockedUntil != nil {
					resp["locked_until"] = result.LockedUntil
				}
			}
			c.JSON(status, resp)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"access_token":  result.AccessToken,
			"refresh_token": result.RefreshToken,
		})
	}
}

func handleGetMe(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		user, err := authService.GetUserByID(userID.(uint))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"role":       user.Role,
			"is_founder": user.IsFounder,
		})
	}
}

func handleRegister(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Username       string `json:"username" binding:"required"`
			Password       string `json:"password" binding:"required"`
			InvitationCode string `json:"invitation_code"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result, err := authService.RegisterWithInvitation(req.Username, req.Password, req.InvitationCode)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		resp := gin.H{
			"id":         result.User.ID,
			"username":   result.User.Username,
			"role":       result.User.Role,
			"is_founder": result.User.IsFounder,
		}
		if result.APIKey != "" {
			resp["api_key"] = result.APIKey
		}

		c.JSON(http.StatusCreated, resp)
	}
}

func handleResetPassword(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			OldPassword string `json:"old_password"`
			NewPassword string `json:"new_password"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if req.OldPassword == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "old_password is required"})
			return
		}
		if req.NewPassword == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "new_password is required"})
			return
		}
		if len(req.NewPassword) < 6 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "new_password must be at least 6 characters"})
			return
		}

		userID := middleware.GetUserID(c)
		if err := authService.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "password reset successfully"})
	}
}

func handleForgotPassword(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "password reset is only available via terminal command",
			"hint":  "Run: ./clawmemory --reset-password NEW_PASSWORD",
		})
	}
}

func handleChangePassword(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		var req struct {
			OldPassword string `json:"old_password"`
			NewPassword string `json:"new_password"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if req.OldPassword == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "old_password is required"})
			return
		}
		if req.NewPassword == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "new_password is required"})
			return
		}
		if len(req.NewPassword) < 6 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "new_password must be at least 6 characters"})
			return
		}
		if err := authService.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "password changed successfully"})
	}
}

func handleRefreshToken(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			RefreshToken string `json:"refresh_token" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "refresh_token is required"})
			return
		}

		result, err := authService.RefreshAccessToken(req.RefreshToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"access_token":  result.AccessToken,
			"refresh_token": result.RefreshToken,
		})
	}
}

func handleLoginStatus(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Username string `json:"username"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if req.Username == "" {
			c.JSON(http.StatusOK, gin.H{
				"requires_captcha": false,
				"locked":           false,
			})
			return
		}

		locked, lockedUntil, failedAttempts := authService.IsAccountLocked(req.Username)
		requiresCaptcha := failedAttempts >= services.MaxFailedAttempts

		resp := gin.H{
			"requires_captcha": requiresCaptcha,
			"locked":           locked,
			"failed_attempts":  failedAttempts,
		}
		if lockedUntil != nil {
			resp["locked_until"] = lockedUntil
		}
		c.JSON(http.StatusOK, resp)
	}
}

func handleInstallStatus(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"checks": gin.H{
				"security_engine": "go",
			},
			"version": config.AppVersion,
		})
	}
}

// Memory handlers
func handleListMemories(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		svc := services.NewMemoryService(db)
		layer := c.Query("layer")
		status := c.Query("status")
		memoryType := c.Query("memory_type")
		sourceAgent := c.Query("source_agent")
		source := c.Query("source")
		visibility := c.Query("visibility")
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
		if page < 1 {
			page = 1
		}
		if size < 1 || size > 100 {
			size = 20
		}

		memories, total, err := svc.List(userID, layer, page, size, status, memoryType, sourceAgent, visibility, source)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"items": memories,
			"total": total,
			"page":  page,
			"size":  size,
			"pages": (total + int64(size) - 1) / int64(size),
		})
	}
}

func handleCreateMemory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		key, _ := req["key"].(string)
		value, _ := req["value"].(string)
		if key == "" || value == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "key and value are required"})
			return
		}

		secretResult := services.ScanSecrets(key + " " + value)

		svc := services.NewMemoryService(db)
		memory, err := svc.Create(userID, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		validationSvc := services.NewValidationService(db)
		validation := validationSvc.ValidateDTO(memory.Key, memory.Value, memory.Layer, memory.Importance)
		if validation.Status != "valid" {
			logDBErr("update validation status on create", db.Model(&models.Memory{}).Where("id = ?", memory.ID).
				Update("validation_status", validation.Status).Error)
		}

		response := gin.H{"memory": memory}
		if secretResult.Found {
			response["secret_warning"] = secretResult
		}
		if validation.Status != "valid" {
			response["validation"] = validation
		}
		c.JSON(http.StatusCreated, response)
	}
}

func handleGetMemory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}
		svc := services.NewMemoryService(db)
		memory, err := svc.Get(userID, uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, memory)
	}
}

func handleUpdateMemory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var secretResult *services.SecretScanResult
		if value, ok := req["value"].(string); ok {
			key, _ := req["key"].(string)
			secretResult = services.ScanSecrets(key + " " + value)
		}

		svc := services.NewMemoryService(db)
		memory, err := svc.Update(userID, uint(id), req)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}

		validationSvc := services.NewValidationService(db)
		validation := validationSvc.ValidateDTO(memory.Key, memory.Value, memory.Layer, memory.Importance)
		logDBErr("update validation status on update", db.Model(&models.Memory{}).Where("id = ?", memory.ID).
			Update("validation_status", validation.Status).Error)

		response := gin.H{"memory": memory}
		if secretResult != nil && secretResult.Found {
			response["secret_warning"] = secretResult
		}
		if validation.Status != "valid" {
			response["validation"] = validation
		}
		c.JSON(http.StatusOK, response)
	}
}

func handleDeleteMemory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}
		svc := services.NewMemoryService(db)
		if err := svc.Delete(userID, uint(id)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	}
}

func handleRestoreMemory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}
		svc := services.NewMemoryService(db)
		if err := svc.Restore(userID, uint(id)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "restored"})
	}
}

func handleSearchMemories(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		q := c.Query("q")
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		if limit <= 0 || limit > 200 {
			limit = 20
		}
		mode := c.DefaultQuery("mode", "keyword")
		userID := middleware.GetUserID(c)

		switch mode {
		case "semantic":
			chromaSvc := services.NewChromaDBService(db)
			if chromaSvc.IsAvailable() {
				results, err := chromaSvc.Search(userID, q, limit)
				if err == nil && len(results) > 0 {
					enriched := enrichChromaResults(db, userID, results, limit)
					c.JSON(http.StatusOK, gin.H{"items": enriched, "engine": "chromadb"})
					return
				}
			}
			svc := services.NewSearchService(db)
			memories, err := svc.SemanticSearch(userID, q, limit)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"items": memories, "engine": "tfidf"})
		case "graph-rag":
			if q == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter q is required"})
				return
			}
			svc := services.NewSearchService(db)
			results, err := svc.GraphRAGSearch(userID, q, limit)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"items":  results,
				"engine": "graph_rag",
				"mode":   "keyword+semantic+graph",
			})
		default:
			svc := services.NewMemoryService(db)
			memories, err := svc.SearchKeyword(userID, q, limit)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"items": memories})
		}
	}
}

func handleMemoryHistory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		memoryIDStr := c.Param("id")
		memoryID, err := strconv.ParseUint(memoryIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid memory id"})
			return
		}

		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

		var history []models.MemoryHistory
		if err := db.Where("user_id = ? AND memory_id = ?", userID, memoryID).
			Order("created_at DESC").Limit(limit).Find(&history).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"items": history,
			"total": len(history),
		})
	}
}

func handleMemoryEvolution(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		memoryIDStr := c.Param("id")
		memoryID, err := strconv.ParseUint(memoryIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid memory id"})
			return
		}

		var memory models.Memory
		if err := db.Where("id = ? AND user_id = ?", memoryID, userID).First(&memory).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "memory not found"})
			return
		}

		var history []models.MemoryHistory
		logDBErr("load memory history", db.Where("user_id = ? AND memory_id = ?", userID, memoryID).
			Order("created_at ASC").Find(&history).Error)

		var relatedEntities []models.Entity
		escapedKey := services.EscapeLikeQuery(memory.Key)
		logDBErr("load related entities", db.Where("user_id = ? AND (name LIKE ? OR description LIKE ?)",
			userID, "%"+escapedKey+"%", "%"+escapedKey+"%").Limit(20).Find(&relatedEntities).Error)

		var relatedRelations []models.Relation
		if len(relatedEntities) > 0 {
			entityIDs := make([]uint, len(relatedEntities))
			for i, e := range relatedEntities {
				entityIDs[i] = e.ID
			}
			logDBErr("load related relations", db.Where("user_id = ? AND (source_id IN ? OR target_id IN ?)",
				userID, entityIDs, entityIDs).Limit(50).Find(&relatedRelations).Error)
		}

		type evolutionStep struct {
			Timestamp  string `json:"timestamp"`
			ChangeType string `json:"change_type"`
			OldValue   string `json:"old_value"`
			NewValue   string `json:"new_value"`
			Reason     string `json:"reason"`
			Source     string `json:"source"`
		}

		steps := make([]evolutionStep, 0, len(history)+1)
		steps = append(steps, evolutionStep{
			Timestamp:  memory.CreatedAt.Format("2006-01-02 15:04:05"),
			ChangeType: "created",
			OldValue:   "",
			NewValue:   memory.Value,
			Reason:     "initial creation",
			Source:     memory.Source,
		})

		for _, h := range history {
			steps = append(steps, evolutionStep{
				Timestamp:  h.CreatedAt.Format("2006-01-02 15:04:05"),
				ChangeType: h.ChangeType,
				OldValue:   h.OldValue,
				NewValue:   h.NewValue,
				Reason:     h.Reason,
				Source:     h.SourceAgent,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"memory":            memory,
			"evolution_steps":   steps,
			"total_changes":     len(history),
			"related_entities":  relatedEntities,
			"related_relations": relatedRelations,
		})
	}
}

func handleExternalReason(db *gorm.DB) gin.HandlerFunc {
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

		svc := services.NewReasoningService(db)
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

func handleGetReasoningConfig(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		svc := services.NewReasoningService(db)
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

func handleSetReasoningConfig(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		svc := services.NewReasoningService(db)
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

func handleTestReasoningConnection(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		svc := services.NewReasoningService(db)
		err := svc.TestConnection(userID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Connection successful"})
	}
}

func handleReason(db *gorm.DB) gin.HandlerFunc {
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

		svc := services.NewReasoningService(db)
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

func handleGetConnectedAgents(c *gin.Context) {
	installed := services.DetectInstalledClients()

	var connected []map[string]interface{}
	for _, client := range installed {
		connected = append(connected, map[string]interface{}{
			"name":         client["name"],
			"display_name": client["display_name"],
			"status":       "connected",
			"found_dirs":   client["found_dirs"],
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"agents": connected,
		"total":  len(connected),
	})
}

func handleOpenClawSyncStatus(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		syncService := services.GetOpenClawSyncService(db)
		status := syncService.GetStatus()
		c.JSON(http.StatusOK, status)
	}
}

func handleOpenClawSyncForce(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		syncService := services.GetOpenClawSyncService(db)
		count := syncService.ForceSync()
		c.JSON(http.StatusOK, gin.H{
			"message":      "sync completed",
			"synced_count": count,
		})
	}
}

func handleOpenClawSyncToggle(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Enabled bool `json:"enabled"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "enabled field required"})
			return
		}

		syncService := services.GetOpenClawSyncService(db)
		syncService.SetAutoSync(req.Enabled)

		status := "enabled"
		if !req.Enabled {
			status = "disabled"
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "auto-sync " + status,
			"enabled": req.Enabled,
		})
	}
}

func handleGetAgentsMD(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		scheme := "http"
		if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		host := c.Request.Host
		if host == "" {
			host = "localhost:8765"
		}
		baseURL := scheme + "://" + host

		var keys []models.APIKey
		logDBErr("load api keys", db.Where("revoked_at IS NULL").Order("created_at DESC").Find(&keys).Error)

		hasKey := len(keys) > 0
		apiKeyHint := "cm_your_api_key_here"
		if hasKey {
			apiKeyHint = keys[0].KeyPrefix + "..."
		}

		content := `## 🧠 ClawMemory Auto-Record

每次回复结束时，自动将当次对话写入 ClawMemory。步骤如下：

1. 读取当前 session 的 jsonl 文件：` + "`~/.openclaw/agents/main/sessions/`" + ` 下的 ` + "`.jsonl`" + ` 文件（找最新修改的）
2. 提取用户消息和我的回复
3. 追加写到 ` + "`~/.openclaw/workspace/memory/auto-conversation.yml`" + `（按时间追加，用 ` + "`---`" + ` 分隔）
4. ClawMemory 的 auto-sync 服务每60秒自动扫描并导入

## 🧠 直接读写 ClawMemory（记忆后端）

ClawMemory 是我的长期记忆后端，我可以直接通过 API 读写它。

### 连接信息
- **地址**: ` + baseURL + `
- **API Key**: ` + apiKeyHint + `
- **初始密码**: ` + "`admin123`（首次登录后必须修改）" + `

### 何时读记忆
- 收到需要回忆的任务时，先搜索 ClawMemory：
` + "```" + `
curl -s "` + baseURL + `/api/v1/external/memories/search?q=关键词" -H "X-API-Key: <your-api-key>"
` + "```" + `
- 替代原有 memory_search 工具，因为 ClawMemory 支持语义搜索，更强

### 何时写记忆
- 每次回复末尾，自动将本轮对话（用户消息 + 我的回复）写入 ClawMemory
- 使用 API：` + "`POST /api/v1/external/conversations`" + `
` + "```" + `
curl -X POST "` + baseURL + `/api/v1/external/conversations" \
  -H "X-API-Key: <your-api-key>" \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "<当前session-id>",
    "agent_name": "openclaw",
    "messages": [
      {"role": "user", "content": "<用户消息>"},
      {"role": "assistant", "content": "<我的回复>"}
    ],
    "summary": "<本轮对话摘要>"
  }'
` + "```" + `
- 不需要等文件同步，直接写，即刻生效
`

		c.JSON(http.StatusOK, gin.H{
			"content":  content,
			"base_url": baseURL,
			"has_key":  hasKey,
		})
	}
}

func handleDecryptMemory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}

		svc := services.NewMemoryService(db)
		memory, err := svc.Get(userID, uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}

		if !memory.IsEncrypted {
			c.JSON(http.StatusOK, gin.H{"value": memory.Value, "encrypted": false})
			return
		}

		secretKey := services.GetEncryptionKey()
		if secretKey == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "SECRET_KEY not configured, cannot decrypt"})
			return
		}

		encryptor, err := services.NewEncryptor(secretKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to init decryptor"})
			return
		}

		decrypted, err := services.DecryptValue(encryptor, memory.Value)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "decryption failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"value": decrypted, "encrypted": true})
	}
}

func enrichChromaResults(db *gorm.DB, userID uint, chromaResults []map[string]interface{}, limit int) []map[string]interface{} {
	memIDs := make([]uint, 0, len(chromaResults))
	for _, r := range chromaResults {
		if memIDStr, ok := r["memory_id"].(string); ok {
			var id uint
			fmt.Sscanf(memIDStr, "%d", &id)
			if id > 0 {
				memIDs = append(memIDs, id)
			}
		}
	}

	if len(memIDs) == 0 {
		return chromaResults
	}

	var memories []models.Memory
	logDBErr("load memories by ids", db.Where("user_id = ? AND id IN ?", userID, memIDs).Find(&memories).Error)

	memMap := make(map[uint]models.Memory)
	for _, m := range memories {
		memMap[m.ID] = m
	}

	enriched := make([]map[string]interface{}, 0, len(chromaResults))
	for _, r := range chromaResults {
		if memIDStr, ok := r["memory_id"].(string); ok {
			var id uint
			fmt.Sscanf(memIDStr, "%d", &id)
			if m, found := memMap[id]; found {
				var tags []string
				if m.Tags != "" {
					json.Unmarshal([]byte(m.Tags), &tags)
				}
				if tags == nil {
					tags = []string{}
				}
				item := map[string]interface{}{
					"id":              m.ID,
					"key":             m.Key,
					"value":           m.Value,
					"layer":           m.Layer,
					"importance":      m.Importance,
					"source":          m.Source,
					"status":          m.Status,
					"summary":         m.Summary,
					"tags":            tags,
					"memory_type":     m.MemoryType,
					"decay_stage":     m.DecayStage,
					"reinforce_count": m.ReinforceCount,
					"access_count":    m.AccessCount,
					"is_encrypted":    m.IsEncrypted,
					"created_at":      m.CreatedAt.Format("2006-01-02 15:04:05"),
					"updated_at":      m.UpdatedAt.Format("2006-01-02 15:04:05"),
				}
				if m.VerifiedAt != nil {
					item["verified_at"] = m.VerifiedAt.Format("2006-01-02 15:04:05")
				}
				if m.TrashedAt != nil {
					item["trashed_at"] = m.TrashedAt.Format("2006-01-02 15:04:05")
				}
				if m.LastAccessedAt != nil {
					item["last_accessed_at"] = m.LastAccessedAt.Format("2006-01-02 15:04:05")
				}
				if score, ok := r["score"].(float64); ok {
					item["score"] = math.Round(score*1000) / 1000
				}
				enriched = append(enriched, item)
			}
		}
	}

	return enriched
}

// Knowledge handlers
func handleListEntities(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		svc := services.NewKnowledgeService(db)
		entityType := c.Query("type")
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
		if page < 1 {
			page = 1
		}
		if size < 1 || size > 100 {
			size = 20
		}

		entities, total, err := svc.ListEntities(userID, entityType, page, size)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": entities, "total": total, "page": page, "size": size, "pages": (total + int64(size) - 1) / int64(size)})
	}
}

func handleCreateEntity(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		svc := services.NewKnowledgeService(db)
		entity, err := svc.CreateEntity(userID, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, entity)
	}
}

func handleListRelations(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		svc := services.NewKnowledgeService(db)
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		size, _ := strconv.Atoi(c.DefaultQuery("size", "100"))
		if page < 1 {
			page = 1
		}
		if size < 1 || size > 200 {
			size = 100
		}

		relations, total, err := svc.ListRelations(userID, page, size)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": relations, "total": total, "page": page, "size": size, "pages": (total + int64(size) - 1) / int64(size)})
	}
}

func handleCreateRelation(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		svc := services.NewKnowledgeService(db)
		relation, err := svc.CreateRelation(userID, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, relation)
	}
}

func handleGetGraph(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		svc := services.NewKnowledgeService(db)
		entities, relations, err := svc.GetGraph(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"entities":  entities,
			"relations": relations,
		})
	}
}

// Wiki handlers
func handleListWiki(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		svc := services.NewWikiService(db)
		category := c.Query("category")
		status := c.Query("status")
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		size, _ := strconv.Atoi(c.DefaultQuery("size", "100"))
		if page < 1 {
			page = 1
		}
		if size < 1 || size > 200 {
			size = 100
		}

		pages, total, err := svc.List(userID, category, status, page, size)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": pages, "total": total, "page": page, "size": size, "pages": (total + int64(size) - 1) / int64(size)})
	}
}

func handleGetEntity(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		svc := services.NewKnowledgeService(db)
		entity, err := svc.GetEntity(userID, uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "entity not found"})
			return
		}
		c.JSON(http.StatusOK, entity)
	}
}

func handleUpdateEntity(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		svc := services.NewKnowledgeService(db)
		entity, err := svc.UpdateEntity(userID, uint(id), req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, entity)
	}
}

func handleDeleteEntity(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		svc := services.NewKnowledgeService(db)
		if err := svc.DeleteEntity(userID, uint(id)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	}
}

func handleDeleteRelation(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		svc := services.NewKnowledgeService(db)
		if err := svc.DeleteRelation(userID, uint(id)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	}
}

func handleCreateWiki(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		svc := services.NewWikiService(db)
		page, err := svc.Create(userID, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, page)
	}
}

func handleGetWiki(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}
		svc := services.NewWikiService(db)
		page, err := svc.Get(userID, uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, page)
	}
}

func handleUpdateWiki(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		svc := services.NewWikiService(db)
		page, err := svc.Update(userID, uint(id), req)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, page)
	}
}

func handleDeleteWiki(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}
		svc := services.NewWikiService(db)
		if err := svc.Delete(userID, uint(id)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	}
}

func handleWikiCategories(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		svc := services.NewWikiService(db)
		categories, err := svc.GetCategories(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, categories)
	}
}

func handleWikiStats(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		svc := services.NewWikiService(db)
		stats, err := svc.GetStats(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, stats)
	}
}

func handleWikiSearch(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		q := c.Query("q")
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		svc := services.NewWikiService(db)
		pages, err := svc.Search(userID, q, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, pages)
	}
}

func handleWikiTree(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		svc := services.NewWikiService(db)
		pages, err := svc.GetTree(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, pages)
	}
}

func handleWikiMarkComplete(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}
		svc := services.NewWikiService(db)
		if err := svc.MarkComplete(userID, uint(id)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "marked complete"})
	}
}

func handleWikiMarkInProgress(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}
		svc := services.NewWikiService(db)
		if err := svc.MarkInProgress(userID, uint(id)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "marked in progress"})
	}
}

func handleWikiConfig(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		svc := services.NewSettingsService(db)
		llmAvailable := false
		if v, err := svc.GetByKey(userID, "ai_provider_id"); err == nil {
			if s, ok := v.(string); ok && s != "" {
				llmAvailable = true
			}
		}
		c.JSON(http.StatusOK, gin.H{"llm_available": llmAvailable})
	}
}

func handleWikiAIExtract(aiSvc *ai.AIService, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		var req struct {
			Conversation string `json:"conversation"`
			IsComplete   bool   `json:"is_complete"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "conversation is required"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
		defer cancel()

		result, err := aiSvc.ExtractFacts(ctx, userID, []map[string]string{
			{"role": "user", "content": req.Conversation},
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		wikiSvc := services.NewWikiService(db)
		facts, _ := result["facts"].([]map[string]interface{})
		prefs, _ := result["preferences"].([]map[string]interface{})
		created := 0
		for _, fact := range facts {
			content, _ := fact["content"].(string)
			if content == "" {
				continue
			}
			title, _ := fact["title"].(string)
			if title == "" {
				cat, _ := fact["category"].(string)
				if cat != "" {
					title = cat + ": " + truncateStr(content, 40)
				} else {
					title = truncateStr(content, 60)
				}
			}
			pageData := map[string]interface{}{
				"title":   title,
				"content": content,
				"status":  "in_progress",
			}
			if cat, ok := fact["category"].(string); ok {
				pageData["category"] = cat
			}
			if _, err := wikiSvc.Create(userID, pageData); err == nil {
				created++
			}
		}
		for _, pref := range prefs {
			topic, _ := pref["topic"].(string)
			value, _ := pref["value"].(string)
			if topic == "" || value == "" {
				continue
			}
			title := "Preference: " + topic
			pageData := map[string]interface{}{
				"title":    title,
				"content":  topic + " → " + value,
				"status":   "in_progress",
				"category": "preferences",
			}
			if _, err := wikiSvc.Create(userID, pageData); err == nil {
				created++
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"extracted": len(facts),
			"created":   created,
			"mode":      "ai",
		})
	}
}

func handleWikiRefine(aiSvc *ai.AIService, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || id == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page id"})
			return
		}

		wikiSvc := services.NewWikiService(db)
		page, err := wikiSvc.Get(userID, uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "page not found"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
		defer cancel()

		result, err := aiSvc.GenerateWiki(ctx, userID, page.Title)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if content, ok := result["content"].(string); ok && content != "" {
			updates := map[string]interface{}{"content": content}
			if summary, ok := result["summary"].(string); ok && summary != "" {
				updates["summary"] = summary
			}
			if category, ok := result["category"].(string); ok && category != "" {
				updates["category"] = category
			}
			if tags, ok := result["tags"].([]interface{}); ok && len(tags) > 0 {
				tagStrs := make([]string, 0, len(tags))
				for _, t := range tags {
					if s, ok := t.(string); ok {
						tagStrs = append(tagStrs, s)
					}
				}
				if len(tagStrs) > 0 {
					updates["tags"] = strings.Join(tagStrs, ",")
				}
			}
			updates["ai_generated"] = true
			if _, err := wikiSvc.Update(userID, uint(id), updates); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save refined content"})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"message": "page refined",
				"content": content,
				"mode":    "ai",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "AI refinement produced no content"})
	}
}

// Report handlers
func handleListReports(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		svc := services.NewDailyReportService(db)
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
		if page < 1 {
			page = 1
		}
		if size < 1 || size > 100 {
			size = 20
		}

		reports, total, err := svc.List(userID, page, size)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": reports, "total": total, "page": page, "size": size, "pages": (total + int64(size) - 1) / int64(size)})
	}
}

func handleCreateReport(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		svc := services.NewDailyReportService(db)
		report, err := svc.Create(userID, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, report)
	}
}

func handleGetReportByDate(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		date := c.Param("date")
		svc := services.NewDailyReportService(db)
		report, err := svc.GetByDate(userID, date)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, report)
	}
}

func handleGenerateReport(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			req = map[string]interface{}{}
		}
		date, _ := req["date"].(string)
		if date == "" {
			date = time.Now().Format("2006-01-02")
		}

		svc := services.NewDailyReportService(db)
		report, err := svc.Generate(userID, date)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, report)
	}
}

// Stats handlers
func handleGetStats(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		var memoryCount, entityCount, relationCount, projectCount int64
		logDBErr("count memories for stats", db.Model(&struct{ ID uint }{}).Table("memories").Where("user_id = ? AND status != ?", userID, "trashed").Count(&memoryCount).Error)
		logDBErr("count entities for stats", db.Model(&struct{ ID uint }{}).Table("entities").Where("user_id = ?", userID).Count(&entityCount).Error)
		logDBErr("count relations for stats", db.Model(&struct{ ID uint }{}).Table("relations").Where("user_id = ?", userID).Count(&relationCount).Error)
		logDBErr("count projects for stats", db.Model(&struct{ ID uint }{}).Table("projects").Where("user_id = ?", userID).Count(&projectCount).Error)

		layerStats := make(map[string]int64)
		rows, err := db.Raw("SELECT COALESCE(layer, 'knowledge') as layer, COUNT(*) as cnt FROM memories WHERE user_id = ? AND status != 'trashed' GROUP BY layer", userID).Rows()
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var layer string
				var cnt int64
				rows.Scan(&layer, &cnt)
				layerStats[layer] = cnt
			}
		}

		if len(layerStats) == 0 {
			layerStats["knowledge"] = 0
		}

		type RecentMemory struct {
			ID        uint      `json:"id"`
			Key       string    `json:"key"`
			Layer     string    `json:"layer"`
			CreatedAt time.Time `json:"created_at"`
		}
		var recentMemories []RecentMemory
		logDBErr("load recent memories for dashboard", db.Table("memories").Where("user_id = ? AND status != ?", userID, "trashed").Order("created_at desc").Limit(10).Find(&recentMemories).Error)

		recentMemoriesJson := make([]map[string]interface{}, 0)
		for _, m := range recentMemories {
			recentMemoriesJson = append(recentMemoriesJson, map[string]interface{}{
				"id":         m.ID,
				"key":        m.Key,
				"layer":      m.Layer,
				"created_at": m.CreatedAt.Format("2006-01-02 15:04:05"),
			})
		}

		var userCount int64
		logDBErr("count users for dashboard", db.Table("users").Count(&userCount).Error)
		passwordSet := userCount > 0

		maxMemories := int64(50000)
		settingsSvc := services.NewSettingsService(db)
		if v, err := settingsSvc.GetByKey(userID, "max_memories"); err == nil {
			if n, ok := v.(float64); ok && n > 0 {
				maxMemories = int64(n)
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"memoryCount":    memoryCount,
			"entityCount":    entityCount,
			"relationCount":  relationCount,
			"projectCount":   projectCount,
			"layerStats":     layerStats,
			"recentMemories": recentMemoriesJson,
			"passwordSet":    passwordSet,
			"maxMemories":    maxMemories,
		})
	}
}

func handleGetSettings(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		svc := services.NewSettingsService(db)
		userID := middleware.GetUserID(c)
		settings, err := svc.Get(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, settings)
	}
}

func handleUpdateSettings(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		svc := services.NewSettingsService(db)
		userID := middleware.GetUserID(c)
		if err := svc.Update(userID, req); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "settings updated"})
	}
}

func handleToolboxConflictScan(aiSvc *ai.AIService, toolbox *services.ToolboxService, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
		defer cancel()

		result, err := aiSvc.ConflictScan(ctx, userID)
		if err != nil {
			result, err = toolbox.ConflictScan(userID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			result["mode"] = "local_fallback"
		}

		c.JSON(http.StatusOK, result)
	}
}

func handleToolboxConflictResolve(toolbox *services.ToolboxService, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		var req struct {
			Strategy string `json:"strategy"`
		}
		c.ShouldBindJSON(&req)
		indexStr := c.Param("index")
		index, err := strconv.Atoi(indexStr)
		if err != nil || index < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conflict index"})
			return
		}
		result, err := toolbox.ConflictResolve(userID, index, req.Strategy)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleToolboxTokenRoute(aiSvc *ai.AIService, toolbox *services.ToolboxService, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		message := c.Query("message")
		contextLength := 0
		if cl := c.Query("context_length"); cl != "" {
			if n, err := strconv.Atoi(cl); err == nil {
				contextLength = n
			}
		}

		if message != "" {
			ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
			defer cancel()

			result, err := aiSvc.SmartRoute(ctx, middleware.GetUserID(c), message)
			if err == nil {
				c.JSON(http.StatusOK, result)
				return
			}
		}

		result, err := toolbox.TokenRoute(message, contextLength)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		result["mode"] = "local_fallback"
		c.JSON(http.StatusOK, result)
	}
}

func handleToolboxTokenStats(toolbox *services.ToolboxService, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		result, err := toolbox.TokenStats(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleToolboxAIExtract(aiSvc *ai.AIService, toolbox *services.ToolboxService, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
		defer cancel()

		result, err := aiSvc.AIExtract(ctx, userID)
		if err != nil {
			result, err = toolbox.ExtractEntities(userID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			result["mode"] = "local_fallback"
		}

		c.JSON(http.StatusOK, result)
	}
}

func handleToolboxAutoGraph(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		var req struct {
			Overwrite bool `json:"overwrite"`
		}
		c.ShouldBindJSON(&req)
		evolutionSvc := services.NewEvolutionService(db)
		result, err := evolutionSvc.AutoGraph(userID, req.Overwrite)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleToolboxCompressPreview(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Level string `json:"level"`
		}
		c.ShouldBindJSON(&req)
		if req.Level == "" {
			req.Level = "light"
		}
		userID := middleware.GetUserID(c)
		svc := services.NewDecayService(db)
		result, err := svc.CompressPreview(userID, req.Level)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleToolboxCompressApply(aiSvc *ai.AIService, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Level   string                 `json:"level"`
			Options map[string]interface{} `json:"options"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if req.Level == "" {
			req.Level = "light"
		}
		userID := middleware.GetUserID(c)

		svc := services.NewDecayService(db)
		preview, previewErr := svc.CompressPreview(userID, req.Level)
		if previewErr == nil {
			if previewItems, ok := preview["preview"].([]map[string]interface{}); ok && len(previewItems) > 0 {
				memoryIDs := make([]uint, 0, len(previewItems))
				for _, item := range previewItems {
					if id, ok := item["memory_id"]; ok {
						switch v := id.(type) {
						case uint:
							memoryIDs = append(memoryIDs, v)
						case float64:
							memoryIDs = append(memoryIDs, uint(v))
						case int:
							memoryIDs = append(memoryIDs, uint(v))
						}
					}
				}

				if len(memoryIDs) > 0 {
					ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
					defer cancel()

					result, err := aiSvc.CompressMemories(ctx, userID, memoryIDs)
					if err == nil {
						result["level"] = req.Level
						result["mode"] = "ai"
						c.JSON(http.StatusOK, result)
						return
					}
				}
			}
		}

		result, err := svc.CompressApply(userID, req.Level)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		result["mode"] = "local_fallback"
		c.JSON(http.StatusOK, result)
	}
}

func handleToolboxCompressConfig(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		svc := services.NewSettingsService(db)
		cfg := map[string]interface{}{
			"auto_compress":      false,
			"threshold":          1000,
			"level":              "light",
			"preserve_important": true,
		}
		if v, err := svc.GetByKey(userID, "compress_config"); err == nil && v != nil {
			if m, ok := v.(map[string]interface{}); ok {
				for k, val := range m {
					cfg[k] = val
				}
			}
		}
		c.JSON(http.StatusOK, map[string]interface{}{"config": cfg})
	}
}

func handleToolboxSetCompressConfig(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		userID := middleware.GetUserID(c)
		svc := services.NewSettingsService(db)
		_ = svc.SetByKey(userID, "compress_config", req)
		c.JSON(http.StatusOK, map[string]interface{}{"updated": true, "config": req})
	}
}

func handleEvolutionInsights(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		svc := services.NewEvolutionService(db)
		result, err := svc.Insights(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleEvolutionRun(aiSvc *ai.AIService, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		var req struct {
			Action  string `json:"action"`
			Context string `json:"context"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		svc := services.NewEvolutionService(db)

		switch req.Action {
		case "discover":
			ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
			defer cancel()
			result, err := aiSvc.DiscoverRelations(ctx, userID)
			if err != nil {
				result, err = svc.Discover(userID)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				result["mode"] = "local_fallback"
			}
			c.JSON(http.StatusOK, result)
		case "infer":
			result, err := svc.HighConfidenceEntities(userID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, result)
		case "importance":
			result, err := svc.ImportanceBuckets(userID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, result)
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid action, must be one of: discover, infer, importance"})
		}
	}
}

func handleGetUsageStats(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
		if days < 1 {
			days = 30
		}
		if days > 365 {
			days = 365
		}

		var memories []struct {
			ID         uint      `json:"id"`
			Key        string    `json:"key"`
			Layer      string    `json:"layer"`
			Source     string    `json:"source"`
			Importance float64   `json:"importance"`
			CreatedAt  time.Time `json:"created_at"`
		}
		logDBErr("load memories for export", db.Table("memories").Where("user_id = ? AND status != ?", userID, "trashed").Order("created_at desc").Find(&memories).Error)

		now := time.Now()

		dailyTrend := make([]map[string]interface{}, 0)
		for i := days - 1; i >= 0; i-- {
			date := now.AddDate(0, 0, -i).Format("2006-01-02")
			dayStart, _ := time.Parse("2006-01-02", date)
			dayEnd := dayStart.AddDate(0, 0, 1)
			count := 0
			for _, m := range memories {
				if m.CreatedAt.After(dayStart) && m.CreatedAt.Before(dayEnd) {
					count++
				}
			}
			dailyTrend = append(dailyTrend, map[string]interface{}{
				"date":  date,
				"count": count,
			})
		}

		sourceDist := make(map[string]int)
		importanceDist := make(map[string]int)
		layerDist := make(map[string]int)
		entityTypeDist := make(map[string]int)

		for _, m := range memories {
			layer := m.Layer
			if layer == "" {
				layer = "knowledge"
			}
			source := m.Source
			if source == "" {
				source = "manual"
			}
			sourceDist[source]++

			if m.Importance >= 0.7 {
				importanceDist["high"]++
			} else if m.Importance >= 0.3 {
				importanceDist["medium"]++
			} else {
				importanceDist["low"]++
			}
			layerDist[layer]++
		}

		var entityCount int64
		logDBErr("count entities for import", db.Table("entities").Where("user_id = ?", userID).Count(&entityCount).Error)

		rows, err := db.Raw("SELECT entity_type, COUNT(*) as cnt FROM entities WHERE user_id = ? GROUP BY entity_type", userID).Rows()
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var etype string
				var cnt int64
				rows.Scan(&etype, &cnt)
				entityTypeDist[etype] = int(cnt)
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"dailyTrend":             dailyTrend,
			"dailyTokenTrend":        []map[string]interface{}{},
			"sourceDistribution":     sourceDist,
			"importanceDistribution": importanceDist,
			"tokenByLayer":           layerDist,
			"totalEstimatedTokens":   len(memories) * 100,
			"topAccessed":            []map[string]interface{}{},
			"operationCounts":        map[string]int{},
			"entityTypeDistribution": entityTypeDist,
			"totalMemories":          len(memories),
			"days":                   days,
		})
	}
}

func handleScanSkills(c *gin.Context) {
	dataDirs := []string{}
	seenDirs := make(map[string]bool)

	addDir := func(d string) {
		abs, err := filepath.Abs(d)
		if err != nil {
			abs = d
		}
		if !seenDirs[abs] {
			seenDirs[abs] = true
			dataDirs = append(dataDirs, abs)
		}
	}

	cfg := config.Load()
	addDir(cfg.SkillsDir)

	exe, _ := os.Executable()
	if exe != "" {
		addDir(filepath.Join(filepath.Dir(exe), "skills"))
	}

	for _, d := range services.GetAllSkillsDirs() {
		addDir(d)
	}

	homeDir, _ := os.UserHomeDir()
	if homeDir != "" {
		addDir(filepath.Join(homeDir, ".agents", "skills"))
	}

	wd, _ := os.Getwd()
	if wd != "" {
		addDir(filepath.Join(wd, "skills"))
		addDir(filepath.Join(wd, ".agents", "skills"))
	}

	globalSkills := make([]map[string]interface{}, 0)
	workspaceSkills := make([]map[string]interface{}, 0)

	globalDir, _ := filepath.Abs(cfg.SkillsDir)

	for _, dir := range dataDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			skillFiles := []string{
				filepath.Join(dir, entry.Name(), "skill.json"),
				filepath.Join(dir, entry.Name(), "skill.yaml"),
				filepath.Join(dir, entry.Name(), "skill.yml"),
				filepath.Join(dir, entry.Name(), "SKILL.md"),
			}
			var content []byte
			var skillFile string
			for _, sf := range skillFiles {
				if data, err := os.ReadFile(sf); err == nil {
					content = data
					skillFile = sf
					break
				}
			}
			if content == nil {
				continue
			}
			var skill map[string]interface{}
			ext := filepath.Ext(skillFile)
			baseName := filepath.Base(skillFile)
			if ext == ".json" {
				json.Unmarshal(content, &skill)
			} else if ext == ".yaml" || ext == ".yml" {
				if parsed, err := parseYAML(content); err == nil {
					skill = parsed
				}
			} else if baseName == "SKILL.md" {
				if parsed, err := parseSKILLMd(content); err == nil {
					skill = parsed
				}
			}
			if skill == nil {
				continue
			}
			skill["skill_dir"] = entry.Name()
			absDir, _ := filepath.Abs(dir)
			if absDir == globalDir {
				skill["scope"] = "global"
				globalSkills = append(globalSkills, skill)
			} else {
				skill["scope"] = "workspace"
				workspaceSkills = append(workspaceSkills, skill)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"global_skills":    globalSkills,
		"workspace_skills": workspaceSkills,
		"clients":          services.DetectInstalledClients(),
	})
}

func handleInstallSkill(c *gin.Context) {
	var req struct {
		RepoURL string `json:"repo_url" binding:"required"`
		Scope   string `json:"scope"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "repo_url is required"})
		return
	}
	if req.Scope == "" {
		req.Scope = "global"
	}

	cfg := config.Load()
	targetDir := cfg.SkillsDir

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create skills directory"})
		return
	}

	repoURL := req.RepoURL
	if !strings.HasPrefix(repoURL, "https://") && !strings.HasPrefix(repoURL, "git@") {
		repoURL = "https://github.com/" + repoURL
	}

	repoName := repoURL
	if idx := strings.LastIndex(repoName, "/"); idx >= 0 {
		repoName = repoName[idx+1:]
	}
	if strings.HasSuffix(repoName, ".git") {
		repoName = repoName[:len(repoName)-4]
	}

	destPath := filepath.Join(targetDir, repoName)
	if _, err := os.Stat(destPath); err == nil {
		c.JSON(http.StatusOK, gin.H{
			"message":   "skill already installed",
			"skill_dir": repoName,
			"path":      destPath,
		})
		return
	}

	cmd := exec.Command("git", "clone", "--depth", "1", repoURL, destPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "failed to clone repository: " + string(output),
			"detail": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "skill installed successfully",
		"skill_dir": repoName,
		"path":      destPath,
	})
}

func parseYAML(data []byte) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			if strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"") {
				val = val[1 : len(val)-1]
			}
			result[key] = val
		}
	}
	return result, nil
}

func parseSKILLMd(data []byte) (map[string]interface{}, error) {
	content := string(data)
	result := make(map[string]interface{})

	if strings.HasPrefix(content, "---") {
		endIdx := strings.Index(content[3:], "---")
		if endIdx >= 0 {
			frontmatter := content[3 : endIdx+3]
			lines := strings.Split(frontmatter, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					val := strings.TrimSpace(parts[1])
					if strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"") {
						val = val[1 : len(val)-1]
					}
					result[key] = val
				}
			}
			bodyStart := endIdx + 6
			if bodyStart < len(content) {
				result["body_full"] = strings.TrimSpace(content[bodyStart:])
			}
		}
	} else {
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "# ") {
				result["name"] = strings.TrimSpace(line[2:])
				break
			}
		}
		result["body_full"] = content
	}

	if _, ok := result["name"]; !ok {
		result["name"] = "unknown"
	}

	return result, nil
}

func handleSkillDetail(c *gin.Context) {
	skillDir := c.Query("skill_dir")
	scope := c.Query("scope")

	searchDirs := []string{}
	if scope == "global" {
		searchDirs = append(searchDirs, services.GetAllSkillsDirs()...)
	}

	cfg := config.Load()
	if cfg.SkillsDir != "" {
		searchDirs = append(searchDirs, cfg.SkillsDir)
	}
	if cfg.DataDir != "" {
		searchDirs = append(searchDirs, filepath.Join(cfg.DataDir, "skills"))
	}

	exe, _ := os.Executable()
	if exe != "" {
		searchDirs = append(searchDirs, filepath.Join(filepath.Dir(exe), "skills"))
	}

	wd, _ := os.Getwd()
	if wd != "" {
		searchDirs = append(searchDirs, filepath.Join(wd, "skills"))
	}

	for _, baseDir := range searchDirs {
		skillFiles := []string{
			filepath.Join(baseDir, skillDir, "skill.json"),
			filepath.Join(baseDir, skillDir, "skill.yaml"),
			filepath.Join(baseDir, skillDir, "skill.yml"),
			filepath.Join(baseDir, skillDir, "SKILL.md"),
		}
		for _, skillFile := range skillFiles {
			content, err := os.ReadFile(skillFile)
			if err != nil {
				continue
			}
			var skill map[string]interface{}
			ext := filepath.Ext(skillFile)
			baseName := filepath.Base(skillFile)
			if ext == ".json" {
				if err := json.Unmarshal(content, &skill); err != nil {
					continue
				}
			} else if ext == ".yaml" || ext == ".yml" {
				if parsed, err := parseYAML(content); err != nil {
					continue
				} else {
					skill = parsed
				}
			} else if baseName == "SKILL.md" {
				if parsed, err := parseSKILLMd(content); err != nil {
					continue
				} else {
					skill = parsed
				}
			}
			if skill != nil {
				skill["skill_dir"] = skillDir
				skill["scope"] = scope
				c.JSON(http.StatusOK, skill)
				return
			}
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "skill not found"})
}

func handleChromaDBStatus(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		svc := services.NewChromaDBService(db)
		status := svc.GetStatus()
		c.JSON(http.StatusOK, status)
	}
}

func handleChromaDBInstall(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		svc := services.NewChromaDBService(db)
		result := svc.Install()
		c.JSON(http.StatusOK, result)
	}
}

func handleChromaDBSync(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		svc := services.NewChromaDBService(db)
		count, err := svc.SyncMemories(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"synced": count, "message": fmt.Sprintf("Synced %d memories to ChromaDB", count)})
	}
}

func handleSmartLoad(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		query := c.Query("q")
		tokenBudget, _ := strconv.Atoi(c.DefaultQuery("token_budget", "2000"))
		loadLevel := c.DefaultQuery("load_level", "auto")

		svc := services.NewSmartLoadService(db)
		result, err := svc.SmartLoad(userID, query, tokenBudget, loadLevel)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleReinforceMemory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}

		svc := services.NewSmartLoadService(db)
		if err := svc.ReinforceMemory(userID, uint(id)); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "memory not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "memory reinforced", "memory_id": id})
	}
}

func handleGenerateSummaries(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		svc := services.NewSmartLoadService(db)
		count, err := svc.BatchGenerateSummaries(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"generated": count, "message": fmt.Sprintf("Generated summaries for %d memories", count)})
	}
}

func handleMCPConfig(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		apiKeySvc := services.NewAPIKeyService(db)
		keys, err := apiKeySvc.List(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var rawKey string
		if len(keys) == 0 {
			_, rk, err := apiKeySvc.Create(userID, "mcp-server")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create API key: " + err.Error()})
				return
			}
			rawKey = rk
		}

		scheme := "http"
		if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		host := c.GetHeader("X-Forwarded-Host")
		if host == "" {
			host = c.Request.Host
		}
		baseURL := fmt.Sprintf("%s://%s", scheme, host)

		if rawKey == "" {
			for _, k := range keys {
				if k.Name == "mcp-server" {
					rawKey = k.KeyPrefix + "••••••••"
					break
				}
			}
			if rawKey == "" && len(keys) > 0 {
				rawKey = keys[0].KeyPrefix + "••••••••"
			}
		}

		mcpServerConfig := map[string]interface{}{
			"command": "npx",
			"args":    []string{"-y", "clawmemory-mcp"},
			"env": map[string]string{
				"CLAWMEMORY_BASE_URL": baseURL,
				"CLAWMEMORY_API_KEY":  rawKey,
			},
		}

		cursorConfig := map[string]interface{}{
			"mcpServers": map[string]interface{}{
				"clawmemory": mcpServerConfig,
			},
		}

		c.JSON(http.StatusOK, gin.H{
			"base_url":    baseURL,
			"api_key":     rawKey,
			"has_new_key": len(keys) == 0,
			"configs": map[string]interface{}{
				"cursor": map[string]interface{}{
					"label":      "Cursor",
					"configPath": "~/.cursor/mcp.json",
					"config":     cursorConfig,
				},
				"claude_desktop": map[string]interface{}{
					"label":      "Claude Desktop",
					"configPath": "~/AppData/Roaming/Claude/claude_desktop_config.json",
					"config":     cursorConfig,
				},
				"windsurf": map[string]interface{}{
					"label":      "Windsurf",
					"configPath": "~/.windsurf/mcp.json",
					"config":     cursorConfig,
				},
				"trae": map[string]interface{}{
					"label":      "Trae",
					"configPath": "~/.trae/mcp.json",
					"config":     cursorConfig,
				},
			},
		})
	}
}

func handleGovernanceStatus(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		svc := services.NewGovernanceService(db)
		status := svc.GetStatus(userID)
		c.JSON(http.StatusOK, status)
	}
}

func handleGovernanceRun(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		svc := services.NewGovernanceService(db)
		result, err := svc.RunFullGovernance(userID)
		if err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleGovernanceConfig(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		var config services.GovernanceConfig
		if err := c.ShouldBindJSON(&config); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		svc := services.NewGovernanceService(db)
		if err := svc.UpdateConfig(userID, config); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "governance config updated"})
	}
}

func getOpenClawSearchDirs() []string {
	return services.GetAllSearchDirs()
}

type memoryPreview struct {
	Key       string `json:"key"`
	Content   string `json:"content"`
	Layer     string `json:"layer"`
	Source    string `json:"source"`
	FilePath  string `json:"file_path"`
	AgentName string `json:"agent_name"`
}

func extractMemoriesFromDir(dir string) ([]memoryPreview, map[string]int) {
	var previews []memoryPreview
	agentCountMap := make(map[string]int)

	memFile := filepath.Join(dir, "MEMORY.md")
	if data, err := os.ReadFile(memFile); err == nil && len(data) > 0 {
		previews, agentCountMap = parseMarkdownMemory(string(data), memFile, "workspace", previews, agentCountMap)
	}

	claudeMdFile := filepath.Join(dir, "CLAUDE.md")
	if data, err := os.ReadFile(claudeMdFile); err == nil && len(data) > 0 {
		previews, agentCountMap = parseMarkdownMemory(string(data), claudeMdFile, "claude", previews, agentCountMap)
	}

	memoryDir := filepath.Join(dir, "memory")
	if info, err := os.Stat(memoryDir); err == nil && info.IsDir() {
		previews, agentCountMap = extractWorkspaceMemory(memoryDir, dir, previews, agentCountMap)
	}

	memoriesDir := filepath.Join(dir, "memories")
	if info, err := os.Stat(memoriesDir); err == nil && info.IsDir() {
		previews, agentCountMap = extractWorkspaceMemory(memoriesDir, dir, previews, agentCountMap)
	}

	sessionsDir := filepath.Join(dir, "agents")
	if info, err := os.Stat(sessionsDir); err == nil && info.IsDir() {
		previews, agentCountMap = extractSessionMemories(sessionsDir, previews, agentCountMap)
	}

	projectsDir := filepath.Join(dir, "projects")
	if info, err := os.Stat(projectsDir); err == nil && info.IsDir() {
		previews, agentCountMap = extractClaudeProjects(projectsDir, previews, agentCountMap)
	}

	sessionsDir2 := filepath.Join(dir, "sessions")
	if info, err := os.Stat(sessionsDir2); err == nil && info.IsDir() {
		previews, agentCountMap = extractSessionDir(sessionsDir2, previews, agentCountMap)
	}

	dataDir := filepath.Join(dir, "data")
	if info, err := os.Stat(dataDir); err == nil && info.IsDir() {
		previews, agentCountMap = extractWorkspaceMemory(dataDir, dir, previews, agentCountMap)
	}

	sqliteFiles, _ := filepath.Glob(filepath.Join(dir, "*.db"))
	sqliteFiles2, _ := filepath.Glob(filepath.Join(dir, "*.sqlite"))
	sqliteFiles3, _ := filepath.Glob(filepath.Join(dir, "*.sqlite3"))
	allSqlite := append(append(sqliteFiles, sqliteFiles2...), sqliteFiles3...)
	for _, dbFile := range allSqlite {
		previews, agentCountMap = extractSqliteMemories(dbFile, previews, agentCountMap)
	}

	vscdbFiles, _ := filepath.Glob(filepath.Join(dir, "*.vscdb"))
	for _, vscdbFile := range vscdbFiles {
		previews, agentCountMap = extractVscdbMemories(vscdbFile, previews, agentCountMap)
	}

	wsStorageDir := filepath.Join(dir, "User", "workspaceStorage")
	if info, err := os.Stat(wsStorageDir); err == nil && info.IsDir() {
		previews, agentCountMap = extractWorkspaceStorage(wsStorageDir, previews, agentCountMap)
	}

	return previews, agentCountMap
}

func extractSessionMemories(sessionsDir string, previews []memoryPreview, agentCountMap map[string]int) ([]memoryPreview, map[string]int) {
	agentDirs, _ := os.ReadDir(sessionsDir)
	for _, ad := range agentDirs {
		if !ad.IsDir() {
			continue
		}
		agentName := ad.Name()
		sessDir := filepath.Join(sessionsDir, agentName, "sessions")
		if info, err := os.Stat(sessDir); err != nil || !info.IsDir() {
			continue
		}
		files, _ := os.ReadDir(sessDir)
		for _, f := range files {
			if f.IsDir() {
				continue
			}
			ext := strings.ToLower(filepath.Ext(f.Name()))
			if ext != ".jsonl" {
				continue
			}
			path := filepath.Join(sessDir, f.Name())
			data, err := os.ReadFile(path)
			if err != nil || len(data) == 0 {
				continue
			}
			pvs, cnt := parseJSONLSession(string(data), path, agentName)
			previews = append(previews, pvs...)
			agentCountMap[agentName] += cnt
		}
	}
	return previews, agentCountMap
}

func parseJSONLSession(content string, filePath string, agentName string) ([]memoryPreview, int) {
	var previews []memoryPreview
	count := 0
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var msg map[string]interface{}
		if json.Unmarshal([]byte(line), &msg) != nil {
			continue
		}
		msgType, _ := msg["type"].(string)
		text, _ := msg["text"].(string)
		if text == "" {
			text, _ = msg["content"].(string)
		}
		if text == "" {
			continue
		}

		isUserMsg := (msgType == "user" || msgType == "human")
		role := msgType
		if role == "" {
			role, _ = msg["role"].(string)
			if role == "" {
				role = "unknown"
			}
		}

		key := role + ": "
		if len(text) > 40 {
			key += text[:40] + "..."
		} else {
			key += text
		}

		preview := text
		if len(preview) > 200 {
			preview = preview[:200] + "..."
		}

		layer := "episodic"
		if isUserMsg {
			layer = "episodic"
		} else {
			layer = "semantic"
		}

		source := "session"
		if strings.Contains(filePath, "sessions") {
			source = "openclaw-session"
		}

		previews = append(previews, memoryPreview{
			Key: key, Content: preview, Layer: layer,
			Source: source, FilePath: filePath, AgentName: agentName,
		})
		count++
	}
	return previews, count
}

func extractWorkspaceMemory(memoryDir string, baseDir string, previews []memoryPreview, agentCountMap map[string]int) ([]memoryPreview, map[string]int) {
	files, _ := os.ReadDir(memoryDir)
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(f.Name()))
		if ext != ".md" {
			continue
		}
		path := filepath.Join(memoryDir, f.Name())
		data, err := os.ReadFile(path)
		if err != nil || len(data) == 0 {
			continue
		}
		previews, agentCountMap = parseMarkdownMemory(string(data), path, "workspace", previews, agentCountMap)
	}
	return previews, agentCountMap
}

func extractWorkspaceStorage(wsDir string, previews []memoryPreview, agentCountMap map[string]int) ([]memoryPreview, map[string]int) {
	entries, err := os.ReadDir(wsDir)
	if err != nil {
		return previews, agentCountMap
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		subDir := filepath.Join(wsDir, entry.Name())
		vscdbFiles, _ := filepath.Glob(filepath.Join(subDir, "*.vscdb"))
		for _, vscdbFile := range vscdbFiles {
			previews, agentCountMap = extractVscdbMemories(vscdbFile, previews, agentCountMap)
		}
		dbFiles, _ := filepath.Glob(filepath.Join(subDir, "*.db"))
		for _, dbFile := range dbFiles {
			previews, agentCountMap = extractSqliteMemories(dbFile, previews, agentCountMap)
		}
	}
	return previews, agentCountMap
}

func extractClaudeProjects(projectsDir string, previews []memoryPreview, agentCountMap map[string]int) ([]memoryPreview, map[string]int) {
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return previews, agentCountMap
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		projectDir := filepath.Join(projectsDir, entry.Name())
		files, err := os.ReadDir(projectDir)
		if err != nil {
			continue
		}
		for _, f := range files {
			if f.IsDir() {
				continue
			}
			ext := strings.ToLower(filepath.Ext(f.Name()))
			if ext != ".jsonl" && ext != ".md" {
				continue
			}
			path := filepath.Join(projectDir, f.Name())
			data, err := os.ReadFile(path)
			if err != nil || len(data) == 0 {
				continue
			}
			if ext == ".jsonl" {
				pvs, cnt := parseJSONLSession(string(data), path, "claude")
				previews = append(previews, pvs...)
				agentCountMap["claude"] += cnt
			} else if ext == ".md" {
				previews, agentCountMap = parseMarkdownMemory(string(data), path, "claude", previews, agentCountMap)
			}
		}
	}
	return previews, agentCountMap
}

func extractSessionDir(sessionsDir string, previews []memoryPreview, agentCountMap map[string]int) ([]memoryPreview, map[string]int) {
	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		return previews, agentCountMap
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".json" && ext != ".jsonl" {
			continue
		}
		path := filepath.Join(sessionsDir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil || len(data) == 0 {
			continue
		}
		agentName := "session"
		if ext == ".jsonl" {
			pvs, cnt := parseJSONLSession(string(data), path, agentName)
			previews = append(previews, pvs...)
			agentCountMap[agentName] += cnt
		} else if ext == ".json" {
			var jsonData map[string]interface{}
			if json.Unmarshal(data, &jsonData) == nil {
				if title, ok := jsonData["title"].(string); ok && title != "" {
					key := title
					content := string(data)
					if len(content) > 2000 {
						content = content[:2000]
					}
					previews = append(previews, memoryPreview{
						Key: key, Content: content, Layer: "episodic",
						Source: "session-json", FilePath: path, AgentName: agentName,
					})
					agentCountMap[agentName]++
				}
			}
		}
	}
	return previews, agentCountMap
}

func extractVscdbMemories(dbPath string, previews []memoryPreview, agentCountMap map[string]int) ([]memoryPreview, map[string]int) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return previews, agentCountMap
	}
	sqlDB, _ := db.DB()
	if sqlDB != nil {
		defer sqlDB.Close()
	}

	agentName := "vscdb"
	lowerPath := strings.ToLower(dbPath)
	if strings.Contains(lowerPath, "trae") {
		agentName = "trae"
	} else if strings.Contains(lowerPath, "codebuddy") {
		agentName = "codebuddy"
	} else if strings.Contains(lowerPath, "cursor") {
		agentName = "cursor"
	} else if strings.Contains(lowerPath, "windsurf") {
		agentName = "windsurf"
	}

	type TableName struct {
		Name string
	}
	var tables []TableName
	db.Raw("SELECT name FROM sqlite_master WHERE type='table'").Scan(&tables)
	hasItemTable := false
	for _, t := range tables {
		if t.Name == "ItemTable" {
			hasItemTable = true
			break
		}
	}
	if !hasItemTable {
		return previews, agentCountMap
	}

	type KV struct {
		Key   string
		Value string
	}

	var chatRows []KV
	db.Raw("SELECT key, value FROM ItemTable WHERE key = 'icube-ai-agent-storage-input-history'").Scan(&chatRows)
	for _, row := range chatRows {
		if row.Value == "" {
			continue
		}
		var entries []struct {
			InputText string `json:"inputText"`
		}
		if json.Unmarshal([]byte(row.Value), &entries) != nil {
			continue
		}
		for _, e := range entries {
			if len(e.InputText) < 10 {
				continue
			}
			key := agentName + "-chat-" + fmt.Sprintf("%x", md5.Sum([]byte(e.InputText)))[:12]
			previews = append(previews, memoryPreview{
				Key: key, Content: e.InputText, Layer: "episodic",
				Source: agentName + "-chat", FilePath: dbPath, AgentName: agentName,
			})
			agentCountMap[agentName]++
		}
	}

	var sessionRows []KV
	db.Raw("SELECT key, value FROM ItemTable WHERE key LIKE 'session:%'").Scan(&sessionRows)
	for _, row := range sessionRows {
		if row.Value == "" || len(row.Value) < 20 {
			continue
		}
		type SessionData struct {
			Title string `json:"title"`
			Cwd   string `json:"cwd"`
		}
		var session SessionData
		if json.Unmarshal([]byte(row.Value), &session) != nil || session.Title == "" {
			continue
		}
		content := session.Title
		if session.Cwd != "" {
			content += " | Project: " + session.Cwd
		}
		key := agentName + "-session-" + fmt.Sprintf("%x", md5.Sum([]byte(row.Value)))[:12]
		previews = append(previews, memoryPreview{
			Key: key, Content: content, Layer: "episodic",
			Source: agentName + "-session", FilePath: dbPath, AgentName: agentName,
		})
		agentCountMap[agentName]++
	}

	var cursorChatRows []KV
	db.Raw("SELECT key, value FROM ItemTable WHERE key LIKE 'cursor%chat%' OR key LIKE 'composerData%'").Scan(&cursorChatRows)
	for _, row := range cursorChatRows {
		if row.Value == "" || len(row.Value) < 50 {
			continue
		}
		key := agentName + "-chat-" + fmt.Sprintf("%x", md5.Sum([]byte(row.Key)))[:12]
		content := row.Value
		if len(content) > 2000 {
			content = content[:2000]
		}
		previews = append(previews, memoryPreview{
			Key: key, Content: content, Layer: "episodic",
			Source: agentName + "-chat", FilePath: dbPath, AgentName: agentName,
		})
		agentCountMap[agentName]++
	}

	return previews, agentCountMap
}

func parseMarkdownMemory(content string, filePath string, agentName string, previews []memoryPreview, agentCountMap map[string]int) ([]memoryPreview, map[string]int) {
	lines := strings.Split(content, "\n")
	currentSection := ""
	currentContent := ""
	for _, line := range lines {
		if strings.HasPrefix(line, "# ") {
			if currentSection != "" || currentContent != "" {
				key := currentSection
				if key == "" {
					key = filepath.Base(filePath)
				}
				preview := strings.TrimSpace(currentContent)
				if len(preview) > 100 {
					preview = preview[:100] + "..."
				}
				previews = append(previews, memoryPreview{Key: key, Content: preview, Layer: "knowledge", Source: "markdown", FilePath: filePath, AgentName: agentName})
				agentCountMap[agentName]++
			}
			currentSection = strings.TrimSpace(line[2:])
			currentContent = ""
		} else {
			currentContent += line + "\n"
		}
	}
	if currentSection != "" || currentContent != "" {
		key := currentSection
		if key == "" {
			key = filepath.Base(filePath)
		}
		preview := strings.TrimSpace(currentContent)
		if len(preview) > 100 {
			preview = preview[:100] + "..."
		}
		previews = append(previews, memoryPreview{Key: key, Content: preview, Layer: "knowledge", Source: "markdown", FilePath: filePath, AgentName: agentName})
		agentCountMap[agentName]++
	}
	return previews, agentCountMap
}

func extractSqliteMemories(dbPath string, previews []memoryPreview, agentCountMap map[string]int) ([]memoryPreview, map[string]int) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return previews, agentCountMap
	}
	sqlDB, err := db.DB()
	if err != nil {
		return previews, agentCountMap
	}
	defer sqlDB.Close()

	type TableName struct {
		Name string
	}
	var tables []TableName
	db.Raw("SELECT name FROM sqlite_master WHERE type='table'").Scan(&tables)

	agentName := filepath.Base(filepath.Dir(dbPath))
	if agentName == "" || agentName == "." {
		agentName = "sqlite-" + filepath.Base(dbPath)
	}

	for _, t := range tables {
		type ColInfo struct {
			CID       int
			Name      string
			Type      string
			NotNull   int
			DfltValue interface{}
			PK        int
		}
		var cols []ColInfo
		db.Raw(fmt.Sprintf("PRAGMA table_info(%s)", t.Name)).Scan(&cols)

		keyCol := ""
		valueCol := ""
		for _, c := range cols {
			cl := strings.ToLower(c.Name)
			if keyCol == "" && (cl == "key" || cl == "name" || cl == "title" || cl == "id") {
				keyCol = c.Name
			}
			if valueCol == "" && (cl == "value" || cl == "content" || cl == "text" || cl == "description" || cl == "body" || cl == "message") {
				valueCol = c.Name
			}
		}

		if keyCol == "" || valueCol == "" {
			continue
		}

		type KV struct {
			Key   string
			Value string
		}
		var kvPairs []KV
		db.Raw(fmt.Sprintf("SELECT %s as key, %s as value FROM %s LIMIT 200", keyCol, valueCol, t.Name)).Scan(&kvPairs)

		for _, kv := range kvPairs {
			if kv.Key == "" || kv.Value == "" {
				continue
			}
			preview := kv.Value
			if len(preview) > 200 {
				preview = preview[:200] + "..."
			}
			previews = append(previews, memoryPreview{
				Key: kv.Key, Content: preview, Layer: "knowledge",
				Source: "sqlite", FilePath: dbPath, AgentName: agentName,
			})
			agentCountMap[agentName]++
		}
	}

	return previews, agentCountMap
}

func handleScanOpenClawMemories(c *gin.Context) {
	searchDirs := getOpenClawSearchDirs()

	var allPreviews []memoryPreview
	agentCountMap := make(map[string]int)
	var foundDirs []string

	for _, dir := range searchDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}

		previews, acm := extractMemoriesFromDir(dir)
		if len(previews) > 0 {
			allPreviews = append(allPreviews, previews...)
			for name, count := range acm {
				agentCountMap[name] += count
			}
			foundDirs = append(foundDirs, dir)
		}
	}

	if len(allPreviews) > 0 {
		agents := make([]map[string]interface{}, 0)
		for name, count := range agentCountMap {
			agentPreviews := make([]map[string]interface{}, 0)
			for _, p := range allPreviews {
				if p.AgentName == name {
					agentPreviews = append(agentPreviews, map[string]interface{}{
						"key":    p.Key,
						"value":  p.Content,
						"layer":  p.Layer,
						"source": p.Source,
					})
				}
			}
			agents = append(agents, map[string]interface{}{
				"agent_name":   name,
				"layout":       "v2",
				"files":        count,
				"memory_count": count,
				"previews":     agentPreviews,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"found":          true,
			"scanned_dirs":   strings.Join(foundDirs, ", "),
			"openclaw_dir":   strings.Join(foundDirs, ", "),
			"clients":        services.DetectInstalledClients(),
			"agents":         agents,
			"total_memories": len(allPreviews),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"found": false,
	})
}

func handleScanOpenClawAgent(c *gin.Context) {
	agentName := c.Param("agentName")
	searchDirs := getOpenClawSearchDirs()

	var allFiltered []map[string]interface{}

	for _, dir := range searchDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}

		previews, _ := extractMemoriesFromDir(dir)

		for _, p := range previews {
			if p.AgentName == agentName {
				allFiltered = append(allFiltered, map[string]interface{}{
					"key":    p.Key,
					"value":  p.Content,
					"layer":  p.Layer,
					"source": p.Source,
				})
			}
		}
	}

	if len(allFiltered) > 0 {
		c.JSON(http.StatusOK, gin.H{
			"agent_name": agentName,
			"preview":    allFiltered,
			"total":      len(allFiltered),
		})
		return
	}

	c.JSON(http.StatusNotFound, gin.H{
		"error": "agent not found",
	})
}

func handleImportOpenClawMemories(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		var req struct {
			AgentName    string `json:"agent_name"`
			Layer        string `json:"layer"`
			SkipExisting bool   `json:"skip_existing"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if req.Layer == "" {
			req.Layer = "knowledge"
		}

		searchDirs := getOpenClawSearchDirs()

		var imported, skipped, errorsCount int

		seenKeys := make(map[string]bool)
		var existingKeys []string
		db.Table("memories").Where("user_id = ? AND status != ?", userID, "trashed").Pluck("key", &existingKeys)
		for _, k := range existingKeys {
			seenKeys[k] = true
		}

		for _, dir := range searchDirs {
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				continue
			}

			previews, _ := extractMemoriesFromDir(dir)

			for _, p := range previews {
				if req.AgentName != "" && p.AgentName != req.AgentName {
					continue
				}

				if p.Key == "" {
					errorsCount++
					continue
				}

				if seenKeys[p.Key] {
					skipped++
					continue
				}

				if req.SkipExisting {
					var count int64
					logDBErr("count memories by key for import", db.Table("memories").Where("user_id = ? AND key = ?", userID, p.Key).Count(&count).Error)
					if count > 0 {
						skipped++
						seenKeys[p.Key] = true
						continue
					}
				}

				fullContent := p.Content
				if strings.HasSuffix(fullContent, "...") {
					data, err := os.ReadFile(p.FilePath)
					if err == nil {
						ext := strings.ToLower(filepath.Ext(p.FilePath))
						if ext == ".jsonl" {
							lines := strings.Split(string(data), "\n")
							for _, line := range lines {
								line = strings.TrimSpace(line)
								if line == "" {
									continue
								}
								var msg map[string]interface{}
								if json.Unmarshal([]byte(line), &msg) != nil {
									continue
								}
								text, _ := msg["text"].(string)
								if text == "" {
									text, _ = msg["content"].(string)
								}
								key := ""
								msgType, _ := msg["type"].(string)
								if text != "" {
									key = msgType + ": "
									if len(text) > 40 {
										key += text[:40] + "..."
									} else {
										key += text
									}
								}
								if key == p.Key && text != "" {
									fullContent = text
									break
								}
							}
						} else if ext == ".json" {
							var jsonItems []map[string]interface{}
							if json.Unmarshal(data, &jsonItems) == nil {
								for _, m := range jsonItems {
									key, _ := m["key"].(string)
									if key == "" {
										key, _ = m["name"].(string)
									}
									if key == p.Key {
										if v, ok := m["content"].(string); ok && v != "" {
											fullContent = v
										} else if v, ok := m["value"].(string); ok && v != "" {
											fullContent = v
										} else if v, ok := m["text"].(string); ok && v != "" {
											fullContent = v
										} else if v, ok := m["description"].(string); ok && v != "" {
											fullContent = v
										}
										break
									}
								}
							} else {
								var single map[string]interface{}
								if json.Unmarshal(data, &single) == nil {
									if v, ok := single["content"].(string); ok && v != "" {
										fullContent = v
									} else if v, ok := single["value"].(string); ok && v != "" {
										fullContent = v
									} else if v, ok := single["text"].(string); ok && v != "" {
										fullContent = v
									}
								}
							}
						} else if ext == ".md" {
							mdLines := strings.Split(string(data), "\n")
							curSection := ""
							curContent := ""
							for _, l := range mdLines {
								if strings.HasPrefix(l, "# ") {
									if curSection == p.Key && curContent != "" {
										fullContent = strings.TrimSpace(curContent)
										break
									}
									curSection = strings.TrimSpace(l[2:])
									curContent = ""
								} else {
									curContent += l + "\n"
								}
							}
							if curSection == p.Key && fullContent == p.Content {
								fullContent = strings.TrimSpace(curContent)
							}
							if fullContent == p.Content {
								fullContent = strings.TrimSpace(string(data))
							}
						} else {
							fullContent = string(data)
						}
					}
				}

				if fullContent == "" {
					errorsCount++
					continue
				}

				layer := req.Layer
				if p.Layer != "" {
					layer = p.Layer
				}

				memSvc := services.NewMemoryService(db)
				_, err := memSvc.Create(userID, map[string]interface{}{
					"key":        p.Key,
					"value":      fullContent,
					"layer":      layer,
					"importance": 0.5,
					"source":     p.Source,
				})
				if err != nil {
					errorsCount++
				} else {
					imported++
					seenKeys[p.Key] = true
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"imported": imported,
			"skipped":  skipped,
			"errors":   errorsCount,
		})
	}
}

func handleAutoImportMemories(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		searchDirs := services.GetAllSearchDirs()

		var imported, skipped, entitiesCreated int
		var foundFiles []string
		seenKeys := make(map[string]bool)

		var existingKeys []string
		db.Table("memories").Where("user_id = ? AND status != ?", userID, "trashed").Pluck("key", &existingKeys)
		for _, k := range existingKeys {
			seenKeys[k] = true
		}

		for _, dir := range searchDirs {
			memFile := filepath.Join(dir, "MEMORY.md")
			if data, err := os.ReadFile(memFile); err == nil && len(data) > 0 {
				foundFiles = append(foundFiles, memFile)
				importFromMarkdown(db, userID, memFile, string(data), seenKeys, &imported, &skipped, &entitiesCreated)
			}

			memoryDir := filepath.Join(dir, "memory")
			if files, err := os.ReadDir(memoryDir); err == nil {
				for _, f := range files {
					if f.IsDir() {
						continue
					}
					ext := strings.ToLower(filepath.Ext(f.Name()))
					if ext != ".md" && ext != ".txt" {
						continue
					}
					path := filepath.Join(memoryDir, f.Name())
					data, err := os.ReadFile(path)
					if err != nil || len(data) == 0 {
						continue
					}
					foundFiles = append(foundFiles, path)
					content := string(data)
					if ext == ".md" {
						importFromMarkdown(db, userID, path, content, seenKeys, &imported, &skipped, &entitiesCreated)
					} else if ext == ".txt" {
						importFromText(db, userID, path, content, seenKeys, &imported, &skipped, &entitiesCreated)
					}
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"imported":         imported,
			"skipped":          skipped,
			"entities_created": entitiesCreated,
			"files_found":      foundFiles,
			"message":          fmt.Sprintf("imported %d memories, created %d entities, skipped %d", imported, entitiesCreated, skipped),
		})
	}
}

func importFromJSON(db *gorm.DB, userID uint, content string, seenKeys map[string]bool, imported, skipped, entitiesCreated *int) {
	var memories []map[string]interface{}
	if json.Unmarshal([]byte(content), &memories) != nil {
		var single map[string]interface{}
		if json.Unmarshal([]byte(content), &single) != nil {
			return
		}
		memories = []map[string]interface{}{single}
	}

	for _, m := range memories {
		key, _ := m["key"].(string)
		contentStr, _ := m["content"].(string)
		if key == "" {
			if name, ok := m["name"].(string); ok {
				key = name
			} else if title, ok := m["title"].(string); ok {
				key = title
			}
		}
		if contentStr == "" {
			contentStr, _ = m["value"].(string)
			if contentStr == "" {
				contentStr, _ = m["text"].(string)
				if contentStr == "" {
					contentStr, _ = m["description"].(string)
				}
			}
		}
		if key == "" || contentStr == "" {
			*skipped++
			continue
		}

		if seenKeys[key] {
			*skipped++
			continue
		}

		layer := classifyLayer(key, contentStr)
		importance := 0.5
		if imp, ok := m["importance"].(float64); ok {
			importance = imp
		}

		tags := extractTags(m)

		source := "auto_import"
		if s, ok := m["source"].(string); ok && s != "" {
			source = s
		}

		memSvc := services.NewMemoryService(db)
		_, err := memSvc.Create(userID, map[string]interface{}{
			"key":        key,
			"value":      contentStr,
			"layer":      layer,
			"importance": importance,
			"tags":       tags,
			"source":     source,
		})
		if err != nil {
			*skipped++
			continue
		}
		seenKeys[key] = true
		*imported++

		tryCreateEntity(db, userID, key, contentStr, entitiesCreated)
	}
}

func importFromMarkdown(db *gorm.DB, userID uint, filePath, content string, seenKeys map[string]bool, imported, skipped, entitiesCreated *int) {
	sections := strings.Split(content, "\n## ")
	for i, section := range sections {
		var key, body string
		if i == 0 {
			lines := strings.SplitN(section, "\n", 2)
			key = strings.TrimPrefix(strings.TrimSpace(lines[0]), "# ")
			if len(lines) > 1 {
				body = strings.TrimSpace(lines[1])
			}
		} else {
			lines := strings.SplitN(section, "\n", 2)
			key = strings.TrimSpace(lines[0])
			if len(lines) > 1 {
				body = strings.TrimSpace(lines[1])
			}
		}

		if key == "" || body == "" {
			continue
		}

		key = fmt.Sprintf("md:%s", key)
		if seenKeys[key] {
			*skipped++
			continue
		}

		layer := classifyLayer(key, body)
		importance := 0.6

		source := "auto_import_md"
		relPath := filePath
		if len(relPath) > 100 {
			relPath = "..." + relPath[len(relPath)-97:]
		}

		memSvc := services.NewMemoryService(db)
		_, err := memSvc.Create(userID, map[string]interface{}{
			"key":        key,
			"value":      body,
			"layer":      layer,
			"importance": importance,
			"tags":       "markdown",
			"source":     source,
		})
		if err != nil {
			*skipped++
			continue
		}
		seenKeys[key] = true
		*imported++

		tryCreateEntity(db, userID, key, body, entitiesCreated)
	}
}

func importFromText(db *gorm.DB, userID uint, filePath, content string, seenKeys map[string]bool, imported, skipped, entitiesCreated *int) {
	lines := strings.Split(content, "\n")
	var buffer []string
	var currentKey string

	flushBuffer := func() {
		if currentKey != "" && len(buffer) > 0 {
			body := strings.Join(buffer, "\n")
			key := fmt.Sprintf("txt:%s", currentKey)

			if !seenKeys[key] {
				layer := classifyLayer(key, body)
				memSvc := services.NewMemoryService(db)
				_, err := memSvc.Create(userID, map[string]interface{}{
					"key":        key,
					"value":      body,
					"layer":      layer,
					"importance": 0.4,
					"tags":       "text",
					"source":     "auto_import_txt",
				})
				if err == nil {
					seenKeys[key] = true
					*imported++
					tryCreateEntity(db, userID, key, body, entitiesCreated)
				} else {
					*skipped++
				}
			} else {
				*skipped++
			}
		}
		buffer = nil
		currentKey = ""
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if len(trimmed) < 80 && !strings.HasPrefix(trimmed, "-") && !strings.HasPrefix(trimmed, "*") && strings.HasSuffix(trimmed, ":") {
			flushBuffer()
			currentKey = strings.TrimSuffix(trimmed, ":")
		} else if len(trimmed) < 60 && !strings.HasPrefix(trimmed, "-") && !strings.HasPrefix(trimmed, "*") && currentKey == "" {
			flushBuffer()
			currentKey = trimmed
		} else {
			buffer = append(buffer, trimmed)
		}
	}
	flushBuffer()

	if *imported == 0 && len(strings.Fields(content)) > 5 {
		key := fmt.Sprintf("txt:%s", filepath.Base(filePath))
		key = strings.TrimSuffix(key, filepath.Ext(key))
		if !seenKeys[key] && len(content) > 10 {
			layer := classifyLayer(key, content)
			memSvc := services.NewMemoryService(db)
			_, err := memSvc.Create(userID, map[string]interface{}{
				"key":        key,
				"value":      content,
				"layer":      layer,
				"importance": 0.3,
				"tags":       "text",
				"source":     "auto_import_txt",
			})
			if err == nil {
				seenKeys[key] = true
				*imported++
				tryCreateEntity(db, userID, key, content, entitiesCreated)
			}
		}
	}
}

func classifyLayer(key, content string) string {
	lowerKey := strings.ToLower(key)
	lowerContent := strings.ToLower(content)

	if strings.Contains(lowerKey, "偏好") || strings.Contains(lowerKey, "preference") ||
		strings.Contains(lowerContent, "我喜欢") || strings.Contains(lowerContent, "i prefer") ||
		strings.Contains(lowerContent, "偏好") || strings.Contains(lowerContent, "preference") {
		return "preference"
	}
	if strings.Contains(lowerKey, "临时") || strings.Contains(lowerKey, "temporary") ||
		strings.Contains(lowerKey, "todo") || strings.Contains(lowerKey, "待办") ||
		strings.Contains(lowerContent, "临时") || strings.Contains(lowerContent, "temporary") {
		return "short_term"
	}
	if strings.Contains(lowerKey, "私密") || strings.Contains(lowerKey, "private") ||
		strings.Contains(lowerKey, "密码") || strings.Contains(lowerKey, "password") ||
		strings.Contains(lowerContent, "私密") || strings.Contains(lowerContent, "private") {
		return "private"
	}
	if strings.Contains(lowerKey, "项目") || strings.Contains(lowerKey, "project") ||
		strings.Contains(lowerContent, "项目") || strings.Contains(lowerContent, "project") {
		return "knowledge"
	}
	if strings.Contains(lowerKey, "工具") || strings.Contains(lowerKey, "tool") ||
		strings.Contains(lowerContent, "工具") || strings.Contains(lowerContent, "software") {
		return "knowledge"
	}
	return "knowledge"
}

func extractTags(m map[string]interface{}) string {
	if t, ok := m["tags"].([]interface{}); ok && len(t) > 0 {
		tagStrs := make([]string, 0, len(t))
		for _, tag := range t {
			if s, ok := tag.(string); ok {
				tagStrs = append(tagStrs, s)
			}
		}
		return strings.Join(tagStrs, ",")
	}
	if t, ok := m["tags"].(string); ok {
		return t
	}
	if cat, ok := m["category"].(string); ok {
		return cat
	}
	return ""
}

func tryCreateEntity(db *gorm.DB, userID uint, key, content string, entitiesCreated *int) {
	if len(content) < 10 || len(content) > 2000 {
		return
	}

	entityType := "concept"
	lowerContent := strings.ToLower(content)
	if strings.Contains(lowerContent, "项目") || strings.Contains(lowerContent, "project") {
		entityType = "organization"
	} else if strings.Contains(lowerContent, "工具") || strings.Contains(lowerContent, "tool") || strings.Contains(lowerContent, "软件") || strings.Contains(lowerContent, "software") {
		entityType = "technology"
	} else if strings.Contains(lowerContent, "人") || strings.Contains(lowerContent, "person") || strings.Contains(lowerContent, "用户") {
		entityType = "person"
	} else if strings.Contains(lowerContent, "地点") || strings.Contains(lowerContent, "location") || strings.Contains(lowerContent, "城市") {
		entityType = "location"
	} else if strings.Contains(lowerContent, "事件") || strings.Contains(lowerContent, "event") {
		entityType = "event"
	}

	name := key
	if strings.HasPrefix(name, "md:") || strings.HasPrefix(name, "txt:") {
		name = name[3:]
	}
	if len(name) > 50 {
		name = name[:50]
	}

	var entityCount int64
	logDBErr("count entities by name", db.Table("entities").Where("user_id = ? AND name = ?", userID, name).Count(&entityCount).Error)
	if entityCount == 0 {
		entity := models.Entity{
			UserID:        userID,
			Name:          name,
			EntityType:    entityType,
			Description:   content,
			Confidence:    0.7,
			ExtractMethod: "auto_import",
		}
		if db.Create(&entity).Error == nil {
			*entitiesCreated++
		}
	}
}

func handleListBackups(c *gin.Context) {
	cfg := config.Load()
	backupDir := cfg.BackupsDir
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		c.JSON(http.StatusOK, gin.H{"backups": []interface{}{}})
		return
	}

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"backups": []interface{}{}})
		return
	}

	backups := make([]map[string]interface{}, 0)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".db") && !strings.HasSuffix(entry.Name(), ".sql") && !strings.HasSuffix(entry.Name(), ".zip") {
			continue
		}
		info, _ := entry.Info()
		backups = append(backups, map[string]interface{}{
			"filename":   entry.Name(),
			"size":       info.Size(),
			"created_at": info.ModTime().Format("2006-01-02 15:04:05"),
		})
	}

	c.JSON(http.StatusOK, gin.H{"backups": backups})
}

func handleCreateBackup(c *gin.Context) {
	cfg := config.Load()
	backupDir := cfg.BackupsDir
	os.MkdirAll(backupDir, 0755)

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("clawmemory_backup_%s.db", timestamp)
	backupPath := filepath.Join(backupDir, filename)

	dbPath := cfg.DatabasePath

	src, err := os.Open(dbPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": fmt.Sprintf("Cannot open database file: %v", err),
		})
		return
	}
	defer src.Close()

	dst, err := os.Create(backupPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": fmt.Sprintf("Cannot create backup file: %v", err),
		})
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": fmt.Sprintf("Backup failed: %v", err),
		})
		return
	}

	fi, _ := dst.Stat()

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"filename": filename,
		"path":     backupPath,
		"size":     fi.Size(),
	})
}

func handleDownloadBackup(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		filename := c.Param("filename")
		if strings.Contains(filename, "..") || strings.ContainsAny(filename, "/\\") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid filename"})
			return
		}
		cfg := config.Load()
		backupPath := filepath.Join(cfg.BackupsDir, filename)

		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "backup not found"})
			return
		}

		c.FileAttachment(backupPath, filename)
	}
}

func handleRestoreBackup(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		filename := c.Param("filename")
		if strings.Contains(filename, "..") || strings.ContainsAny(filename, "/\\") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid filename"})
			return
		}
		cfg := config.Load()
		backupPath := filepath.Join(cfg.BackupsDir, filename)

		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "backup not found"})
			return
		}

		dbPath := cfg.DatabasePath
		preRestorePath := dbPath + ".pre-restore"
		if _, err := os.Stat(dbPath); err == nil {
			if err := copyFile(dbPath, preRestorePath); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create safety backup before restore"})
				return
			}
		}

		src, err := os.Open(backupPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot read backup file"})
			return
		}
		defer src.Close()

		dst, err := os.Create(dbPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot write database file"})
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			os.Rename(preRestorePath, dbPath)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "restore failed, rolled back to previous state"})
			return
		}

		os.Remove(preRestorePath)

		c.JSON(http.StatusOK, gin.H{"message": "backup restored successfully", "filename": filename})
	}
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func handleDeleteBackup(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		filename := c.Param("filename")
		if strings.Contains(filename, "..") || strings.ContainsAny(filename, "/\\") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid filename"})
			return
		}
		cfg := config.Load()
		backupPath := filepath.Join(cfg.BackupsDir, filename)

		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "backup not found"})
			return
		}

		if err := os.Remove(backupPath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot delete backup file"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "backup deleted", "filename": filename})
	}
}

func handleDecayStats(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		svc := services.NewDecayService(db)
		userID := middleware.GetUserID(c)
		stats, err := svc.GetStats(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, stats)
	}
}

func handleDecayApply(aiSvc *ai.AIService, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
		defer cancel()

		result, err := aiSvc.DecayEvaluate(ctx, userID)
		if err != nil {
			svc := services.NewDecayService(db)
			result, err = svc.ApplyDecay(userID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			result["mode"] = "local_fallback"
		} else {
			evaluations, _ := result["evaluations"].([]map[string]interface{})
			archived := 0
			trashed := 0
			kept := 0
			for _, ev := range evaluations {
				action, _ := ev["action"].(string)
				memID, _ := ev["id"].(float64)
				switch action {
				case "archive":
					db.Model(&models.Memory{}).Where("id = ? AND user_id = ?", uint(memID), userID).Update("status", "archived")
					archived++
				case "delete":
					db.Model(&models.Memory{}).Where("id = ? AND user_id = ?", uint(memID), userID).Update("status", "trashed")
					trashed++
				default:
					if newImp, ok := ev["new_importance"].(float64); ok {
						db.Model(&models.Memory{}).Where("id = ? AND user_id = ?", uint(memID), userID).Update("importance", newImp)
					}
					kept++
				}
			}
			result["processed"] = len(evaluations)
			result["archived"] = archived
			result["trashed"] = trashed
			result["adjusted"] = kept
			result["algorithm"] = "ai_decay_v1"
		}

		c.JSON(http.StatusOK, result)
	}
}

func handleDecaySettingsGet(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		svc := services.NewDecayService(db)
		userID := middleware.GetUserID(c)
		settings, err := svc.GetDecaySettings(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, settings)
	}
}

func handleDecaySettingsUpdate(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Enabled bool                   `json:"enabled"`
			Config  map[string]interface{} `json:"config"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		svc := services.NewDecayService(db)
		userID := middleware.GetUserID(c)
		if err := svc.UpdateDecaySettings(userID, req.Enabled, req.Config); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "decay settings updated"})
	}
}

func handleEmptyTrash(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		svc := services.NewDecayService(db)
		userID := middleware.GetUserID(c)
		count, err := svc.EmptyTrash(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"deleted": count})
	}
}

func handleListTrash(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var memories []models.Memory
		userID := middleware.GetUserID(c)
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
		if limit < 1 || limit > 500 {
			limit = 100
		}
		logDBErr("load trashed memories", db.Where("user_id = ? AND status = ?", userID, "trashed").Order("trashed_at DESC").Limit(limit).Find(&memories).Error)
		items := make([]*services.MemoryModel, 0, len(memories))
		for i := range memories {
			items = append(items, services.ToMemoryModel(&memories[i]))
		}
		c.JSON(http.StatusOK, gin.H{"items": items})
	}
}

func handleExportData(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "password confirmation required"})
			return
		}

		userID := middleware.GetUserID(c)
		var user models.User
		if err := db.First(&user, userID).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "incorrect password"})
			return
		}

		exportData := map[string]interface{}{}

		var memories []models.Memory
		if err := db.Where("user_id = ?", userID).Find(&memories).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "export failed: " + err.Error()})
			return
		}
		memoryModels := make([]*services.MemoryModel, 0, len(memories))
		for i := range memories {
			memoryModels = append(memoryModels, services.ToMemoryModel(&memories[i]))
		}
		exportData["memories"] = memoryModels

		var entities []models.Entity
		logDBErr("load entities for export", db.Where("user_id = ?", userID).Find(&entities).Error)
		exportData["entities"] = entities

		var relations []models.Relation
		logDBErr("load relations for export", db.Where("user_id = ?", userID).Find(&relations).Error)
		exportData["relations"] = relations

		var wikiPages []models.WikiPage
		logDBErr("load wiki pages for export", db.Where("user_id = ?", userID).Find(&wikiPages).Error)
		exportData["wiki_pages"] = wikiPages

		var projects []models.Project
		logDBErr("load projects for export", db.Where("user_id = ?", userID).Find(&projects).Error)
		exportData["projects"] = projects

		var projectNotes []models.ProjectNote
		logDBErr("load project notes for export", db.Where("user_id = ?", userID).Find(&projectNotes).Error)
		exportData["project_notes"] = projectNotes

		var reports []models.DailyReport
		logDBErr("load reports for export", db.Where("user_id = ?", userID).Find(&reports).Error)
		exportData["daily_reports"] = reports

		exportData["exported_at"] = time.Now().Format(time.RFC3339)
		exportData["version"] = config.AppVersion

		c.JSON(http.StatusOK, exportData)
	}
}

func handleImportData(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 50*1024*1024)

		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userID := middleware.GetUserID(c)
		imported := 0
		failed := 0
		maxImport := 5000

		memSvc := services.NewMemoryService(db)
		knowSvc := services.NewKnowledgeService(db)
		wikiSvc := services.NewWikiService(db)
		projSvc := services.NewProjectService(db)

		if memories, ok := req["memories"].([]interface{}); ok {
			for _, m := range memories {
				if imported >= maxImport {
					break
				}
				if data, ok := m.(map[string]interface{}); ok {
					if _, err := memSvc.Create(userID, data); err == nil {
						imported++
					} else {
						failed++
					}
				}
			}
		}

		if entities, ok := req["entities"].([]interface{}); ok {
			for _, e := range entities {
				if imported >= maxImport {
					break
				}
				if data, ok := e.(map[string]interface{}); ok {
					if _, err := knowSvc.CreateEntity(userID, data); err == nil {
						imported++
					} else {
						failed++
					}
				}
			}
		}

		if wikiPages, ok := req["wiki_pages"].([]interface{}); ok {
			for _, w := range wikiPages {
				if imported >= maxImport {
					break
				}
				if data, ok := w.(map[string]interface{}); ok {
					if _, err := wikiSvc.Create(userID, data); err == nil {
						imported++
					} else {
						failed++
					}
				}
			}
		}

		if projects, ok := req["projects"].([]interface{}); ok {
			for _, p := range projects {
				if imported >= maxImport {
					break
				}
				if data, ok := p.(map[string]interface{}); ok {
					if _, err := projSvc.Create(userID, data); err == nil {
						imported++
					} else {
						failed++
					}
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"imported": imported,
			"failed":   failed,
			"message":  "data imported successfully",
		})
	}
}

func handleDedupScan(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		svc := services.NewDedupService(db)
		result, err := svc.Scan(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleDedupMerge(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			SourceID uint `json:"source_id" binding:"required"`
			TargetID uint `json:"target_id" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		userID := middleware.GetUserID(c)
		svc := services.NewDedupService(db)
		result, err := svc.Merge(userID, req.SourceID, req.TargetID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleMemoryHealth(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		svc := services.NewHealthService(db)
		result, err := svc.GetHealthScore(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleMemoryQuality(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		svc := services.NewHealthService(db)
		result, err := svc.AssessQuality(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleMemoryAutoFix(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		var data struct {
			IssueTypes []string `json:"issue_types"`
		}
		c.ShouldBindJSON(&data)

		svc := services.NewHealthService(db)
		result, err := svc.AutoFix(userID, data.IssueTypes)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleListProjects(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
		if page < 1 {
			page = 1
		}
		if size < 1 || size > 100 {
			size = 20
		}
		status := c.Query("status")
		category := c.Query("category")

		svc := services.NewProjectService(db)
		projects, total, err := svc.List(userID, page, size, status, category)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": projects, "total": total, "page": page, "size": size, "pages": (total + int64(size) - 1) / int64(size)})
	}
}

func handleGetProject(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}

		svc := services.NewProjectService(db)
		project, err := svc.Get(userID, uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
			return
		}
		c.JSON(http.StatusOK, project)
	}
}

func handleCreateProject(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		svc := services.NewProjectService(db)
		project, err := svc.Create(userID, data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, project)
	}
}

func handleUpdateProject(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}
		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		svc := services.NewProjectService(db)
		project, err := svc.Update(userID, uint(id), data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, project)
	}
}

func handleDeleteProject(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}

		svc := services.NewProjectService(db)
		if err := svc.Delete(userID, uint(id)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	}
}

func handleProjectNotes(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		projectID, ok := parseIDParam(c, "id")
		if !ok {
			return
		}

		svc := services.NewProjectService(db)
		notes, err := svc.GetNotes(userID, uint(projectID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": notes})
	}
}

func handleAddProjectNote(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		projectID, ok := parseIDParam(c, "id")
		if !ok {
			return
		}
		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		svc := services.NewProjectService(db)
		note, err := svc.AddNote(userID, uint(projectID), data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, note)
	}
}

func handleUpdateProjectNote(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		noteID, ok := parseIDParam(c, "noteId")
		if !ok {
			return
		}
		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		svc := services.NewProjectService(db)
		note, err := svc.UpdateNote(userID, uint(noteID), data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, note)
	}
}

func handleDeleteProjectNote(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		noteID, ok := parseIDParam(c, "noteId")
		if !ok {
			return
		}

		svc := services.NewProjectService(db)
		if err := svc.DeleteNote(userID, uint(noteID)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	}
}

func handleProjectCategories(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		svc := services.NewProjectService(db)
		categories, err := svc.GetCategories(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"categories": categories})
	}
}

func handleProjectExtractMemories(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}

		svc := services.NewProjectService(db)
		extracted, err := svc.ExtractFromMemories(userID, uint(id))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"extracted": extracted})
	}
}

func handleProjectContext(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		name := c.Query("name")
		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name parameter required"})
			return
		}

		svc := services.NewProjectService(db)
		context, err := svc.GetContextForOpenClaw(userID, name)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"context": context})
	}
}

func handleProjectSearch(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		query := c.Query("q")
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

		svc := services.NewProjectService(db)
		projects, err := svc.Search(userID, query, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": projects})
	}
}

func auditLog(db *gorm.DB, c *gin.Context, action, target, detail string) {
	userID := middleware.GetUserID(c)
	log := models.AuditLog{
		UserID: userID,
		Action: action,
		Target: target,
		Detail: detail,
		IP:     c.ClientIP(),
	}
	logDBErr("create audit log", db.Create(&log).Error)
}

func getString(m map[string]interface{}, key, def string) string {
	if v, ok := m[key].(string); ok && v != "" {
		return v
	}
	return def
}

func getFloat(m map[string]interface{}, key string, def float64) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	return def
}

func getBool(m map[string]interface{}, key string, def bool) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return def
}

func handleExtractMemories(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Content string `json:"content" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "content is required"})
			return
		}

		userID := middleware.GetUserID(c)
		svc := services.NewExtractionService(db, userID)
		result := svc.ExtractFromConversation(req.Content)
		c.JSON(http.StatusOK, result)
	}
}

func handleExtractAndSave(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Content  string `json:"content" binding:"required"`
			AutoSave bool   `json:"auto_save"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "content is required"})
			return
		}

		userID := middleware.GetUserID(c)
		svc := services.NewExtractionService(db, userID)
		result := svc.ExtractFromConversation(req.Content)

		if req.AutoSave && result.Count > 0 {
			saved := 0
			for _, em := range result.Memories {
				data := map[string]interface{}{
					"key":         em.Key,
					"value":       em.Value,
					"layer":       em.Layer,
					"memory_type": em.MemoryType,
					"importance":  em.Importance,
					"tags":        em.Tags,
					"source":      em.Source,
				}
				memSvc := services.NewMemoryService(db)
				if _, err := memSvc.Create(userID, data); err == nil {
					saved++
				}
			}
			c.JSON(http.StatusOK, gin.H{
				"extracted": result.Count,
				"saved":     saved,
				"memories":  result.Memories,
				"warnings":  result.Warnings,
			})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func handleVerifyMemory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}

		var memory models.Memory
		if err := db.Where("id = ? AND user_id = ?", id, userID).First(&memory).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "memory not found"})
			return
		}

		now := time.Now()
		memory.VerifiedAt = &now
		memory.VerifyCount++
		if err := db.Save(&memory).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":           memory.ID,
			"verified_at":  memory.VerifiedAt.Format("2006-01-02 15:04:05"),
			"verify_count": memory.VerifyCount,
			"message":      "memory verified successfully",
		})
	}
}

func handleBatchValidate(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		svc := services.NewValidationService(db)
		result, err := svc.BatchValidate(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleListTemplates(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		svc := services.NewTemplateService(db)
		templates, err := svc.ListTemplates(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": templates})
	}
}

func handleCreateTemplate(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		var req services.MemoryTemplate
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		svc := services.NewTemplateService(db)
		if err := svc.CreateTemplate(userID, req); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"created": true})
	}
}

func handleDeleteTemplate(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		name := c.Param("name")
		svc := services.NewTemplateService(db)
		if err := svc.DeleteTemplate(userID, name); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"deleted": true})
	}
}

func handleApplyTemplate(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		name := c.Param("name")
		var req struct {
			Values map[string]string `json:"values"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		svc := services.NewTemplateService(db)
		key, value, layer, err := svc.ApplyTemplate(userID, name, req.Values)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"key":   key,
			"value": value,
			"layer": layer,
		})
	}
}

func handleScanSecrets(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Content string `json:"content" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "content is required"})
			return
		}

		if len(req.Content) > 1*1024*1024 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "content too large (max 1MB)"})
			return
		}

		result := services.ScanSecrets(req.Content)
		c.JSON(http.StatusOK, result)
	}
}

func handleCreateSessionMemory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		session := models.SessionMemory{
			UserID:        userID,
			SessionID:     getString(data, "session_id", ""),
			Title:         getString(data, "title", ""),
			CurrentState:  getString(data, "current_state", ""),
			TaskSpec:      getString(data, "task_spec", ""),
			FilesAndFuncs: getString(data, "files_and_funcs", ""),
			Workflow:      getString(data, "workflow", ""),
			Errors:        getString(data, "errors", ""),
			Docs:          getString(data, "docs", ""),
			Learnings:     getString(data, "learnings", ""),
			KeyResults:    getString(data, "key_results", ""),
			Worklog:       getString(data, "worklog", ""),
			Status:        getString(data, "status", "active"),
		}

		if v, ok := data["token_count"].(float64); ok {
			session.TokenCount = int(v)
		}
		if v, ok := data["compressed_from"].(string); ok {
			session.CompressedFrom = v
		}

		ttlHours := 24
		if v, ok := data["ttl_hours"].(float64); ok && v > 0 {
			ttlHours = int(v)
		}
		expiresAt := time.Now().Add(time.Duration(ttlHours) * time.Hour)
		session.ExpiresAt = &expiresAt

		if err := db.Create(&session).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, session)
	}
}

func handleListSessionMemories(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		sessionID := c.Query("session_id")
		status := c.Query("status")

		var sessions []models.SessionMemory
		query := db.Where("user_id = ?", userID)
		if sessionID != "" {
			query = query.Where("session_id = ?", sessionID)
		}
		if status != "" {
			query = query.Where("status = ?", status)
		}
		showExpired := c.Query("show_expired") == "true"
		if !showExpired {
			query = query.Where("expires_at IS NULL OR expires_at > ?", time.Now())
		}
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
		if limit < 1 || limit > 200 {
			limit = 50
		}
		logDBErr("load sessions", query.Order("updated_at DESC").Limit(limit).Find(&sessions).Error)

		c.JSON(http.StatusOK, gin.H{"items": sessions})
	}
}

func handleGetSessionMemory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}

		var session models.SessionMemory
		if err := db.Where("id = ? AND user_id = ?", id, userID).First(&session).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "session memory not found"})
			return
		}
		c.JSON(http.StatusOK, session)
	}
}

func handleUpdateSessionMemory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}
		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var session models.SessionMemory
		if err := db.Where("id = ? AND user_id = ?", id, userID).First(&session).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "session memory not found"})
			return
		}

		updatables := []string{
			"title", "current_state", "task_spec", "files_and_funcs",
			"workflow", "errors", "docs", "learnings", "key_results",
			"worklog", "status", "session_id", "compressed_from",
		}
		for _, field := range updatables {
			if v, ok := data[field].(string); ok {
				switch field {
				case "title":
					session.Title = v
				case "current_state":
					session.CurrentState = v
				case "task_spec":
					session.TaskSpec = v
				case "files_and_funcs":
					session.FilesAndFuncs = v
				case "workflow":
					session.Workflow = v
				case "errors":
					session.Errors = v
				case "docs":
					session.Docs = v
				case "learnings":
					session.Learnings = v
				case "key_results":
					session.KeyResults = v
				case "worklog":
					session.Worklog = v
				case "status":
					session.Status = v
				case "session_id":
					session.SessionID = v
				case "compressed_from":
					session.CompressedFrom = v
				}
			}
		}
		if v, ok := data["token_count"].(float64); ok {
			session.TokenCount = int(v)
		}

		if err := db.Save(&session).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, session)
	}
}

func handleDeleteSessionMemory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}

		if err := db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.SessionMemory{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	}
}

func toJSONStr(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case []interface{}:
		b, _ := json.Marshal(val)
		return string(b)
	case []string:
		b, _ := json.Marshal(val)
		return string(b)
	default:
		b, _ := json.Marshal(val)
		return string(b)
	}
}

func handleListAPIKeys(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		svc := services.NewAPIKeyService(db)
		keys, err := svc.List(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": keys})
	}
}

func handleCreateAPIKey(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		var req struct {
			Name        string `json:"name" binding:"required"`
			Permissions string `json:"permissions"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
			return
		}

		svc := services.NewAPIKeyService(db)
		var apiKey *models.APIKey
		var rawKey string
		var err error
		if req.Permissions != "" {
			apiKey, rawKey, err = svc.CreateWithPermissions(userID, req.Name, req.Permissions, "")
		} else {
			apiKey, rawKey, err = svc.Create(userID, req.Name)
		}
		if err != nil {
			status := http.StatusInternalServerError
			if strings.Contains(err.Error(), "maximum") {
				status = http.StatusConflict
			}
			c.JSON(status, gin.H{"error": err.Error()})
			return
		}

		remaining := services.MaxAPIKeysPerUser - int(svc.Count(userID))

		auditLog(db, c, "api_key.create", fmt.Sprintf("id:%d", apiKey.ID), fmt.Sprintf("name:%s prefix:%s perms:%s", apiKey.Name, apiKey.KeyPrefix, apiKey.Permissions))

		c.JSON(http.StatusCreated, gin.H{
			"id":          apiKey.ID,
			"name":        apiKey.Name,
			"key_prefix":  apiKey.KeyPrefix,
			"permissions": apiKey.Permissions,
			"key":         rawKey,
			"created_at":  apiKey.CreatedAt,
			"remaining":   remaining,
			"message":     "please save the API key securely, it will not be shown again",
		})
	}
}

func handleDeleteAPIKey(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}

		svc := services.NewAPIKeyService(db)
		if err := svc.Delete(userID, uint(id)); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		auditLog(db, c, "api_key.delete", fmt.Sprintf("id:%d", id), "")

		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
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
			"layer":        getString(data, "layer", "episodic"),
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
				"layer":        getString(m, "layer", "episodic"),
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
