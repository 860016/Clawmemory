package api

import (
	"clawmemory/internal/middleware"
	"clawmemory/internal/models"
	"clawmemory/internal/services"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func handleListMemories(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		svc := services.NewMemoryService(db)
		layer := c.Query("layer")
		status := c.Query("status")
		memoryType := c.Query("memory_type")
		sourceAgent := c.Query("source_agent")
		source := c.Query("source")
		visibility := c.Query("visibility")
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
		if page < 1 {
			page = 1
		}
		if size < 1 || size > 100 {
			size = 20
		}

		memories, total, err := svc.List(userID, layer, page, size, status, memoryType, sourceAgent, visibility, source)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"items": memories,
			"total": total,
			"page":  page,
			"size":  size,
			"pages": (total + int64(size) - 1) / int64(size),
		})
	}
}

func handleCreateMemory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		var dto MemoryCreateRequest
		if err := c.ShouldBindJSON(&dto); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		secretResult := services.ScanSecrets(dto.Key + " " + dto.Value)

		req := dto.ToMap()
		svc := services.NewMemoryService(db)
		memory, err := svc.Create(userID, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		validationSvc := services.NewValidationService(db)
		validation := validationSvc.ValidateDTO(memory.Key, memory.Value, memory.Layer, memory.Importance)
		if validation.Status != "valid" {
			logDBErr("update validation status on create", db.Model(&models.Memory{}).Where("id = ?", memory.ID).
				Update("validation_status", validation.Status).Error)
		}

		response := gin.H{"memory": memory}
		if secretResult.Found {
			response["secret_warning"] = secretResult
		}
		if validation.Status != "valid" {
			response["validation"] = validation
		}
		c.JSON(http.StatusCreated, response)
	}
}

func handleGetMemory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}
		svc := services.NewMemoryService(db)
		memory, err := svc.Get(userID, uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, memory)
	}
}

func handleUpdateMemory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}
		var dto MemoryUpdateRequest
		if err := c.ShouldBindJSON(&dto); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		req := dto.ToMap()
		if len(req) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
			return
		}

		var secretResult *services.SecretScanResult
		if dto.Value != nil {
			keyStr := ""
			if dto.Key != nil {
				keyStr = *dto.Key
			}
			secretResult = services.ScanSecrets(keyStr + " " + *dto.Value)
		}

		svc := services.NewMemoryService(db)
		memory, err := svc.Update(userID, uint(id), req)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}

		validationSvc := services.NewValidationService(db)
		validation := validationSvc.ValidateDTO(memory.Key, memory.Value, memory.Layer, memory.Importance)
		logDBErr("update validation status on update", db.Model(&models.Memory{}).Where("id = ?", memory.ID).
			Update("validation_status", validation.Status).Error)

		response := gin.H{"memory": memory}
		if secretResult != nil && secretResult.Found {
			response["secret_warning"] = secretResult
		}
		if validation.Status != "valid" {
			response["validation"] = validation
		}
		c.JSON(http.StatusOK, response)
	}
}

func handleDeleteMemory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}
		svc := services.NewMemoryService(db)
		if err := svc.Delete(userID, uint(id)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	}
}

func handleRestoreMemory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}
		svc := services.NewMemoryService(db)
		if err := svc.Restore(userID, uint(id)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "restored"})
	}
}

func handleSearchMemories(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		q := c.Query("q")
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		if limit <= 0 || limit > 200 {
			limit = 20
		}
		mode := c.DefaultQuery("mode", "keyword")
		userID := middleware.GetUserID(c)

		switch mode {
		case "semantic":
			chromaSvc := services.NewChromaDBService(db)
			if chromaSvc.IsAvailable() {
				results, err := chromaSvc.Search(userID, q, limit)
				if err == nil && len(results) > 0 {
					enriched := enrichChromaResults(db, userID, results, limit)
					c.JSON(http.StatusOK, gin.H{"items": enriched, "engine": "chromadb"})
					return
				}
			}
			svc := services.NewSearchService(db)
			memories, err := svc.SemanticSearch(userID, q, limit)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"items": memories, "engine": "tfidf"})
		case "graph-rag":
			if q == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter q is required"})
				return
			}
			svc := services.NewSearchService(db)
			results, err := svc.GraphRAGSearch(userID, q, limit)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"items":  results,
				"engine": "graph_rag",
				"mode":   "keyword+semantic+graph",
			})
		case "smart":
			if q == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter q is required for smart search"})
				return
			}
			searchSvc := services.NewSearchService(db)
			results, err := searchSvc.GraphRAGSearch(userID, q, limit)
			if err != nil || len(results) == 0 {
				chromaSvc := services.NewChromaDBService(db)
				if chromaSvc.IsAvailable() {
					chromaResults, chromaErr := chromaSvc.Search(userID, q, limit)
					if chromaErr == nil && len(chromaResults) > 0 {
						enriched := enrichChromaResults(db, userID, chromaResults, limit)
						c.JSON(http.StatusOK, gin.H{"items": enriched, "engine": "chromadb"})
						return
					}
				}
				memSvc := services.NewMemoryService(db)
				memories, kwErr := memSvc.SearchKeyword(userID, q, limit)
				if kwErr != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": kwErr.Error()})
					return
				}
				c.JSON(http.StatusOK, gin.H{"items": memories, "engine": "keyword"})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"items":  results,
				"engine": "smart",
				"mode":   "keyword+semantic+graph",
			})
		default:
			svc := services.NewMemoryService(db)
			memories, err := svc.SearchKeyword(userID, q, limit)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"items": memories})
		}
	}
}

