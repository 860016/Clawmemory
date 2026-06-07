package services

import (
	"testing"
	"time"
)

// --- tokenize ---

func TestTokenize_English(t *testing.T) {
	tokens := tokenize("Hello World")
	if len(tokens) != 2 {
		t.Fatalf("expected 2 tokens, got %d: %v", len(tokens), tokens)
	}
	if tokens[0] != "hello" || tokens[1] != "world" {
		t.Fatalf("unexpected tokens: %v", tokens)
	}
}

func TestTokenize_Chinese(t *testing.T) {
	tokens := tokenize("编程语言")
	// "编程语言" → bigrams: ["编程", "程语", "语言"]
	if len(tokens) < 2 {
		t.Fatalf("expected at least 2 tokens for CJK bigram, got %d: %v", len(tokens), tokens)
	}
}

func TestTokenize_ShortChinese(t *testing.T) {
	tokens := tokenize("语言")
	// 2 chars: output whole
	if len(tokens) != 1 || tokens[0] != "语言" {
		t.Fatalf("expected ['语言'], got %v", tokens)
	}
}

func TestTokenize_SingleChineseChar(t *testing.T) {
	tokens := tokenize("语")
	if len(tokens) != 1 || tokens[0] != "语" {
		t.Fatalf("expected ['语'], got %v", tokens)
	}
}

func TestTokenize_Mixed(t *testing.T) {
	tokens := tokenize("Go编程语言")
	if len(tokens) == 0 {
		t.Fatal("expected some tokens for mixed text")
	}
}

func TestTokenize_StopWords(t *testing.T) {
	tokens := tokenize("the quick brown fox")
	// "the" is a stop word, should be filtered
	for _, tok := range tokens {
		if tok == "the" {
			t.Fatal("stop word 'the' should be filtered")
		}
	}
}

func TestTokenize_Empty(t *testing.T) {
	tokens := tokenize("")
	if len(tokens) != 0 {
		t.Fatalf("expected 0 tokens for empty string, got %d", len(tokens))
	}
}

// --- isCJK ---

func TestIsCJK(t *testing.T) {
	tests := []struct {
		r    rune
		want bool
	}{
		{'编', true},
		{'A', false},
		{'0', false},
		{' ', false},
		{0x4E00, true}, // CJK start
		{0x9FFF, true}, // CJK end
		{0x9FFF + 1, false},
	}
	for _, tt := range tests {
		if got := isCJK(tt.r); got != tt.want {
			t.Errorf("isCJK(%U) = %v, want %v", tt.r, got, tt.want)
		}
	}
}

// --- cjkBigram ---

func TestCjkBigram_Long(t *testing.T) {
	// "编程语言" → ["编程", "程语", "语言"]
	tokens := cjkBigram("编程语言")
	if len(tokens) != 3 {
		t.Fatalf("expected 3 bigrams, got %d: %v", len(tokens), tokens)
	}
}

func TestCjkBigram_TwoChars(t *testing.T) {
	tokens := cjkBigram("语言")
	if len(tokens) != 1 || tokens[0] != "语言" {
		t.Fatalf("expected ['语言'], got %v", tokens)
	}
}

func TestCjkBigram_OneChar(t *testing.T) {
	tokens := cjkBigram("语")
	if len(tokens) != 1 || tokens[0] != "语" {
		t.Fatalf("expected ['语'], got %v", tokens)
	}
}

// --- jaccardSimilarity ---

func TestJaccardSimilarity_Identical(t *testing.T) {
	sim := jaccardSimilarity("hello world", "hello world")
	if sim != 1.0 {
		t.Fatalf("expected 1.0 for identical strings, got %f", sim)
	}
}

func TestJaccardSimilarity_Different(t *testing.T) {
	sim := jaccardSimilarity("cat dog", "fish bird")
	if sim != 0.0 {
		t.Fatalf("expected 0.0 for completely different strings, got %f", sim)
	}
}

func TestJaccardSimilarity_Partial(t *testing.T) {
	sim := jaccardSimilarity("hello world", "hello there")
	if sim <= 0.0 || sim >= 1.0 {
		t.Fatalf("expected 0 < sim < 1 for partial overlap, got %f", sim)
	}
}

