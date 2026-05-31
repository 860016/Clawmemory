package services

import (
	"strings"
	"unicode/utf8"

	"clawmemory/internal/models"

	"gorm.io/gorm"
)

type ValidationService struct {
	db *gorm.DB
}

func NewValidationService(db *gorm.DB) *ValidationService {
	return &ValidationService{db: db}
}

type ValidationResult struct {
	Status string   `json:"status"`
	Errors []string `json:"errors,omitempty"`
	Warns  []string `json:"warns,omitempty"`
	Score  float64  `json:"score"`
}

func (s *ValidationService) Validate(memory *models.Memory) ValidationResult {
	var errors []string
	var warns []string
	score := 1.0

	if strings.TrimSpace(memory.Key) == "" {
		errors = append(errors, "key is empty")
		score -= 0.3
	}
	if strings.TrimSpace(memory.Value) == "" {
		errors = append(errors, "value is empty")
		score -= 0.3
	}
	if utf8.RuneCountInString(memory.Key) > 200 {
		warns = append(warns, "key exceeds 200 characters")
		score -= 0.1
	}
	if utf8.RuneCountInString(memory.Value) > 50000 {
		warns = append(warns, "value exceeds 50000 characters, consider splitting")
		score -= 0.1
	}

	if memory.Layer == "" {
		errors = append(errors, "layer is not specified")
		score -= 0.2
	}

	validLayers := map[string]bool{"episodic": true, "semantic": true, "procedural": true}
	if memory.Layer != "" && !validLayers[memory.Layer] {
		warns = append(warns, "layer '"+memory.Layer+"' is not a standard layer (episodic/semantic/procedural)")
		score -= 0.05
	}

	if memory.Importance < 0 || memory.Importance > 1 {
		errors = append(errors, "importance must be between 0 and 1")
		score -= 0.2
	}

	keyScan := ScanSecrets(memory.Key)
	if keyScan.Found {
		for _, m := range keyScan.Matches {
			errors = append(errors, "key contains "+m.Description+" ("+m.Type+")")
		}
		score -= 0.5
	}
	valueScan := ScanSecrets(memory.Value)
	if valueScan.Found {
		for _, m := range valueScan.Matches {
			errors = append(errors, "value contains "+m.Description+" ("+m.Type+")")
		}
		score -= 0.5
	}

	if score < 0 {
		score = 0
	}

	status := "valid"
	if len(errors) > 0 {
		status = "invalid"
	} else if len(warns) > 0 {
		status = "warning"
	}

	return ValidationResult{
		Status: status,
		Errors: errors,
		Warns:  warns,
		Score:  score,
	}
}

func (s *ValidationService) ValidateDTO(key, value, layer string, importance float64) ValidationResult {
	m := &models.Memory{
		Key:        key,
		Value:      value,
		Layer:      layer,
		Importance: importance,
	}
	return s.Validate(m)
}

func (s *ValidationService) ValidateAndMark(memory *models.Memory) ValidationResult {
	result := s.Validate(memory)
	memory.ValidationStatus = result.Status
	return result
}

func (s *ValidationService) BatchValidate(userID uint) (map[string]interface{}, error) {
	var memories []models.Memory
	logDBErr("load memories for batch validation", s.db.Where("user_id = ? AND status = ? AND validation_status = ?", userID, "active", "pending").
		Limit(5000).Find(&memories).Error)

	validated := 0
	invalid := 0
	warning := 0

	for i := range memories {
		result := s.Validate(&memories[i])
		logDBErr("update validation status", s.db.Model(&models.Memory{}).Where("id = ?", memories[i].ID).
			Update("validation_status", result.Status).Error)
		switch result.Status {
		case "valid":
			validated++
		case "invalid":
			invalid++
		case "warning":
			warning++
		}
	}

	return map[string]interface{}{
		"total":   len(memories),
		"valid":   validated,
		"invalid": invalid,
		"warning": warning,
	}, nil
}
