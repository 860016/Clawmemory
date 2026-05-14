package services

import (
	"context"

	"clawmemory/internal/models"
)

type MemoryServiceInterface interface {
	Create(userID uint, data map[string]interface{}) (*MemoryModel, error)
	Get(userID, id uint) (*MemoryModel, error)
	List(userID uint, page, pageSize int, filters map[string]interface{}) ([]MemoryModel, int64, error)
	Update(userID, id uint, data map[string]interface{}) (*MemoryModel, error)
	Delete(userID, id uint) error
	Search(userID uint, query string, limit int) ([]MemoryModel, error)
}

type AuthServiceInterface interface {
	Login(username, password string) (string, error)
	RegisterWithInvitation(username, password, invitationCode string) (*models.User, error)
	ChangePassword(userID uint, oldPassword, newPassword string) error
	ResetPassword(userID uint, newPassword string) error
	IncrementTokenVersion(userID uint) error
	GetUserByID(userID uint) (*models.User, error)
}

type APIKeyServiceInterface interface {
	CreateWithPermissions(userID uint, name, permissions, agentName string) (*models.APIKey, string, error)
	List(userID uint) ([]models.APIKey, error)
	Delete(userID, id uint) error
	Validate(rawKey string) (*models.APIKey, error)
	HasPermission(apiKey *models.APIKey, permission string) bool
}

type MemoryShareServiceInterface interface {
	ShareMemory(memoryID, fromUserID uint, toAgent string, shareType string) (*models.MemoryShare, error)
	ApproveShare(shareID, approverID uint) error
	RejectShare(shareID, rejectorID uint) error
	RevokeShare(shareID, ownerID uint) error
	GetPendingShares(userID uint) ([]models.MemoryShare, error)
	GetSharedMemories(agent string) ([]models.Memory, error)
	GetOutboundShares(userID uint) ([]models.MemoryShare, error)
	CreateRule(rule *models.ShareRule) error
	ListRules(userID uint) ([]models.ShareRule, error)
	DeleteRule(ruleID, userID uint) error
}

type EmbeddingProviderInterface interface {
	Embed(ctx context.Context, texts []string) ([][]float64, error)
}
