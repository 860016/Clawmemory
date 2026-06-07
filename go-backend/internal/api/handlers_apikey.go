package api

import (
	"clawmemory/internal/middleware"
	"clawmemory/internal/models"
	"clawmemory/internal/services"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func handleListAPIKeys(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		svc := services.NewAPIKeyService(db)
		keys, err := svc.List(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": keys})
	}
}

func handleCreateAPIKey(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		var req struct {
			Name        string `json:"name" binding:"required"`
			Permissions string `json:"permissions"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
			return
		}

		svc := services.NewAPIKeyService(db)
		var apiKey *models.APIKey
		var rawKey string
		var err error
		if req.Permissions != "" {
			apiKey, rawKey, err = svc.CreateWithPermissions(userID, req.Name, req.Permissions, "")
		} else {
			apiKey, rawKey, err = svc.Create(userID, req.Name)
		}
		if err != nil {
			status := http.StatusInternalServerError
			if strings.Contains(err.Error(), "maximum") {
				status = http.StatusConflict
			}
			c.JSON(status, gin.H{"error": err.Error()})
			return
		}

		remaining := services.MaxAPIKeysPerUser - int(svc.Count(userID))

		auditLog(db, c, "api_key.create", fmt.Sprintf("id:%d", apiKey.ID), fmt.Sprintf("name:%s prefix:%s perms:%s", apiKey.Name, apiKey.KeyPrefix, apiKey.Permissions))

		c.JSON(http.StatusCreated, gin.H{
			"id":          apiKey.ID,
			"name":        apiKey.Name,
			"key_prefix":  apiKey.KeyPrefix,
			"permissions": apiKey.Permissions,
			"key":         rawKey,
			"created_at":  apiKey.CreatedAt,
			"remaining":   remaining,
			"message":     "please save the API key securely, it will not be shown again",
		})
	}
}

func handleDeleteAPIKey(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}

		svc := services.NewAPIKeyService(db)
		if err := svc.Delete(userID, uint(id)); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		auditLog(db, c, "api_key.delete", fmt.Sprintf("id:%d", id), "")

		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	}
}
