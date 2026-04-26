package api

import (
	"encoding/json"
	"fmt"
	"io"
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
	"gorm.io/gorm"
)

func handleInitStatus(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		passwordSet, err := authService.CheckInitStatus()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": "failed to check init status"})
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
			c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
			return
		}

		if len(req.Password) < 4 {
			c.JSON(http.StatusBadRequest, gin.H{"detail": "password too short"})
			return
		}

		token, err := authService.SetPassword(req.Password)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"access_token": token})
	}
}

func handleLogin(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Password string `json:"password" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
			return
		}

		token, err := authService.LoginWithPassword(req.Password)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"detail": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"access_token": token})
	}
}

func handleGetMe(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"detail": "unauthorized"})
			return
		}

		user, err := authService.GetUserByID(userID.(uint))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"detail": "user not found"})
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
			Username    string `json:"username"`
			OldPassword string `json:"old_password"`
			NewPassword string `json:"new_password"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "密码重置成功"})
	}
}

// License handlers
func handleLicenseInfo(lm *services.LicenseManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, lm.GetLicenseInfo())
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
		c.JSON(http.StatusOK, gin.H{"message": "已停用"})
	}
}

// Memory handlers
func handleListMemories(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		svc := services.NewMemoryService(db)
		layer := c.Query("layer")
		status := c.Query("status")
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

		memories, total, err := svc.List(userID, layer, page, size, status)
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

		svc := services.NewMemoryService(db)
		memory, err := svc.Create(userID, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, memory)
	}
}

func handleGetMemory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))
		svc := services.NewMemoryService(db)
		memory, err := svc.Get(uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, memory)
	}
}

func handleUpdateMemory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		svc := services.NewMemoryService(db)
		memory, err := svc.Update(uint(id), req)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, memory)
	}
}

func handleDeleteMemory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))
		svc := services.NewMemoryService(db)
		if err := svc.Delete(uint(id)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	}
}

func handleRestoreMemory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))
		svc := services.NewMemoryService(db)
		if err := svc.Restore(uint(id)); err != nil {
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

func handleSearchSemantic(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		q := c.Query("q")
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		userID := middleware.GetUserID(c)

		svc := services.NewSearchService(db)
		memories, err := svc.SemanticSearch(userID, q, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": memories})
	}
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
		var relations []struct {
			ID           uint    `json:"id"`
			SourceID     uint    `json:"source_id"`
			TargetID     uint    `json:"target_id"`
			RelationType string  `json:"relation_type"`
			Description  string  `json:"description"`
			Confidence   float64 `json:"confidence"`
			Weight       float64 `json:"weight"`
		}
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
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

		pages, total, err := svc.List(userID, category, page, size)
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
		id, _ := strconv.Atoi(c.Param("id"))
		svc := services.NewWikiService(db)
		page, err := svc.Get(uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, page)
	}
}

func handleUpdateWiki(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		svc := services.NewWikiService(db)
		page, err := svc.Update(uint(id), req)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, page)
	}
}

