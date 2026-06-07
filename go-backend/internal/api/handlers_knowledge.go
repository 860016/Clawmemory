package api

import (
	"clawmemory/internal/ai"
	"clawmemory/internal/middleware"
	"clawmemory/internal/services"
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func handleListEntities(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		svc := services.NewKnowledgeService(db)
		entityType := c.Query("type")
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
		if page < 1 {
			page = 1
		}
		if size < 1 || size > 100 {
			size = 20
		}

		entities, total, err := svc.ListEntities(userID, entityType, page, size)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": entities, "total": total, "page": page, "size": size, "pages": (total + int64(size) - 1) / int64(size)})
	}
}

func handleCreateEntity(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		svc := services.NewKnowledgeService(db)
		entity, err := svc.CreateEntity(userID, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, entity)
	}
}

func handleListRelations(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		svc := services.NewKnowledgeService(db)
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		size, _ := strconv.Atoi(c.DefaultQuery("size", "100"))
		if page < 1 {
			page = 1
		}
		if size < 1 || size > 200 {
			size = 100
		}

		relations, total, err := svc.ListRelations(userID, page, size)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": relations, "total": total, "page": page, "size": size, "pages": (total + int64(size) - 1) / int64(size)})
	}
}

func handleCreateRelation(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		svc := services.NewKnowledgeService(db)
		relation, err := svc.CreateRelation(userID, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, relation)
	}
}

func handleGetGraph(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		svc := services.NewKnowledgeService(db)
		entities, relations, err := svc.GetGraph(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"entities":  entities,
			"relations": relations,
		})
	}
}

func handleListWiki(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		svc := services.NewWikiService(db)
		category := c.Query("category")
		status := c.Query("status")
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		size, _ := strconv.Atoi(c.DefaultQuery("size", "100"))
		if page < 1 {
			page = 1
		}
		if size < 1 || size > 200 {
			size = 100
		}

		pages, total, err := svc.List(userID, category, status, page, size)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": pages, "total": total, "page": page, "size": size, "pages": (total + int64(size) - 1) / int64(size)})
	}
}

func handleGetEntity(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		svc := services.NewKnowledgeService(db)
		entity, err := svc.GetEntity(userID, uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "entity not found"})
			return
		}
		c.JSON(http.StatusOK, entity)
	}
}

func handleUpdateEntity(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		svc := services.NewKnowledgeService(db)
		entity, err := svc.UpdateEntity(userID, uint(id), req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, entity)
	}
}

func handleDeleteEntity(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		svc := services.NewKnowledgeService(db)
		if err := svc.DeleteEntity(userID, uint(id)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	}
}

func handleDeleteRelation(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		svc := services.NewKnowledgeService(db)
		if err := svc.DeleteRelation(userID, uint(id)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	}
}

func handleCreateWiki(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		svc := services.NewWikiService(db)
		page, err := svc.Create(userID, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, page)
	}
}

func handleGetWiki(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}
		svc := services.NewWikiService(db)
		page, err := svc.Get(userID, uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, page)
	}
}

func handleUpdateWiki(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		svc := services.NewWikiService(db)
		page, err := svc.Update(userID, uint(id), req)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, page)
	}
}

func handleDeleteWiki(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}
		svc := services.NewWikiService(db)
		if err := svc.Delete(userID, uint(id)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	}
}

func handleWikiCategories(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		svc := services.NewWikiService(db)
		categories, err := svc.GetCategories(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, categories)
	}
}

func handleWikiStats(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		svc := services.NewWikiService(db)
		stats, err := svc.GetStats(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, stats)
	}
}

func handleWikiSearch(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		q := c.Query("q")
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		svc := services.NewWikiService(db)
		pages, err := svc.Search(userID, q, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, pages)
	}
}

func handleWikiTree(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		svc := services.NewWikiService(db)
		pages, err := svc.GetTree(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, pages)
	}
}

func handleWikiMarkComplete(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}
		svc := services.NewWikiService(db)
		if err := svc.MarkComplete(userID, uint(id)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "marked complete"})
	}
}

func handleWikiMarkInProgress(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}
		svc := services.NewWikiService(db)
		if err := svc.MarkInProgress(userID, uint(id)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "marked in progress"})
	}
}

func handleWikiConfig(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		svc := services.NewSettingsService(db)
		llmAvailable := false
		if v, err := svc.GetByKey(userID, "ai_provider_id"); err == nil {
			if s, ok := v.(string); ok && s != "" {
				llmAvailable = true
			}
		}
		c.JSON(http.StatusOK, gin.H{"llm_available": llmAvailable})
	}
}

func handleWikiAIExtract(aiSvc *ai.AIService, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		var req struct {
			Conversation string `json:"conversation"`
			IsComplete   bool   `json:"is_complete"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "conversation is required"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
		defer cancel()

		result, err := aiSvc.ExtractFacts(ctx, userID, []map[string]string{
			{"role": "user", "content": req.Conversation},
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		wikiSvc := services.NewWikiService(db)
		facts, _ := result["facts"].([]map[string]interface{})
		prefs, _ := result["preferences"].([]map[string]interface{})
		created := 0
		for _, fact := range facts {
			content, _ := fact["content"].(string)
			if content == "" {
				continue
			}
			title, _ := fact["title"].(string)
			if title == "" {
				cat, _ := fact["category"].(string)
				if cat != "" {
					title = cat + ": " + truncateStr(content, 40)
				} else {
					title = truncateStr(content, 60)
				}
			}
			pageData := map[string]interface{}{
				"title":   title,
				"content": content,
				"status":  "in_progress",
			}
			if cat, ok := fact["category"].(string); ok {
				pageData["category"] = cat
			}
			if _, err := wikiSvc.Create(userID, pageData); err == nil {
				created++
			}
		}
		for _, pref := range prefs {
			topic, _ := pref["topic"].(string)
			value, _ := pref["value"].(string)
			if topic == "" || value == "" {
				continue
			}
			title := "Preference: " + topic
			pageData := map[string]interface{}{
				"title":    title,
				"content":  topic + " → " + value,
				"status":   "in_progress",
				"category": "preferences",
			}
			if _, err := wikiSvc.Create(userID, pageData); err == nil {
				created++
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"extracted": len(facts),
			"created":   created,
			"mode":      "ai",
		})
	}
}

func handleWikiRefine(aiSvc *ai.AIService, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || id == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page id"})
			return
		}

		wikiSvc := services.NewWikiService(db)
		page, err := wikiSvc.Get(userID, uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "page not found"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
		defer cancel()

		result, err := aiSvc.GenerateWiki(ctx, userID, page.Title)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if content, ok := result["content"].(string); ok && content != "" {
			updates := map[string]interface{}{"content": content}
			if summary, ok := result["summary"].(string); ok && summary != "" {
				updates["summary"] = summary
			}
			if category, ok := result["category"].(string); ok && category != "" {
				updates["category"] = category
			}
			if tags, ok := result["tags"].([]interface{}); ok && len(tags) > 0 {
				tagStrs := make([]string, 0, len(tags))
				for _, t := range tags {
					if s, ok := t.(string); ok {
						tagStrs = append(tagStrs, s)
					}
				}
				if len(tagStrs) > 0 {
					updates["tags"] = strings.Join(tagStrs, ",")
				}
			}
			updates["ai_generated"] = true
			if _, err := wikiSvc.Update(userID, uint(id), updates); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save refined content"})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"message": "page refined",
				"content": content,
				"mode":    "ai",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "AI refinement produced no content"})
	}
}
