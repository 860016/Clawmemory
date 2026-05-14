package services

import (
	"encoding/json"
	"sync"

	"clawmemory/internal/models"

	"gorm.io/gorm"
)

type RiskSwitch string

const (
	RiskShareAutoApprove       RiskSwitch = "risk_share_auto_approve"
	RiskCrossAgentWrite        RiskSwitch = "risk_cross_agent_write"
	RiskMemoryVisibilityChange RiskSwitch = "risk_memory_visibility_change"
	RiskAutoImportMemories     RiskSwitch = "risk_auto_import_memories"
	RiskAgentMemoryAccess      RiskSwitch = "risk_agent_memory_access"
	RiskBulkDelete             RiskSwitch = "risk_bulk_delete"
	RiskDecayAutoApply         RiskSwitch = "risk_decay_auto_apply"
	RiskCompressAutoApply      RiskSwitch = "risk_compress_auto_apply"
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
		Key:          RiskShareAutoApprove,
		Label:        "自动审批记忆共享",
		Description:  "允许自动审批来自其他 Agent 的记忆共享请求，无需人工确认",
		Category:     "sharing",
		DefaultValue: false,
	},
	{
		Key:          RiskCrossAgentWrite,
		Label:        "跨 Agent 写入记忆",
		Description:  "允许一个 Agent 向另一个 Agent 的私有记忆写入数据",
		Category:     "sharing",
		DefaultValue: false,
	},
	{
		Key:          RiskMemoryVisibilityChange,
		Label:        "记忆可见性变更",
		Description:  "允许将私有记忆变更为共享或公开可见性",
		Category:     "visibility",
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
		Key:          RiskAgentMemoryAccess,
		Label:        "Agent 跨用户访问记忆",
		Description:  "允许 Agent 通过 API Key 访问其他用户的共享记忆",
		Category:     "access",
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
		Key:          RiskDecayAutoApply,
		Label:        "自动衰减应用",
		Description:  "允许系统自动应用记忆衰减策略，可能自动归档或删除低重要性记忆",
		Category:     "destructive",
		DefaultValue: false,
	},
	{
		Key:          RiskCompressAutoApply,
		Label:        "自动压缩应用",
		Description:  "允许系统自动压缩长记忆内容，可能导致信息损失",
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

func (s *RiskSwitchService) IsEnabled(switchKey RiskSwitch) bool {
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
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]map[string]interface{}, 0, len(RiskSwitchDefinitions))
	for _, def := range RiskSwitchDefinitions {
		enabled := s.cache[def.Key]
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
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cache[switchKey] = enabled

	settingsSvc := NewSettingsService(s.db)
	riskSettings := map[string]interface{}{
		string(switchKey): enabled,
	}
	return settingsSvc.SetByKey(userID, "risk_switches", riskSettings)
}

func (s *RiskSwitchService) BatchSet(switches map[RiskSwitch]bool, userID uint) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for k, v := range switches {
		s.cache[k] = v
	}

	settingsSvc := NewSettingsService(s.db)
	data := make(map[string]interface{})
	for k, v := range switches {
		data[string(k)] = v
	}
	return settingsSvc.SetByKey(userID, "risk_switches", data)
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
