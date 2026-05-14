package api

import (
	"net/http"
	"strconv"
	"time"

	"clawmemory/internal/middleware"
	"clawmemory/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func handleRegisterWithInvitation(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Username       string `json:"username" binding:"required"`
			Password       string `json:"password" binding:"required,min=6"`
			InvitationCode string `json:"invitation_code"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		user, err := authService.RegisterWithInvitation(req.Username, req.Password, req.InvitationCode)
		if err != nil {
			status := http.StatusBadRequest
			if err.Error() == "username already exists" {
				status = http.StatusConflict
			}
			c.JSON(status, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"id":       user.ID,
			"username": user.Username,
			"role":     user.Role,
			"message":  "registration successful",
		})
	}
}

func handleCreateInvitation(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		invSvc := services.NewInvitationService(db)
		if !invSvc.IsAdmin(userID) {
			c.JSON(http.StatusForbidden, gin.H{"error": "only admins can create invitation codes"})
			return
		}

		var req struct {
			MaxUses   int  `json:"max_uses"`
			ExpiresIn int  `json:"expires_in_hours"`
		}
		c.ShouldBindJSON(&req)

		if req.MaxUses <= 0 {
			req.MaxUses = 1
		}

		var expiresAt *time.Time
		if req.ExpiresIn > 0 {
			t := time.Now().Add(time.Duration(req.ExpiresIn) * time.Hour)
			expiresAt = &t
		}

		inv, err := invSvc.GenerateCode(userID, req.MaxUses, expiresAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		auditLog(db, c, "invitation.create", strconv.Itoa(int(inv.ID)), "code:"+inv.Code)
		c.JSON(http.StatusCreated, inv)
	}
}

func handleListInvitations(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		invSvc := services.NewInvitationService(db)
		if !invSvc.IsAdmin(userID) {
			c.JSON(http.StatusForbidden, gin.H{"error": "only admins can list invitation codes"})
			return
		}

		invitations, err := invSvc.ListCodes(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"items": invitations})
	}
}

func handleDeleteInvitation(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}

		invSvc := services.NewInvitationService(db)
		if !invSvc.IsAdmin(userID) {
			c.JSON(http.StatusForbidden, gin.H{"error": "only admins can delete invitation codes"})
			return
		}

		if err := invSvc.DeleteCode(uint(id), userID); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		auditLog(db, c, "invitation.delete", strconv.Itoa(id), "")
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	}
}

func handleRevokeAllTokens(db *gorm.DB, authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		if err := authService.IncrementTokenVersion(userID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		auditLog(db, c, "auth.revoke_all", strconv.Itoa(int(userID)), "all tokens revoked")
		c.JSON(http.StatusOK, gin.H{"message": "all tokens have been revoked, please login again"})
	}
}
