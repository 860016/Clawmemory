package services

import (
	"container/list"
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"gorm.io/gorm"
)

type EmbeddingProvider interface {
	Embed(ctx context.Context, texts []string) ([][]float64, error)
}

type AIEmbedRouter interface {
	AIEmbed(ctx context.Context, userID uint, isPro bool, texts []string) ([][]float64, error)
}

type lruEntry struct {
	key       string
	vec       []float64
	createdAt time.Time
}

type lruCache struct {
	mu      sync.RWMutex
	items   map[string]*list.Element
	order   *list.List
	maxSize int
	ttl     time.Duration
}

func newLRUCache(maxSize int, ttl time.Duration) *lruCache {
	return &lruCache{
		items:   make(map[string]*list.Element),
		order:   list.New(),
		maxSize: maxSize,
		ttl:     ttl,
	}
}

func (c *lruCache) Get(key string) ([]float64, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[key]; ok {
		entry := elem.Value.(*lruEntry)
		if time.Since(entry.createdAt) < c.ttl {
			c.order.MoveToFront(elem)
			return entry.vec, true
		}
		c.order.Remove(elem)
		delete(c.items, key)
	}
	return nil, false
}

func (c *lruCache) Set(key string, vec []float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[key]; ok {
		c.order.Remove(elem)
		delete(c.items, key)
	}

	for c.order.Len() >= c.maxSize {
		oldest := c.order.Back()
		if oldest != nil {
			c.order.Remove(oldest)
			delete(c.items, oldest.Value.(*lruEntry).key)
		}
	}

	entry := &lruEntry{key: key, vec: vec, createdAt: time.Now()}
	elem := c.order.PushFront(entry)
	c.items[key] = elem
}

func (c *lruCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

func (c *lruCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]*list.Element)
	c.order.Init()
}

type EmbeddingService struct {
	db       *gorm.DB
	provider EmbeddingProvider
	aiRouter AIEmbedRouter
	lru      *lruCache
}

var (
	globalEmbeddingService *EmbeddingService
	embeddingOnce          sync.Once
)

func InitEmbeddingService(db *gorm.DB, provider EmbeddingProvider) *EmbeddingService {
	embeddingOnce.Do(func() {
		globalEmbeddingService = &EmbeddingService{
			db:       db,
			provider: provider,
			lru:      newLRUCache(10000, 24*time.Hour),
		}
	})
	return globalEmbeddingService
}

func SetEmbeddingAIRouter(router AIEmbedRouter) {
	embeddingOnce.Do(func() {
		if globalEmbeddingService == nil {
			globalEmbeddingService = &EmbeddingService{
				lru: newLRUCache(10000, 24*time.Hour),
			}
		}
	})
	globalEmbeddingService.aiRouter = router
}

func GetEmbeddingService() *EmbeddingService {
	if globalEmbeddingService == nil {
		return nil
	}
	return globalEmbeddingService
}

func NewEmbeddingService(db *gorm.DB, provider EmbeddingProvider) *EmbeddingService {
	return &EmbeddingService{
		db:       db,
		provider: provider,
		lru:      newLRUCache(10000, 24*time.Hour),
	}
}

func (s *EmbeddingService) GetEmbedding(text string) []float64 {
	if vec, ok := s.cacheGet(text); ok {
		return vec
	}

	vec := s.generateEmbedding(text)
	s.cacheSet(text, vec)
	return vec
}

func (s *EmbeddingService) GetEmbeddings(texts []string) [][]float64 {
	results := make([][]float64, len(texts))
	uncached := make([]int, 0)
	uncachedTexts := make([]string, 0)

	for i, text := range texts {
		if vec, ok := s.cacheGet(text); ok {
			results[i] = vec
		} else {
			uncached = append(uncached, i)
			uncachedTexts = append(uncachedTexts, text)
		}
	}

	if len(uncachedTexts) > 0 {
		vectors, err := s.tryAIEmbed(uncachedTexts)
		if err != nil || len(vectors) == 0 {
			if err != nil {
				log.Printf("AI Embedding unavailable, using enhanced local: %v", err)
			}
			for _, idx := range uncached {
				results[idx] = enhancedTextToVector(texts[idx])
			}
		} else {
			for i, idx := range uncached {
				if i < len(vectors) && len(vectors[i]) > 0 {
					results[idx] = vectors[i]
					s.cacheSet(uncachedTexts[i], vectors[i])
				} else {
					results[idx] = enhancedTextToVector(texts[idx])
				}
			}
		}
	}

	return results
}

func (s *EmbeddingService) tryAIEmbed(texts []string) ([][]float64, error) {
	if s.aiRouter != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		vectors, err := s.aiRouter.AIEmbed(ctx, 1, false, texts)
		if err == nil && len(vectors) > 0 {
			return vectors, nil
		}
		if err != nil {
			return nil, err
		}
	}

	if s.provider != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		vectors, err := s.provider.Embed(ctx, texts)
		if err == nil && len(vectors) > 0 {
			return vectors, nil
		}
		if err != nil {
			return nil, err
		}
	}

	return nil, fmt.Errorf("no embedding provider available")
}

func (s *EmbeddingService) generateEmbedding(text string) []float64 {
	vectors, err := s.tryAIEmbed([]string{text})
	if err == nil && len(vectors) > 0 && len(vectors[0]) > 0 {
		return vectors[0]
	}
	if err != nil {
		log.Printf("AI Embedding unavailable, using enhanced local: %v", err)
	}
	return enhancedTextToVector(text)
}

func (s *EmbeddingService) cacheGet(key string) ([]float64, bool) {
	return s.lru.Get(key)
}

func (s *EmbeddingService) cacheSet(key string, vec []float64) {
	s.lru.Set(key, vec)
}

func (s *EmbeddingService) ClearCache() {
	s.lru.Clear()
}

func (s *EmbeddingService) CacheStats() map[string]interface{} {
	return map[string]interface{}{
		"size":     s.lru.Len(),
		"ttl":      "24h0m0s",
		"max_size": 10000,
		"type":     "lru",
	}
}
