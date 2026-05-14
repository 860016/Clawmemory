package api

import (
	"net/http"

	"clawmemory/internal/middleware"
	"clawmemory/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func handleGetWritebackTargets(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		svc := services.NewMemoryWritebackService(db)
		c.JSON(http.StatusOK, gin.H{"targets": svc.GetWritebackTargets()})
	}
}

func handlePreviewWriteback(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		var req struct {
			AgentName string `json:"agent_name"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		svc := services.NewMemoryWritebackService(db)
		memories, err := svc.PreviewWriteback(userID, req.AgentName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"memories": memories, "count": len(memories)})
	}
}

func handleExecuteWriteback(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		var req struct {
			AgentName   string `json:"agent_name"`
			ProjectPath string `json:"project_path"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		svc := services.NewMemoryWritebackService(db)
		result, err := svc.Writeback(userID, req.AgentName, req.ProjectPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "result": result})
			return
		}
		c.JSON(http.StatusOK, gin.H{"result": result})
	}
}
