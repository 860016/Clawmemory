package services

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type Validator struct {
	errors []ValidationError
}

func NewValidator() *Validator {
	return &Validator{errors: make([]ValidationError, 0)}
}

func (v *Validator) HasErrors() bool {
	return len(v.errors) > 0
}

func (v *Validator) Errors() []ValidationError {
	return v.errors
}

func (v *Validator) AddError(field, message string) {
	v.errors = append(v.errors, ValidationError{Field: field, Message: message})
}

func (v *Validator) Error() string {
	if len(v.errors) == 0 {
		return ""
	}
	var msgs []string
	for _, e := range v.errors {
		msgs = append(msgs, fmt.Sprintf("%s: %s", e.Field, e.Message))
	}
	return strings.Join(msgs, "; ")
}

func (v *Validator) Required(field, value string) *Validator {
	if strings.TrimSpace(value) == "" {
		v.AddError(field, fmt.Sprintf("%s is required", field))
	}
	return v
}

func (v *Validator) MaxLength(field, value string, max int) *Validator {
	if utf8.RuneCountInString(value) > max {
		v.AddError(field, fmt.Sprintf("%s must be at most %d characters", field, max))
	}
	return v
}

func (v *Validator) MinLength(field, value string, min int) *Validator {
	if utf8.RuneCountInString(value) < min {
		v.AddError(field, fmt.Sprintf("%s must be at least %d characters", field, min))
	}
	return v
}

func (v *Validator) RangeInt(field string, value, min, max int) *Validator {
	if value < min || value > max {
		v.AddError(field, fmt.Sprintf("%s must be between %d and %d", field, min, max))
	}
	return v
}

func (v *Validator) RangeFloat(field string, value, min, max float64) *Validator {
	if value < min || value > max {
		v.AddError(field, fmt.Sprintf("%s must be between %.2f and %.2f", field, min, max))
	}
	return v
}

func (v *Validator) OneOf(field, value string, allowed []string) *Validator {
	for _, a := range allowed {
		if value == a {
			return v
		}
	}
	v.AddError(field, fmt.Sprintf("%s must be one of: %s", field, strings.Join(allowed, ", ")))
	return v
}

func (v *Validator) Regex(field, value, pattern, message string) *Validator {
	if matched, _ := regexp.MatchString(pattern, value); !matched {
		v.AddError(field, message)
	}
	return v
}

func ValidateMemoryCreate(data map[string]interface{}) *Validator {
	v := NewValidator()

	key, _ := data["key"].(string)
	value, _ := data["value"].(string)

	v.Required("key", key).MaxLength("key", key, 500)
	v.Required("value", value).MaxLength("value", value, 50000)

	if layer, ok := data["layer"].(string); ok && layer != "" {
		v.OneOf("layer", layer, []string{"core", "context", "detail", "episodic", "semantic", "preference", "knowledge", "short_term", "private"})
	}

	if memoryType, ok := data["memory_type"].(string); ok && memoryType != "" {
		v.OneOf("memory_type", memoryType, []string{"knowledge", "preference", "instruction", "context", "fact", "episodic", "conversation"})
	}

	if importance, ok := data["importance"].(float64); ok {
		v.RangeFloat("importance", importance, 0, 1)
	}

	if visibility, ok := data["visibility"].(string); ok && visibility != "" {
		v.OneOf("visibility", visibility, []string{"private", "shared", "public"})
	}

	return v
}

func ValidateUsername(username string) *Validator {
	v := NewValidator()
	v.Required("username", username).
		MinLength("username", username, 3).
		MaxLength("username", username, 50).
		Regex("username", username, `^[a-zA-Z0-9_-]+$`, "username can only contain letters, numbers, underscores, and hyphens")
	return v
}

func ValidatePassword(password string) *Validator {
	v := NewValidator()
	v.Required("password", password).
		MinLength("password", password, 6).
		MaxLength("password", password, 128)
	return v
}
