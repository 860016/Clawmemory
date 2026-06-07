package api

import (
	"clawmemory/internal/middleware"
	"clawmemory/internal/models"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func handleCreateSessionMemory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		session := models.SessionMemory{
			UserID:        userID,
			SessionID:     getString(data, "session_id", ""),
			Title:         getString(data, "title", ""),
			CurrentState:  getString(data, "current_state", ""),
			TaskSpec:      getString(data, "task_spec", ""),
			FilesAndFuncs: getString(data, "files_and_funcs", ""),
			Workflow:      getString(data, "workflow", ""),
			Errors:        getString(data, "errors", ""),
			Docs:          getString(data, "docs", ""),
			Learnings:     getString(data, "learnings", ""),
			KeyResults:    getString(data, "key_results", ""),
			Worklog:       getString(data, "worklog", ""),
			Status:        getString(data, "status", "active"),
		}

		if v, ok := data["token_count"].(float64); ok {
			session.TokenCount = int(v)
		}
		if v, ok := data["compressed_from"].(string); ok {
			session.CompressedFrom = v
		}

		ttlHours := 24
		if v, ok := data["ttl_hours"].(float64); ok && v > 0 {
			ttlHours = int(v)
		}
		expiresAt := time.Now().Add(time.Duration(ttlHours) * time.Hour)
		session.ExpiresAt = &expiresAt

		if err := db.Create(&session).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, session)
	}
}

func handleListSessionMemories(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		sessionID := c.Query("session_id")
		status := c.Query("status")

		var sessions []models.SessionMemory
		query := db.Where("user_id = ?", userID)
		if sessionID != "" {
			query = query.Where("session_id = ?", sessionID)
		}
		if status != "" {
			query = query.Where("status = ?", status)
		}
		showExpired := c.Query("show_expired") == "true"
		if !showExpired {
			query = query.Where("expires_at IS NULL OR expires_at > ?", time.Now())
		}
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
		if limit < 1 || limit > 200 {
			limit = 50
		}
		logDBErr("load sessions", query.Order("updated_at DESC").Limit(limit).Find(&sessions).Error)

		c.JSON(http.StatusOK, gin.H{"items": sessions})
	}
}

func handleGetSessionMemory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}

		var session models.SessionMemory
		if err := db.Where("id = ? AND user_id = ?", id, userID).First(&session).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "session memory not found"})
			return
		}
		c.JSON(http.StatusOK, session)
	}
}

func handleUpdateSessionMemory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}
		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var session models.SessionMemory
		if err := db.Where("id = ? AND user_id = ?", id, userID).First(&session).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "session memory not found"})
			return
		}

		updatables := []string{
			"title", "current_state", "task_spec", "files_and_funcs",
			"workflow", "errors", "docs", "learnings", "key_results",
			"worklog", "status", "session_id", "compressed_from",
		}
		for _, field := range updatables {
			if v, ok := data[field].(string); ok {
				switch field {
				case "title":
					session.Title = v
				case "current_state":
					session.CurrentState = v
				case "task_spec":
					session.TaskSpec = v
				case "files_and_funcs":
					session.FilesAndFuncs = v
				case "workflow":
					session.Workflow = v
				case "errors":
					session.Errors = v
				case "docs":
					session.Docs = v
				case "learnings":
					session.Learnings = v
				case "key_results":
					session.KeyResults = v
				case "worklog":
					session.Worklog = v
				case "status":
					session.Status = v
				case "session_id":
					session.SessionID = v
				case "compressed_from":
					session.CompressedFrom = v
				}
			}
		}
		if v, ok := data["token_count"].(float64); ok {
			session.TokenCount = int(v)
		}

		if err := db.Save(&session).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, session)
	}
}

func handleDeleteSessionMemory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}

		if err := db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.SessionMemory{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	}
}

func toJSONStr(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case []interface{}:
		b, _ := json.Marshal(val)
		return string(b)
	case []string:
		b, _ := json.Marshal(val)
		return string(b)
	default:
		b, _ := json.Marshal(val)
		return string(b)
	}
}
