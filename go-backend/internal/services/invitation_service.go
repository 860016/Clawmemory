package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"clawmemory/internal/models"

	"gorm.io/gorm"
)

type InvitationService struct {
	db *gorm.DB
}

func NewInvitationService(db *gorm.DB) *InvitationService {
	return &InvitationService{db: db}
}

func (s *InvitationService) GenerateCode(createdBy uint, maxUses int, expiresAt *time.Time) (*models.Invitation, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return nil, err
	}
	code := hex.EncodeToString(bytes)

	inv := &models.Invitation{
		Code:      code,
		CreatedBy: createdBy,
		MaxUses:   maxUses,
		ExpiresAt: expiresAt,
	}

	if err := s.db.Create(inv).Error; err != nil {
		return nil, err
	}

	return inv, nil
}

func (s *InvitationService) ValidateCode(code string) (*models.Invitation, error) {
	var inv models.Invitation
	if err := s.db.Where("code = ?", code).First(&inv).Error; err != nil {
		return nil, errors.New("invalid invitation code")
	}

	if inv.UsedCount >= inv.MaxUses {
		return nil, errors.New("invitation code has been fully used")
	}

	if inv.ExpiresAt != nil && time.Now().After(*inv.ExpiresAt) {
		return nil, errors.New("invitation code has expired")
	}

	return &inv, nil
}

func (s *InvitationService) UseCode(code string, userID uint) error {
	var inv models.Invitation
	if err := s.db.Where("code = ?", code).First(&inv).Error; err != nil {
		return errors.New("invalid invitation code")
	}

	if inv.UsedCount >= inv.MaxUses {
		return errors.New("invitation code has been fully used")
	}

	if inv.ExpiresAt != nil && time.Now().After(*inv.ExpiresAt) {
		return errors.New("invitation code has expired")
	}

	now := time.Now()
	updates := map[string]interface{}{
		"used_count": gorm.Expr("used_count + 1"),
	}

	if inv.MaxUses == 1 {
		updates["used_by"] = userID
		updates["used_at"] = now
	}

	return s.db.Model(&inv).Updates(updates).Error
}

func (s *InvitationService) ListCodes(createdBy uint) ([]models.Invitation, error) {
	var invitations []models.Invitation
	err := s.db.Where("created_by = ?", createdBy).Order("created_at DESC").Find(&invitations).Error
	return invitations, err
}

func (s *InvitationService) DeleteCode(codeID, createdBy uint) error {
	return s.db.Where("id = ? AND created_by = ?", codeID, createdBy).Delete(&models.Invitation{}).Error
}

func (s *InvitationService) IsAdmin(userID uint) bool {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return false
	}
	return user.Role == "admin"
}
