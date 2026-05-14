package services

import (
	"sync"

	"gorm.io/gorm"
)

type AppContainer struct {
	mu sync.RWMutex

	db               *gorm.DB
	embeddingService *EmbeddingService
	searchService    *SearchService
	riskSwitchSvc    *RiskSwitchService
	syncService      *OpenClawSyncService
}

var globalContainer *AppContainer
var containerOnce sync.Once

func InitContainer(db *gorm.DB) *AppContainer {
	containerOnce.Do(func() {
		globalContainer = &AppContainer{
			db: db,
		}
	})
	return globalContainer
}

func GetContainer() *AppContainer {
	if globalContainer == nil {
		return nil
	}
	return globalContainer
}

func (c *AppContainer) DB() *gorm.DB {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.db
}

func (c *AppContainer) SetDB(db *gorm.DB) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.db = db
}

func (c *AppContainer) Embedding() *EmbeddingService {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.embeddingService
}

func (c *AppContainer) SetEmbeddingService(svc *EmbeddingService) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.embeddingService = svc
}

func (c *AppContainer) Search() *SearchService {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.searchService
}

func (c *AppContainer) SetSearchService(svc *SearchService) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.searchService = svc
}

func (c *AppContainer) RiskSwitch() *RiskSwitchService {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.riskSwitchSvc
}

func (c *AppContainer) SetRiskSwitchService(svc *RiskSwitchService) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.riskSwitchSvc = svc
}

func (c *AppContainer) SyncService() *OpenClawSyncService {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.syncService
}

func (c *AppContainer) SetSyncService(svc *OpenClawSyncService) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.syncService = svc
}
