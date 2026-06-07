package api

import (
	"log"
	"net/http"
	"strconv"

	"clawmemory/internal/config"
	"clawmemory/internal/middleware"
	"clawmemory/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func logDBErr(context string, err error) {
	if err != nil {
		log.Printf("[DB] %s: %v", context, err)
	}
}

func truncateStr(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}

func parseIDParam(c *gin.Context, name string) (int, bool) {
	id, err := strconv.Atoi(c.Param(name))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid " + name})
		return 0, false
	}
	return id, true
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

func auditLog(db *gorm.DB, c *gin.Context, action, target, detail string) {
	userID := middleware.GetUserID(c)
	log := models.AuditLog{
		UserID: userID,
		Action: action,
		Target: target,
		Detail: detail,
		IP:     c.ClientIP(),
	}
	logDBErr("create audit log", db.Create(&log).Error)
}

func getString(m map[string]interface{}, key, def string) string {
	if v, ok := m[key].(string); ok && v != "" {
		return v
	}
	return def
}

func getFloat(m map[string]interface{}, key string, def float64) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	return def
}

func getBool(m map[string]interface{}, key string, def bool) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return def
}
