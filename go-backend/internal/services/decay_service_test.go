package services

import (
	"testing"
)

// --- DecayDefaults ---

func TestDecayDefaults_LayerHalfLife(t *testing.T) {
	svc := NewDecayService(nil)
	cfg := svc.config

	tests := []struct {
		layer    string
		expected float64
	}{
		{"core", cfg.CoreHalfLife},
		{"context", cfg.ContextHalfLife},
		{"detail", cfg.DetailHalfLife},
		{"unknown", cfg.DefaultHalfLife},
	}
	for _, tt := range tests {
		got := svc.layerHalfLife(tt.layer, 0)
		if got != tt.expected {
			t.Errorf("layerHalfLife(%q, 0) = %f, want %f", tt.layer, got, tt.expected)
		}
	}
}

func TestDecayDefaults_LayerHalfLife_WithReinforce(t *testing.T) {
	svc := NewDecayService(nil)
	cfg := svc.config

	got := svc.layerHalfLife("detail", 3)
	want := cfg.DetailHalfLife + 3*cfg.ReinforceBonus
	if got != want {
		t.Errorf("layerHalfLife(%q, 3) = %f, want %f", "detail", got, want)
	}
}

// --- layerThresholds ---

func TestLayerThresholds(t *testing.T) {
	svc := NewDecayService(nil)

	tests := []struct {
		layer           string
		wantArchive     float64
		wantTrash       float64
	}{
		{"core", 0.5, 0.2},
		{"context", 0.3, 0.1},
		{"detail", 0.2, 0.05},
		{"unknown", 0.3, 0.1}, // default
	}
	for _, tt := range tests {
		th := svc.layerThresholds(tt.layer)
		if th.archive != tt.wantArchive || th.trash != tt.wantTrash {
			t.Errorf("layerThresholds(%q) = {archive: %f, trash: %f}, want {archive: %f, trash: %f}",
				tt.layer, th.archive, th.trash, tt.wantArchive, tt.wantTrash)
		}
	}
}

// --- DecayConfig values ---

func TestDecayConfig_Values(t *testing.T) {
	cfg := defaultDecayConfig

	if cfg.CoreHalfLife <= 0 {
		t.Error("CoreHalfLife should be positive")
	}
	if cfg.ContextHalfLife >= cfg.CoreHalfLife {
		t.Error("ContextHalfLife should be less than CoreHalfLife")
	}
	if cfg.DetailHalfLife >= cfg.ContextHalfLife {
		t.Error("DetailHalfLife should be less than ContextHalfLife")
	}
	if cfg.GlobalMinImportance <= 0 {
		t.Error("GlobalMinImportance should be positive")
	}
	if cfg.CoreMinImportance < cfg.GlobalMinImportance {
		t.Error("CoreMinImportance should be >= GlobalMinImportance")
	}
}

// --- CompressThresholds ---

func TestCompressThresholds(t *testing.T) {
	cfg := defaultDecayConfig

	levels := []string{"light", "medium", "heavy", "default"}
	for _, level := range levels {
		th, ok := cfg.CompressThresholds[level]
		if !ok {
			t.Errorf("missing compress threshold for level %q", level)
		}
		if th <= 0 || th > 1 {
			t.Errorf("compress threshold for %q = %f, want (0,1]", level, th)
		}
	}

	// heavier compression should have higher threshold
	if cfg.CompressThresholds["light"] >= cfg.CompressThresholds["medium"] {
		t.Error("light threshold should be less than medium")
	}
	if cfg.CompressThresholds["medium"] >= cfg.CompressThresholds["heavy"] {
		t.Error("medium threshold should be less than heavy")
	}
}

// --- LayerThresholds completeness ---

func TestLayerThresholds_AllLayers(t *testing.T) {
	cfg := defaultDecayConfig

	for _, layer := range []string{"core", "context", "detail", "default"} {
		th, ok := cfg.LayerThresholds[layer]
		if !ok {
			t.Errorf("missing layer threshold for %q", layer)
		}
		if th.archive <= th.trash {
			t.Errorf("archive threshold (%f) should be > trash threshold (%f) for %q", th.archive, th.trash, layer)
		}
	}
}
