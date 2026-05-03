package api

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"clawmemory/internal/config"
	"clawmemory/internal/middleware"
	"clawmemory/internal/models"
	"clawmemory/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

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
			Password string `json:"password" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if len(req.Password) < 4 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "password too short"})
			return
		}

		token, err := authService.SetPassword(req.Password)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"access_token": token})
	}
}

func handleLogin(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if req.Username == "" {
			req.Username = "admin"
		}

		token, err := authService.Login(req.Username, req.Password)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"access_token": token})
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
			"id":       user.ID,
			"username": user.Username,
		})
	}
}

func handleRegister(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		user, err := authService.Register(req.Username, req.Password)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"id":       user.ID,
			"username": user.Username,
		})
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
		if len(req.NewPassword) < 4 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "new_password too short"})
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
		passwordSet, err := authService.CheckInitStatus()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check init status"})
			return
		}
		if !passwordSet {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no password set yet, please set a password first"})
			return
		}

		var req struct {
			Username    string `json:"username"`
			NewPassword string `json:"new_password"`
			Confirm     bool   `json:"confirm"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if !req.Confirm {
			c.JSON(http.StatusBadRequest, gin.H{"error": "confirm is required"})
			return
		}
		if req.Username == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "username is required", "hint": "use terminal command: ./clawmemory --reset-password NEW_PASSWORD"})
			return
		}
		if req.NewPassword == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "new_password is required"})
			return
		}
		if len(req.NewPassword) < 4 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "new_password too short"})
			return
		}

		var user models.User
		if err := authService.FindFirstUser(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no user found", "hint": "use terminal command: ./clawmemory --reset-password NEW_PASSWORD"})
			return
		}

		if user.Username != req.Username {
			c.JSON(http.StatusBadRequest, gin.H{"error": "username mismatch", "hint": "use terminal command: ./clawmemory --reset-password NEW_PASSWORD"})
			return
		}

		if err := authService.ChangePassword(user.ID, "", req.NewPassword); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "password reset successful"})
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
		if len(req.NewPassword) < 4 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "new_password too short"})
			return
		}
		if err := authService.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "password changed successfully"})
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

// License handlers
func handleLicenseInfo(proxy *services.ProProxy) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, proxy.GetLicenseInfo())
	}
}

