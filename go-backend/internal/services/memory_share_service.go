package services

import (
	"fmt"
	"time"

	"clawmemory/internal/models"

	"gorm.io/gorm"
)

type MemoryShareService struct {
	db *gorm.DB
}

func NewMemoryShareService(db *gorm.DB) *MemoryShareService {
	return &MemoryShareService{db: db}
}

func (s *MemoryShareService) ShareMemory(memoryID, fromUserID uint, toAgent string, shareType string) (*models.MemoryShare, error) {
	var memory models.Memory
	if err := s.db.First(&memory, memoryID).Error; err != nil {
		return nil, fmt.Errorf("memory not found: %w", err)
	}

	if memory.UserID != fromUserID {
		return nil, fmt.Errorf("not authorized to share this memory")
	}

	if memory.Visibility == "private" && shareType != "manual" {
		return nil, fmt.Errorf("private memories can only be shared manually")
	}

	if shareType == "auto" {
		riskSvc := GetRiskSwitchService()
		if riskSvc != nil && riskSvc.IsDisabled(RiskShareAutoApprove) {
			shareType = "manual"
		}
	}

	share := &models.MemoryShare{
		MemoryID:   memoryID,
		FromUserID: fromUserID,
		ToAgent:    toAgent,
		ShareType:  shareType,
		Status:     "pending",
	}

	if shareType == "auto" {
		share.Status = "approved"
		now := time.Now()
		share.ApprovedAt = &now
		share.ApprovedBy = &fromUserID
	}

	if err := s.db.Create(share).Error; err != nil {
		return nil, fmt.Errorf("failed to create share: %w", err)
	}

	return share, nil
}

func (s *MemoryShareService) ApproveShare(shareID, approverID uint) error {
	var share models.MemoryShare
	if err := s.db.First(&share, shareID).Error; err != nil {
		return fmt.Errorf("share not found: %w", err)
	}

	if share.Status != "pending" {
		return fmt.Errorf("share is not pending (current: %s)", share.Status)
	}

	if share.ToUserID != 0 && share.ToUserID != approverID {
		return fmt.Errorf("not authorized to approve this share")
	}

	now := time.Now()
	share.Status = "approved"
	share.ApprovedBy = &approverID
	share.ApprovedAt = &now

	return s.db.Save(&share).Error
}

func (s *MemoryShareService) RejectShare(shareID, rejectorID uint) error {
	var share models.MemoryShare
	if err := s.db.First(&share, shareID).Error; err != nil {
		return fmt.Errorf("share not found: %w", err)
	}

	if share.Status != "pending" {
		return fmt.Errorf("share is not pending (current: %s)", share.Status)
	}

	if share.ToUserID != 0 && share.ToUserID != rejectorID {
		return fmt.Errorf("not authorized to reject this share")
	}

	share.Status = "rejected"
	return s.db.Save(&share).Error
}

func (s *MemoryShareService) RevokeShare(shareID, ownerID uint) error {
	var share models.MemoryShare
	if err := s.db.First(&share, shareID).Error; err != nil {
		return fmt.Errorf("share not found: %w", err)
	}

	if share.FromUserID != ownerID {
		return fmt.Errorf("not authorized to revoke this share")
	}

	share.Status = "revoked"
	return s.db.Save(&share).Error
}

func (s *MemoryShareService) GetPendingShares(userID uint) ([]models.MemoryShare, error) {
	var shares []models.MemoryShare
	err := s.db.Where("to_user_id = ? AND status = ?", userID, "pending").
		Preload("Memory").
		Find(&shares).Error
	return shares, err
}

func (s *MemoryShareService) GetSharedMemories(agent string) ([]models.Memory, error) {
	var shares []models.MemoryShare
	err := s.db.Where("to_agent = ? AND status = ?", agent, "approved").
		Find(&shares).Error
	if err != nil {
		return nil, err
	}

	memoryIDs := make([]uint, len(shares))
	for i, s := range shares {
		memoryIDs[i] = s.MemoryID
	}

	if len(memoryIDs) == 0 {
		return []models.Memory{}, nil
	}

	var memories []models.Memory
	err = s.db.Where("id IN ?", memoryIDs).Find(&memories).Error
	return memories, err
}

func (s *MemoryShareService) GetOutboundShares(userID uint) ([]models.MemoryShare, error) {
	var shares []models.MemoryShare
	err := s.db.Where("from_user_id = ?", userID).
		Preload("Memory").
		Find(&shares).Error
	return shares, err
}

func (s *MemoryShareService) ProcessAutoShareRules(userID uint, memory *models.Memory) error {
	if memory.Visibility == "public" {
		var rules []models.ShareRule
		_ = s.db.Where("user_id = ? AND enabled = ? AND auto_approve = ?", userID, true, true).Find(&rules).Error

		for _, rule := range rules {
			if rule.SourceAgent != "" && rule.SourceAgent != memory.SourceAgent {
				continue
			}
			if rule.Layer != "" && rule.Layer != memory.Layer {
				continue
			}
			if memory.Importance < rule.MinImportance {
				continue
			}

			s.ShareMemory(memory.ID, userID, rule.TargetAgent, "auto")
		}
	}

	return nil
}

func (s *MemoryShareService) CreateRule(rule *models.ShareRule) error {
	return s.db.Create(rule).Error
}

func (s *MemoryShareService) UpdateRule(rule *models.ShareRule) error {
	return s.db.Save(rule).Error
}

func (s *MemoryShareService) DeleteRule(ruleID, userID uint) error {
	return s.db.Where("id = ? AND user_id = ?", ruleID, userID).Delete(&models.ShareRule{}).Error
}

func (s *MemoryShareService) ListRules(userID uint) ([]models.ShareRule, error) {
	var rules []models.ShareRule
	err := s.db.Where("user_id = ?", userID).Find(&rules).Error
	return rules, err
}
