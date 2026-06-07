package services

import (
	"fmt"
	"math"
	"time"

	"clawmemory/internal/models"

	"gorm.io/gorm"
)

type decayThresholds struct {
	archive float64
	trash   float64
}

type DecayService struct {
	db     *gorm.DB
	config DecayDefaults
}

type DecayDefaults struct {
	CoreHalfLife          float64
	ContextHalfLife       float64
	DetailHalfLife        float64
	DefaultHalfLife       float64
	ReinforceBonus        float64
	CoreMinImportance     float64
	GlobalMinImportance   float64
	CoreYearlyDecayFactor float64
	CoreReinforceMin      int
	DecayAdjustThreshold  float64
	PruneImportanceBelow  float64
	BatchSize             int
	CompressThresholds    map[string]float64
	LayerThresholds       map[string]decayThresholds
}

var defaultDecayConfig = DecayDefaults{
	CoreHalfLife:          3650.0,
	ContextHalfLife:       60.0,
	DetailHalfLife:        15.0,
	DefaultHalfLife:       30.0,
	ReinforceBonus:        10.0,
	CoreMinImportance:     0.7,
	GlobalMinImportance:   0.05,
	CoreYearlyDecayFactor: 0.98,
	CoreReinforceMin:      5,
	DecayAdjustThreshold:  0.95,
	PruneImportanceBelow:  0.2,
	BatchSize:             200,
	CompressThresholds: map[string]float64{
		"light":   0.2,
		"medium":  0.35,
		"heavy":   0.5,
		"default": 0.2,
	},
	LayerThresholds: map[string]decayThresholds{
		"core":    {archive: 0.5, trash: 0.2},
		"context": {archive: 0.3, trash: 0.1},
		"detail":  {archive: 0.2, trash: 0.05},
		"default": {archive: 0.3, trash: 0.1},
	},
}

func NewDecayService(db *gorm.DB) *DecayService {
	return &DecayService{db: db, config: defaultDecayConfig}
}

type DecayStatsResult struct {
	Total           int64                    `json:"total"`
	Active          int64                    `json:"active"`
	Archived        int64                    `json:"archived"`
	Trashed         int64                    `json:"trashed"`
	PruneCandidates []map[string]interface{} `json:"prune_candidates"`
}

func (s *DecayService) GetStats(userID uint) (*DecayStatsResult, error) {
	var stats DecayStatsResult
	s.db.Model(&models.Memory{}).Where("user_id = ?", userID).Count(&stats.Total)
	s.db.Model(&models.Memory{}).Where("user_id = ? AND status = ?", userID, "active").Count(&stats.Active)
	s.db.Model(&models.Memory{}).Where("user_id = ? AND status = ?", userID, "archived").Count(&stats.Archived)
	s.db.Model(&models.Memory{}).Where("user_id = ? AND status = ?", userID, "trashed").Count(&stats.Trashed)

	var memories []models.Memory
	s.db.Where("user_id = ? AND status = ?", userID, "active").Limit(5000).Find(&memories)
	stats.PruneCandidates = []map[string]interface{}{}
	for _, m := range memories {
		if m.Importance < s.config.PruneImportanceBelow {
			stats.PruneCandidates = append(stats.PruneCandidates, map[string]interface{}{
				"id":         m.ID,
				"key":        m.Key,
				"layer":      m.Layer,
				"importance": m.Importance,
				"reason":     "low_importance",
			})
		}
	}

	return &stats, nil
}