func handleLicenseActivate(lm *services.LicenseManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			LicenseKey string `json:"license_key" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result, err := lm.Activate(req.LicenseKey)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func handleLicenseDeactivate(lm *services.LicenseManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "deactivated"})
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
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

		memories, total, err := svc.List(userID, layer, page, size, status, memoryType)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"items": memories,
			"total": total,
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

		response := gin.H{"memory": memory}
		if secretResult.Found {
			response["secret_warning"] = secretResult
		}
		c.JSON(http.StatusCreated, response)
	}
}

func handleGetMemory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, _ := strconv.Atoi(c.Param("id"))
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
		id, _ := strconv.Atoi(c.Param("id"))
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

		response := gin.H{"memory": memory}
		if secretResult != nil && secretResult.Found {
			response["secret_warning"] = secretResult
		}
		c.JSON(http.StatusOK, response)
	}
}

func handleDeleteMemory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, _ := strconv.Atoi(c.Param("id"))
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
		id, _ := strconv.Atoi(c.Param("id"))
		svc := services.NewMemoryService(db)
		if err := svc.Restore(userID, uint(id)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "restored"})
	}
}

func handleSearchKeyword(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		q := c.Query("q")
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		userID := middleware.GetUserID(c)

		svc := services.NewMemoryService(db)
		memories, err := svc.SearchKeyword(userID, q, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": memories})
	}
}

func handleExternalPushConversation(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		var req services.ConversationPushRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if req.AgentName == "" {
			req.AgentName = "openclaw"
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

		auditLog(db, c, "external.conversation_push", req.SessionID, fmt.Sprintf("agent:%s messages:%d created:%d", req.AgentName, len(req.Messages), created))

		c.JSON(http.StatusOK, gin.H{
			"created":  created,
			"messages": len(req.Messages),
			"agent":    req.AgentName,
		})
	}
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
			"message":     "sync completed",
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

func handleDecryptMemory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, _ := strconv.Atoi(c.Param("id"))

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

func handleSearchSemantic(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		q := c.Query("q")
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		userID := middleware.GetUserID(c)

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
	db.Where("user_id = ? AND id IN ?", userID, memIDs).Find(&memories)

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

		entities, total, err := svc.ListEntities(userID, entityType, page, size)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": entities, "total": total})
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
		var relations []models.Relation
		db.Where("user_id = ?", userID).Find(&relations)
		c.JSON(http.StatusOK, gin.H{"items": relations})
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

		pages, total, err := svc.List(userID, category, status, page, size)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": pages, "total": total})
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
		id, _ := strconv.Atoi(c.Param("id"))
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
		id, _ := strconv.Atoi(c.Param("id"))
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
		id, _ := strconv.Atoi(c.Param("id"))
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
		id, _ := strconv.Atoi(c.Param("id"))
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
		id, _ := strconv.Atoi(c.Param("id"))
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
		c.JSON(http.StatusOK, gin.H{"llm_available": false})
	}
}

func handleWikiAIExtract(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "AI extraction not available in OSS version"})
	}
}

func handleWikiRefine(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "AI refinement not available in OSS version"})
	}
}

// Report handlers
func handleListReports(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		svc := services.NewDailyReportService(db)
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

		reports, total, err := svc.List(userID, page, size)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": reports, "total": total})
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
		db.Model(&struct{ ID uint }{}).Table("memories").Where("user_id = ? AND status != ?", userID, "trashed").Count(&memoryCount)
		db.Model(&struct{ ID uint }{}).Table("entities").Where("user_id = ?", userID).Count(&entityCount)
		db.Model(&struct{ ID uint }{}).Table("relations").Where("user_id = ?", userID).Count(&relationCount)
		db.Model(&struct{ ID uint }{}).Table("projects").Where("user_id = ?", userID).Count(&projectCount)

		layerStats := make(map[string]int64)
		rows, _ := db.Raw("SELECT COALESCE(layer, 'knowledge') as layer, COUNT(*) as cnt FROM memories WHERE user_id = ? AND status != 'trashed' GROUP BY layer", userID).Rows()
		for rows.Next() {
			var layer string
			var cnt int64
			rows.Scan(&layer, &cnt)
			layerStats[layer] = cnt
		}
		rows.Close()

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
		db.Table("memories").Where("user_id = ? AND status != ?", userID, "trashed").Order("created_at desc").Limit(10).Find(&recentMemories)

		recentMemoriesJson := make([]map[string]interface{}, 0)
		for _, m := range recentMemories {
			recentMemoriesJson = append(recentMemoriesJson, map[string]interface{}{
				"id":         m.ID,
				"key":        m.Key,
				"layer":      m.Layer,
				"created_at": m.CreatedAt.Format("2006-01-02 15:04:05"),
			})
		}

		var licenseInfo map[string]interface{}
		licenseInfo = map[string]interface{}{
			"tier":       "oss",
			"active":     false,
			"type":       "",
			"expires_at": "",
			"device_slot": "",
		}

		c.JSON(http.StatusOK, gin.H{
			"memoryCount":    memoryCount,
			"entityCount":    entityCount,
			"relationCount":  relationCount,
			"projectCount":  projectCount,
			"layerStats":     layerStats,
			"recentMemories": recentMemoriesJson,
			"license":        licenseInfo,
			"passwordSet":    true,
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

func proErrorHandler(c *gin.Context, err error) {
	if proErr, ok := err.(*services.ProError); ok {
		c.JSON(proErr.Code, gin.H{"error": proErr.Message})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}

func checkPro(proxy *services.ProProxy, c *gin.Context) bool {
	if !proxy.IsPro() {
		c.JSON(http.StatusForbidden, gin.H{"error": "Pro license required"})
		return false
	}
	return true
}

func handleProDecayStats(proxy *services.ProProxy, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !checkPro(proxy, c) { return }
		userID := middleware.GetUserID(c)
		svc := services.NewProLocalService(db)
		result, err := svc.DecayStats(userID)
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProDecayApply(proxy *services.ProProxy, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !checkPro(proxy, c) { return }
		userID := middleware.GetUserID(c)
		svc := services.NewProLocalService(db)
		result, err := svc.DecayApply(userID)
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProReinforce(proxy *services.ProProxy, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !checkPro(proxy, c) { return }
		userID := middleware.GetUserID(c)
		idStr := c.Param("id")
		id, _ := strconv.ParseUint(idStr, 10, 64)
		svc := services.NewProLocalService(db)
		result, err := svc.ReinforceMemory(userID, uint(id))
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProPruneSuggest(proxy *services.ProProxy, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !checkPro(proxy, c) { return }
		userID := middleware.GetUserID(c)
		svc := services.NewProLocalService(db)
		result, err := svc.PruneSuggest(userID)
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProConflictScan(proxy *services.ProProxy, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !checkPro(proxy, c) { return }
		userID := middleware.GetUserID(c)
		svc := services.NewProLocalService(db)
		result, err := svc.ConflictScan(userID)
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProConflictResolve(proxy *services.ProProxy, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !checkPro(proxy, c) { return }
		userID := middleware.GetUserID(c)
		var req struct {
			Strategy string `json:"strategy"`
		}
		c.ShouldBindJSON(&req)
		indexStr := c.Param("index")
		index, _ := strconv.Atoi(indexStr)
		svc := services.NewProLocalService(db)
		result, err := svc.ConflictResolve(userID, index, req.Strategy)
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProTokenRoute(proxy *services.ProProxy, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !checkPro(proxy, c) { return }
		message := c.Query("message")
		contextLength := 0
		if cl := c.Query("context_length"); cl != "" {
			if n, err := strconv.Atoi(cl); err == nil {
				contextLength = n
			}
		}
		svc := services.NewProLocalService(db)
		result, err := svc.TokenRoute(message, contextLength)
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProTokenStats(proxy *services.ProProxy, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !checkPro(proxy, c) { return }
		userID := middleware.GetUserID(c)
		svc := services.NewProLocalService(db)
		result, err := svc.TokenStats(userID)
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProAIExtract(proxy *services.ProProxy, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !checkPro(proxy, c) { return }
		userID := middleware.GetUserID(c)
		svc := services.NewProLocalService(db)
		result, err := svc.AIExtract(userID)
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProAutoGraph(proxy *services.ProProxy, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !checkPro(proxy, c) { return }
		userID := middleware.GetUserID(c)
		var req struct {
			Overwrite bool `json:"overwrite"`
		}
		c.ShouldBindJSON(&req)
		svc := services.NewProLocalService(db)
		result, err := svc.AutoGraph(userID, req.Overwrite)
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProBackupSchedule(proxy *services.ProProxy, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !checkPro(proxy, c) { return }
		svc := services.NewProLocalService(db)
		result, err := svc.BackupSchedule()
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProSetBackupSchedule(proxy *services.ProProxy, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !checkPro(proxy, c) { return }
		var req struct {
			Enabled       bool `json:"enabled"`
			IntervalHours int  `json:"interval_hours"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		svc := services.NewProLocalService(db)
		result, err := svc.SetBackupSchedule(req.Enabled, req.IntervalHours)
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProCompressPreview(proxy *services.ProProxy, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !checkPro(proxy, c) { return }
		var req struct {
			Level string `json:"level"`
		}
		c.ShouldBindJSON(&req)
		if req.Level == "" {
			req.Level = "light"
		}
		userID := middleware.GetUserID(c)
		svc := services.NewProLocalService(db)
		result, err := svc.CompressPreview(userID, req.Level)
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProCompressApply(proxy *services.ProProxy, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !checkPro(proxy, c) { return }
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
		svc := services.NewProLocalService(db)
		result, err := svc.CompressApply(userID, req.Level)
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProCompressConfig(proxy *services.ProProxy, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !checkPro(proxy, c) { return }
		svc := services.NewProLocalService(db)
		result, err := svc.CompressConfig()
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProSetCompressConfig(proxy *services.ProProxy, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !checkPro(proxy, c) { return }
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		svc := services.NewProLocalService(db)
		result, err := svc.SetCompressConfig(req)
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProEvolutionInsights(proxy *services.ProProxy, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !checkPro(proxy, c) { return }
		userID := middleware.GetUserID(c)
		svc := services.NewProLocalService(db)
		result, err := svc.EvolutionInsights(userID)
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProEvolutionDiscover(proxy *services.ProProxy, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !checkPro(proxy, c) { return }
		userID := middleware.GetUserID(c)
		svc := services.NewProLocalService(db)
		result, err := svc.EvolutionDiscover(userID)
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProEvolutionInfer(proxy *services.ProProxy, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !checkPro(proxy, c) { return }
		userID := middleware.GetUserID(c)
		svc := services.NewProLocalService(db)
		result, err := svc.EvolutionInfer(userID)
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProEvolutionImportance(proxy *services.ProProxy, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !checkPro(proxy, c) { return }
		userID := middleware.GetUserID(c)
		svc := services.NewProLocalService(db)
		result, err := svc.EvolutionImportance(userID)
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProEvolutionPrefetch(proxy *services.ProProxy, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !checkPro(proxy, c) { return }
		var req struct {
			Context string `json:"context"`
		}
		c.ShouldBindJSON(&req)
		userID := middleware.GetUserID(c)
		svc := services.NewProLocalService(db)
		result, err := svc.EvolutionPrefetch(userID, req.Context)
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
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
			ID        uint      `json:"id"`
			Key       string    `json:"key"`
			Layer     string    `json:"layer"`
			Source    string    `json:"source"`
			Importance float64  `json:"importance"`
			CreatedAt time.Time `json:"created_at"`
		}
		db.Table("memories").Where("user_id = ? AND status != ?", userID, "trashed").Order("created_at desc").Find(&memories)

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

		db.Table("entities").Where("user_id = ?", userID).Find(&struct{}{})
		var entityCount int64
		db.Table("entities").Where("user_id = ?", userID).Count(&entityCount)

		rows, _ := db.Raw("SELECT entity_type, COUNT(*) as cnt FROM entities WHERE user_id = ? GROUP BY entity_type", userID).Rows()
		for rows.Next() {
			var etype string
			var cnt int64
			rows.Scan(&etype, &cnt)
			entityTypeDist[etype] = int(cnt)
		}
		rows.Close()

		c.JSON(http.StatusOK, gin.H{
			"dailyTrend":            dailyTrend,
			"dailyTokenTrend":       []map[string]interface{}{},
			"sourceDistribution":   sourceDist,
			"importanceDistribution": importanceDist,
			"tokenByLayer":          layerDist,
			"totalEstimatedTokens": len(memories) * 100,
			"topAccessed":           []map[string]interface{}{},
			"operationCounts":      map[string]int{},
			"entityTypeDistribution": entityTypeDist,
			"totalMemories":         len(memories),
			"days":                 days,
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

	homeDir, _ := os.UserHomeDir()
	if homeDir != "" {
		addDir(filepath.Join(homeDir, ".openclaw", "skills"))
		addDir(filepath.Join(homeDir, ".openclaw", "workspace", "skills"))
		addDir(filepath.Join(homeDir, ".agents", "skills"))

		openclawWorkspace := filepath.Join(homeDir, ".openclaw", "workspace")
		if info, err := os.Stat(openclawWorkspace); err == nil && info.IsDir() {
			addDir(filepath.Join(openclawWorkspace, "skills"))
			addDir(filepath.Join(openclawWorkspace, ".agents", "skills"))
		}
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
	homeDir, _ := os.UserHomeDir()
	if homeDir != "" {
		if scope == "global" {
			searchDirs = append(searchDirs, filepath.Join(homeDir, ".openclaw", "skills"))
			searchDirs = append(searchDirs, filepath.Join(homeDir, ".clawmemory", "skills"))
		}
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
		searchDirs = append(searchDirs, filepath.Join(wd, ".openclaw", "skills"))
		searchDirs = append(searchDirs, filepath.Join(wd, ".clawmemory", "skills"))
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
		id, _ := strconv.Atoi(c.Param("id"))

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

func getOpenClawSearchDirs() []string {
	dirs := []string{}
	seenDirs := make(map[string]bool)

	addDir := func(d string) {
		abs, err := filepath.Abs(d)
		if err != nil {
			abs = d
		}
		if !seenDirs[abs] {
			seenDirs[abs] = true
			dirs = append(dirs, abs)
		}
	}

	homeDir, _ := os.UserHomeDir()
	if homeDir != "" {
		addDir(filepath.Join(homeDir, ".openclaw"))
		addDir(filepath.Join(homeDir, ".clawmemory"))
	}

	exePath, _ := os.Executable()
	if exePath != "" {
		exeDir := filepath.Dir(exePath)
		addDir(filepath.Join(exeDir, "openclaw"))
		addDir(filepath.Join(exeDir, "data"))
	}

	wd, _ := os.Getwd()
	if wd != "" {
		addDir(filepath.Join(wd, ".openclaw"))
		addDir(filepath.Join(wd, ".clawmemory"))
		addDir(filepath.Join(wd, "data"))
	}

	return dirs
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

	memoryDir := filepath.Join(dir, "memory")
	if info, err := os.Stat(memoryDir); err == nil && info.IsDir() {
		previews, agentCountMap = extractWorkspaceMemory(memoryDir, dir, previews, agentCountMap)
	}

	sessionsDir := filepath.Join(dir, "agents")
	if info, err := os.Stat(sessionsDir); err == nil && info.IsDir() {
		previews, agentCountMap = extractSessionMemories(sessionsDir, previews, agentCountMap)
	}

	sqliteFiles, _ := filepath.Glob(filepath.Join(dir, "*.db"))
	sqliteFiles2, _ := filepath.Glob(filepath.Join(dir, "*.sqlite"))
	sqliteFiles3, _ := filepath.Glob(filepath.Join(dir, "*.sqlite3"))
	allSqlite := append(append(sqliteFiles, sqliteFiles2...), sqliteFiles3...)
	for _, dbFile := range allSqlite {
		previews, agentCountMap = extractSqliteMemories(dbFile, previews, agentCountMap)
	}

	return previews, agentCountMap
}

func extractSessionMemories(sessionsDir string, previews []memoryPreview, agentCountMap map[string]int) ([]memoryPreview, map[string]int) {
	agentDirs, _ := os.ReadDir(sessionsDir)
	for _, ad := range agentDirs {
		if !ad.IsDir() { continue }
		agentName := ad.Name()
		sessDir := filepath.Join(sessionsDir, agentName, "sessions")
		if info, err := os.Stat(sessDir); err != nil || !info.IsDir() { continue }
		files, _ := os.ReadDir(sessDir)
		for _, f := range files {
			if f.IsDir() { continue }
			ext := strings.ToLower(filepath.Ext(f.Name()))
			if ext != ".jsonl" { continue }
			path := filepath.Join(sessDir, f.Name())
			data, err := os.ReadFile(path)
			if err != nil || len(data) == 0 { continue }
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
		if line == "" { continue }
		var msg map[string]interface{}
		if json.Unmarshal([]byte(line), &msg) != nil { continue }
		msgType, _ := msg["type"].(string)
		text, _ := msg["text"].(string)
		if text == "" { text, _ = msg["content"].(string) }
		if text == "" { continue }

		isUserMsg := (msgType == "user" || msgType == "human")
		role := msgType
		if role == "" { role, _ = msg["role"].(string); if role == "" { role = "unknown" } }

		key := role + ": "
		if len(text) > 40 {
			key += text[:40] + "..."
		} else {
			key += text
		}

		preview := text
		if len(preview) > 200 { preview = preview[:200] + "..." }

		layer := "episodic"
		if isUserMsg { layer = "episodic" } else { layer = "semantic" }

		source := "session"
		if strings.Contains(filePath, "sessions") { source = "openclaw-session" }

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
		if f.IsDir() { continue }
		ext := strings.ToLower(filepath.Ext(f.Name()))
		if ext != ".md" { continue }
		path := filepath.Join(memoryDir, f.Name())
		data, err := os.ReadFile(path)
		if err != nil || len(data) == 0 { continue }
		previews, agentCountMap = parseMarkdownMemory(string(data), path, "workspace", previews, agentCountMap)
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
				if key == "" { key = filepath.Base(filePath) }
				preview := strings.TrimSpace(currentContent)
				if len(preview) > 100 { preview = preview[:100] + "..." }
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
		if key == "" { key = filepath.Base(filePath) }
		preview := strings.TrimSpace(currentContent)
		if len(preview) > 100 { preview = preview[:100] + "..." }
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
	sqlDB, _ := db.DB()
	if sqlDB != nil {
		defer sqlDB.Close()
	}

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
			"openclaw_dir":   strings.Join(foundDirs, ", "),
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
					db.Table("memories").Where("user_id = ? AND key = ?", userID, p.Key).Count(&count)
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
								if line == "" { continue }
								var msg map[string]interface{}
								if json.Unmarshal([]byte(line), &msg) != nil { continue }
								text, _ := msg["text"].(string)
								if text == "" { text, _ = msg["content"].(string) }
								key := ""
								msgType, _ := msg["type"].(string)
								if text != "" {
									key = msgType + ": "
									if len(text) > 40 { key += text[:40] + "..." } else { key += text }
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
		homeDir, _ := os.UserHomeDir()
		searchDirs := []string{}
		if homeDir != "" {
			searchDirs = append(searchDirs, filepath.Join(homeDir, ".openclaw", "workspace"))
		}

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
	db.Table("entities").Where("user_id = ? AND name = ?", userID, name).Count(&entityCount)
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
	userID := middleware.GetUserID(c)
	backupDir := filepath.Join(".", "backups", fmt.Sprintf("%d", userID))
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
			"filename":    entry.Name(),
			"size":        info.Size(),
			"created_at":  info.ModTime().Format("2006-01-02 15:04:05"),
		})
	}

	c.JSON(http.StatusOK, gin.H{"backups": backups})
}

func handleCreateBackup(c *gin.Context) {
	userID := middleware.GetUserID(c)
	backupDir := filepath.Join(".", "backups", fmt.Sprintf("%d", userID))
	os.MkdirAll(backupDir, 0755)

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("clawmemory_backup_%s.db", timestamp)
	backupPath := filepath.Join(backupDir, filename)

	dbPath := "clawmemory.db"
	if envDb := os.Getenv("DB_PATH"); envDb != "" {
		dbPath = envDb
	}

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
		userID := middleware.GetUserID(c)
		filename := c.Param("filename")
		if strings.Contains(filename, "..") || strings.ContainsAny(filename, "/\\") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid filename"})
			return
		}
		backupPath := filepath.Join(".", "backups", fmt.Sprintf("%d", userID), filename)

		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "backup not found"})
			return
		}

		c.FileAttachment(backupPath, filename)
	}
}

func handleRestoreBackup(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		filename := c.Param("filename")
		if strings.Contains(filename, "..") || strings.ContainsAny(filename, "/\\") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid filename"})
			return
		}
		backupPath := filepath.Join(".", "backups", fmt.Sprintf("%d", userID), filename)

		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "backup not found"})
			return
		}

		src, err := os.Open(backupPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot read backup file"})
			return
		}
		defer src.Close()

		dbPath := "clawmemory.db"
		if envDb := os.Getenv("DB_PATH"); envDb != "" {
			dbPath = envDb
		}
		dst, err := os.Create(dbPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot write database file"})
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "restore failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "backup restored successfully", "filename": filename})
	}
}

func handleDeleteBackup(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		filename := c.Param("filename")
		if strings.Contains(filename, "..") || strings.ContainsAny(filename, "/\\") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid filename"})
			return
		}
		backupPath := filepath.Join(".", "backups", fmt.Sprintf("%d", userID), filename)

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

func handleDecayApply(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		svc := services.NewDecayService(db)
		userID := middleware.GetUserID(c)
		result, err := svc.ApplyDecay(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
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
		db.Where("user_id = ? AND status = ?", userID, "trashed").Order("trashed_at DESC").Find(&memories)
		items := make([]*services.MemoryModel, 0, len(memories))
		for i := range memories {
			items = append(items, services.ToMemoryModel(&memories[i]))
		}
		c.JSON(http.StatusOK, gin.H{"items": items})
	}
}

func handleExportData(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		exportData := map[string]interface{}{}

		var memories []models.Memory
		db.Where("user_id = ?", userID).Find(&memories)
		memoryModels := make([]*services.MemoryModel, 0, len(memories))
		for i := range memories {
			memoryModels = append(memoryModels, services.ToMemoryModel(&memories[i]))
		}
		exportData["memories"] = memoryModels

		var entities []models.Entity
		db.Where("user_id = ?", userID).Find(&entities)
		exportData["entities"] = entities

		var relations []models.Relation
		db.Where("user_id = ?", userID).Find(&relations)
		exportData["relations"] = relations

		var wikiPages []models.WikiPage
		db.Where("user_id = ?", userID).Find(&wikiPages)
		exportData["wiki_pages"] = wikiPages

		var projects []models.Project
		db.Where("user_id = ?", userID).Find(&projects)
		exportData["projects"] = projects

		var projectNotes []models.ProjectNote
		db.Where("user_id = ?", userID).Find(&projectNotes)
		exportData["project_notes"] = projectNotes

		var reports []models.DailyReport
		db.Where("user_id = ?", userID).Find(&reports)
		exportData["daily_reports"] = reports

		exportData["exported_at"] = time.Now().Format(time.RFC3339)
		exportData["version"] = config.AppVersion

		c.JSON(http.StatusOK, exportData)
	}
}

func handleCheckUpdate(c *gin.Context) {
	type GitHubRelease struct {
		TagName string `json:"tag_name"`
		HTMLURL string `json:"html_url"`
		Body    string `json:"body"`
	}

	resp, err := http.Get("https://api.github.com/repos/" + config.GitHubRepo + "/releases/latest")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"current_version": config.AppVersion,
			"latest_version":  config.AppVersion,
			"has_update":      false,
			"error":           "failed to check update",
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		c.JSON(http.StatusOK, gin.H{
			"current_version": config.AppVersion,
			"latest_version":  config.AppVersion,
			"has_update":      false,
			"error":           "github api error",
		})
		return
	}

	var release GitHubRelease
	if json.NewDecoder(resp.Body).Decode(&release) != nil {
		c.JSON(http.StatusOK, gin.H{
			"current_version": config.AppVersion,
			"latest_version":  config.AppVersion,
			"has_update":      false,
			"error":           "failed to parse release",
		})
		return
	}

	latestVer := strings.TrimPrefix(release.TagName, "v")
	hasUpdate := latestVer != "" && latestVer != config.AppVersion && latestVer > config.AppVersion

	c.JSON(http.StatusOK, gin.H{
		"current_version": config.AppVersion,
		"latest_version":  latestVer,
		"has_update":      hasUpdate,
		"download_url":    release.HTMLURL,
		"release_notes":   release.Body,
		"github_repo":     config.GitHubRepoURL,
	})
}

func handleImportData(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userID := middleware.GetUserID(c)
		imported := 0

		if memories, ok := req["memories"].([]interface{}); ok {
			for _, m := range memories {
				if data, ok := m.(map[string]interface{}); ok {
					svc := services.NewMemoryService(db)
					if _, err := svc.Create(userID, data); err == nil {
						imported++
					}
				}
			}
		}

		if entities, ok := req["entities"].([]interface{}); ok {
			for _, e := range entities {
				if data, ok := e.(map[string]interface{}); ok {
					svc := services.NewKnowledgeService(db)
					if _, err := svc.CreateEntity(userID, data); err == nil {
						imported++
					}
				}
			}
		}

		if wikiPages, ok := req["wiki_pages"].([]interface{}); ok {
			for _, w := range wikiPages {
				if data, ok := w.(map[string]interface{}); ok {
					svc := services.NewWikiService(db)
					if _, err := svc.Create(userID, data); err == nil {
						imported++
					}
				}
			}
		}

		if projects, ok := req["projects"].([]interface{}); ok {
			for _, p := range projects {
				if data, ok := p.(map[string]interface{}); ok {
					svc := services.NewProjectService(db)
					if _, err := svc.Create(userID, data); err == nil {
						imported++
					}
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"imported": imported,
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

func handleMemoryRecommend(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		memoryIDStr := c.Query("memory_id")
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

		svc := services.NewRecommendService(db)
		var result map[string]interface{}
		var err error

		if memoryIDStr != "" {
			memoryID, _ := strconv.Atoi(memoryIDStr)
			result, err = svc.RecommendForMemory(userID, uint(memoryID), limit)
		} else {
			context := c.Query("context")
			result, err = svc.RecommendByContext(userID, context, limit)
		}

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
		status := c.Query("status")
		category := c.Query("category")

		svc := services.NewProjectService(db)
		projects, total, err := svc.List(userID, page, size, status, category)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": projects, "total": total})
	}
}

func handleGetProject(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, _ := strconv.Atoi(c.Param("id"))

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
		id, _ := strconv.Atoi(c.Param("id"))
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
		id, _ := strconv.Atoi(c.Param("id"))

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
		projectID, _ := strconv.Atoi(c.Param("id"))

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
		projectID, _ := strconv.Atoi(c.Param("id"))
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
		noteID, _ := strconv.Atoi(c.Param("noteId"))
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
		noteID, _ := strconv.Atoi(c.Param("noteId"))

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
		id, _ := strconv.Atoi(c.Param("id"))

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
	db.Create(&log)
}

func getString(m map[string]interface{}, key, def string) string {
	if v, ok := m[key].(string); ok && v != "" {
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
		id, _ := strconv.Atoi(c.Param("id"))

		var memory models.Memory
		if err := db.Where("id = ? AND user_id = ?", id, userID).First(&memory).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "memory not found"})
			return
		}

		now := time.Now()
		memory.VerifiedAt = &now
		memory.ReinforceCount++
		if err := db.Save(&memory).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":          memory.ID,
			"verified_at": memory.VerifiedAt.Format("2006-01-02 15:04:05"),
			"message":     "memory verified successfully",
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
		query.Order("updated_at DESC").Find(&sessions)

		c.JSON(http.StatusOK, gin.H{"items": sessions})
	}
}

func handleGetSessionMemory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, _ := strconv.Atoi(c.Param("id"))

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
		id, _ := strconv.Atoi(c.Param("id"))
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
		id, _ := strconv.Atoi(c.Param("id"))

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
			Name string `json:"name" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
			return
		}

		svc := services.NewAPIKeyService(db)
		apiKey, rawKey, err := svc.Create(userID, req.Name)
		if err != nil {
			status := http.StatusInternalServerError
			if strings.Contains(err.Error(), "maximum") {
				status = http.StatusConflict
			}
			c.JSON(status, gin.H{"error": err.Error()})
			return
		}

		remaining := services.MaxAPIKeysPerUser - int(svc.Count(userID))

		auditLog(db, c, "api_key.create", fmt.Sprintf("id:%d", apiKey.ID), fmt.Sprintf("name:%s prefix:%s", apiKey.Name, apiKey.KeyPrefix))

		c.JSON(http.StatusCreated, gin.H{
			"id":         apiKey.ID,
			"name":       apiKey.Name,
			"key_prefix": apiKey.KeyPrefix,
			"key":        rawKey,
			"created_at": apiKey.CreatedAt,
			"remaining":  remaining,
			"message":    "please save the API key securely, it will not be shown again",
		})
	}
}

func handleDeleteAPIKey(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, _ := strconv.Atoi(c.Param("id"))

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
		userID := middleware.GetUserID(c)
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
			source = "openclaw"
		}

		svc := services.NewMemoryService(db)
		memory, err := svc.Create(userID, map[string]interface{}{
			"key":         key,
			"value":       value,
			"layer":       getString(data, "layer", "episodic"),
			"importance":  data["importance"],
			"source":      source,
			"memory_type": getString(data, "memory_type", "knowledge"),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		auditLog(db, c, "external.memory_create", key, fmt.Sprintf("source:%s", source))

		c.JSON(http.StatusCreated, memory)
	}
}

func handleExternalBatchCreateMemories(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
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

		svc := services.NewMemoryService(db)
		var created, errorsCount int

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
				source = "openclaw"
			}

			_, err := svc.Create(userID, map[string]interface{}{
				"key":         key,
				"value":       value,
				"layer":       getString(m, "layer", "episodic"),
				"importance":  m["importance"],
				"source":      source,
				"memory_type": getString(m, "memory_type", "knowledge"),
			})
			if err != nil {
				errorsCount++
				continue
			}
			created++
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
		userID := middleware.GetUserID(c)
		q := c.Query("q")
		if q == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "q is required"})
			return
		}
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

		svc := services.NewMemoryService(db)
		memories, err := svc.SearchKeyword(userID, q, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": memories})
	}
}
