package api

import (
	"clawmemory/internal/ai"
	"clawmemory/internal/middleware"
	"clawmemory/internal/services"
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func handleEvolutionInsights(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		svc := services.NewEvolutionService(db)
		result, err := svc.Insights(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleEvolutionRun(aiSvc *ai.AIService, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		var req struct {
			Action  string `json:"action"`
			Context string `json:"context"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		svc := services.NewEvolutionService(db)

		switch req.Action {
		case "discover":
			ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
			defer cancel()
			result, err := aiSvc.DiscoverRelations(ctx, userID)
			if err != nil {
				result, err = svc.Discover(userID)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				result["mode"] = "local_fallback"
			}
			c.JSON(http.StatusOK, result)
		case "infer":
			result, err := svc.HighConfidenceEntities(userID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, result)
		case "importance":
			result, err := svc.ImportanceBuckets(userID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, result)
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid action, must be one of: discover, infer, importance"})
		}
	}
}

func handleEvolutionGraphReasoning(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		svc := services.NewEvolutionService(db)
		result, err := svc.GraphReasoning(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleEvolutionCentrality(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		svc := services.NewEvolutionService(db)
		result, err := svc.CentralityAnalysis(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleEvolutionCommunityDiscovery(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		svc := services.NewEvolutionService(db)
		result, err := svc.CommunityDiscovery(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleEvolutionCommunitiesToWiki(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		svc := services.NewEvolutionService(db)
		result, err := svc.CommunitiesToWiki(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}
