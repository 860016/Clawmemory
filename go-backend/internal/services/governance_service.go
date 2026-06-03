package services

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"clawmemory/internal/models"

	"gorm.io/gorm"
)

type GovernanceService struct {
	db *gorm.DB
	mu sync.Mutex
}

func NewGovernanceService(db *gorm.DB) *GovernanceService {
	return &GovernanceService{db: db}
}

type GovernanceResult struct {
	SummaryGenerated int               `json:"summary_generated"`
	AutoFixed        int               `json:"auto_fixed"`
	MergedGroups     int               `json:"merged_groups"`
	DecayApplied     int               `json:"decay_applied"`
	TrashCleaned     int64             `json:"trash_cleaned"`
	DurationMs       int64             `json:"duration_ms"`
	StepErrors       map[string]string `json:"step_errors,omitempty"`
	StepDurations    map[string]int64  `json:"step_durations,omitempty"`
}

type GovernanceConfig struct {
	Enabled          bool    `json:"enabled"`
	Interval         string  `json:"interval"`
	AutoMergeSimilar bool    `json:"auto_merge_similar"`
	MergeThreshold   float64 `json:"merge_threshold"`
	AutoDecay        bool    `json:"auto_decay"`
	AutoCleanup      bool    `json:"auto_cleanup"`
	AutoSummary      bool    `json:"auto_summary"`
	AutoFix          bool    `json:"auto_fix"`
}

type GovernanceStatus struct {
	LastRunAt  *time.Time        `json:"last_run_at"`
	LastResult *GovernanceResult `json:"last_result"`
	NextRunAt  *time.Time        `json:"next_run_at"`
	Config     GovernanceConfig  `json:"config"`
	IsRunning  bool              `json:"is_running"`
}

var (
	governanceStatusMap = make(map[uint]*GovernanceStatus)
	governanceMu        sync.RWMutex
)

func getGovernanceStatus(userID uint) *GovernanceStatus {
	governanceMu.Lock()
	defer governanceMu.Unlock()
	if s, ok := governanceStatusMap[userID]; ok {
		return s
	}
	defaultConfig := GovernanceConfig{
		Enabled:          false,
		Interval:         "daily",
		AutoMergeSimilar: false,
		MergeThreshold:   0.9,
		AutoDecay:        true,
		AutoCleanup:      true,
		AutoSummary:      true,
		AutoFix:          true,
	}
	s := &GovernanceStatus{
		Config: defaultConfig,
	}
	governanceStatusMap[userID] = s
	return s
}

func getGovernanceStatusWithDB(db *gorm.DB, userID uint) *GovernanceStatus {
	status := getGovernanceStatus(userID)
	if config := loadGovernanceConfigFromDB(db, userID); config != nil {
		status.Config = *config
	}
	return status
}

func (s *GovernanceService) GetStatus(userID uint) *GovernanceStatus {
	return getGovernanceStatusWithDB(s.db, userID)
}

func (s *GovernanceService) UpdateConfig(userID uint, config GovernanceConfig) error {
	status := getGovernanceStatus(userID)
	status.Config = config

	settingsSvc := NewSettingsService(s.db)
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal governance config: %w", err)
	}
	settingsSvc.SetByKey(userID, "governance_config", string(configJSON))

	return nil
}

func loadGovernanceConfigFromDB(db *gorm.DB, userID uint) *GovernanceConfig {
	settingsSvc := NewSettingsService(db)
	val, err := settingsSvc.GetByKey(userID, "governance_config")
	if err != nil || val == nil {
		return nil
	}
	var config GovernanceConfig
	switch v := val.(type) {
	case string:
		if err := json.Unmarshal([]byte(v), &config); err != nil {
			return nil
		}
		return &config
	default:
		data, err := json.Marshal(v)
		if err != nil {
			return nil
		}
		if err := json.Unmarshal(data, &config); err != nil {
			return nil
		}
		return &config
	}
}

