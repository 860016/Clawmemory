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
		svc := services.NewMemoryService(db)
		layer := c.Query("layer")
		status := c.Query("status")
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

		memories, total, err := svc.List(layer, page, size, status)
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
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		svc := services.NewMemoryService(db)
		memory, err := svc.Create(req)
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
		svc := services.NewKnowledgeService(db)
		entityType := c.Query("type")
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

		entities, total, err := svc.ListEntities(entityType, page, size)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": entities, "total": total})
	}
}

func handleCreateEntity(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		svc := services.NewKnowledgeService(db)
		entity, err := svc.CreateEntity(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, entity)
	}
}

func handleListRelations(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var relations []struct {
			ID           uint    `json:"id"`
			SourceID     uint    `json:"source_id"`
			TargetID     uint    `json:"target_id"`
			RelationType string  `json:"relation_type"`
			Description  string  `json:"description"`
			Confidence   float64 `json:"confidence"`
			Weight       float64 `json:"weight"`
		}
		db.Where("user_id = ?", 1).Find(&relations)
		c.JSON(http.StatusOK, gin.H{"items": relations})
	}
}

func handleCreateRelation(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		svc := services.NewKnowledgeService(db)
		relation, err := svc.CreateRelation(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, relation)
	}
}

func handleGetGraph(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		svc := services.NewKnowledgeService(db)
		entities, relations, err := svc.GetGraph()
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
		svc := services.NewWikiService(db)
		category := c.Query("category")
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

		pages, total, err := svc.List(category, page, size)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": pages, "total": total})
	}
}

func handleGetEntity(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		svc := services.NewKnowledgeService(db)
		entity, err := svc.GetEntity(uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "entity not found"})
			return
		}
		c.JSON(http.StatusOK, entity)
	}
}

func handleUpdateEntity(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		svc := services.NewKnowledgeService(db)
		entity, err := svc.UpdateEntity(uint(id), req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, entity)
	}
}

func handleDeleteEntity(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		svc := services.NewKnowledgeService(db)
		if err := svc.DeleteEntity(uint(id)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	}
}

func handleDeleteRelation(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		svc := services.NewKnowledgeService(db)
		if err := svc.DeleteRelation(uint(id)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	}
}

func handleCreateWiki(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		svc := services.NewWikiService(db)
		page, err := svc.Create(req)
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
		svc := services.NewDailyReportService(db)
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

		reports, total, err := svc.List(page, size)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": reports, "total": total})
	}
}

func handleCreateReport(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		svc := services.NewDailyReportService(db)
		report, err := svc.Create(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, report)
	}
}

