package api

import (
	"clawmemory/internal/ai"
	"clawmemory/internal/middleware"
	"clawmemory/internal/services"
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func handleToolboxConflictScan(aiSvc *ai.AIService, toolbox *services.ToolboxService, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
		defer cancel()

		result, err := aiSvc.ConflictScan(ctx, userID)
		if err != nil {
			result, err = toolbox.ConflictScan(userID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			result["mode"] = "local_fallback"
		}

		c.JSON(http.StatusOK, result)
	}
}

func handleToolboxConflictResolve(toolbox *services.ToolboxService, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		var req struct {
			Strategy string `json:"strategy"`
		}
		c.ShouldBindJSON(&req)
		indexStr := c.Param("index")
		index, err := strconv.Atoi(indexStr)
		if err != nil || index < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conflict index"})
			return
		}
		result, err := toolbox.ConflictResolve(userID, index, req.Strategy)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleToolboxTokenRoute(aiSvc *ai.AIService, toolbox *services.ToolboxService, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		message := c.Query("message")
		contextLength := 0
		if cl := c.Query("context_length"); cl != "" {
			if n, err := strconv.Atoi(cl); err == nil {
				contextLength = n
			}
		}

		if message != "" {
			ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
			defer cancel()

			result, err := aiSvc.SmartRoute(ctx, middleware.GetUserID(c), message)
			if err == nil {
				c.JSON(http.StatusOK, result)
				return
			}
		}

		result, err := toolbox.TokenRoute(message, contextLength)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		result["mode"] = "local_fallback"
		c.JSON(http.StatusOK, result)
	}
}

func handleToolboxTokenStats(toolbox *services.ToolboxService, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		result, err := toolbox.TokenStats(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleToolboxAIExtract(aiSvc *ai.AIService, toolbox *services.ToolboxService, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
		defer cancel()

		result, err := aiSvc.AIExtract(ctx, userID)
		if err != nil {
			result, err = toolbox.ExtractEntities(userID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			result["mode"] = "local_fallback"
		}

		c.JSON(http.StatusOK, result)
	}
}

func handleToolboxAutoGraph(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		var req struct {
			Overwrite bool `json:"overwrite"`
		}
		c.ShouldBindJSON(&req)
		evolutionSvc := services.NewEvolutionService(db)
		result, err := evolutionSvc.AutoGraph(userID, req.Overwrite)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleToolboxCompressPreview(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Level string `json:"level"`
		}
		c.ShouldBindJSON(&req)
		if req.Level == "" {
			req.Level = "light"
		}
		userID := middleware.GetUserID(c)
		svc := services.NewDecayService(db)
		result, err := svc.CompressPreview(userID, req.Level)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleToolboxCompressApply(aiSvc *ai.AIService, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Level   string                 `json:"level"`
			Options map[string]interface{} `json:"options"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if req.Level == "" {
			req.Level = "light"
		}
		userID := middleware.GetUserID(c)

		svc := services.NewDecayService(db)
		preview, previewErr := svc.CompressPreview(userID, req.Level)
		if previewErr == nil {
			if previewItems, ok := preview["preview"].([]map[string]interface{}); ok && len(previewItems) > 0 {
				memoryIDs := make([]uint, 0, len(previewItems))
				for _, item := range previewItems {
					if id, ok := item["memory_id"]; ok {
						switch v := id.(type) {
						case uint:
							memoryIDs = append(memoryIDs, v)
						case float64:
							memoryIDs = append(memoryIDs, uint(v))
						case int:
							memoryIDs = append(memoryIDs, uint(v))
						}
					}
				}

				if len(memoryIDs) > 0 {
					ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
					defer cancel()

					result, err := aiSvc.CompressMemories(ctx, userID, memoryIDs)
					if err == nil {
						result["level"] = req.Level
						result["mode"] = "ai"
						c.JSON(http.StatusOK, result)
						return
					}
				}
			}
		}

		result, err := svc.CompressApply(userID, req.Level)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		result["mode"] = "local_fallback"
		c.JSON(http.StatusOK, result)
	}
}

func handleToolboxCompressConfig(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		svc := services.NewSettingsService(db)
		cfg := map[string]interface{}{
			"auto_compress":      false,
			"threshold":          1000,
			"level":              "light",
			"preserve_important": true,
		}
		if v, err := svc.GetByKey(userID, "compress_config"); err == nil && v != nil {
			if m, ok := v.(map[string]interface{}); ok {
				for k, val := range m {
					cfg[k] = val
				}
			}
		}
		c.JSON(http.StatusOK, map[string]interface{}{"config": cfg})
	}
}

func handleToolboxSetCompressConfig(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		userID := middleware.GetUserID(c)
		svc := services.NewSettingsService(db)
		_ = svc.SetByKey(userID, "compress_config", req)
		c.JSON(http.StatusOK, map[string]interface{}{"updated": true, "config": req})
	}
}

func handleScanSecrets(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Content string `json:"content" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "content is required"})
			return
		}

		if len(req.Content) > 1*1024*1024 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "content too large (max 1MB)"})
			return
		}

		result := services.ScanSecrets(req.Content)
		c.JSON(http.StatusOK, result)
	}
}