func (s *GovernanceService) RunFullGovernance(userID uint) (*GovernanceResult, error) {
	s.mu.Lock()
	status := getGovernanceStatusWithDB(s.db, userID)
	if status.IsRunning {
		s.mu.Unlock()
		return nil, fmt.Errorf("governance is already running")
	}
	status.IsRunning = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		status.IsRunning = false
		s.mu.Unlock()
	}()

	start := time.Now()
	result := &GovernanceResult{
		StepErrors:    make(map[string]string),
		StepDurations: make(map[string]int64),
	}

	if status.Config.AutoFix {
		stepStart := time.Now()
		healthSvc := NewHealthService(s.db)
		fixResult, err := healthSvc.AutoFix(userID, nil)
		result.StepDurations["auto_fix"] = time.Since(stepStart).Milliseconds()
		if err != nil {
			result.StepErrors["auto_fix"] = err.Error()
			log.Printf("Governance: AutoFix error: %v", err)
		} else if fixed, ok := fixResult["fixed"].(int); ok {
			result.AutoFixed = fixed
		}
	}

	if status.Config.AutoSummary {
		stepStart := time.Now()
		smartLoadSvc := NewSmartLoadService(s.db)
		count, err := smartLoadSvc.BatchGenerateSummaries(userID)
		result.StepDurations["auto_summary"] = time.Since(stepStart).Milliseconds()
		if err != nil {
			result.StepErrors["auto_summary"] = err.Error()
			log.Printf("Governance: BatchGenerateSummaries error: %v", err)
		} else {
			result.SummaryGenerated = count
		}
	}

	if status.Config.AutoMergeSimilar {
		stepStart := time.Now()
		dedupSvc := NewDedupService(s.db)
		merged, err := dedupSvc.AutoMergeSimilar(userID, status.Config.MergeThreshold)
		result.StepDurations["auto_merge"] = time.Since(stepStart).Milliseconds()
		if err != nil {
			result.StepErrors["auto_merge"] = err.Error()
			log.Printf("Governance: AutoMergeSimilar error: %v", err)
		} else {
			result.MergedGroups = merged
		}
	}

	if status.Config.AutoDecay {
		stepStart := time.Now()
		decaySvc := NewDecayService(s.db)
		decayResult, err := decaySvc.ApplyDecay(userID)
		result.StepDurations["auto_decay"] = time.Since(stepStart).Milliseconds()
		if err != nil {
			result.StepErrors["auto_decay"] = err.Error()
			log.Printf("Governance: ApplyDecay error: %v", err)
		} else {
			archived, _ := decayResult["archived"].(int)
			trashed, _ := decayResult["trashed"].(int)
			adjusted, _ := decayResult["adjusted"].(int)
			result.DecayApplied = archived + trashed + adjusted
		}
	}

	if status.Config.AutoCleanup {
		stepStart := time.Now()
		decaySvc := NewDecayService(s.db)
		cleaned, err := decaySvc.AutoCleanupTrash(userID)
		result.StepDurations["auto_cleanup"] = time.Since(stepStart).Milliseconds()
		if err != nil {
			result.StepErrors["auto_cleanup"] = err.Error()
			log.Printf("Governance: AutoCleanupTrash error: %v", err)
		} else {
			result.TrashCleaned = cleaned
		}
	}

	result.DurationMs = time.Since(start).Milliseconds()

	now := time.Now()
	status.LastRunAt = &now
	status.LastResult = result

	if status.Config.Enabled {
		var next time.Time
		switch status.Config.Interval {
		case "weekly":
			next = now.Add(7 * 24 * time.Hour)
		case "daily":
			next = now.Add(24 * time.Hour)
		default:
			next = now.Add(24 * time.Hour)
		}
		status.NextRunAt = &next
	}

	return result, nil
}

func (s *GovernanceService) RunAutoGovernanceForAllUsers() {
	var users []models.User
	if err := s.db.Find(&users).Error; err != nil {
		log.Printf("Governance: failed to load users: %v", err)
		return
	}

	for _, user := range users {
		status := getGovernanceStatus(user.ID)
		if !status.Config.Enabled {
			continue
		}
		if status.IsRunning {
			continue
		}
		if status.NextRunAt != nil && time.Now().Before(*status.NextRunAt) {
			continue
		}

		log.Printf("Governance: running auto governance for user %d", user.ID)
		result, err := s.RunFullGovernance(user.ID)
		if err != nil {
			log.Printf("Governance: auto governance failed for user %d: %v", user.ID, err)
			continue
		}
		log.Printf("Governance: user %d done - summaries:%d fixed:%d merged:%d decay:%d cleaned:%d",
			user.ID, result.SummaryGenerated, result.AutoFixed, result.MergedGroups, result.DecayApplied, result.TrashCleaned)
	}
}
