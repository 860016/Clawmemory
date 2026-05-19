package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"time"

	"clawmemory/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const DefaultUsername = "admin"

const (
	MaxFailedAttempts = 5
	LockoutDuration   = 15 * time.Minute
	AccessTokenTTL    = 24 * time.Hour
	RefreshTokenTTL   = 30 * 24 * time.Hour
)

type LoginResult struct {
	AccessToken     string     `json:"access_token"`
	RefreshToken    string     `json:"refresh_token"`
	RequiresCaptcha bool       `json:"requires_captcha"`
	LockedUntil     *time.Time `json:"locked_until,omitempty"`
}

type AuthService struct {
	db        *gorm.DB
	jwtSecret []byte
}

func NewAuthService(db *gorm.DB, jwtSecret string) *AuthService {
	return &AuthService{
		db:        db,
		jwtSecret: []byte(jwtSecret),
	}
}

func (s *AuthService) IsPasswordSet() bool {
	var count int64
	s.db.Model(&models.User{}).Count(&count)
	return count > 0
}

func (s *AuthService) CheckInitStatus() (bool, error) {
	var count int64
	if err := s.db.Model(&models.User{}).Count(&count).Error; err != nil {
		log.Printf("CheckInitStatus: database error: %v", err)
		return false, err
	}
	return count > 0, nil
}

func (s *AuthService) SetPassword(password string) (*LoginResult, error) {
	return s.SetPasswordWithUsername(DefaultUsername, password)
}

func (s *AuthService) SetPasswordWithUsername(username, password string) (*LoginResult, error) {
	if s.IsPasswordSet() {
		log.Println("SetPassword: password already set")
		return nil, errors.New("password already set")
	}

	if len(username) < 2 || len(username) > 50 {
		return nil, errors.New("username must be between 2 and 50 characters")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("SetPassword: failed to hash password: %v", err)
		return nil, err
	}

	user := &models.User{
		Username:  username,
		Password:  string(hashedPassword),
		Role:      "admin",
		IsFounder: true,
	}

	if err := s.db.Create(user).Error; err != nil {
		log.Printf("SetPassword: failed to create user: %v", err)
		return nil, err
	}

	log.Printf("SetPassword: founder user created, ID=%d, username=%s", user.ID, username)

	accessToken, err := s.generateAccessToken(user.ID)
	if err != nil {
		return nil, err
	}
	refreshToken, err := s.generateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) LoginWithPassword(password string) (*LoginResult, error) {
	var user models.User
	if err := s.db.Where("username = ?", DefaultUsername).First(&user).Error; err != nil {
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	return s.completeLogin(&user)
}

func (s *AuthService) Register(username, password string) (*models.User, error) {
	return s.RegisterWithInvitation(username, password, "")
}

func (s *AuthService) RegisterWithInvitation(username, password, invitationCode string) (*models.User, error) {
	if len(username) < 2 || len(username) > 50 {
		return nil, errors.New("username must be between 2 and 50 characters")
	}
	if len(password) < 6 {
		return nil, errors.New("password must be at least 6 characters")
	}

	var existing models.User
	if err := s.db.Where("username = ?", username).First(&existing).Error; err == nil {
		return nil, errors.New("username already exists")
	}

	invSvc := NewInvitationService(s.db)
	var userCount int64
	s.db.Model(&models.User{}).Count(&userCount)

	role := "user"
	isFounder := false
	if userCount == 0 {
		role = "admin"
		isFounder = true
	} else {
		if invitationCode == "" {
			return nil, errors.New("invitation code is required for registration")
		}
		inv, err := invSvc.ValidateCode(invitationCode)
		if err != nil {
			return nil, err
		}
		_ = inv
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Username:  username,
		Password:  string(hashedPassword),
		Role:      role,
		IsFounder: isFounder,
	}

	if err := s.db.Create(user).Error; err != nil {
		return nil, err
	}

	if invitationCode != "" && userCount > 0 {
		invSvc.UseCode(invitationCode, user.ID)
	}

	return user, nil
}

func (s *AuthService) GetUserByID(userID uint) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *AuthService) FindFirstUser(user *models.User) error {
	return s.db.First(user).Error
}