// --- computeSimilarity ---

func TestComputeSimilarity_SameKey(t *testing.T) {
	// When keys are very similar (>0.8), key gets higher weight
	sim := computeSimilarity("project alpha", "project alpha", "desc a", "desc b")
	// "project alpha" vs "project alpha" → jaccard=1.0, key_sim > 0.8
	// But "desc a" vs "desc b" → jaccard may be 0 due to stop words filtering single chars
	if sim < 0 || sim > 1 {
		t.Fatalf("expected 0 <= sim <= 1, got %f", sim)
	}
}

func TestComputeSimilarity_DifferentKey(t *testing.T) {
	sim := computeSimilarity("alpha", "beta", "desc a", "desc b")
	if sim < 0 || sim > 1 {
		t.Fatalf("expected 0 <= sim <= 1, got %f", sim)
	}
}

// --- keyPrefix ---

func TestKeyPrefix(t *testing.T) {
	tests := []struct {
		key  string
		want string
	}{
		{"project alpha", "proj"},
		{"test", "test"},
		{"ab", "ab"},
		{"", "_empty"},
		{"  ", "_empty"},
	}
	for _, tt := range tests {
		got := keyPrefix(tt.key)
		if got != tt.want {
			t.Errorf("keyPrefix(%q) = %q, want %q", tt.key, got, tt.want)
		}
	}
}

// --- SearchCache ---

func TestSearchCache_SetGet(t *testing.T) {
	cache := NewSearchCache(1e9, 100) // long TTL
	results := []GraphRAGResult{
		{MemoryID: 1, Key: "test", Score: 0.9},
	}
	cache.Set(1, "test query", 10, results)

	got, ok := cache.Get(1, "test query", 10)
	if !ok {
		t.Fatal("expected cache hit")
	}
	if len(got) != 1 || got[0].MemoryID != 1 {
		t.Fatalf("unexpected cache result: %v", got)
	}
}

func TestSearchCache_Miss(t *testing.T) {
	cache := NewSearchCache(1e9, 100)
	_, ok := cache.Get(1, "nonexistent", 10)
	if ok {
		t.Fatal("expected cache miss")
	}
}

func TestSearchCache_Expiry(t *testing.T) {
	cache := NewSearchCache(1, 100) // 1ns TTL = instant expiry
	cache.Set(1, "test", 10, []GraphRAGResult{{MemoryID: 1}})
	// Wait briefly for expiry
	time.Sleep(2 * time.Millisecond)
	_, ok := cache.Get(1, "test", 10)
	if ok {
		t.Fatal("expected cache miss after expiry")
	}
}

func TestSearchCache_Invalidate(t *testing.T) {
	cache := NewSearchCache(1e9, 100)
	cache.Set(1, "test", 10, []GraphRAGResult{{MemoryID: 1}})
	cache.Set(1, "other", 10, []GraphRAGResult{{MemoryID: 2}})
	cache.Set(2, "test", 10, []GraphRAGResult{{MemoryID: 3}})

	cache.Invalidate(1)

	_, ok1 := cache.Get(1, "test", 10)
	_, ok2 := cache.Get(1, "other", 10)
	_, ok3 := cache.Get(2, "test", 10)

	if ok1 || ok2 {
		t.Fatal("user 1 entries should be invalidated")
	}
	if !ok3 {
		t.Fatal("user 2 entries should remain")
	}
}

func TestSearchCache_MaxSize(t *testing.T) {
	cache := NewSearchCache(1e9, 3)
	for i := 0; i < 5; i++ {
		cache.Set(1, string(rune('a'+i)), 10, []GraphRAGResult{{MemoryID: uint(i)}})
	}
	// Cache should have evicted some entries to stay under maxSize
	count := 0
	for i := 0; i < 5; i++ {
		if _, ok := cache.Get(1, string(rune('a'+i)), 10); ok {
			count++
		}
	}
	if count > 3 {
		t.Fatalf("cache should have at most 3 entries, found %d", count)
	}
}