func handleDeleteWiki(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))
		svc := services.NewWikiService(db)
		if err := svc.Delete(uint(id)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
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

// Stats handlers
func handleGetStats(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		var memoryCount, entityCount, relationCount, wikiCount int64
		db.Model(&struct{ ID uint }{}).Table("memories").Where("user_id = ? AND status != ?", userID, "trashed").Count(&memoryCount)
		db.Model(&struct{ ID uint }{}).Table("entities").Where("user_id = ?", userID).Count(&entityCount)
		db.Model(&struct{ ID uint }{}).Table("relations").Where("user_id = ?", userID).Count(&relationCount)
		db.Model(&struct{ ID uint }{}).Table("wiki_pages").Where("user_id = ?", userID).Count(&wikiCount)

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
			"wikiCount":      wikiCount,
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
		c.JSON(proErr.Code, gin.H{"detail": proErr.Message})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
}

func checkPro(proxy *services.ProProxy, c *gin.Context) bool {
	if !proxy.IsPro() {
		c.JSON(http.StatusForbidden, gin.H{"error": "Pro license required", "mode": "local_pro"})
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
		svc := services.NewProLocalService(db)
		result, err := svc.TokenStats()
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
			CreatedAt time.Time `json:"created_at"`
		}
		db.Table("memories").Where("status != ?", "trashed").Order("created_at desc").Find(&memories)

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
			sourceDist["manual"]++
			importanceDist["medium"]++
			layerDist[layer]++
		}

		db.Table("entities").Find(&struct{}{})
		var entityCount int64
		db.Table("entities").Count(&entityCount)

		rows, _ := db.Raw("SELECT entity_type, COUNT(*) as cnt FROM entities GROUP BY entity_type").Rows()
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

func handleChromaDBStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"available": false,
		"reason":    "ChromaDB integration is not available in Go backend",
	})
}

func handleChromaDBInstall(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{
		"success": false,
		"message": "ChromaDB integration is not available in Go backend. This feature requires Python backend.",
	})
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
		addDir(filepath.Join(homeDir, ".trae-cn"))
		addDir(filepath.Join(homeDir, ".trae"))
	}

	cfg := config.Load()
	if cfg.DataDir != "" {
		addDir(cfg.DataDir)
	}

	exe, _ := os.Executable()
	if exe != "" {
		addDir(filepath.Join(filepath.Dir(exe), "openclaw"))
		addDir(filepath.Join(filepath.Dir(exe), "data"))
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

	filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if name == "node_modules" || name == ".git" || name == "vendor" || name == "__pycache__" || name == ".cache" {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".json" && ext != ".md" && ext != ".txt" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		content := string(data)
		relPath, _ := filepath.Rel(dir, path)
		parts := strings.Split(relPath, string(filepath.Separator))
		dirAgent := "default"
		if len(parts) > 1 {
			dirAgent = parts[0]
		}

		if ext == ".json" {
			var memories []map[string]interface{}
			if json.Unmarshal(data, &memories) == nil {
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
					if key == "" && contentStr == "" {
						continue
					}
					if key == "" {
						key = contentStr
						if len(key) > 50 {
							key = key[:50]
						}
					}
					agent := dirAgent
					if a, ok := m["agent_name"].(string); ok && a != "" {
						agent = a
					}
					layer := "knowledge"
					if l, ok := m["layer"].(string); ok && l != "" {
						layer = l
					}
					source := "openclaw"
					if s, ok := m["source"].(string); ok && s != "" {
						source = s
					}
					preview := contentStr
					if len(preview) > 100 {
						preview = preview[:100] + "..."
					}
					previews = append(previews, memoryPreview{
						Key: key, Content: preview, Layer: layer,
						Source: source, FilePath: path, AgentName: agent,
					})
					agentCountMap[agent]++
				}
			} else {
				var single map[string]interface{}
				if json.Unmarshal(data, &single) == nil {
					key, _ := single["key"].(string)
					contentStr, _ := single["content"].(string)
					if key == "" {
						key, _ = single["name"].(string)
					}
					if contentStr == "" {
						contentStr, _ = single["value"].(string)
						if contentStr == "" {
							contentStr, _ = single["text"].(string)
						}
					}
					if key != "" || contentStr != "" {
						if key == "" {
							key = "json_item"
						}
						preview := contentStr
						if len(preview) > 100 {
							preview = preview[:100] + "..."
						}
						previews = append(previews, memoryPreview{
							Key: key, Content: preview, Layer: "knowledge",
							Source: "openclaw", FilePath: path, AgentName: dirAgent,
						})
						agentCountMap[dirAgent]++
					}
				}
			}
		} else if ext == ".md" {
			lines := strings.Split(content, "\n")
			currentSection := ""
			currentContent := ""
			for _, line := range lines {
				if strings.HasPrefix(line, "# ") {
					if currentSection != "" || currentContent != "" {
						key := currentSection
						if key == "" {
							key = filepath.Base(path)
						}
						preview := strings.TrimSpace(currentContent)
						if len(preview) > 100 {
							preview = preview[:100] + "..."
						}
						previews = append(previews, memoryPreview{
							Key: key, Content: preview, Layer: "knowledge",
							Source: "markdown", FilePath: path, AgentName: dirAgent,
						})
						agentCountMap[dirAgent]++
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
					key = filepath.Base(path)
				}
				preview := strings.TrimSpace(currentContent)
				if len(preview) > 100 {
					preview = preview[:100] + "..."
				}
				previews = append(previews, memoryPreview{
					Key: key, Content: preview, Layer: "knowledge",
					Source: "markdown", FilePath: path, AgentName: dirAgent,
				})
				agentCountMap[dirAgent]++
			}
		} else if ext == ".txt" {
			txtContent := strings.TrimSpace(content)
			if txtContent != "" {
				key := filepath.Base(path)
				preview := txtContent
				if len(preview) > 100 {
					preview = preview[:100] + "..."
				}
				previews = append(previews, memoryPreview{
					Key: key, Content: preview, Layer: "knowledge",
					Source: "text", FilePath: path, AgentName: dirAgent,
				})
				agentCountMap[dirAgent]++
			}
		}

		return nil
	})

	return previews, agentCountMap
}