func (s *AuthService) IsAccountLocked(username string) (bool, *time.Time, int) {
	var user models.User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		return false, nil, 0
	}
	if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
		return true, user.LockedUntil, user.FailedAttempts
	}
	if user.LockedUntil != nil && !user.LockedUntil.After(time.Now()) {
		s.db.Model(&user).Updates(map[string]interface{}{
			"locked_until":    nil,
			"failed_attempts": 0,
		})
		return false, nil, 0
	}
	return false, nil, user.FailedAttempts
}

func (s *AuthService) RequiresCaptcha(username string) bool {
	var user models.User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		return false
	}
	return user.FailedAttempts >= MaxFailedAttempts
}

func (s *AuthService) Login(username, password, captcha string) (*LoginResult, error) {
	var user models.User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, errors.New("invalid credentials")
	}

	if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
		return &LoginResult{
			RequiresCaptcha: true,
			LockedUntil:     user.LockedUntil,
		}, errors.New("account is temporarily locked, please try again later or use terminal command to reset password")
	}

	if user.FailedAttempts >= MaxFailedAttempts && captcha == "" {
		return &LoginResult{
			RequiresCaptcha: true,
		}, errors.New("captcha verification required after multiple failed attempts")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		newAttempts := user.FailedAttempts + 1
		updates := map[string]interface{}{
			"failed_attempts": newAttempts,
		}
		if newAttempts >= MaxFailedAttempts {
			lockedUntil := time.Now().Add(LockoutDuration)
			updates["locked_until"] = lockedUntil
			s.db.Model(&user).Updates(updates)
			return &LoginResult{
				RequiresCaptcha: true,
				LockedUntil:     &lockedUntil,
			}, errors.New("account locked for 15 minutes due to too many failed attempts, please use terminal command './clawmemory --reset-password NEW_PASSWORD' to reset")
		}
		s.db.Model(&user).Updates(updates)
		remaining := MaxFailedAttempts - newAttempts
		return &LoginResult{
			RequiresCaptcha: newAttempts >= MaxFailedAttempts,
		}, errors.New("invalid credentials, remaining attempts: " + fmt.Sprintf("%d", remaining))
	}

	return s.completeLogin(&user)
}

func (s *AuthService) completeLogin(user *models.User) (*LoginResult, error) {
	s.db.Model(user).Updates(map[string]interface{}{
		"failed_attempts": 0,
		"locked_until":    nil,
	})

	accessToken, err := s.generateAccessToken(user.ID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) RefreshAccessToken(refreshToken string) (*LoginResult, error) {
	var user models.User
	if err := s.db.Where("refresh_token = ?", refreshToken).First(&user).Error; err != nil {
		return nil, errors.New("invalid refresh token")
	}
	if user.RefreshTokenExp != nil && user.RefreshTokenExp.Before(time.Now()) {
		return nil, errors.New("refresh token expired, please login again")
	}

	accessToken, err := s.generateAccessToken(user.ID)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := s.generateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *AuthService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return errors.New("invalid old password")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.db.Model(&user).Updates(map[string]interface{}{
		"password":      string(hashedPassword),
		"token_version": gorm.Expr("token_version + 1"),
		"refresh_token": "",
	}).Error
}

func (s *AuthService) ResetPassword(userID uint, newPassword string) error {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.db.Model(&user).Updates(map[string]interface{}{
		"password":        string(hashedPassword),
		"token_version":   gorm.Expr("token_version + 1"),
		"failed_attempts": 0,
		"locked_until":    nil,
		"refresh_token":   "",
	}).Error
}

func (s *AuthService) IncrementTokenVersion(userID uint) error {
	return s.db.Model(&models.User{}).Where("id = ?", userID).
		Update("token_version", gorm.Expr("token_version + 1")).Error
}

func (s *AuthService) generateAccessToken(userID uint) (string, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":       userID,
		"token_version": user.TokenVersion,
		"type":          "access",
		"exp":           time.Now().Add(AccessTokenTTL).Unix(),
	})

	return token.SignedString(s.jwtSecret)
}

func (s *AuthService) generateRefreshToken(userID uint) (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	token := hex.EncodeToString(bytes)

	exp := time.Now().Add(RefreshTokenTTL)
	s.db.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"refresh_token":     token,
		"refresh_token_exp": exp,
	})

	return token, nil
}

func (s *AuthService) GetFounderCount() int64 {
	var count int64
	s.db.Model(&models.User{}).Where("is_founder = ?", true).Count(&count)
	return count
}
