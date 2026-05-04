package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"clawmemory/internal/services"

	"gorm.io/gorm"
)

type AIRouter struct {
	db            *gorm.DB
	freeProvider  Provider
	mu            sync.RWMutex
	providerCache map[uint]cachedProvider
}

type cachedProvider struct {
	provider  Provider
	expiredAt time.Time
}

func NewAIRouter(db *gorm.DB) *AIRouter {
	return &AIRouter{
		db:            db,
		freeProvider:  NewNVIDIAFreeProvider(),
		providerCache: make(map[uint]cachedProvider),
	}
}

func (r *AIRouter) Chat(ctx context.Context, userID uint, isPro bool, messages []Message, opts ChatOptions) (*ChatResponse, error) {
	provider, err := r.getProvider(userID, isPro)
	if err != nil {
		return nil, err
	}

	if err := provider.Validate(); err != nil {
		if isPro {
			return nil, fmt.Errorf("AI provider config error: %w (please check your AI settings)", err)
		}
		return nil, fmt.Errorf("free AI provider unavailable: %w", err)
	}

	return provider.Chat(ctx, messages, opts)
}

func (r *AIRouter) Embed(ctx context.Context, userID uint, isPro bool, texts []string) (*EmbeddingResponse, error) {
	provider, err := r.getProvider(userID, isPro)
	if err != nil {
		return nil, err
	}

	return provider.Embed(ctx, texts)
}

func (r *AIRouter) getProvider(userID uint, isPro bool) (Provider, error) {
	if !isPro {
		return r.freeProvider, nil
	}

	r.mu.RLock()
	if cached, ok := r.providerCache[userID]; ok && time.Now().Before(cached.expiredAt) {
		r.mu.RUnlock()
		return cached.provider, nil
	}
	r.mu.RUnlock()

	provider, err := r.loadProProvider(userID)
	if err != nil {
		return nil, err
	}

	r.mu.Lock()
	r.providerCache[userID] = cachedProvider{
		provider:  provider,
		expiredAt: time.Now().Add(5 * time.Minute),
	}
	r.mu.Unlock()

	return provider, nil
}

func (r *AIRouter) loadProProvider(userID uint) (Provider, error) {
	settingsSvc := services.NewSettingsService(r.db)

	providerID, _ := settingsSvc.GetByKey(userID, "ai_provider_id")
	if providerID == nil {
		return r.freeProvider, nil
	}

	pid, ok := providerID.(string)
	if !ok || pid == "" {
		return r.freeProvider, nil
	}

	if pid == "nvidia-nim" {
		return NewNVIDIAFreeProvider(), nil
	}

	cfg := ProviderConfig{ID: pid}

	if v, _ := settingsSvc.GetByKey(userID, "ai_provider_type"); v != nil {
		if s, ok := v.(string); ok {
			cfg.Type = s
		}
	}
	if cfg.Type == "" {
		cfg.Type = pid
	}

	if v, _ := settingsSvc.GetByKey(userID, "ai_api_key"); v != nil {
		if s, ok := v.(string); ok {
			cfg.APIKey = s
		}
	}

	if v, _ := settingsSvc.GetByKey(userID, "ai_base_url"); v != nil {
		if s, ok := v.(string); ok {
			cfg.BaseURL = s
		}
	}

	if v, _ := settingsSvc.GetByKey(userID, "ai_model"); v != nil {
		if s, ok := v.(string); ok {
			cfg.Model = s
		}
	}
	if cfg.Model == "" {
		cfg.Model = defaultModel(pid)
	}

	if v, _ := settingsSvc.GetByKey(userID, "ai_embed_model"); v != nil {
		if s, ok := v.(string); ok {
			cfg.EmbedModel = s
		}
	}

	return NewOpenAICompatibleProvider(cfg), nil
}

func (r *AIRouter) InvalidateUserCache(userID uint) {
	r.mu.Lock()
	delete(r.providerCache, userID)
	r.mu.Unlock()
}

func (r *AIRouter) GetCurrentUserConfig(userID uint, isPro bool) map[string]interface{} {
	result := map[string]interface{}{
		"provider_id":         "nvidia-nim",
		"model":               "nvidia/llama-3.1-nemotron-70b-instruct",
		"is_pro":              isPro,
		"available_providers": AllProviders,
	}

	if !isPro {
		result["free_only"] = true
		result["available_providers"] = func() []ProviderInfo {
			var free []ProviderInfo
			for _, p := range AllProviders {
				if p.Free {
					free = append(free, p)
				}
			}
			return free
		}()
		return result
	}

	settingsSvc := services.NewSettingsService(r.db)

	if v, _ := settingsSvc.GetByKey(userID, "ai_provider_id"); v != nil {
		if s, ok := v.(string); ok {
			result["provider_id"] = s
		}
	}
	if v, _ := settingsSvc.GetByKey(userID, "ai_model"); v != nil {
		if s, ok := v.(string); ok {
			result["model"] = s
		}
	}
	if v, _ := settingsSvc.GetByKey(userID, "ai_base_url"); v != nil {
		if s, ok := v.(string); ok {
			result["base_url"] = s
		}
	}
	if v, _ := settingsSvc.GetByKey(userID, "ai_embed_model"); v != nil {
		if s, ok := v.(string); ok {
			result["embed_model"] = s
		}
	}

	return result
}

