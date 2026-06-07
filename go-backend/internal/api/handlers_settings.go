package api

import (
	"clawmemory/internal/middleware"
	"clawmemory/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

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
