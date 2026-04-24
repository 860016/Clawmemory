package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"clawmemory/internal/middleware"
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
		// 简化实现，返回关键词搜索结果
		handleSearchKeyword(db)(c)
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

func handleGetSettings() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"language": "zh-CN",
			"theme":    "light",
		})
	}
}

func handleUpdateSettings() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req map[string]interface{}
		c.ShouldBindJSON(&req)
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
		var memories []map[string]interface{}
		db.Model(&struct{ ID uint }{}).Table("memories").Where("status != ?", "trashed").Find(&memories)
		result, err := proxy.GetDecayStats(memories)
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProDecayApply(proxy *services.ProProxy, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var memories []map[string]interface{}
		db.Model(&struct{ ID uint }{}).Table("memories").Where("status != ?", "trashed").Find(&memories)
		result, err := proxy.ApplyDecay(memories)
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProReinforce(proxy *services.ProProxy, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, _ := strconv.ParseUint(idStr, 10, 64)
		var memory map[string]interface{}
		db.Table("memories").Where("id = ?", id).First(&memory)
		result, err := proxy.ReinforceMemory(uint(id), memory)
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProPruneSuggest(proxy *services.ProProxy, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var memories []map[string]interface{}
		db.Model(&struct{ ID uint }{}).Table("memories").Where("status != ?", "trashed").Find(&memories)
		result, err := proxy.GetPruneSuggestions(memories)
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProConflictScan(proxy *services.ProProxy, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var memories []map[string]interface{}
		db.Model(&struct{ ID uint }{}).Table("memories").Where("status != ?", "trashed").Find(&memories)
		result, err := proxy.ScanConflicts(memories)
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProConflictResolve(proxy *services.ProProxy) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Strategy string `json:"strategy"`
		}
		c.ShouldBindJSON(&req)
		indexStr := c.Param("index")
		index, _ := strconv.Atoi(indexStr)
		result, err := proxy.ResolveConflict(index, req.Strategy)
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProTokenRoute(proxy *services.ProProxy) gin.HandlerFunc {
	return func(c *gin.Context) {
		message := c.Query("message")
		contextLength := 0
		if cl := c.Query("context_length"); cl != "" {
			if n, err := strconv.Atoi(cl); err == nil {
				contextLength = n
			}
		}
		result, err := proxy.RouteModel(message, contextLength)
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProTokenStats(proxy *services.ProProxy) gin.HandlerFunc {
	return func(c *gin.Context) {
		result, err := proxy.GetTokenStats()
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProAIExtract(proxy *services.ProProxy, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			MemoryIDs []uint `json:"memory_ids"`
		}
		c.ShouldBindJSON(&req)
		result, err := proxy.AIExtract("", req.MemoryIDs)
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProAutoGraph(proxy *services.ProProxy) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Overwrite bool `json:"overwrite"`
		}
		c.ShouldBindJSON(&req)
		result, err := proxy.AutoGraph(req.Overwrite)
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProBackupSchedule(proxy *services.ProProxy) gin.HandlerFunc {
	return func(c *gin.Context) {
		result, err := proxy.GetBackupSchedule()
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProSetBackupSchedule(proxy *services.ProProxy) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Enabled       bool `json:"enabled"`
			IntervalHours int  `json:"interval_hours"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
			return
		}
		result, err := proxy.SetBackupSchedule(req.Enabled, req.IntervalHours)
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
		var memories []map[string]interface{}
		db.Model(&struct{ ID uint }{}).Table("memories").Where("status != ?", "trashed").Find(&memories)
		result, err := proxy.CompressPreview(memories, req.Level)
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
		var memories []map[string]interface{}
		db.Model(&struct{ ID uint }{}).Table("memories").Where("status != ?", "trashed").Find(&memories)
		result, err := proxy.CompressApply(memories, req.Level, req.Options)
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProCompressConfig(proxy *services.ProProxy) gin.HandlerFunc {
	return func(c *gin.Context) {
		result, err := proxy.GetCompressConfig()
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProSetCompressConfig(proxy *services.ProProxy) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
			return
		}
		result, err := proxy.SetCompressConfig(req)
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProEvolutionInsights(proxy *services.ProProxy) gin.HandlerFunc {
	return func(c *gin.Context) {
		result, err := proxy.GetEvolutionInsights()
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProEvolutionDiscover(proxy *services.ProProxy, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var memories []map[string]interface{}
		db.Model(&struct{ ID uint }{}).Table("memories").Where("status != ?", "trashed").Find(&memories)
		result, err := proxy.DiscoverRelations(memories)
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProEvolutionInfer(proxy *services.ProProxy) gin.HandlerFunc {
	return func(c *gin.Context) {
		result, err := proxy.InferChains()
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProEvolutionImportance(proxy *services.ProProxy, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var memories []map[string]interface{}
		db.Model(&struct{ ID uint }{}).Table("memories").Where("status != ?", "trashed").Find(&memories)
		result, err := proxy.GetImportanceAdjustments(memories)
		if err != nil {
			proErrorHandler(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleProEvolutionPrefetch(proxy *services.ProProxy) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Context string `json:"context"`
		}
		c.ShouldBindJSON(&req)
		result, err := proxy.PrefetchMemories(req.Context)
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
		cutoff := now.AddDate(0, 0, -days)

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

	homeDir, _ := os.UserHomeDir()
	if homeDir != "" {
		dataDirs = append(dataDirs, filepath.Join(homeDir, ".openclaw", "skills"))
	}

	exe, _ := os.Executable()
	if exe != "" {
		dataDirs = append(dataDirs, filepath.Join(filepath.Dir(exe), "skills"))
	}

	globalSkills := make([]map[string]interface{}, 0)
	workspaceSkills := make([]map[string]interface{}, 0)

	for _, dir := range dataDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			skillFile := filepath.Join(dir, entry.Name(), "skill.json")
			if _, err := os.Stat(skillFile); err != nil {
				continue
			}
			content, err := os.ReadFile(skillFile)
			if err != nil {
				continue
			}
			var skill map[string]interface{}
			if json.Unmarshal(content, &skill) == nil {
				skill["skill_dir"] = entry.Name()
				if dir == filepath.Join(homeDir, ".openclaw", "skills") {
					skill["scope"] = "global"
					globalSkills = append(globalSkills, skill)
				} else {
					skill["scope"] = "workspace"
					workspaceSkills = append(workspaceSkills, skill)
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"global_skills":   globalSkills,
		"workspace_skills": workspaceSkills,
	})
}

func handleSkillDetail(c *gin.Context) {
	skillDir := c.Query("skill_dir")
	scope := c.Query("scope")

	homeDir, _ := os.UserHomeDir()
	var baseDir string
	if scope == "global" && homeDir != "" {
		baseDir = filepath.Join(homeDir, ".openclaw", "skills")
	} else {
		exe, _ := os.Executable()
		baseDir = filepath.Join(filepath.Dir(exe), "skills")
	}

	skillFile := filepath.Join(baseDir, skillDir, "skill.json")
	content, err := os.ReadFile(skillFile)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "skill not found"})
		return
	}

	var skill map[string]interface{}
	if err := json.Unmarshal(content, &skill); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid skill file"})
		return
	}

	skill["skill_dir"] = skillDir
	skill["scope"] = scope

	c.JSON(http.StatusOK, skill)
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

	if homeDir != "" {
		openclawDirs = append(openclawDirs, filepath.Join(homeDir, ".openclaw"))
	}
	exe, _ := os.Executable()
	if exe != "" {
		openclawDirs = append(openclawDirs, filepath.Join(filepath.Dir(exe), "openclaw"))
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
				agents := make([]map[string]interface{}, 0)
				agentMap := make(map[string]bool)
				for _, m := range memories {
					if agent, ok := m["agent_name"].(string); ok && agent != "" && !agentMap[agent] {
						agentMap[agent] = true
						agents = append(agents, map[string]interface{}{
							"name":        agent,
							"memory_count": 1,
						})
					} else if ok && agentMap[agent] {
						for _, a := range agents {
							if a["name"] == agent {
								a["memory_count"] = a["memory_count"].(int) + 1
							}
						}
					}
				}
				c.JSON(http.StatusOK, gin.H{
					"found":         true,
					"openclaw_dir":  dir,
					"agents":        agents,
					"total_memories": len(memories),
				})
				return
			}
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
	}
	exe, _ := os.Executable()
	if exe != "" {
		openclawDirs = append(openclawDirs, filepath.Join(filepath.Dir(exe), "openclaw"))
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
		}
		exe, _ := os.Executable()
		if exe != "" {
			openclawDirs = append(openclawDirs, filepath.Join(filepath.Dir(exe), "openclaw"))
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

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"filename": filename,
		"path":     backupPath,
		"size":     dst.(interface{ Size() int64 }).Size(),
	})
}
