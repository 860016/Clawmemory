package api

import (
	"net/http"
	"strconv"

	"clawmemory/internal/middleware"
	"clawmemory/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func handleHealthCheck(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		checker := services.NewHealthChecker(db)
		status := checker.Check()

		httpStatus := http.StatusOK
		if status.Status == "unhealthy" {
			httpStatus = http.StatusServiceUnavailable
		}

		c.JSON(httpStatus, status)
	}
}

func handleMetrics(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		metrics := services.GetMetrics()
		if metrics == nil {
			c.JSON(http.StatusOK, gin.H{"message": "metrics not initialized"})
			return
		}
		c.JSON(http.StatusOK, metrics.GetStats())
	}
}

func handleAuditLog(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

		if limit > 200 {
			limit = 200
		}
		if limit < 1 {
			limit = 50
		}

		sec := services.GetSecurity()
		if sec == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "security service not initialized"})
			return
		}

		logs, total := sec.GetAuditLog(userID, limit, offset)
		c.JSON(http.StatusOK, gin.H{
			"items":  logs,
			"total":  total,
			"limit":  limit,
			"offset": offset,
		})
	}
}
