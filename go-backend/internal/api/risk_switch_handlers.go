package api

import (
	"net/http"

	"clawmemory/internal/middleware"
	"clawmemory/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func handleGetRiskSwitches(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		svc := services.GetRiskSwitchService()
		if svc == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "risk switch service not initialized"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": svc.GetAllWithMeta(userID)})
	}
}

func handleSetRiskSwitches(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		var outer struct {
			Switches map[string]bool `json:"switches"`
		}
		if err := c.ShouldBindJSON(&outer); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		svc := services.GetRiskSwitchService()
		if svc == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "risk switch service not initialized"})
			return
		}

		switches := make(map[services.RiskSwitch]bool)
		for k, v := range outer.Switches {
			switches[services.RiskSwitch(k)] = v
		}

		if err := svc.BatchSet(switches, userID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "risk switches updated", "items": svc.GetAllWithMeta(userID)})
	}
}
