package api

import (
	"context"
	"net/http"
	"time"

	"clawmemory/internal/ai"
	"clawmemory/internal/middleware"
	"clawmemory/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func handleAIConfig(aiRouter *ai.AIRouter, proProxy *services.ProProxy) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		isPro := proProxy.IsPro()
		config := aiRouter.GetCurrentUserConfig(userID, isPro)
		c.JSON(http.StatusOK, config)
	}
}

func handleAIConfigUpdate(aiRouter *ai.AIRouter, proProxy *services.ProProxy) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !proProxy.IsPro() {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Pro license required",
				"message": "AI provider customization requires a Pro license. Free users can only use NVIDIA NIM free models.",
			})
			return
		}

		guard := services.GetProGuard()
		if guard != nil && !guard.IsProFeatureEnabled("ai-config-update") {
			c.JSON(http.StatusForbidden, gin.H{"error": "License verification failed"})
			return
		}

		userID := middleware.GetUserID(c)
		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		if err := aiRouter.UpdateProConfig(userID, data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "AI configuration updated",
			"config":  aiRouter.GetCurrentUserConfig(userID, true),
		})
	}
}

func handleAITestConnection(aiRouter *ai.AIRouter, proProxy *services.ProProxy) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		isPro := proProxy.IsPro()

		result, err := aiRouter.TestConnection(userID, isPro)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func handleAIUsage(aiRouter *ai.AIRouter) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		stats := aiRouter.GetUsageStats(userID)
		c.JSON(http.StatusOK, stats)
	}
}

func handleAIProviders(proProxy *services.ProProxy) gin.HandlerFunc {
	return func(c *gin.Context) {
		isPro := proProxy.IsPro()

		providers := ai.AllProviders
		if !isPro {
			providers = func() []ai.ProviderInfo {
				var free []ai.ProviderInfo
				for _, p := range ai.AllProviders {
					if p.Free {
						free = append(free, p)
					}
				}
				return free
			}()
		}

		c.JSON(http.StatusOK, gin.H{
			"providers": providers,
			"is_pro":    isPro,
		})
	}
}

func handleAIExtract(aiSvc *ai.AIService, proProxy *services.ProProxy) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		isPro := proProxy.IsPro()

		ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
		defer cancel()

		result, err := aiSvc.AIExtract(ctx, userID, isPro)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func checkProWithGuard(proxy *services.ProProxy, feature string, c *gin.Context) bool {
	if !proxy.IsPro() {
		c.JSON(http.StatusForbidden, gin.H{"error": "Pro license required"})
		return false
	}

	guard := services.GetProGuard()
	if guard != nil && !guard.IsProFeatureEnabled(feature) {
		c.JSON(http.StatusForbidden, gin.H{"error": "License verification failed"})
		return false
	}

	return true
}

func handleAIConflictScan(aiSvc *ai.AIService, proProxy *services.ProProxy) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !checkProWithGuard(proProxy, "ai-conflict-scan", c) {
			return
		}

		userID := middleware.GetUserID(c)

		ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
		defer cancel()

		result, err := aiSvc.ConflictScan(ctx, userID, true)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func handleAIDecayEvaluate(aiSvc *ai.AIService, proProxy *services.ProProxy) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !checkProWithGuard(proProxy, "ai-decay-evaluate", c) {
			return
		}

		userID := middleware.GetUserID(c)

		ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
		defer cancel()

		result, err := aiSvc.DecayEvaluate(ctx, userID, true)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func handleAIDailyReport(aiSvc *ai.AIService, proProxy *services.ProProxy, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		isPro := proProxy.IsPro()

		date := c.Query("date")
		if date == "" {
			date = time.Now().Format("2006-01-02")
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
		defer cancel()

		result, err := aiSvc.GenerateDailyReport(ctx, userID, isPro, date)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func handleAIWikiGenerate(aiSvc *ai.AIService, proProxy *services.ProProxy) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !checkProWithGuard(proProxy, "ai-wiki-generate", c) {
			return
		}

		userID := middleware.GetUserID(c)

		var data struct {
			Topic string `json:"topic"`
		}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "topic is required"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
		defer cancel()

		result, err := aiSvc.GenerateWiki(ctx, userID, true, data.Topic)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func handleAICompress(aiSvc *ai.AIService, proProxy *services.ProProxy) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !checkProWithGuard(proProxy, "ai-compress", c) {
			return
		}

		userID := middleware.GetUserID(c)

		var data struct {
			MemoryIDs []uint `json:"memory_ids"`
		}
		if err := c.ShouldBindJSON(&data); err != nil || len(data.MemoryIDs) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "memory_ids is required"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
		defer cancel()

		result, err := aiSvc.CompressMemories(ctx, userID, true, data.MemoryIDs)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func handleAIDiscoverRelations(aiSvc *ai.AIService, proProxy *services.ProProxy) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !checkProWithGuard(proProxy, "ai-discover-relations", c) {
			return
		}

		userID := middleware.GetUserID(c)

		ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
		defer cancel()

		result, err := aiSvc.DiscoverRelations(ctx, userID, true)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func handleAISmartRoute(aiSvc *ai.AIService, proProxy *services.ProProxy) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !checkProWithGuard(proProxy, "ai-smart-route", c) {
			return
		}

		userID := middleware.GetUserID(c)

		var data struct {
			Text string `json:"text"`
		}
		if err := c.ShouldBindJSON(&data); err != nil || data.Text == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "text is required"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
		defer cancel()

		result, err := aiSvc.SmartRoute(ctx, userID, true, data.Text)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}
