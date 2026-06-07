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

func handleInitStatus(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		passwordSet, err := authService.CheckInitStatus()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check init status"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"password_set": passwordSet,
		})
	}
}

func handleSetPassword(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if len(req.Password) < 6 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "password must be at least 6 characters"})
			return
		}

		username := req.Username
		if username == "" {
			username = "admin"
		}

		result, err := authService.SetPasswordWithUsername(username, req.Password)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		resp := gin.H{
			"access_token":  result.AccessToken,
			"refresh_token": result.RefreshToken,
		}
		if result.APIKey != "" {
			resp["api_key"] = result.APIKey
		}

		c.JSON(http.StatusOK, resp)
	}
}

func handleLogin(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password" binding:"required"`
			Captcha  string `json:"captcha"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if req.Username == "" {
			req.Username = "admin"
		}

		result, err := authService.Login(req.Username, req.Password, req.Captcha)
		if err != nil {
			status := http.StatusUnauthorized
			resp := gin.H{"error": err.Error()}
			if result != nil {
				resp["requires_captcha"] = result.RequiresCaptcha
				if result.LockedUntil != nil {
					resp["locked_until"] = result.LockedUntil
				}
			}
			c.JSON(status, resp)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"access_token":  result.AccessToken,
			"refresh_token": result.RefreshToken,
		})
	}
}

func handleGetMe(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		user, err := authService.GetUserByID(userID.(uint))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"role":       user.Role,
			"is_founder": user.IsFounder,
		})
	}
}

func handleRegister(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Username       string `json:"username" binding:"required"`
			Password       string `json:"password" binding:"required"`
			InvitationCode string `json:"invitation_code"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result, err := authService.RegisterWithInvitation(req.Username, req.Password, req.InvitationCode)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		resp := gin.H{
			"id":         result.User.ID,
			"username":   result.User.Username,
			"role":       result.User.Role,
			"is_founder": result.User.IsFounder,
		}
		if result.APIKey != "" {
			resp["api_key"] = result.APIKey
		}

		c.JSON(http.StatusCreated, resp)
	}
}

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

		result, err := authService.RegisterWithInvitation(req.Username, req.Password, req.InvitationCode)
		if err != nil {
			status := http.StatusBadRequest
			if err.Error() == "username already exists" {
				status = http.StatusConflict
			}
			c.JSON(status, gin.H{"error": err.Error()})
			return
		}

		resp := gin.H{
			"id":       result.User.ID,
			"username": result.User.Username,
			"role":     result.User.Role,
			"message":  "registration successful",
		}
		if result.APIKey != "" {
			resp["api_key"] = result.APIKey
		}

		c.JSON(http.StatusCreated, resp)
	}
}

func handleResetPassword(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			OldPassword string `json:"old_password"`
			NewPassword string `json:"new_password"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if req.OldPassword == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "old_password is required"})
			return
		}
		if req.NewPassword == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "new_password is required"})
			return
		}
		if len(req.NewPassword) < 6 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "new_password must be at least 6 characters"})
			return
		}

		userID := middleware.GetUserID(c)
		if err := authService.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "password reset successfully"})
	}
}

func handleForgotPassword(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "password reset is only available via terminal command",
			"hint":  "Run: ./clawmemory --reset-password NEW_PASSWORD",
		})
	}
}

func handleChangePassword(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		var req struct {
			OldPassword string `json:"old_password"`
			NewPassword string `json:"new_password"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if req.OldPassword == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "old_password is required"})
			return
		}
		if req.NewPassword == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "new_password is required"})
			return
		}
		if len(req.NewPassword) < 6 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "new_password must be at least 6 characters"})
			return
		}
		if err := authService.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "password changed successfully"})
	}
}

func handleRefreshToken(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			RefreshToken string `json:"refresh_token" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "refresh_token is required"})
			return
		}

		result, err := authService.RefreshAccessToken(req.RefreshToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"access_token":  result.AccessToken,
			"refresh_token": result.RefreshToken,
		})
	}
}

func handleLoginStatus(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Username string `json:"username"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if req.Username == "" {
			c.JSON(http.StatusOK, gin.H{
				"requires_captcha": false,
				"locked":           false,
			})
			return
		}

		locked, lockedUntil, failedAttempts := authService.IsAccountLocked(req.Username)
		requiresCaptcha := failedAttempts >= services.MaxFailedAttempts

		resp := gin.H{
			"requires_captcha": requiresCaptcha,
			"locked":           locked,
			"failed_attempts":  failedAttempts,
		}
		if lockedUntil != nil {
			resp["locked_until"] = lockedUntil
		}
		c.JSON(http.StatusOK, resp)
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
			MaxUses   int `json:"max_uses"`
			ExpiresIn int `json:"expires_in_hours"`
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

func handleListUsers(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		invSvc := services.NewInvitationService(db)
		if !invSvc.IsAdmin(userID) {
			c.JSON(http.StatusForbidden, gin.H{"error": "only admins can list users"})
			return
		}

		type UserInfo struct {
			ID          uint       `json:"id"`
			Username    string     `json:"username"`
			Role        string     `json:"role"`
			IsFounder   bool       `json:"is_founder"`
			CreatedAt   time.Time  `json:"created_at"`
			LockedUntil *time.Time `json:"locked_until,omitempty"`
		}

		var users []UserInfo
		if err := db.Table("users").Select("id, username, role, is_founder, created_at, locked_until").Order("id ASC").Find(&users).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"items": users})
	}
}

func handleAdminResetUserPassword(db *gorm.DB, authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		adminID := middleware.GetUserID(c)

		invSvc := services.NewInvitationService(db)
		if !invSvc.IsAdmin(adminID) {
			c.JSON(http.StatusForbidden, gin.H{"error": "only admins can reset user passwords"})
			return
		}

		var req struct {
			UserID      uint   `json:"user_id" binding:"required"`
			NewPassword string `json:"new_password" binding:"required,min=6"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := authService.ResetPassword(req.UserID, req.NewPassword); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		auditLog(db, c, "auth.admin_reset_password", strconv.Itoa(int(req.UserID)), "password reset by admin")
		c.JSON(http.StatusOK, gin.H{"message": "password reset successfully"})
	}
}
