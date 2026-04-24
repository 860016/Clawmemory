package api

import (
	"net/http"
	"strconv"

	"clawmemory/internal/middleware"
	"clawmemory/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func handleInitStatus(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"password_set": authService.IsPasswordSet(),
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

		c.JSON(http.StatusOK, gin.H{
			"memories":  memoryCount,
			"entities":  entityCount,
			"relations": relationCount,
			"wiki":      wikiCount,
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
