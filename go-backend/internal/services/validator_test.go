package services

import (
	"testing"
)

// --- Validator ---

func TestValidator_Required(t *testing.T) {
	v := NewValidator()
	v.Required("name", "")
	if !v.HasErrors() {
		t.Error("expected error for empty required field")
	}

	v2 := NewValidator()
	v2.Required("name", "hello")
	if v2.HasErrors() {
		t.Error("expected no error for non-empty field")
	}
}

func TestValidator_MaxLength(t *testing.T) {
	v := NewValidator()
	v.MaxLength("key", "short", 10)
	if v.HasErrors() {
		t.Error("short value should pass max length check")
	}

	v2 := NewValidator()
	v2.MaxLength("key", "this is way too long", 5)
	if !v2.HasErrors() {
		t.Error("long value should fail max length check")
	}
}

func TestValidator_MinLength(t *testing.T) {
	v := NewValidator()
	v.MinLength("password", "abc", 6)
	if !v.HasErrors() {
		t.Error("short value should fail min length check")
	}

	v2 := NewValidator()
	v2.MinLength("password", "longenough", 6)
	if v2.HasErrors() {
		t.Error("long enough value should pass min length check")
	}
}

func TestValidator_RangeFloat(t *testing.T) {
	v := NewValidator()
	v.RangeFloat("importance", 0.5, 0, 1)
	if v.HasErrors() {
		t.Error("0.5 should be in [0,1]")
	}

	v2 := NewValidator()
	v2.RangeFloat("importance", 1.5, 0, 1)
	if !v2.HasErrors() {
		t.Error("1.5 should be out of [0,1]")
	}
}

func TestValidator_OneOf(t *testing.T) {
	v := NewValidator()
	v.OneOf("layer", "core", []string{"core", "context", "detail"})
	if v.HasErrors() {
		t.Error("core should be a valid layer")
	}

	v2 := NewValidator()
	v2.OneOf("layer", "invalid", []string{"core", "context", "detail"})
	if !v2.HasErrors() {
		t.Error("invalid should not be a valid layer")
	}
}

func TestValidator_Chaining(t *testing.T) {
	v := NewValidator()
	v.Required("key", "").MaxLength("key", "", 500)
	if len(v.Errors()) != 1 {
		t.Errorf("expected 1 error (Required fails, MaxLength on empty is ok), got %d", len(v.Errors()))
	}
}

// --- ValidateMemoryCreate ---

func TestValidateMemoryCreate_Valid(t *testing.T) {
	err := ValidateMemoryCreate(map[string]interface{}{
		"key":   "test-key",
		"value": "test value",
		"layer": "core",
	})
	if err != nil {
		t.Errorf("valid memory should pass: %v", err)
	}
}

func TestValidateMemoryCreate_MissingKey(t *testing.T) {
	err := ValidateMemoryCreate(map[string]interface{}{
		"value": "test value",
	})
	if err == nil {
		t.Error("missing key should fail validation")
	}
}

func TestValidateMemoryCreate_InvalidLayer(t *testing.T) {
	err := ValidateMemoryCreate(map[string]interface{}{
		"key":   "test",
		"value": "val",
		"layer": "invalid",
	})
	if err == nil {
		t.Error("invalid layer should fail validation")
	}
}

func TestValidateMemoryCreate_InvalidImportance(t *testing.T) {
	err := ValidateMemoryCreate(map[string]interface{}{
		"key":        "test",
		"value":      "val",
		"importance": 5.0,
	})
	if err == nil {
		t.Error("importance > 1 should fail validation")
	}
}

// --- ValidateUsername ---

func TestValidateUsername_Valid(t *testing.T) {
	if err := ValidateUsername("testuser"); err != nil {
		t.Errorf("valid username should pass: %v", err)
	}
}

func TestValidateUsername_TooShort(t *testing.T) {
	if err := ValidateUsername("ab"); err == nil {
		t.Error("short username should fail")
	}
}

func TestValidateUsername_SpecialChars(t *testing.T) {
	if err := ValidateUsername("user@name"); err == nil {
		t.Error("special chars in username should fail")
	}
}

// --- ValidatePassword ---

func TestValidatePassword_Valid(t *testing.T) {
	if err := ValidatePassword("secret123"); err != nil {
		t.Errorf("valid password should pass: %v", err)
	}
}

func TestValidatePassword_TooShort(t *testing.T) {
	if err := ValidatePassword("abc"); err == nil {
		t.Error("short password should fail")
	}
}

// --- APIKey format validation ---

func TestAPIKeyFormatValidation(t *testing.T) {
	// Test that the API key format constants are correct
	if APIKeyPrefix != "cm" {
		t.Errorf("APIKeyPrefix should be 'cm', got %q", APIKeyPrefix)
	}
	if APIKeyLength != 50 {
		t.Errorf("APIKeyLength should be 50, got %d", APIKeyLength)
	}
	if MaxAPIKeysPerUser != 5 {
		t.Errorf("MaxAPIKeysPerUser should be 5, got %d", MaxAPIKeysPerUser)
	}
}

func TestValidPermissions_Contains(t *testing.T) {
	required := []string{"memories:read", "memories:write", "admin"}
	for _, p := range required {
		if !ValidPermissions[p] {
			t.Errorf("ValidPermissions should contain %q", p)
		}
	}
}

func TestValidPermissions_Excludes(t *testing.T) {
	invalid := []string{"invalid:perm", "hack", ""}
	for _, p := range invalid {
		if ValidPermissions[p] {
			t.Errorf("ValidPermissions should not contain %q", p)
		}
	}
}
