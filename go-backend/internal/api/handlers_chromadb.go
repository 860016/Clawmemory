package api

import (
	"clawmemory/internal/middleware"
	"clawmemory/internal/services"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func handleChromaDBStatus(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		svc := services.NewChromaDBService(db)
		status := svc.GetStatus()
		c.JSON(http.StatusOK, status)
	}
}

func handleChromaDBInstall(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		svc := services.NewChromaDBService(db)
		result := svc.Install()
		c.JSON(http.StatusOK, result)
	}
}

func handleChromaDBSync(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		svc := services.NewChromaDBService(db)
		count, err := svc.SyncMemories(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"synced": count, "message": fmt.Sprintf("Synced %d memories to ChromaDB", count)})
	}
}
