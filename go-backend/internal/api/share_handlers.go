package api

import (
	"net/http"
	"strconv"

	"clawmemory/internal/middleware"
	"clawmemory/internal/models"
	"clawmemory/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func handleShareMemory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		var req struct {
			MemoryID  uint   `json:"memory_id" binding:"required"`
			ToAgent   string `json:"to_agent" binding:"required"`
			ShareType string `json:"share_type"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "memory_id and to_agent are required"})
			return
		}

		if req.ShareType == "" {
			req.ShareType = "manual"
		}

		svc := services.NewMemoryShareService(db)
		share, err := svc.ShareMemory(req.MemoryID, userID, req.ToAgent, req.ShareType)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		auditLog(db, c, "memory.share", strconv.Itoa(int(req.MemoryID)), "agent:"+req.ToAgent)
		c.JSON(http.StatusCreated, share)
	}
}

func handleApproveShare(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}

		svc := services.NewMemoryShareService(db)
		if err := svc.ApproveShare(uint(id), userID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		auditLog(db, c, "share.approve", strconv.Itoa(id), "")
		c.JSON(http.StatusOK, gin.H{"message": "share approved"})
	}
}

func handleRejectShare(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}

		svc := services.NewMemoryShareService(db)
		if err := svc.RejectShare(uint(id), userID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		auditLog(db, c, "share.reject", strconv.Itoa(id), "")
		c.JSON(http.StatusOK, gin.H{"message": "share rejected"})
	}
}

func handleRevokeShare(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}

		svc := services.NewMemoryShareService(db)
		if err := svc.RevokeShare(uint(id), userID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		auditLog(db, c, "share.revoke", strconv.Itoa(id), "")
		c.JSON(http.StatusOK, gin.H{"message": "share revoked"})
	}
}

func handleListPendingShares(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		svc := services.NewMemoryShareService(db)
		shares, err := svc.GetPendingShares(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": shares})
	}
}

func handleListOutboundShares(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		svc := services.NewMemoryShareService(db)
		shares, err := svc.GetOutboundShares(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": shares})
	}
}

func handleGetAgentMemories(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		agent := c.Param("agent")
		if agent == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "agent name is required"})
			return
		}

		svc := services.NewMemoryShareService(db)
		memories, err := svc.GetSharedMemories(agent)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": memories})
	}
}

func handleCreateShareRule(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		var rule models.ShareRule
		if err := c.ShouldBindJSON(&rule); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		rule.UserID = userID

		svc := services.NewMemoryShareService(db)
		if err := svc.CreateRule(&rule); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		auditLog(db, c, "share_rule.create", strconv.Itoa(int(rule.ID)), "name:"+rule.Name)
		c.JSON(http.StatusCreated, rule)
	}
}

func handleListShareRules(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		svc := services.NewMemoryShareService(db)
		rules, err := svc.ListRules(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": rules})
	}
}

func handleUpdateShareRule(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		var rule models.ShareRule
		if err := c.ShouldBindJSON(&rule); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		rule.UserID = userID

		svc := services.NewMemoryShareService(db)
		if err := svc.UpdateRule(&rule); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		auditLog(db, c, "share_rule.update", strconv.Itoa(int(rule.ID)), "")
		c.JSON(http.StatusOK, rule)
	}
}

func handleDeleteShareRule(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}

		svc := services.NewMemoryShareService(db)
		if err := svc.DeleteRule(uint(id), userID); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		auditLog(db, c, "share_rule.delete", strconv.Itoa(id), "")
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	}
}
