package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"clawmemory/internal/models"

	"gorm.io/gorm"
)

const (
	MaxAPIKeysPerUser = 5
	APIKeyLength      = 50
	APIKeyPrefix      = "cm"
)

type APIKeyService struct {
	db *gorm.DB
}

func NewAPIKeyService(db *gorm.DB) *APIKeyService {
	return &APIKeyService{db: db}
}

func (s *APIKeyService) Create(userID uint, name string) (*models.APIKey, string, error) {
	return s.CreateWithPermissions(userID, name, "read,write", "")
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
		permissions = "read,write"
	}

	rawKey, err := generateAPIKey()
	if err != nil {
		return nil, "", err
	}

	hash := sha256.Sum256([]byte(rawKey))
	keyHash := hex.EncodeToString(hash[:])
	prefix := rawKey[:8]

	apiKey := &models.APIKey{
		UserID:      userID,
		Name:        name,
		KeyHash:     keyHash,
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

	hash := sha256.Sum256([]byte(rawKey))
	keyHash := hex.EncodeToString(hash[:])

	var apiKey models.APIKey
	if err := s.db.Where("key_hash = ?", keyHash).First(&apiKey).Error; err != nil {
		return nil, fmt.Errorf("invalid API key")
	}

	if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("API key expired")
	}

	now := time.Now()
	s.db.Model(&apiKey).Update("last_used_at", now)

	return &apiKey, nil
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
