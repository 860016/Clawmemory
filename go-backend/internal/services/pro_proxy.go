package services

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"clawmemory/internal/config"
	"clawmemory/internal/models"

	"gorm.io/gorm"
)

type ProProxy struct {
	db       *gorm.DB
	cfg      *config.Config
	client   *http.Client
}

func NewProProxy(db *gorm.DB, cfg *config.Config) *ProProxy {
	return &ProProxy{
		db:     db,
		cfg:    cfg,
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

func (p *ProProxy) getLicenseKey() string {
	var license models.License
	if err := p.db.Where("status = ?", "active").First(&license).Error; err != nil {
		return ""
	}
	return license.LicenseKey
}

func (p *ProProxy) IsPro() bool {
	return p.getLicenseKey() != ""
}

func (p *ProProxy) proxyRequest(path string, body interface{}) (map[string]interface{}, error) {
	licenseKey := p.getLicenseKey()
	if licenseKey == "" {
		return nil, ErrProRequired
	}

	var bodyReader io.Reader
	if body != nil {
		payload := map[string]interface{}{
			"license_key": licenseKey,
		}
		if m, ok := body.(map[string]interface{}); ok {
			for k, v := range m {
				payload[k] = v
			}
		}
		b, _ := json.Marshal(payload)
		bodyReader = strings.NewReader(string(b))
	} else {
		b, _ := json.Marshal(map[string]string{"license_key": licenseKey})
		bodyReader = strings.NewReader(string(b))
	}

	url := strings.TrimRight(p.cfg.LicenseServerURL, "/") + "/api/v1/pro/" + path
	resp, err := p.client.Post(url, "application/json", bodyReader)
	if err != nil {
		return nil, ErrProServerUnreachable
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusForbidden {
		return nil, ErrProNotAuthorized
	}

	if resp.StatusCode != http.StatusOK {
		if msg, ok := result["error"].(string); ok {
			return nil, &ProError{Message: msg, Code: resp.StatusCode}
		}
		return nil, ErrProServerUnreachable
	}

	return result, nil
}

func (p *ProProxy) GetDecayStats(memories []map[string]interface{}) (map[string]interface{}, error) {
	return p.proxyRequest("decay/stats", map[string]interface{}{
		"memories": memories,
	})
}

func (p *ProProxy) ApplyDecay(memories []map[string]interface{}) (map[string]interface{}, error) {
	return p.proxyRequest("decay/apply", map[string]interface{}{
		"memories": memories,
	})
}

func (p *ProProxy) ReinforceMemory(memoryId uint, memory map[string]interface{}) (map[string]interface{}, error) {
	return p.proxyRequest("reinforce", map[string]interface{}{
		"memory_id": memoryId,
		"memory":    memory,
	})
}

func (p *ProProxy) GetPruneSuggestions(memories []map[string]interface{}) (map[string]interface{}, error) {
	return p.proxyRequest("prune-suggest", map[string]interface{}{
		"memories": memories,
	})
}

func (p *ProProxy) ScanConflicts(memories []map[string]interface{}) (map[string]interface{}, error) {
	return p.proxyRequest("conflicts/scan", map[string]interface{}{
		"memories": memories,
	})
}

func (p *ProProxy) ResolveConflict(index int, strategy string) (map[string]interface{}, error) {
	return p.proxyRequest("conflicts/resolve", map[string]interface{}{
		"index":    index,
		"strategy": strategy,
	})
}

func (p *ProProxy) RouteModel(text string, contextLength int) (map[string]interface{}, error) {
	return p.proxyRequest("token/route", map[string]interface{}{
		"text":            text,
		"context_length":  contextLength,
	})
}

func (p *ProProxy) GetTokenStats() (map[string]interface{}, error) {
	return p.proxyRequest("token/stats", nil)
}

func (p *ProProxy) AIExtract(text string, memoryIds []uint) (map[string]interface{}, error) {
	return p.proxyRequest("ai/extract", map[string]interface{}{
		"text":        text,
		"memory_ids":  memoryIds,
	})
}

func (p *ProProxy) AutoGraph(overwrite bool) (map[string]interface{}, error) {
	return p.proxyRequest("auto-graph", map[string]interface{}{
		"overwrite": overwrite,
	})
}

func (p *ProProxy) GetBackupSchedule() (map[string]interface{}, error) {
	return p.proxyRequest("backup/schedule", nil)
}

func (p *ProProxy) SetBackupSchedule(enabled bool, intervalHours int) (map[string]interface{}, error) {
	return p.proxyRequest("backup/schedule", map[string]interface{}{
		"enabled":        enabled,
		"interval_hours": intervalHours,
	})
}

func (p *ProProxy) CompressPreview(memories []map[string]interface{}, level string) (map[string]interface{}, error) {
	return p.proxyRequest("compress/preview", map[string]interface{}{
		"memories": memories,
		"level":    level,
	})
}

func (p *ProProxy) CompressApply(memories []map[string]interface{}, level string, options map[string]interface{}) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"memories": memories,
		"level":    level,
	}
	for k, v := range options {
		payload[k] = v
	}
	return p.proxyRequest("compress/apply", payload)
}

func (p *ProProxy) GetCompressConfig() (map[string]interface{}, error) {
	return p.proxyRequest("compress/config", nil)
}

func (p *ProProxy) SetCompressConfig(config map[string]interface{}) (map[string]interface{}, error) {
	return p.proxyRequest("compress/config", config)
}

func (p *ProProxy) GetEvolutionInsights() (map[string]interface{}, error) {
	return p.proxyRequest("evolution/insights", nil)
}

func (p *ProProxy) DiscoverRelations(memories []map[string]interface{}) (map[string]interface{}, error) {
	return p.proxyRequest("evolution/discover", map[string]interface{}{
		"memories": memories,
	})
}

func (p *ProProxy) InferChains() (map[string]interface{}, error) {
	return p.proxyRequest("evolution/infer", nil)
}

func (p *ProProxy) GetImportanceAdjustments(memories []map[string]interface{}) (map[string]interface{}, error) {
	return p.proxyRequest("evolution/importance", map[string]interface{}{
		"memories": memories,
	})
}

func (p *ProProxy) PrefetchMemories(context string) (map[string]interface{}, error) {
	return p.proxyRequest("evolution/prefetch", map[string]interface{}{
		"context": context,
	})
}

var (
	ErrProRequired         = &ProError{Message: "Pro license required", Code: 403}
	ErrProNotAuthorized    = &ProError{Message: "Pro feature not authorized", Code: 403}
	ErrProServerUnreachable = &ProError{Message: "Pro server unreachable", Code: 503}
)

type ProError struct {
	Message string
	Code    int
}

func (e *ProError) Error() string {
	return e.Message
}
