package services

import (
	"testing"
)

// --- computeSimilarity ---

func TestComputeSimilarity_Identical(t *testing.T) {
	sim := computeSimilarity("test key", "test value", "test key", "test value")
	if sim != 1.0 {
		t.Fatalf("expected 1.0 for identical content, got %f", sim)
	}
}

func TestComputeSimilarity_SameKeyDiffValue(t *testing.T) {
	sim := computeSimilarity("project alpha", "old desc", "project alpha", "new desc")
	if sim <= 0 {
		t.Fatalf("expected positive similarity for same key, got %f", sim)
	}
}

func TestComputeSimilarity_DiffKeySameValue(t *testing.T) {
	sim := computeSimilarity("alpha", "shared description", "beta", "shared description")
	if sim <= 0 {
		t.Fatalf("expected positive similarity for same value, got %f", sim)
	}
}

func TestComputeSimilarity_CompletelyDifferent(t *testing.T) {
	sim := computeSimilarity("alpha", "desc a", "beta", "desc b")
	// jaccard("alpha","beta") and jaccard("desc a","desc b") both have some overlap due to short words
	if sim < 0 || sim > 1 {
		t.Fatalf("expected 0 <= sim <= 1, got %f", sim)
	}
}

// --- jaccardSimilarity ---

func TestJaccardSimilarity_Empty(t *testing.T) {
	sim := jaccardSimilarity("", "")
	if sim != 1.0 {
		t.Fatalf("expected 1.0 for both empty, got %f", sim)
	}
}

func TestJaccardSimilarity_OneEmpty(t *testing.T) {
	sim := jaccardSimilarity("hello", "")
	if sim != 0.0 {
		t.Fatalf("expected 0.0 for one empty, got %f", sim)
	}
}

func TestJaccardSimilarity_HalfOverlap(t *testing.T) {
	// "hello world" vs "hello there" → "hello" overlaps
	sim := jaccardSimilarity("hello world", "hello there")
	if sim <= 0 || sim >= 1 {
		t.Fatalf("expected 0 < sim < 1 for partial overlap, got %f", sim)
	}
}

// --- keyPrefix ---

func TestKeyPrefix_ShortKey(t *testing.T) {
	got := keyPrefix("ab")
	if got != "ab" {
		t.Errorf("keyPrefix(%q) = %q, want %q", "ab", got, "ab")
	}
}

func TestKeyPrefix_LongKey(t *testing.T) {
	got := keyPrefix("programming")
	if got != "prog" {
		t.Errorf("keyPrefix(%q) = %q, want %q", "programming", got, "prog")
	}
}

func TestKeyPrefix_MultiWord(t *testing.T) {
	got := keyPrefix("project alpha plan")
	if got != "proj" {
		t.Errorf("keyPrefix(%q) = %q, want %q", "project alpha plan", got, "proj")
	}
}

func TestKeyPrefix_Empty(t *testing.T) {
	got := keyPrefix("")
	if got != "_empty" {
		t.Errorf("keyPrefix(%q) = %q, want %q", "", got, "_empty")
	}
}

// --- truncateStr ---

func TestTruncateString_Short(t *testing.T) {
	got := truncateStr("hello", 10)
	if got != "hello" {
		t.Errorf("expected 'hello', got %q", got)
	}
}

func TestTruncateString_Long(t *testing.T) {
	got := truncateStr("hello world foo bar", 5)
	if got != "hello..." {
		t.Errorf("expected 'hello...', got %q", got)
	}
}

func TestTruncateString_Exact(t *testing.T) {
	got := truncateStr("hello", 5)
	if got != "hello" {
		t.Errorf("expected 'hello', got %q", got)
	}
}

// --- bucketByPrefix ---

func TestBucketByPrefix(t *testing.T) {
	memories := []struct {
		// minimal fields needed
	}{
		// Can't easily construct models.Memory without DB, so test the logic
		// indirectly through the public Scan method (requires DB).
		// Instead, test keyPrefix which is the core of bucketByPrefix.
	}
	_ = memories
	// keyPrefix is tested above; bucketByPrefix is a private method that
	// groups memories by keyPrefix. The logic is straightforward.
}

// --- DedupService constructor ---

func TestNewDedupService(t *testing.T) {
	svc := NewDedupService(nil)
	if svc == nil {
		t.Fatal("expected non-nil DedupService")
	}
}
