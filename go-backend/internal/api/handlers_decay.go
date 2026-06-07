package api

import (
	"clawmemory/internal/ai"
	"clawmemory/internal/middleware"
	"clawmemory/internal/models"
	"clawmemory/internal/services"
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

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