func (r *AIRouter) UpdateProConfig(userID uint, data map[string]interface{}) error {
	settingsSvc := services.NewSettingsService(r.db)

	if providerID, ok := data["provider_id"].(string); ok {
		info := GetProviderInfo(providerID)
		if info == nil {
			return fmt.Errorf("unknown provider: %s", providerID)
		}
		settingsSvc.SetByKey(userID, "ai_provider_id", providerID)
		settingsSvc.SetByKey(userID, "ai_provider_type", info.Type)
	}

	if apiKey, ok := data["api_key"].(string); ok {
		settingsSvc.SetByKey(userID, "ai_api_key", apiKey)
	}
	if baseURL, ok := data["base_url"].(string); ok {
		settingsSvc.SetByKey(userID, "ai_base_url", baseURL)
	}
	if model, ok := data["model"].(string); ok {
		settingsSvc.SetByKey(userID, "ai_model", model)
	}
	if embedModel, ok := data["embed_model"].(string); ok {
		settingsSvc.SetByKey(userID, "ai_embed_model", embedModel)
	}

	r.InvalidateUserCache(userID)
	return nil
}

func (r *AIRouter) TestConnection(userID uint, isPro bool) (map[string]interface{}, error) {
	provider, err := r.getProvider(userID, isPro)
	if err != nil {
		return nil, err
	}

	if err := provider.Validate(); err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := provider.Chat(ctx, []Message{
		{Role: "user", Content: "Hello, respond with exactly: OK"},
	}, ChatOptions{MaxTokens: 10, Temperature: 0})

	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}, nil
	}

	return map[string]interface{}{
		"success":    true,
		"model":      resp.Model,
		"provider":   resp.ProviderID,
		"latency_ms": resp.Latency.Milliseconds(),
		"tokens_in":  resp.TokensIn,
		"tokens_out": resp.TokensOut,
	}, nil
}

func (r *AIRouter) GetUsageStats(userID uint) map[string]interface{} {
	settingsSvc := services.NewSettingsService(r.db)

	var totalCalls int64
	if v, _ := settingsSvc.GetByKey(userID, "ai_total_calls"); v != nil {
		if f, ok := v.(float64); ok {
			totalCalls = int64(f)
		}
	}

	var totalTokens int64
	if v, _ := settingsSvc.GetByKey(userID, "ai_total_tokens"); v != nil {
		if f, ok := v.(float64); ok {
			totalTokens = int64(f)
		}
	}

	return map[string]interface{}{
		"total_calls":  totalCalls,
		"total_tokens": totalTokens,
	}
}

func (r *AIRouter) IncrementUsage(userID uint, tokensIn, tokensOut int) {
	settingsSvc := services.NewSettingsService(r.db)

	var totalCalls float64
	if v, _ := settingsSvc.GetByKey(userID, "ai_total_calls"); v != nil {
		if f, ok := v.(float64); ok {
			totalCalls = f
		}
	}
	settingsSvc.SetByKey(userID, "ai_total_calls", totalCalls+1)

	var totalTokens float64
	if v, _ := settingsSvc.GetByKey(userID, "ai_total_tokens"); v != nil {
		if f, ok := v.(float64); ok {
			totalTokens = f
		}
	}
	settingsSvc.SetByKey(userID, "ai_total_tokens", totalTokens+float64(tokensIn+tokensOut))
}

func defaultModel(providerID string) string {
	switch providerID {
	case "nvidia-nim":
		return "nvidia/llama-3.1-nemotron-70b-instruct"
	case "deepseek":
		return "deepseek-chat"
	case "openai":
		return "gpt-4o-mini"
	default:
		return "gpt-4o-mini"
	}
}

func ParseAIResponse(content string, target interface{}) error {
	cleaned := content
	if idx := indexOf(cleaned, "```json"); idx >= 0 {
		cleaned = cleaned[idx+7:]
		if endIdx := indexOf(cleaned, "```"); endIdx >= 0 {
			cleaned = cleaned[:endIdx]
		}
	} else if idx := indexOf(cleaned, "```"); idx >= 0 {
		cleaned = cleaned[idx+3:]
		if endIdx := indexOf(cleaned, "```"); endIdx >= 0 {
			cleaned = cleaned[:endIdx]
		}
	}

	cleaned = trimBrackets(cleaned)

	return json.Unmarshal([]byte(cleaned), target)
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func trimBrackets(s string) string {
	start := 0
	end := len(s)
	for start < len(s) && (s[start] == ' ' || s[start] == '\n' || s[start] == '\r' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\n' || s[end-1] == '\r' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}
