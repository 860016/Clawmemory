package api

import (
	"clawmemory/internal/config"
	"clawmemory/internal/middleware"
	"clawmemory/internal/models"
	"clawmemory/internal/services"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

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