func (s *DecayService) ApplyDecay(userID uint) (map[string]interface{}, error) {
	now := time.Now()

	archived := 0
	trashed := 0
	adjusted := 0
	reinforced := 0
	locked := 0

	batchSize := s.config.BatchSize
	var lastID uint
	totalProcessed := 0

	for {
		var memories []models.Memory
		query := s.db.Where("user_id = ? AND status != ?", userID, "trashed")
		if lastID > 0 {
			query = query.Where("id > ?", lastID)
		}
		if err := query.Order("id ASC").Limit(batchSize).Find(&memories).Error; err != nil {
			return nil, fmt.Errorf("failed to load memories: %w", err)
		}
		if len(memories) == 0 {
			break
		}

		for i := range memories {
			m := &memories[i]
			lastID = m.ID
			daysSinceAccess := now.Sub(m.UpdatedAt).Hours() / 24
			if m.LastAccessedAt != nil {
				daysSinceAccess = now.Sub(*m.LastAccessedAt).Hours() / 24
			}

			newImportance := m.Importance
			newStatus := m.Status
			newDecayStage := m.DecayStage

			if m.ReinforceCount >= s.config.CoreReinforceMin {
				reinforced++
				continue
			}

			if m.Layer == "core" {
				locked++
				if daysSinceAccess > 365 {
					newImportance = m.Importance * s.config.CoreYearlyDecayFactor
					newDecayStage = 1
					adjusted++
				}
				newImportance = math.Max(newImportance, s.config.CoreMinImportance)
				if newImportance != m.Importance || newDecayStage != m.DecayStage {
					updates := map[string]interface{}{
						"importance":  newImportance,
						"decay_stage": newDecayStage,
					}
					s.db.Model(m).Updates(updates)
				}
				continue
			}

			halfLife := s.layerHalfLife(m.Layer, m.ReinforceCount)
			decayFactor := math.Pow(0.5, daysSinceAccess/halfLife)
			newImportance = m.Importance * decayFactor

			thresholds := s.layerThresholds(m.Layer)

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
			} else if newImportance < m.Importance*s.config.DecayAdjustThreshold {
				newDecayStage = 1
				adjusted++
			}

			newImportance = math.Max(newImportance, s.config.GlobalMinImportance)

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

		totalProcessed += len(memories)
		if len(memories) < batchSize {
			break
		}
	}

	return map[string]interface{}{
		"processed":  totalProcessed,
		"archived":   archived,
		"trashed":    trashed,
		"adjusted":   adjusted,
		"reinforced": reinforced,
		"locked":     locked,
		"algorithm":  "layered_exponential_decay",
	}, nil
}

func (s *DecayService) layerHalfLife(layer string, reinforceCount int) float64 {
	base := s.config.DefaultHalfLife
	switch layer {
	case "core":
		return s.config.CoreHalfLife
	case "context":
		base = s.config.ContextHalfLife
	case "detail":
		base = s.config.DetailHalfLife
	}
	return base + float64(reinforceCount)*s.config.ReinforceBonus
}

func (s *DecayService) layerThresholds(layer string) decayThresholds {
	if t, ok := s.config.LayerThresholds[layer]; ok {
		return t
	}
	return s.config.LayerThresholds["default"]
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

func (s *DecayService) CompressPreview(userID uint, level string) (map[string]interface{}, error) {
	var memories []models.Memory
	logDBErr("load memories for compress preview", s.db.Where("user_id = ? AND status != ?", userID, "trashed").Limit(5000).Find(&memories).Error)

	threshold := s.config.CompressThresholds["default"]
	if t, ok := s.config.CompressThresholds[level]; ok {
		threshold = t
	}
	if level == "deep" {
		if t, ok := s.config.CompressThresholds["heavy"]; ok {
			threshold = t
		}
	}

	preview := []map[string]interface{}{}
	for _, m := range memories {
		if m.Importance < threshold {
			preview = append(preview, map[string]interface{}{
				"memory_id":  m.ID,
				"key":        m.Key,
				"value_len":  len(m.Value),
				"importance": m.Importance,
				"action":     "archive",
			})
		}
	}

	return map[string]interface{}{
		"mode":      "local",
		"level":     level,
		"threshold": threshold,
		"preview":   preview,
		"total":     len(preview),
	}, nil
}

func (s *DecayService) CompressApply(userID uint, level string) (map[string]interface{}, error) {
	preview, err := s.CompressPreview(userID, level)
	if err != nil {
		return nil, err
	}

	previewItems, _ := preview["preview"].([]map[string]interface{})
	archived := 0
	for _, item := range previewItems {
		if id, ok := item["memory_id"]; ok {
			var memoryID uint
			switch v := id.(type) {
			case uint:
				memoryID = v
			case float64:
				memoryID = uint(v)
			case int:
				memoryID = uint(v)
			}
			if err := s.db.Model(&models.Memory{}).Where("id = ? AND user_id = ?", memoryID, userID).
				Update("status", "archived").Error; err == nil {
				archived++
			}
		}
	}

	return map[string]interface{}{
		"mode":     "local",
		"level":    level,
		"archived": archived,
		"total":    len(previewItems),
	}, nil
}