func handleMemoryHistory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		memoryIDStr := c.Param("id")
		memoryID, err := strconv.ParseUint(memoryIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid memory id"})
			return
		}

		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

		var history []models.MemoryHistory
		if err := db.Where("user_id = ? AND memory_id = ?", userID, memoryID).
			Order("created_at DESC").Limit(limit).Find(&history).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"items": history,
			"total": len(history),
		})
	}
}

func handleMemoryEvolution(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		memoryIDStr := c.Param("id")
		memoryID, err := strconv.ParseUint(memoryIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid memory id"})
			return
		}

		var memory models.Memory
		if err := db.Where("id = ? AND user_id = ?", memoryID, userID).First(&memory).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "memory not found"})
			return
		}

		var history []models.MemoryHistory
		logDBErr("load memory history", db.Where("user_id = ? AND memory_id = ?", userID, memoryID).
			Order("created_at ASC").Find(&history).Error)

		var relatedEntities []models.Entity
		escapedKey := services.EscapeLikeQuery(memory.Key)
		logDBErr("load related entities", db.Where("user_id = ? AND (name LIKE ? OR description LIKE ?)",
			userID, "%"+escapedKey+"%", "%"+escapedKey+"%").Limit(20).Find(&relatedEntities).Error)

		var relatedRelations []models.Relation
		if len(relatedEntities) > 0 {
			entityIDs := make([]uint, len(relatedEntities))
			for i, e := range relatedEntities {
				entityIDs[i] = e.ID
			}
			logDBErr("load related relations", db.Where("user_id = ? AND (source_id IN ? OR target_id IN ?)",
				userID, entityIDs, entityIDs).Limit(50).Find(&relatedRelations).Error)
		}

		type evolutionStep struct {
			Timestamp  string `json:"timestamp"`
			ChangeType string `json:"change_type"`
			OldValue   string `json:"old_value"`
			NewValue   string `json:"new_value"`
			Reason     string `json:"reason"`
			Source     string `json:"source"`
		}

		steps := make([]evolutionStep, 0, len(history)+1)
		steps = append(steps, evolutionStep{
			Timestamp:  memory.CreatedAt.Format("2006-01-02 15:04:05"),
			ChangeType: "created",
			OldValue:   "",
			NewValue:   memory.Value,
			Reason:     "initial creation",
			Source:     memory.Source,
		})

		for _, h := range history {
			steps = append(steps, evolutionStep{
				Timestamp:  h.CreatedAt.Format("2006-01-02 15:04:05"),
				ChangeType: h.ChangeType,
				OldValue:   h.OldValue,
				NewValue:   h.NewValue,
				Reason:     h.Reason,
				Source:     h.SourceAgent,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"memory":            memory,
			"evolution_steps":   steps,
			"total_changes":     len(history),
			"related_entities":  relatedEntities,
			"related_relations": relatedRelations,
		})
	}
}

