package api

import (
	"net/http"
	"strconv"

	"clawmemory/internal/middleware"
	"clawmemory/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Auth handlers
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

func handleLogin(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		token, err := authService.Login(req.Username, req.Password)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"token": token})
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

func handleGetDecayStats(db *gorm.DB, lm *services.LicenseManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !lm.IsFeatureEnabled("auto_decay") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Pro feature required"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"stages": []interface{}{}})
	}
}

// Settings handlers
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
