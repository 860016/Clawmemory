package services

import (
	"encoding/json"
	"sync"

	"clawmemory/internal/models"

	"gorm.io/gorm"
)

type RiskSwitch string

const (
	RiskCrossAgentAccess   RiskSwitch = "risk_cross_agent_access"
	RiskAutoImportMemories RiskSwitch = "risk_auto_import_memories"
	RiskBulkDelete         RiskSwitch = "risk_bulk_delete"
	RiskAutoDestructive    RiskSwitch = "risk_auto_destructive"
)

type RiskSwitchConfig struct {
	Key          RiskSwitch `json:"key"`
	Label        string     `json:"label"`
	Description  string     `json:"description"`
	Category     string     `json:"category"`
	DefaultValue bool       `json:"default_value"`
}

var RiskSwitchDefinitions = []RiskSwitchConfig{
	{
		Key:          RiskCrossAgentAccess,
		Label:        "跨 Agent 访问控制",
		Description:  "允许跨 Agent 读取、写入、共享记忆，以及变更记忆可见性",
		Category:     "access",
		DefaultValue: false,
	},
	{
		Key:          RiskAutoImportMemories,
		Label:        "自动导入外部记忆",
		Description:  "允许自动从扫描到的 Agent 本地文件导入记忆，无需逐一确认",
		Category:     "import",
		DefaultValue: false,
	},
	{
		Key:          RiskBulkDelete,
		Label:        "批量删除记忆",
		Description:  "允许一次删除多条记忆，包括批量清空回收站",
		Category:     "destructive",
		DefaultValue: false,
	},
	{
		Key:          RiskAutoDestructive,
		Label:        "自动破坏性操作",
		Description:  "允许系统自动执行衰减归档、压缩等可能丢失信息的操作",
		Category:     "destructive",
		DefaultValue: false,
	},
}

type RiskSwitchService struct {
	db         *gorm.DB
	mu         sync.RWMutex
	cache      map[RiskSwitch]bool
	cacheValid bool
}

var (
	globalRiskSwitchService *RiskSwitchService
	riskSwitchOnce          sync.Once
)

func InitRiskSwitchService(db *gorm.DB) *RiskSwitchService {
	riskSwitchOnce.Do(func() {
		globalRiskSwitchService = &RiskSwitchService{
			db:    db,
			cache: make(map[RiskSwitch]bool),
		}
		globalRiskSwitchService.loadDefaults()
	})
	return globalRiskSwitchService
}

func GetRiskSwitchService() *RiskSwitchService {
	return globalRiskSwitchService
}

func (s *RiskSwitchService) loadDefaults() {
	for _, def := range RiskSwitchDefinitions {
		s.cache[def.Key] = def.DefaultValue
	}
	s.cacheValid = true
}

func (s *RiskSwitchService) loadUserCache(userID uint) map[RiskSwitch]bool {
	result := make(map[RiskSwitch]bool)
	for _, def := range RiskSwitchDefinitions {
		result[def.Key] = def.DefaultValue
	}

	settingsSvc := NewSettingsService(s.db)
	if v, err := settingsSvc.GetByKey(userID, "risk_switches"); err == nil && v != nil {
		if m, ok := v.(map[string]interface{}); ok {
			for k, val := range m {
				if b, ok := val.(bool); ok {
					result[RiskSwitch(k)] = b
				}
			}
		}
	}
	return result
}

func (s *RiskSwitchService) loadUserCacheRaw(userID uint) map[string]interface{} {
	existing := make(map[string]interface{})

	settingsSvc := NewSettingsService(s.db)
	if v, err := settingsSvc.GetByKey(userID, "risk_switches"); err == nil && v != nil {
		if m, ok := v.(map[string]interface{}); ok {
			for k, val := range m {
				existing[k] = val
			}
		}
	}
	return existing
}

func (s *RiskSwitchService) IsEnabled(switchKey RiskSwitch) bool {
	return s.IsEnabledForUser(switchKey, 0)
}

func (s *RiskSwitchService) IsEnabledForUser(switchKey RiskSwitch, userID uint) bool {
	if userID > 0 {
		userCache := s.loadUserCache(userID)
		return userCache[switchKey]
	}

	s.mu.RLock()
	cacheValid := s.cacheValid
	val, ok := s.cache[switchKey]
	s.mu.RUnlock()

	if !cacheValid {
		s.refresh()
		s.mu.RLock()
		val, ok = s.cache[switchKey]
		s.mu.RUnlock()
		return ok && val
	}

	return ok && val
}

func (s *RiskSwitchService) IsDisabled(switchKey RiskSwitch) bool {
	return !s.IsEnabled(switchKey)
}

func (s *RiskSwitchService) GetAll(userID uint) map[RiskSwitch]bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[RiskSwitch]bool)
	for k, v := range s.cache {
		result[k] = v
	}
	return result
}

func (s *RiskSwitchService) GetAllWithMeta(userID uint) []map[string]interface{} {
	userCache := s.loadUserCache(userID)

	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]map[string]interface{}, 0, len(RiskSwitchDefinitions))
	for _, def := range RiskSwitchDefinitions {
		enabled := def.DefaultValue
		if v, ok := userCache[def.Key]; ok {
			enabled = v
		}
		result = append(result, map[string]interface{}{
			"key":           string(def.Key),
			"label":         def.Label,
			"description":   def.Description,
			"category":      def.Category,
			"default_value": def.DefaultValue,
			"enabled":       enabled,
		})
	}
	return result
}

func (s *RiskSwitchService) Set(switchKey RiskSwitch, enabled bool, userID uint) error {
	settingsSvc := NewSettingsService(s.db)

	existing := s.loadUserCacheRaw(userID)
	existing[string(switchKey)] = enabled

	return settingsSvc.SetByKey(userID, "risk_switches", existing)
}

func (s *RiskSwitchService) BatchSet(switches map[RiskSwitch]bool, userID uint) error {
	settingsSvc := NewSettingsService(s.db)

	existing := s.loadUserCacheRaw(userID)
	for k, v := range switches {
		existing[string(k)] = v
	}

	return settingsSvc.SetByKey(userID, "risk_switches", existing)
}

func (s *RiskSwitchService) refresh() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.loadDefaults()

	var setting models.Setting
	if err := s.db.Where("key = ?", "risk_switches").First(&setting).Error; err != nil {
		return
	}

	var stored map[string]bool
	if err := json.Unmarshal([]byte(setting.Value), &stored); err != nil {
		return
	}

	for k, v := range stored {
		s.cache[RiskSwitch(k)] = v
	}
	s.cacheValid = true
}

func (s *RiskSwitchService) InvalidateCache() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cacheValid = false
}

func CheckRiskSwitch(switchKey RiskSwitch) error {
	svc := GetRiskSwitchService()
	if svc == nil {
		return nil
	}
	if svc.IsDisabled(switchKey) {
		return &RiskSwitchError{SwitchKey: switchKey}
	}
	return nil
}

type RiskSwitchError struct {
	SwitchKey RiskSwitch
}

func (e *RiskSwitchError) Error() string {
	for _, def := range RiskSwitchDefinitions {
		if def.Key == e.SwitchKey {
			return "operation blocked by risk switch: " + def.Label + " (" + string(e.SwitchKey) + ") is disabled"
		}
	}
	return "operation blocked by risk switch: " + string(e.SwitchKey) + " is disabled"
}