func enrichChromaResults(db *gorm.DB, userID uint, chromaResults []map[string]interface{}, limit int) []map[string]interface{} {
	memIDs := make([]uint, 0, len(chromaResults))
	for _, r := range chromaResults {
		if memIDStr, ok := r["memory_id"].(string); ok {
			var id uint
			fmt.Sscanf(memIDStr, "%d", &id)
			if id > 0 {
				memIDs = append(memIDs, id)
			}
		}
	}

	if len(memIDs) == 0 {
		return chromaResults
	}

	var memories []models.Memory
	logDBErr("load memories by ids", db.Where("user_id = ? AND id IN ?", userID, memIDs).Find(&memories).Error)

	memMap := make(map[uint]models.Memory)
	for _, m := range memories {
		memMap[m.ID] = m
	}

	enriched := make([]map[string]interface{}, 0, len(chromaResults))
	for _, r := range chromaResults {
		if memIDStr, ok := r["memory_id"].(string); ok {
			var id uint
			fmt.Sscanf(memIDStr, "%d", &id)
			if m, found := memMap[id]; found {
				var tags []string
				if m.Tags != "" {
					json.Unmarshal([]byte(m.Tags), &tags)
				}
				if tags == nil {
					tags = []string{}
				}
				item := map[string]interface{}{
					"id":              m.ID,
					"key":             m.Key,
					"value":           m.Value,
					"layer":           m.Layer,
					"importance":      m.Importance,
					"source":          m.Source,
					"status":          m.Status,
					"summary":         m.Summary,
					"tags":            tags,
					"memory_type":     m.MemoryType,
					"decay_stage":     m.DecayStage,
					"reinforce_count": m.ReinforceCount,
					"access_count":    m.AccessCount,
					"is_encrypted":    m.IsEncrypted,
					"created_at":      m.CreatedAt.Format("2006-01-02 15:04:05"),
					"updated_at":      m.UpdatedAt.Format("2006-01-02 15:04:05"),
				}
				if m.VerifiedAt != nil {
					item["verified_at"] = m.VerifiedAt.Format("2006-01-02 15:04:05")
				}
				if m.TrashedAt != nil {
					item["trashed_at"] = m.TrashedAt.Format("2006-01-02 15:04:05")
				}
				if m.LastAccessedAt != nil {
					item["last_accessed_at"] = m.LastAccessedAt.Format("2006-01-02 15:04:05")
				}
				if score, ok := r["score"].(float64); ok {
					item["score"] = math.Round(score*1000) / 1000
				}
				enriched = append(enriched, item)
			}
		}
	}

	return enriched
}

func handleSmartLoad(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		query := c.Query("q")
		tokenBudget, _ := strconv.Atoi(c.DefaultQuery("token_budget", "2000"))
		loadLevel := c.DefaultQuery("load_level", "auto")

		svc := services.NewSmartLoadService(db)
		result, err := svc.SmartLoad(userID, query, tokenBudget, loadLevel, "api", "")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func handleReinforceMemory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}

		svc := services.NewSmartLoadService(db)
		if err := svc.ReinforceMemory(userID, uint(id)); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "memory not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "memory reinforced", "memory_id": id})
	}
}

func handleGenerateSummaries(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		svc := services.NewSmartLoadService(db)
		count, err := svc.BatchGenerateSummaries(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"generated": count, "message": fmt.Sprintf("Generated summaries for %d memories", count)})
	}
}

func handleExtractMemories(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Content string `json:"content" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "content is required"})
			return
		}

		userID := middleware.GetUserID(c)
		svc := services.NewExtractionService(db, userID)
		result := svc.ExtractFromConversation(req.Content)
		c.JSON(http.StatusOK, result)
	}
}

func handleExtractAndSave(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Content  string `json:"content" binding:"required"`
			AutoSave bool   `json:"auto_save"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "content is required"})
			return
		}

		userID := middleware.GetUserID(c)
		svc := services.NewExtractionService(db, userID)
		result := svc.ExtractFromConversation(req.Content)

		if req.AutoSave && result.Count > 0 {
			saved := 0
			for _, em := range result.Memories {
				tags := make([]interface{}, len(em.Tags))
				for i, t := range em.Tags {
					tags[i] = t
				}
				data := map[string]interface{}{
					"key":         em.Key,
					"value":       em.Value,
					"layer":       em.Layer,
					"memory_type": em.MemoryType,
					"importance":  em.Importance,
					"tags":        tags,
					"source":      em.Source,
				}
				memSvc := services.NewMemoryService(db)
				if _, err := memSvc.Create(userID, data); err == nil {
					saved++
				}
			}
			c.JSON(http.StatusOK, gin.H{
				"extracted": result.Count,
				"saved":     saved,
				"memories":  result.Memories,
				"warnings":  result.Warnings,
			})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func handleVerifyMemory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}

		var memory models.Memory
		if err := db.Where("id = ? AND user_id = ?", id, userID).First(&memory).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "memory not found"})
			return
		}

		now := time.Now()
		memory.VerifiedAt = &now
		memory.VerifyCount++
		if err := db.Save(&memory).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":           memory.ID,
			"verified_at":  memory.VerifiedAt.Format("2006-01-02 15:04:05"),
			"verify_count": memory.VerifyCount,
			"message":      "memory verified successfully",
		})
	}
}

func handleBatchValidate(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		svc := services.NewValidationService(db)
		result, err := svc.BatchValidate(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}
