package services

import (
	"encoding/json"
	"math"
	"time"

	"clawmemory/internal/models"

	"gorm.io/gorm"
)

type DecayService struct {
	db *gorm.DB
}

func NewDecayService(db *gorm.DB) *DecayService {
	return &DecayService{db: db}
}

type DecayStatsResult struct {
	Total    int64 `json:"total"`
	Active   int64 `json:"active"`
	Archived int64 `json:"archived"`
	Trashed  int64 `json:"trashed"`
}

func (s *DecayService) GetStats(userID uint) (*DecayStatsResult, error) {
	var stats DecayStatsResult
	s.db.Model(&models.Memory{}).Where("user_id = ?", userID).Count(&stats.Total)
	s.db.Model(&models.Memory{}).Where("user_id = ? AND status = ?", userID, "active").Count(&stats.Active)
	s.db.Model(&models.Memory{}).Where("user_id = ? AND status = ?", userID, "archived").Count(&stats.Archived)
	s.db.Model(&models.Memory{}).Where("user_id = ? AND status = ?", userID, "trashed").Count(&stats.Trashed)
	return &stats, nil
}

func (s *DecayService) ApplyDecay(userID uint) (map[string]interface{}, error) {
	now := time.Now()
	var memories []models.Memory
	s.db.Where("user_id = ? AND status != ?", userID, "trashed").Find(&memories)

	archived := 0
	trashed := 0
	adjusted := 0

	for i := range memories {
		m := &memories[i]
		daysSinceAccess := now.Sub(m.UpdatedAt).Hours() / 24
		if m.LastAccessedAt != nil {
			daysSinceAccess = now.Sub(*m.LastAccessedAt).Hours() / 24
		}

		newImportance := m.Importance
		newStatus := m.Status
		newDecayStage := m.DecayStage

		if daysSinceAccess > 60 {
			newImportance = m.Importance * 0.3
			newStatus = "trashed"
			newDecayStage = 3
			trashedNow := now
			m.TrashedAt = &trashedNow
			trashed++
		} else if daysSinceAccess > 30 {
			newImportance = m.Importance * 0.7
			newStatus = "archived"
			newDecayStage = 2
			archived++
		} else if daysSinceAccess > 15 {
			newImportance = m.Importance * 0.9
			newDecayStage = 1
			adjusted++
		}

		newImportance = math.Max(newImportance, 0.05)

		if newImportance != m.Importance || newStatus != m.Status || newDecayStage != m.DecayStage {
			updates := map[string]interface{}{
				"importance":  newImportance,
				"status":      newStatus,
				"decay_stage": newDecayStage,
			}
			if m.TrashedAt != nil {
				updates["trashed_at"] = *m.TrashedAt
			}
			s.db.Model(m).Updates(updates)
		}
	}

	return map[string]interface{}{
		"processed":   len(memories),
		"archived":    archived,
		"trashed":     trashed,
		"adjusted":    adjusted,
		"algorithm":   "local_time_decay_v1",
	}, nil
}

func (s *DecayService) EmptyTrash(userID uint) (int64, error) {
	result := s.db.Where("user_id = ? AND status = ?", userID, "trashed").Delete(&models.Memory{})
	return result.RowsAffected, nil
}

func (s *DecayService) GetDecaySettings(userID uint) (map[string]interface{}, error) {
	settingsSvc := NewSettingsService(s.db)
	enabled, _ := settingsSvc.GetByKey(userID, "auto_decay")
	if enabled == nil {
		enabled = false
	}
	config, _ := settingsSvc.GetByKey(userID, "decay_config")
	if config == nil {
		config = map[string]interface{}{
			"stage1_days": 15,
			"stage1_rate": 0.1,
			"stage2_days": 30,
			"stage2_rate": 0.3,
			"stage3_days": 60,
			"trash_retain_days": 30,
		}
	}
	return map[string]interface{}{
		"enabled": enabled,
		"config":  config,
	}, nil
}

func (s *DecayService) UpdateDecaySettings(userID uint, enabled bool, config map[string]interface{}) error {
	settingsSvc := NewSettingsService(s.db)
	settingsSvc.SetByKey(userID, "auto_decay", enabled)
	if config != nil {
		settingsSvc.SetByKey(userID, "decay_config", config)
	}
	return nil
}

func (s *DecayService) AutoCleanupTrash(userID uint) (int64, error) {
	retainDays := 30
	settingsSvc := NewSettingsService(s.db)
	config, err := settingsSvc.GetByKey(userID, "decay_config")
	if err == nil {
		if configMap, ok := config.(map[string]interface{}); ok {
			if rd, ok := configMap["trash_retain_days"].(float64); ok {
				retainDays = int(rd)
			}
		}
	}

	cutoff := time.Now().AddDate(0, 0, -retainDays)
	var trashMemories []models.Memory
	s.db.Where("user_id = ? AND status = ? AND trashed_at < ?", userID, "trashed", cutoff).Find(&trashMemories)

	count := int64(len(trashMemories))
	if count > 0 {
		ids := make([]uint, len(trashMemories))
		for i, m := range trashMemories {
			ids[i] = m.ID
		}
		s.db.Where("id IN ?", ids).Delete(&models.Memory{})
	}
	return count, nil
}

func (s *DecayService) GetPruneSuggestions(userID uint) ([]map[string]interface{}, error) {
	var memories []models.Memory
	s.db.Where("user_id = ? AND status = ?", userID, "active").Find(&memories)

	suggestions := []map[string]interface{}{}
	for _, m := range memories {
		if m.Importance < 0.2 {
			suggestions = append(suggestions, map[string]interface{}{
				"id":                m.ID,
				"key":               m.Key,
				"layer":             m.Layer,
				"importance":        m.Importance,
				"decayed_importance": m.Importance * 0.7,
				"reason":            "low_importance",
			})
		}
	}
	return suggestions, nil
}

func init() {
	_ = json.Marshal
	_ = math.Max
}