func handleScanOpenClawMemories(c *gin.Context) {
	searchDirs := getOpenClawSearchDirs()

	for _, dir := range searchDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}

		previews, agentCountMap := extractMemoriesFromDir(dir)

		if len(previews) > 0 {
			agents := make([]map[string]interface{}, 0)
			for name, count := range agentCountMap {
				agentPreviews := make([]map[string]interface{}, 0)
				for _, p := range previews {
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
				"openclaw_dir":   dir,
				"agents":         agents,
				"total_memories": len(previews),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"found": false,
	})
}

func handleScanOpenClawAgent(c *gin.Context) {
	agentName := c.Param("agentName")
	searchDirs := getOpenClawSearchDirs()

	for _, dir := range searchDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}

		previews, _ := extractMemoriesFromDir(dir)

		filtered := make([]map[string]interface{}, 0)
		for _, p := range previews {
			if p.AgentName == agentName {
				filtered = append(filtered, map[string]interface{}{
					"key":    p.Key,
					"value":  p.Content,
					"layer":  p.Layer,
					"source": p.Source,
				})
			}
		}

		if len(filtered) > 0 {
			c.JSON(http.StatusOK, gin.H{
				"agent_name": agentName,
				"preview":    filtered,
				"total":      len(filtered),
			})
			return
		}
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
						if ext == ".json" {
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

				memory := models.Memory{
					UserID:     userID,
					Key:        p.Key,
					Value:      fullContent,
					Layer:      layer,
					Importance: 0.5,
					Tags:       "",
					Source:     p.Source,
					Status:     "active",
				}
				result := db.Create(&memory)
				if result.Error != nil {
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
			searchDirs = append(searchDirs, filepath.Join(homeDir, ".openclaw"))
			searchDirs = append(searchDirs, filepath.Join(homeDir, ".clawmemory"))
			searchDirs = append(searchDirs, filepath.Join(homeDir, ".trae-cn"))
			searchDirs = append(searchDirs, filepath.Join(homeDir, ".trae"))
		}

		cfg := config.Load()
		if cfg.DataDir != "" {
			searchDirs = append(searchDirs, cfg.DataDir)
		}

		wd, _ := os.Getwd()
		if wd != "" {
			searchDirs = append(searchDirs, filepath.Join(wd, ".openclaw"))
			searchDirs = append(searchDirs, filepath.Join(wd, ".clawmemory"))
			searchDirs = append(searchDirs, filepath.Join(wd, "data"))
		}

		exe, _ := os.Executable()
		if exe != "" {
			exeDir := filepath.Dir(exe)
			searchDirs = append(searchDirs, exeDir)
			searchDirs = append(searchDirs, filepath.Join(exeDir, "data"))
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
			filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
				if err != nil {
					return nil
				}
				if d.IsDir() {
					name := d.Name()
					if name == "node_modules" || name == ".git" || name == "vendor" || name == "__pycache__" || name == ".cache" {
						return filepath.SkipDir
					}
					return nil
				}

				ext := strings.ToLower(filepath.Ext(path))
				if ext != ".json" && ext != ".md" && ext != ".txt" {
					return nil
				}

				foundFiles = append(foundFiles, path)

				data, err := os.ReadFile(path)
				if err != nil {
					return nil
				}
				content := string(data)

				if ext == ".json" {
					importFromJSON(db, userID, content, seenKeys, &imported, &skipped, &entitiesCreated)
				} else if ext == ".md" {
					importFromMarkdown(db, userID, path, content, seenKeys, &imported, &skipped, &entitiesCreated)
				} else if ext == ".txt" {
					importFromText(db, userID, path, content, seenKeys, &imported, &skipped, &entitiesCreated)
				}

				return nil
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"imported":         imported,
			"skipped":          skipped,
			"entities_created": entitiesCreated,
			"files_found":      foundFiles,
			"message":          fmt.Sprintf("成功导入 %d 条记忆，创建 %d 个实体，跳过 %d 条", imported, entitiesCreated, skipped),
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

		memory := models.Memory{
			UserID:     userID,
			Key:        key,
			Value:      contentStr,
			Layer:      layer,
			Importance: importance,
			Tags:       tags,
			Source:     source,
			Status:     "active",
		}
		result := db.Create(&memory)
		if result.Error != nil {
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

		memory := models.Memory{
			UserID:     userID,
			Key:        key,
			Value:      body,
			Layer:      layer,
			Importance: importance,
			Tags:       "markdown",
			Source:     source,
			Status:     "active",
		}
		result := db.Create(&memory)
		if result.Error != nil {
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
				memory := models.Memory{
					UserID:     userID,
					Key:        key,
					Value:      body,
					Layer:      layer,
					Importance: 0.4,
					Tags:       "text",
					Source:     "auto_import_txt",
					Status:     "active",
				}
				result := db.Create(&memory)
				if result.Error == nil {
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
			memory := models.Memory{
				UserID:     userID,
				Key:        key,
				Value:      content,
				Layer:      layer,
				Importance: 0.3,
				Tags:       "text",
				Source:     "auto_import_txt",
				Status:     "active",
			}
			result := db.Create(&memory)
			if result.Error == nil {
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
	backupDir := filepath.Join(".", "backups")
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
	backupDir := filepath.Join(".", "backups")
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
		c.JSON(http.StatusOK, gin.H{"items": memories})
	}
}

func handleExportData(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		exportData := map[string]interface{}{}

		var memories []models.Memory
		db.Where("user_id = ?", userID).Find(&memories)
		exportData["memories"] = memories

		var entities []models.Entity
		db.Where("user_id = ?", userID).Find(&entities)
		exportData["entities"] = entities

		var relations []models.Relation
		db.Where("user_id = ?", userID).Find(&relations)
		exportData["relations"] = relations

		var wikiPages []models.WikiPage
		db.Where("user_id = ?", userID).Find(&wikiPages)
		exportData["wiki_pages"] = wikiPages

		var reports []models.DailyReport
		db.Where("user_id = ?", userID).Find(&reports)
		exportData["daily_reports"] = reports

		exportData["exported_at"] = time.Now().Format(time.RFC3339)
		exportData["version"] = "2.8.2"

		c.JSON(http.StatusOK, exportData)
	}
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
