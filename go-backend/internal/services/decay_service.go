package services

import (
	"encoding/json"
	"fmt"
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

	archived := 0
	trashed := 0
	adjusted := 0
	reinforced := 0
	locked := 0

	batchSize := 200
	offset := 0

	for {
		var memories []models.Memory
		if err := s.db.Where("user_id = ? AND status != ?", userID, "trashed").
			Order("importance ASC").Limit(batchSize).Offset(offset).Find(&memories).Error; err != nil {
			return nil, fmt.Errorf("failed to load memories: %w", err)
		}
		if len(memories) == 0 {
			break
		}

		for i := range memories {
			m := &memories[i]
			daysSinceAccess := now.Sub(m.UpdatedAt).Hours() / 24
			if m.LastAccessedAt != nil {
				daysSinceAccess = now.Sub(*m.LastAccessedAt).Hours() / 24
			}

			newImportance := m.Importance
			newStatus := m.Status
			newDecayStage := m.DecayStage

			if m.ReinforceCount >= 5 {
				reinforced++
				continue
			}

			if m.Layer == "core" {
				locked++
				if daysSinceAccess > 365 {
					newImportance = m.Importance * 0.98
					newDecayStage = 1
					adjusted++
				}
				newImportance = math.Max(newImportance, 0.7)
				if newImportance != m.Importance || newDecayStage != m.DecayStage {
					updates := map[string]interface{}{
						"importance":  newImportance,
						"decay_stage": newDecayStage,
					}
					s.db.Model(m).Updates(updates)
				}
				continue
			}

			halfLife := layerHalfLife(m.Layer, m.ReinforceCount)
			decayFactor := math.Pow(0.5, daysSinceAccess/halfLife)
			newImportance = m.Importance * decayFactor

			thresholds := layerThresholds(m.Layer)

			if newImportance < thresholds.trash {
				newStatus = "trashed"
				newDecayStage = 3
				trashedNow := now
				m.TrashedAt = &trashedNow
				trashed++
			} else if newImportance < thresholds.archive {
				newStatus = "archived"
				newDecayStage = 2
				archived++
			} else if newImportance < m.Importance*0.95 {
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

		if len(memories) < batchSize {
			break
		}
		offset += batchSize
	}

	return map[string]interface{}{
		"processed":  len(memories),
		"archived":   archived,
		"trashed":    trashed,
		"adjusted":   adjusted,
		"reinforced": reinforced,
		"locked":     locked,
		"algorithm":  "layered_exponential_decay",
	}, nil
}

type decayThresholds struct {
	archive float64
	trash   float64
}

func layerHalfLife(layer string, reinforceCount int) float64 {
	base := 30.0
	switch layer {
	case "core":
		return 3650.0
	case "context":
		base = 60.0
	case "detail":
		base = 15.0
	default:
		base = 30.0
	}
	return base + float64(reinforceCount)*10.0
}

func layerThresholds(layer string) decayThresholds {
	switch layer {
	case "core":
		return decayThresholds{archive: 0.5, trash: 0.2}
	case "context":
		return decayThresholds{archive: 0.3, trash: 0.1}
	case "detail":
		return decayThresholds{archive: 0.2, trash: 0.05}
	default:
		return decayThresholds{archive: 0.3, trash: 0.1}
	}
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
			"stage1_days":       15,
			"stage1_rate":       0.1,
			"stage2_days":       30,
			"stage2_rate":       0.3,
			"stage3_days":       60,
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
	result := s.db.Where("user_id = ? AND status = ? AND trashed_at < ?", userID, "trashed", cutoff).Delete(&models.Memory{})
	return result.RowsAffected, nil
}

func (s *DecayService) GetPruneSuggestions(userID uint) ([]map[string]interface{}, error) {
	var memories []models.Memory
	if err := s.db.Where("user_id = ? AND status = ?", userID, "active").Limit(5000).Find(&memories).Error; err != nil {
		return nil, err
	}
	suggestions := []map[string]interface{}{}
	for _, m := range memories {
		if m.Importance < 0.2 {
			suggestions = append(suggestions, map[string]interface{}{
				"id":                 m.ID,
				"key":                m.Key,
				"layer":              m.Layer,
				"importance":         m.Importance,
				"decayed_importance": m.Importance * 0.7,
				"reason":             "low_importance",
			})
		}
	}
	return suggestions, nil
}

func init() {
	_ = json.Marshal
	_ = math.Max
}