func handleGetReportByDate(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		date := c.Param("date")
		svc := services.NewDailyReportService(db)
		report, err := svc.GetByDate(date)
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
		var memoryCount, entityCount, relationCount, wikiCount int64
		db.Model(&struct{ ID uint }{}).Table("memories").Where("status != ?", "trashed").Count(&memoryCount)
		db.Model(&struct{ ID uint }{}).Table("entities").Count(&entityCount)
		db.Model(&struct{ ID uint }{}).Table("relations").Count(&relationCount)
		db.Model(&struct{ ID uint }{}).Table("wiki_pages").Count(&wikiCount)

		layerStats := make(map[string]int64)
		rows, _ := db.Raw("SELECT COALESCE(layer, 'knowledge') as layer, COUNT(*) as cnt FROM memories WHERE status != 'trashed' GROUP BY layer").Rows()
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
		db.Table("memories").Where("status != ?", "trashed").Order("created_at desc").Limit(10).Find(&recentMemories)

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

func handleProDecayStats(proxy *services.ProProxy, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		if proxy.IsPro() {
			var memories []map[string]interface{}
			db.Model(&struct{ ID uint }{}).Table("memories").Where("status != ?", "trashed").Find(&memories)
			result, err := proxy.GetDecayStats(memories)
			if err == nil {
				c.JSON(http.StatusOK, result)
				return
			}
		}
		fallback := services.NewProFallbackService(db)
		result, err := fallback.DecayStats(userID)
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProDecayApply(proxy *services.ProProxy, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		if proxy.IsPro() {
			var memories []map[string]interface{}
			db.Model(&struct{ ID uint }{}).Table("memories").Where("status != ?", "trashed").Find(&memories)
			result, err := proxy.ApplyDecay(memories)
			if err == nil {
				c.JSON(http.StatusOK, result)
				return
			}
		}
		fallback := services.NewProFallbackService(db)
		result, err := fallback.DecayApply(userID)
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProReinforce(proxy *services.ProProxy, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		idStr := c.Param("id")
		id, _ := strconv.ParseUint(idStr, 10, 64)
		if proxy.IsPro() {
			var memory map[string]interface{}
			db.Table("memories").Where("id = ?", id).First(&memory)
			result, err := proxy.ReinforceMemory(uint(id), memory)
			if err == nil {
				c.JSON(http.StatusOK, result)
				return
			}
		}
		fallback := services.NewProFallbackService(db)
		result, err := fallback.ReinforceMemory(userID, uint(id))
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProPruneSuggest(proxy *services.ProProxy, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		if proxy.IsPro() {
			var memories []map[string]interface{}
			db.Model(&struct{ ID uint }{}).Table("memories").Where("status != ?", "trashed").Find(&memories)
			result, err := proxy.GetPruneSuggestions(memories)
			if err == nil {
				c.JSON(http.StatusOK, result)
				return
			}
		}
		fallback := services.NewProFallbackService(db)
		result, err := fallback.PruneSuggest(userID)
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProConflictScan(proxy *services.ProProxy, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		if proxy.IsPro() {
			var memories []map[string]interface{}
			db.Model(&struct{ ID uint }{}).Table("memories").Where("status != ?", "trashed").Find(&memories)
			result, err := proxy.ScanConflicts(memories)
			if err == nil {
				c.JSON(http.StatusOK, result)
				return
			}
		}
		fallback := services.NewProFallbackService(db)
		result, err := fallback.ConflictScan(userID)
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProConflictResolve(proxy *services.ProProxy, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		var req struct {
			Strategy string `json:"strategy"`
		}
		c.ShouldBindJSON(&req)
		indexStr := c.Param("index")
		index, _ := strconv.Atoi(indexStr)
		if proxy.IsPro() {
			result, err := proxy.ResolveConflict(index, req.Strategy)
			if err == nil {
				c.JSON(http.StatusOK, result)
				return
			}
		}
		fallback := services.NewProFallbackService(db)
		result, err := fallback.ConflictResolve(userID, index, req.Strategy)
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProTokenRoute(proxy *services.ProProxy, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		message := c.Query("message")
		contextLength := 0
		if cl := c.Query("context_length"); cl != "" {
			if n, err := strconv.Atoi(cl); err == nil {
				contextLength = n
			}
		}
		if proxy.IsPro() {
			result, err := proxy.RouteModel(message, contextLength)
			if err == nil {
				c.JSON(http.StatusOK, result)
				return
			}
		}
		fallback := services.NewProFallbackService(db)
		result, err := fallback.TokenRoute(message, contextLength)
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProTokenStats(proxy *services.ProProxy, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if proxy.IsPro() {
			result, err := proxy.GetTokenStats()
			if err == nil {
				c.JSON(http.StatusOK, result)
				return
			}
		}
		fallback := services.NewProFallbackService(db)
		result, err := fallback.TokenStats()
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProAIExtract(proxy *services.ProProxy, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		var req struct {
			MemoryIDs []uint `json:"memory_ids"`
		}
		c.ShouldBindJSON(&req)
		if proxy.IsPro() {
			result, err := proxy.AIExtract("", req.MemoryIDs)
			if err == nil {
				c.JSON(http.StatusOK, result)
				return
			}
		}
		fallback := services.NewProFallbackService(db)
		result, err := fallback.AIExtract(userID)
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProAutoGraph(proxy *services.ProProxy, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		var req struct {
			Overwrite bool `json:"overwrite"`
		}
		c.ShouldBindJSON(&req)
		if proxy.IsPro() {
			result, err := proxy.AutoGraph(req.Overwrite)
			if err == nil {
				c.JSON(http.StatusOK, result)
				return
			}
		}
		fallback := services.NewProFallbackService(db)
		result, err := fallback.AutoGraph(userID, req.Overwrite)
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProBackupSchedule(proxy *services.ProProxy, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if proxy.IsPro() {
			result, err := proxy.GetBackupSchedule()
			if err == nil {
				c.JSON(http.StatusOK, result)
				return
			}
		}
		fallback := services.NewProFallbackService(db)
		result, err := fallback.BackupSchedule()
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProSetBackupSchedule(proxy *services.ProProxy, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Enabled       bool `json:"enabled"`
			IntervalHours int  `json:"interval_hours"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
			return
		}
		if proxy.IsPro() {
			result, err := proxy.SetBackupSchedule(req.Enabled, req.IntervalHours)
			if err == nil {
				c.JSON(http.StatusOK, result)
				return
			}
		}
		fallback := services.NewProFallbackService(db)
		result, err := fallback.SetBackupSchedule(req.Enabled, req.IntervalHours)
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProCompressPreview(proxy *services.ProProxy, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Level string `json:"level"`
		}
		c.ShouldBindJSON(&req)
		if req.Level == "" {
			req.Level = "light"
		}
		userID := middleware.GetUserID(c)
		if proxy.IsPro() {
			var memories []map[string]interface{}
			db.Model(&struct{ ID uint }{}).Table("memories").Where("status != ?", "trashed").Find(&memories)
			result, err := proxy.CompressPreview(memories, req.Level)
			if err == nil {
				c.JSON(http.StatusOK, result)
				return
			}
		}
		fallback := services.NewProFallbackService(db)
		result, err := fallback.CompressPreview(userID, req.Level)
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProCompressApply(proxy *services.ProProxy, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Level    string                 `json:"level"`
			Options  map[string]interface{} `json:"options"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
			return
		}
		if req.Level == "" {
			req.Level = "light"
		}
		userID := middleware.GetUserID(c)
		if proxy.IsPro() {
			var memories []map[string]interface{}
			db.Model(&struct{ ID uint }{}).Table("memories").Where("status != ?", "trashed").Find(&memories)
			result, err := proxy.CompressApply(memories, req.Level, req.Options)
			if err == nil {
				c.JSON(http.StatusOK, result)
				return
			}
		}
		fallback := services.NewProFallbackService(db)
		result, err := fallback.CompressApply(userID, req.Level)
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProCompressConfig(proxy *services.ProProxy, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if proxy.IsPro() {
			result, err := proxy.GetCompressConfig()
			if err == nil {
				c.JSON(http.StatusOK, result)
				return
			}
		}
		fallback := services.NewProFallbackService(db)
		result, err := fallback.CompressConfig()
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProSetCompressConfig(proxy *services.ProProxy, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
			return
		}
		if proxy.IsPro() {
			result, err := proxy.SetCompressConfig(req)
			if err == nil {
				c.JSON(http.StatusOK, result)
				return
			}
		}
		fallback := services.NewProFallbackService(db)
		result, err := fallback.SetCompressConfig(req)
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProEvolutionInsights(proxy *services.ProProxy, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		if proxy.IsPro() {
			result, err := proxy.GetEvolutionInsights()
			if err == nil {
				c.JSON(http.StatusOK, result)
				return
			}
		}
		fallback := services.NewProFallbackService(db)
		result, err := fallback.EvolutionInsights(userID)
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProEvolutionDiscover(proxy *services.ProProxy, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		if proxy.IsPro() {
			var memories []map[string]interface{}
			db.Model(&struct{ ID uint }{}).Table("memories").Where("status != ?", "trashed").Find(&memories)
			result, err := proxy.DiscoverRelations(memories)
			if err == nil {
				c.JSON(http.StatusOK, result)
				return
			}
		}
		fallback := services.NewProFallbackService(db)
		result, err := fallback.EvolutionDiscover(userID)
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProEvolutionInfer(proxy *services.ProProxy, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		if proxy.IsPro() {
			result, err := proxy.InferChains()
			if err == nil {
				c.JSON(http.StatusOK, result)
				return
			}
		}
		fallback := services.NewProFallbackService(db)
		result, err := fallback.EvolutionInfer(userID)
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProEvolutionImportance(proxy *services.ProProxy, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		if proxy.IsPro() {
			var memories []map[string]interface{}
			db.Model(&struct{ ID uint }{}).Table("memories").Where("status != ?", "trashed").Find(&memories)
			result, err := proxy.GetImportanceAdjustments(memories)
			if err == nil {
				c.JSON(http.StatusOK, result)
				return
			}
		}
		fallback := services.NewProFallbackService(db)
		result, err := fallback.EvolutionImportance(userID)
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProEvolutionPrefetch(proxy *services.ProProxy, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Context string `json:"context"`
		}
		c.ShouldBindJSON(&req)
		userID := middleware.GetUserID(c)
		if proxy.IsPro() {
			result, err := proxy.PrefetchMemories(req.Context)
			if err == nil {
				c.JSON(http.StatusOK, result)
				return
			}
		}
		fallback := services.NewProFallbackService(db)
		result, err := fallback.EvolutionPrefetch(userID, req.Context)
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

func handleScanOpenClawMemories(c *gin.Context) {
	homeDir, _ := os.UserHomeDir()
	openclawDirs := []string{}
	seenDirs := make(map[string]bool)

	addScanDir := func(d string) {
		abs, err := filepath.Abs(d)
		if err != nil {
			abs = d
		}
		if !seenDirs[abs] {
			seenDirs[abs] = true
			openclawDirs = append(openclawDirs, abs)
		}
	}

	if homeDir != "" {
		addScanDir(filepath.Join(homeDir, ".openclaw"))
		addScanDir(filepath.Join(homeDir, ".clawmemory"))
		addScanDir(filepath.Join(homeDir, ".trae-cn"))
		addScanDir(filepath.Join(homeDir, ".trae"))
	}

	cfg := config.Load()
	if cfg.DataDir != "" {
		addScanDir(cfg.DataDir)
	}

	exe, _ := os.Executable()
	if exe != "" {
		addScanDir(filepath.Join(filepath.Dir(exe), "openclaw"))
		addScanDir(filepath.Join(filepath.Dir(exe), "data"))
	}

	wd, _ := os.Getwd()
	if wd != "" {
		addScanDir(filepath.Join(wd, ".openclaw"))
		addScanDir(filepath.Join(wd, ".clawmemory"))
		addScanDir(filepath.Join(wd, "data"))
	}

	for _, dir := range openclawDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}

		agents := make([]map[string]interface{}, 0)
		agentCountMap := make(map[string]int)
		totalMemories := 0
		var foundFiles []string

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
				var memories []map[string]interface{}
				if json.Unmarshal(data, &memories) == nil {
					totalMemories += len(memories)
					for _, m := range memories {
						agent := "default"
						if a, ok := m["agent_name"].(string); ok && a != "" {
							agent = a
						}
						agentCountMap[agent]++
					}
				} else {
					totalMemories++
					relPath, _ := filepath.Rel(dir, path)
					parts := strings.Split(relPath, string(filepath.Separator))
					agent := "default"
					if len(parts) > 1 {
						agent = parts[0]
					}
					agentCountMap[agent]++
				}
			} else if ext == ".md" {
				sections := strings.Count(content, "\n## ") + 1
				if sections == 0 {
					sections = 1
				}
				totalMemories += sections
				relPath, _ := filepath.Rel(dir, path)
				parts := strings.Split(relPath, string(filepath.Separator))
				agent := "markdown"
				if len(parts) > 1 {
					agent = parts[0]
				}
				agentCountMap[agent] += sections
			} else if ext == ".txt" {
				lines := len(strings.Split(content, "\n"))
				totalMemories += lines / 3
				if lines/3 == 0 && lines > 0 {
					totalMemories++
				}
				relPath, _ := filepath.Rel(dir, path)
				parts := strings.Split(relPath, string(filepath.Separator))
				agent := "text"
				if len(parts) > 1 {
					agent = parts[0]
				}
				agentCountMap[agent]++
			}

			return nil
		})

		if totalMemories > 0 {
			for name, count := range agentCountMap {
				agents = append(agents, map[string]interface{}{
					"name":         name,
					"memory_count": count,
				})
			}
			c.JSON(http.StatusOK, gin.H{
				"found":          true,
				"openclaw_dir":   dir,
				"agents":         agents,
				"total_memories": totalMemories,
				"files_found":    foundFiles,
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
	
	homeDir, _ := os.UserHomeDir()
	openclawDirs := []string{}
	if homeDir != "" {
		openclawDirs = append(openclawDirs, filepath.Join(homeDir, ".openclaw"))
		openclawDirs = append(openclawDirs, filepath.Join(homeDir, ".clawmemory"))
	}

	cfg := config.Load()
	if cfg.DataDir != "" {
		openclawDirs = append(openclawDirs, cfg.DataDir)
	}

	exe, _ := os.Executable()
	if exe != "" {
		openclawDirs = append(openclawDirs, filepath.Join(filepath.Dir(exe), "openclaw"))
		openclawDirs = append(openclawDirs, filepath.Join(filepath.Dir(exe), "data"))
	}

	wd, _ := os.Getwd()
	if wd != "" {
		openclawDirs = append(openclawDirs, filepath.Join(wd, ".openclaw"))
		openclawDirs = append(openclawDirs, filepath.Join(wd, ".clawmemory"))
		openclawDirs = append(openclawDirs, filepath.Join(wd, "data"))
	}

	for _, dir := range openclawDirs {
		memFile := filepath.Join(dir, "memories.json")
		if _, err := os.Stat(memFile); err == nil {
			content, err := os.ReadFile(memFile)
			if err != nil {
				continue
			}
			var memories []map[string]interface{}
			if json.Unmarshal(content, &memories) == nil {
				filtered := make([]map[string]interface{}, 0)
				for _, m := range memories {
					if m["agent_name"] == agentName {
						filtered = append(filtered, m)
					}
				}
				c.JSON(http.StatusOK, gin.H{
					"agent":          agentName,
					"memories":       filtered,
					"total_filtered": len(filtered),
				})
				return
			}
		}
	}

	c.JSON(http.StatusNotFound, gin.H{
		"error": "agent not found",
	})
}

func handleImportOpenClawMemories(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			AgentName      string `json:"agent_name"`
			TargetAgentID  *int   `json:"target_agent_id"`
			Layer          string `json:"layer"`
			SkipExisting   bool   `json:"skip_existing"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
			return
		}
		if req.AgentName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"detail": "agent_name is required"})
			return
		}
		if req.Layer == "" {
			req.Layer = "knowledge"
		}

		homeDir, _ := os.UserHomeDir()
		openclawDirs := []string{}
		if homeDir != "" {
			openclawDirs = append(openclawDirs, filepath.Join(homeDir, ".openclaw"))
			openclawDirs = append(openclawDirs, filepath.Join(homeDir, ".clawmemory"))
		}

		cfg := config.Load()
		if cfg.DataDir != "" {
			openclawDirs = append(openclawDirs, cfg.DataDir)
		}

		exe, _ := os.Executable()
		if exe != "" {
			openclawDirs = append(openclawDirs, filepath.Join(filepath.Dir(exe), "openclaw"))
			openclawDirs = append(openclawDirs, filepath.Join(filepath.Dir(exe), "data"))
		}

		wd, _ := os.Getwd()
		if wd != "" {
			openclawDirs = append(openclawDirs, filepath.Join(wd, ".openclaw"))
			openclawDirs = append(openclawDirs, filepath.Join(wd, ".clawmemory"))
			openclawDirs = append(openclawDirs, filepath.Join(wd, "data"))
		}

		var imported, skipped, errors int
		for _, dir := range openclawDirs {
			memFile := filepath.Join(dir, "memories.json")
			if _, err := os.Stat(memFile); err != nil {
				continue
			}
			content, err := os.ReadFile(memFile)
			if err != nil {
				continue
			}
			var memories []map[string]interface{}
			if json.Unmarshal(content, &memories) != nil {
				continue
			}

			for _, m := range memories {
				if m["agent_name"] != req.AgentName {
					skipped++
					continue
				}

				key, _ := m["key"].(string)
				contentStr, _ := m["content"].(string)
				if key == "" || contentStr == "" {
					errors++
					continue
				}

				if req.SkipExisting {
					var count int64
					db.Table("memories").Where("key = ?", key).Count(&count)
					if count > 0 {
						skipped++
						continue
					}
				}

				importance := float64(0.5)
				if imp, ok := m["importance"].(float64); ok {
					importance = imp
				} else if imp, ok := m["importance"].(int64); ok {
					importance = float64(imp)
				}

				tags := ""
				if t, ok := m["tags"].([]interface{}); ok && len(t) > 0 {
					tagStrs := make([]string, 0, len(t))
					for _, tag := range t {
						if s, ok := tag.(string); ok {
							tagStrs = append(tagStrs, s)
						}
					}
					tags = strings.Join(tagStrs, ",")
				} else if t, ok := m["tags"].(string); ok {
					tags = t
				}

				source := "openclaw"
				if s, ok := m["source"].(string); ok && s != "" {
					source = s
				}

				result := db.Exec(`INSERT INTO memories (key, content, layer, importance, tags, source, status, created_at, updated_at)
					VALUES (?, ?, ?, ?, ?, ?, 'active', datetime('now'), datetime('now'))`,
					key, contentStr, req.Layer, importance, tags, source)
				if result.Error != nil {
					errors++
				} else {
					imported++
				}
			}
			break
		}

		c.JSON(http.StatusOK, gin.H{
			"imported": imported,
			"skipped":  skipped,
			"errors":   errors,
		})
	}
}

func handleAutoImportMemories(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
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
		db.Table("memories").Where("status != ?", "trashed").Pluck("key", &existingKeys)
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
					importFromJSON(db, content, seenKeys, &imported, &skipped, &entitiesCreated)
				} else if ext == ".md" {
					importFromMarkdown(db, path, content, seenKeys, &imported, &skipped, &entitiesCreated)
				} else if ext == ".txt" {
					importFromText(db, path, content, seenKeys, &imported, &skipped, &entitiesCreated)
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

func importFromJSON(db *gorm.DB, content string, seenKeys map[string]bool, imported, skipped, entitiesCreated *int) {
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

		result := db.Exec(`INSERT INTO memories (key, content, layer, importance, tags, source, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, 'active', datetime('now'), datetime('now'))`, key, contentStr, layer, importance, tags, source)
		if result.Error != nil {
			*skipped++
			continue
		}
		seenKeys[key] = true
		*imported++

		tryCreateEntity(db, key, contentStr, entitiesCreated)
	}
}

func importFromMarkdown(db *gorm.DB, filePath, content string, seenKeys map[string]bool, imported, skipped, entitiesCreated *int) {
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

		result := db.Exec(`INSERT INTO memories (key, content, layer, importance, tags, source, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, 'active', datetime('now'), datetime('now'))`, key, body, layer, importance, "markdown", source)
		if result.Error != nil {
			*skipped++
			continue
		}
		seenKeys[key] = true
		*imported++

		tryCreateEntity(db, key, body, entitiesCreated)
	}
}

func importFromText(db *gorm.DB, filePath, content string, seenKeys map[string]bool, imported, skipped, entitiesCreated *int) {
	lines := strings.Split(content, "\n")
	var buffer []string
	var currentKey string

	flushBuffer := func() {
		if currentKey != "" && len(buffer) > 0 {
			body := strings.Join(buffer, "\n")
			key := fmt.Sprintf("txt:%s", currentKey)

			if !seenKeys[key] {
				layer := classifyLayer(key, body)
				result := db.Exec(`INSERT INTO memories (key, content, layer, importance, tags, source, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, 'active', datetime('now'), datetime('now'))`, key, body, layer, 0.4, "text", "auto_import_txt")
				if result.Error == nil {
					seenKeys[key] = true
					*imported++
					tryCreateEntity(db, key, body, entitiesCreated)
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
			result := db.Exec(`INSERT INTO memories (key, content, layer, importance, tags, source, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, 'active', datetime('now'), datetime('now'))`, key, content, layer, 0.3, "text", "auto_import_txt")
			if result.Error == nil {
				seenKeys[key] = true
				*imported++
				tryCreateEntity(db, key, content, entitiesCreated)
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

func tryCreateEntity(db *gorm.DB, key, content string, entitiesCreated *int) {
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
	db.Table("entities").Where("name = ?", name).Count(&entityCount)
	if entityCount == 0 {
		db.Exec(`INSERT INTO entities (user_id, name, entity_type, description, confidence, extract_method, created_at, updated_at) VALUES (1, ?, ?, ?, 0.7, 'auto_import', datetime('now'), datetime('now'))`, name, entityType, content)
		*entitiesCreated++
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
					data["user_id"] = userID
					if _, err := svc.Create(data); err == nil {
						imported++
					}
				}
			}
		}

		if entities, ok := req["entities"].([]interface{}); ok {
			for _, e := range entities {
				if data, ok := e.(map[string]interface{}); ok {
					svc := services.NewKnowledgeService(db)
					data["user_id"] = userID
					if _, err := svc.CreateEntity(data); err == nil {
						imported++
					}
				}
			}
		}

		if wikiPages, ok := req["wiki_pages"].([]interface{}); ok {
			for _, w := range wikiPages {
				if data, ok := w.(map[string]interface{}); ok {
					svc := services.NewWikiService(db)
					data["user_id"] = userID
					if _, err := svc.Create(data); err == nil {
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
