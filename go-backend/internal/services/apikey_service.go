package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"clawmemory/internal/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	MaxAPIKeysPerUser = 5
	APIKeyLength      = 50
	APIKeyPrefix      = "cm"
	bcryptCost        = 10
)

var ValidPermissions = map[string]bool{
	"memories:read":       true,
	"memories:write":      true,
	"conversations:write": true,
	"sessions:write":      true,
	"reason:execute":      true,
	"ai:execute":          true,
	"read":                true,
	"write":               true,
	"admin":               true,
}

type APIKeyService struct {
	db *gorm.DB
}

func NewAPIKeyService(db *gorm.DB) *APIKeyService {
	return &APIKeyService{db: db}
}

func (s *APIKeyService) Create(userID uint, name string) (*models.APIKey, string, error) {
	return s.CreateWithPermissions(userID, name, "memories:read,memories:write,conversations:write,sessions:write,reason:execute,ai:execute", "")
}

func (s *APIKeyService) CreateWithPermissions(userID uint, name, permissions, agentName string) (*models.APIKey, string, error) {
	var count int64
	s.db.Model(&models.APIKey{}).Where("user_id = ?", userID).Count(&count)
	if count >= MaxAPIKeysPerUser {
		return nil, "", fmt.Errorf("maximum %d API keys per user", MaxAPIKeysPerUser)
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return nil, "", fmt.Errorf("name is required")
	}
	if len(name) > 100 {
		name = name[:100]
	}

	if permissions == "" {
		permissions = "memories:read,memories:write,conversations:write,sessions:write,reason:execute"
	}

	for _, p := range strings.Split(permissions, ",") {
		p = strings.TrimSpace(p)
		if p != "" && !ValidPermissions[p] {
			return nil, "", fmt.Errorf("invalid permission: %s", p)
		}
	}

	rawKey, err := generateAPIKey()
	if err != nil {
		return nil, "", err
	}

	keyHash, err := bcrypt.GenerateFromPassword([]byte(rawKey), bcryptCost)
	if err != nil {
		return nil, "", fmt.Errorf("failed to hash API key: %w", err)
	}
	prefix := rawKey[:8]

	apiKey := &models.APIKey{
		UserID:      userID,
		Name:        name,
		KeyHash:     string(keyHash),
		KeyPrefix:   prefix,
		Permissions: permissions,
		AgentName:   agentName,
	}

	if err := s.db.Create(apiKey).Error; err != nil {
		return nil, "", err
	}

	return apiKey, rawKey, nil
}

func (s *APIKeyService) List(userID uint) ([]models.APIKey, error) {
	var keys []models.APIKey
	err := s.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&keys).Error
	return keys, err
}

func (s *APIKeyService) Delete(userID uint, id uint) error {
	result := s.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.APIKey{})
	if result.RowsAffected == 0 {
		return fmt.Errorf("API key not found")
	}
	return result.Error
}

func (s *APIKeyService) Validate(rawKey string) (*models.APIKey, error) {
	if !strings.HasPrefix(rawKey, APIKeyPrefix) || len(rawKey) != APIKeyLength {
		return nil, fmt.Errorf("invalid API key format")
	}

	// Look up by prefix first to avoid full-table scan
	prefix := rawKey[:8]
	var apiKey models.APIKey
	if err := s.db.Where("key_prefix = ?", prefix).First(&apiKey).Error; err != nil {
		return nil, fmt.Errorf("invalid API key")
	}

	// Verify key hash: try bcrypt first, fall back to legacy SHA-256
	matched := false
	if err := bcrypt.CompareHashAndPassword([]byte(apiKey.KeyHash), []byte(rawKey)); err == nil {
		matched = true
	} else {
		// Legacy SHA-256 fallback
		hash := sha256.Sum256([]byte(rawKey))
		shaHash := hex.EncodeToString(hash[:])
		if apiKey.KeyHash == shaHash {
			matched = true
			// Auto-upgrade: re-hash with bcrypt
			if newHash, err := bcrypt.GenerateFromPassword([]byte(rawKey), bcryptCost); err == nil {
				s.db.Model(&apiKey).Update("key_hash", string(newHash))
			}
		}
	}

	if !matched {
		return nil, fmt.Errorf("invalid API key")
	}

	// Expiry check
	if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("API key expired")
	}

	// Enabled check
	if !apiKey.IsEnabled {
		return nil, fmt.Errorf("API key is disabled")
	}

	// Lock check
	if apiKey.LockedUntil != nil && apiKey.LockedUntil.After(time.Now()) {
		return nil, fmt.Errorf("API key is temporarily locked, try again after %s", apiKey.LockedUntil.Format(time.RFC3339))
	}

	// Failed attempts threshold
	if apiKey.FailedAttempts >= 10 {
		lockedUntil := time.Now().Add(30 * time.Minute)
		s.db.Model(&apiKey).Updates(map[string]interface{}{
			"locked_until":    lockedUntil,
			"failed_attempts": 0,
		})
		return nil, fmt.Errorf("API key locked due to too many failed attempts")
	}

	now := time.Now()
	s.db.Model(&apiKey).Updates(map[string]interface{}{
		"last_used_at":    now,
		"failed_attempts": 0,
	})

	return &apiKey, nil
}

// IncrementFailedAttempts increases the failed attempt counter for an API key
func (s *APIKeyService) IncrementFailedAttempts(rawKey string) {
	if !strings.HasPrefix(rawKey, APIKeyPrefix) || len(rawKey) != APIKeyLength {
		return
	}
	prefix := rawKey[:8]
	s.db.Model(&models.APIKey{}).Where("key_prefix = ?", prefix).
		UpdateColumn("failed_attempts", gorm.Expr("failed_attempts + 1"))
}

func (s *APIKeyService) Count(userID uint) int64 {
	var count int64
	s.db.Model(&models.APIKey{}).Where("user_id = ?", userID).Count(&count)
	return count
}

func (s *APIKeyService) HasPermission(apiKey *models.APIKey, permission string) bool {
	perms := strings.Split(apiKey.Permissions, ",")
	for _, p := range perms {
		p = strings.TrimSpace(p)
		if p == "admin" || p == permission {
			return true
		}
	}
	return false
}

func (s *APIKeyService) GetAgentName(apiKey *models.APIKey) string {
	if apiKey.AgentName != "" {
		return apiKey.AgentName
	}
	return "unknown"
}

func generateAPIKey() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return APIKeyPrefix + hex.EncodeToString(b), nil
}
