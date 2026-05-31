package services

import (
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
	SummaryGenerated int   `json:"summary_generated"`
	AutoFixed        int   `json:"auto_fixed"`
	MergedGroups     int   `json:"merged_groups"`
	DecayApplied     int   `json:"decay_applied"`
	TrashCleaned     int64 `json:"trash_cleaned"`
	DurationMs       int64 `json:"duration_ms"`
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
	s := &GovernanceStatus{
		Config: GovernanceConfig{
			Enabled:          false,
			Interval:         "daily",
			AutoMergeSimilar: false,
			MergeThreshold:   0.9,
			AutoDecay:        true,
			AutoCleanup:      true,
			AutoSummary:      true,
			AutoFix:          true,
		},
	}
	governanceStatusMap[userID] = s
	return s
}

func (s *GovernanceService) GetStatus(userID uint) *GovernanceStatus {
	return getGovernanceStatus(userID)
}

func (s *GovernanceService) UpdateConfig(userID uint, config GovernanceConfig) error {
	status := getGovernanceStatus(userID)
	status.Config = config
	return nil
}

func (s *GovernanceService) RunFullGovernance(userID uint) (*GovernanceResult, error) {
	s.mu.Lock()
	status := getGovernanceStatus(userID)
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
	result := &GovernanceResult{}

	if status.Config.AutoFix {
		healthSvc := NewHealthService(s.db)
		fixResult, err := healthSvc.AutoFix(userID, nil)
		if err != nil {
			log.Printf("Governance: AutoFix error: %v", err)
		} else if fixed, ok := fixResult["fixed"].(int); ok {
			result.AutoFixed = fixed
		}
	}

	if status.Config.AutoSummary {
		smartLoadSvc := NewSmartLoadService(s.db)
		count, err := smartLoadSvc.BatchGenerateSummaries(userID)
		if err != nil {
			log.Printf("Governance: BatchGenerateSummaries error: %v", err)
		} else {
			result.SummaryGenerated = count
		}
	}

	if status.Config.AutoMergeSimilar {
		dedupSvc := NewDedupService(s.db)
		merged, err := dedupSvc.AutoMergeSimilar(userID, status.Config.MergeThreshold)
		if err != nil {
			log.Printf("Governance: AutoMergeSimilar error: %v", err)
		} else {
			result.MergedGroups = merged
		}
	}

	if status.Config.AutoDecay {
		decaySvc := NewDecayService(s.db)
		decayResult, err := decaySvc.ApplyDecay(userID)
		if err != nil {
			log.Printf("Governance: ApplyDecay error: %v", err)
		} else if affected, ok := decayResult["affected"].(int); ok {
			result.DecayApplied = affected
		}
	}

	if status.Config.AutoCleanup {
		decaySvc := NewDecayService(s.db)
		cleaned, err := decaySvc.AutoCleanupTrash(userID)
		if err != nil {
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
