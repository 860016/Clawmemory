package services

import (
	"sync"

	"gorm.io/gorm"
)

// AppContainer is the central service registry. All services are lazily
// initialized on first access and cached for subsequent calls.
type AppContainer struct {
	mu sync.RWMutex
	db *gorm.DB

	// lazily-initialized services
	memorySvc      *MemoryService
	searchSvc      *SearchService
	embeddingSvc   *EmbeddingService
	riskSwitchSvc  *RiskSwitchService
	syncSvc        *OpenClawSyncService
	evolutionSvc   *EvolutionService
	dedupSvc       *DedupService
	decaySvc       *DecayService
	governanceSvc  *GovernanceService
	chromaDBSvc    *ChromaDBService
	smartLoadSvc   *SmartLoadService
	toolboxSvc     *ToolboxService
	projectSvc     *ProjectService
	templateSvc    *TemplateService
	healthSvc      *HealthService
	validationSvc  *ValidationService
	writebackSvc   *MemoryWritebackService
	shareSvc       *MemoryShareService
	skillSvc       *SkillLearningService
	wikiSvc        *WikiService
	apiKeySvc      *APIKeyService
	knowledgeSvc   *KnowledgeService
	dailyReportSvc *DailyReportService
	settingsSvc    *SettingsService
	invitationSvc  *InvitationService
}

// NewAppContainer creates a new container with the given DB.
func NewAppContainer(db *gorm.DB) *AppContainer {
	return &AppContainer{db: db}
}

func (c *AppContainer) DB() *gorm.DB {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.db
}

// --- Lazy getters ---

func (c *AppContainer) MemoryService() *MemoryService {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.memorySvc == nil {
		c.memorySvc = NewMemoryService(c.db)
	}
	return c.memorySvc
}

func (c *AppContainer) SearchService() *SearchService {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.searchSvc == nil {
		c.searchSvc = NewSearchService(c.db)
	}
	return c.searchSvc
}

func (c *AppContainer) EmbeddingService() *EmbeddingService {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.embeddingSvc
}

func (c *AppContainer) SetEmbeddingService(svc *EmbeddingService) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.embeddingSvc = svc
}

func (c *AppContainer) RiskSwitchService() *RiskSwitchService {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.riskSwitchSvc == nil {
		c.riskSwitchSvc = InitRiskSwitchService(c.db)
	}
	return c.riskSwitchSvc
}

func (c *AppContainer) SyncService() *OpenClawSyncService {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.syncSvc == nil {
		c.syncSvc = GetOpenClawSyncService(c.db)
	}
	return c.syncSvc
}

func (c *AppContainer) EvolutionService() *EvolutionService {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.evolutionSvc == nil {
		c.evolutionSvc = NewEvolutionService(c.db)
	}
	return c.evolutionSvc
}

func (c *AppContainer) DedupService() *DedupService {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.dedupSvc == nil {
		c.dedupSvc = NewDedupService(c.db)
	}
	return c.dedupSvc
}

func (c *AppContainer) DecayService() *DecayService {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.decaySvc == nil {
		c.decaySvc = NewDecayService(c.db)
	}
	return c.decaySvc
}

func (c *AppContainer) GovernanceService() *GovernanceService {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.governanceSvc == nil {
		c.governanceSvc = NewGovernanceService(c.db)
	}
	return c.governanceSvc
}

func (c *AppContainer) ChromaDBService() *ChromaDBService {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.chromaDBSvc == nil {
		c.chromaDBSvc = NewChromaDBService(c.db)
	}
	return c.chromaDBSvc
}

func (c *AppContainer) SmartLoadService() *SmartLoadService {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.smartLoadSvc == nil {
		c.smartLoadSvc = NewSmartLoadService(c.db)
	}
	return c.smartLoadSvc
}

func (c *AppContainer) ToolboxService() *ToolboxService {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.toolboxSvc == nil {
		c.toolboxSvc = NewToolboxService(c.db)
	}
	return c.toolboxSvc
}

func (c *AppContainer) ProjectService() *ProjectService {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.projectSvc == nil {
		c.projectSvc = NewProjectService(c.db)
	}
	return c.projectSvc
}

func (c *AppContainer) TemplateService() *TemplateService {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.templateSvc == nil {
		c.templateSvc = NewTemplateService(c.db)
	}
	return c.templateSvc
}

func (c *AppContainer) HealthService() *HealthService {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.healthSvc == nil {
		c.healthSvc = NewHealthService(c.db)
	}
	return c.healthSvc
}

func (c *AppContainer) ValidationService() *ValidationService {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.validationSvc == nil {
		c.validationSvc = NewValidationService(c.db)
	}
	return c.validationSvc
}

func (c *AppContainer) WritebackService() *MemoryWritebackService {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.writebackSvc == nil {
		c.writebackSvc = NewMemoryWritebackService(c.db)
	}
	return c.writebackSvc
}

func (c *AppContainer) ShareService() *MemoryShareService {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.shareSvc == nil {
		c.shareSvc = NewMemoryShareService(c.db)
	}
	return c.shareSvc
}

func (c *AppContainer) SkillService() *SkillLearningService {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.skillSvc == nil {
		c.skillSvc = NewSkillLearningService(c.db)
	}
	return c.skillSvc
}

func (c *AppContainer) WikiService() *WikiService {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.wikiSvc == nil {
		c.wikiSvc = NewWikiService(c.db)
	}
	return c.wikiSvc
}

func (c *AppContainer) APIKeyService() *APIKeyService {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.apiKeySvc == nil {
		c.apiKeySvc = NewAPIKeyService(c.db)
	}
	return c.apiKeySvc
}

func (c *AppContainer) KnowledgeService() *KnowledgeService {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.knowledgeSvc == nil {
		c.knowledgeSvc = NewKnowledgeService(c.db)
	}
	return c.knowledgeSvc
}

func (c *AppContainer) DailyReportService() *DailyReportService {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.dailyReportSvc == nil {
		c.dailyReportSvc = NewDailyReportService(c.db)
	}
	return c.dailyReportSvc
}

func (c *AppContainer) SettingsService() *SettingsService {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.settingsSvc == nil {
		c.settingsSvc = NewSettingsService(c.db)
	}
	return c.settingsSvc
}

func (c *AppContainer) InvitationService() *InvitationService {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.invitationSvc == nil {
		c.invitationSvc = NewInvitationService(c.db)
	}
	return c.invitationSvc
}
