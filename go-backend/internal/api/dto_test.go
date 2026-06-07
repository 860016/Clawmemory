package api

import (
	"testing"
)

func TestMemoryCreateRequest_ToMap(t *testing.T) {
	imp := 0.9
	r := MemoryCreateRequest{
		Key:         "test-key",
		Value:       "test value",
		Layer:       "core",
		Importance:  &imp,
		Tags:        []string{"go", "test"},
		MemoryType:  "knowledge",
		Visibility:  "private",
		SourceAgent: "unit-test",
	}

	m := r.ToMap()

	if m["key"] != "test-key" {
		t.Errorf("key mismatch: %v", m["key"])
	}
	if m["value"] != "test value" {
		t.Errorf("value mismatch: %v", m["value"])
	}
	if m["layer"] != "core" {
		t.Errorf("layer mismatch: %v", m["layer"])
	}
	if m["importance"] != 0.9 {
		t.Errorf("importance mismatch: %v", m["importance"])
	}
	if m["visibility"] != "private" {
		t.Errorf("visibility mismatch: %v", m["visibility"])
	}
	// Verify tags are []interface{} (not []string) for service layer compatibility
	tags, ok := m["tags"].([]interface{})
	if !ok {
		t.Fatalf("tags should be []interface{}, got %T", m["tags"])
	}
	if len(tags) != 2 || tags[0] != "go" || tags[1] != "test" {
		t.Errorf("tags mismatch: %v", tags)
	}
}

func TestMemoryCreateRequest_ToMap_OmitsEmpty(t *testing.T) {
	r := MemoryCreateRequest{
		Key:   "k",
		Value: "v",
	}

	m := r.ToMap()

	if _, ok := m["layer"]; ok {
		t.Error("empty layer should be omitted")
	}
	if _, ok := m["importance"]; ok {
		t.Error("nil importance should be omitted")
	}
	if _, ok := m["tags"]; ok {
		t.Error("empty tags should be omitted")
	}
	if _, ok := m["is_encrypted"]; ok {
		t.Error("false is_encrypted should be omitted")
	}
}

func TestMemoryUpdateRequest_ToMap(t *testing.T) {
	key := "new-key"
	val := "new-value"
	layer := "detail"
	imp := 0.5

	r := MemoryUpdateRequest{
		Key:        &key,
		Value:      &val,
		Layer:      &layer,
		Importance: &imp,
	}

	m := r.ToMap()

	if m["key"] != "new-key" {
		t.Errorf("key mismatch: %v", m["key"])
	}
	if m["value"] != "new-value" {
		t.Errorf("value mismatch: %v", m["value"])
	}
	if m["layer"] != "detail" {
		t.Errorf("layer mismatch: %v", m["layer"])
	}
	if m["importance"] != 0.5 {
		t.Errorf("importance mismatch: %v", m["importance"])
	}
}

func TestMemoryUpdateRequest_ToMap_NilOmitted(t *testing.T) {
	r := MemoryUpdateRequest{}

	m := r.ToMap()

	if len(m) != 0 {
		t.Errorf("all-nil request should produce empty map, got %d keys", len(m))
	}
}

func TestMemoryUpdateRequest_ToMap_PartialUpdate(t *testing.T) {
	val := "only-value-updated"
	r := MemoryUpdateRequest{
		Value: &val,
	}

	m := r.ToMap()

	if len(m) != 1 {
		t.Errorf("partial update should have 1 key, got %d", len(m))
	}
	if m["value"] != "only-value-updated" {
		t.Errorf("value mismatch: %v", m["value"])
	}
}
