package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"clawmemory/internal/models"
	"clawmemory/internal/services"

	"gorm.io/gorm"
)

type AIRouter struct {
	db            *gorm.DB
	mu            sync.RWMutex
	providerCache map[uint]cachedProvider
	usageMu       sync.Mutex
	usageBuffer   map[uint]*usageEntry
	lastFlush     time.Time
	stopCh        chan struct{}
}

type usageEntry struct {
	Calls  int
	Tokens int
}

type cachedProvider struct {
	provider  Provider
	expiredAt time.Time
}

func NewAIRouter(db *gorm.DB) *AIRouter {
	r := &AIRouter{
		db:            db,
		providerCache: make(map[uint]cachedProvider),
		usageBuffer:   make(map[uint]*usageEntry),
		lastFlush:     time.Now(),
		stopCh:        make(chan struct{}),
	}
	go r.flushLoop()
	return r
}

func (r *AIRouter) Stop() {
	close(r.stopCh)
}

func (r *AIRouter) flushLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			r.flushUsage()
			r.cleanExpiredCache()
		case <-r.stopCh:
			r.flushUsage()
			return
		}
	}
}

func (r *AIRouter) cleanExpiredCache() {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	for k, v := range r.providerCache {
		if now.After(v.expiredAt) {
			delete(r.providerCache, k)
		}
	}
}

func (r *AIRouter) flushUsage() {
	r.usageMu.Lock()
	if len(r.usageBuffer) == 0 {
		r.usageMu.Unlock()
		return
	}
	buffer := r.usageBuffer
	r.usageBuffer = make(map[uint]*usageEntry)
	r.lastFlush = time.Now()
	r.usageMu.Unlock()

	settingsSvc := services.NewSettingsService(r.db)
	for userID, entry := range buffer {
		var totalCalls float64
		if v, _ := settingsSvc.GetByKey(userID, "ai_total_calls"); v != nil {
			if f, ok := v.(float64); ok {
				totalCalls = f
			}
		}
		settingsSvc.SetByKey(userID, "ai_total_calls", totalCalls+float64(entry.Calls))

		var totalTokens float64
		if v, _ := settingsSvc.GetByKey(userID, "ai_total_tokens"); v != nil {
			if f, ok := v.(float64); ok {
				totalTokens = f
			}
		}
		settingsSvc.SetByKey(userID, "ai_total_tokens", totalTokens+float64(entry.Tokens))
	}
}

func (r *AIRouter) Chat(ctx context.Context, userID uint, messages []Message, opts ChatOptions) (*ChatResponse, error) {
	provider, err := r.getProvider(userID)
	if err != nil {
		return nil, err
	}

	if err := provider.Validate(); err != nil {
		return nil, fmt.Errorf("AI provider config error: %w (please configure an AI provider in Settings > Reasoning)", err)
	}

	return provider.Chat(ctx, messages, opts)
}

func (r *AIRouter) Embed(ctx context.Context, userID uint, texts []string) (*EmbeddingResponse, error) {
	provider, err := r.getProvider(userID)
	if err != nil {
		return nil, err
	}

	return provider.Embed(ctx, texts)
}

func (r *AIRouter) getProvider(userID uint) (Provider, error) {
	r.mu.RLock()
	if cached, ok := r.providerCache[userID]; ok && time.Now().Before(cached.expiredAt) {
		r.mu.RUnlock()
		return cached.provider, nil
	}
	r.mu.RUnlock()

	provider, err := r.loadProvider(userID)
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

func (r *AIRouter) loadProvider(userID uint) (Provider, error) {
	provider, err := r.loadFromReasoningConfig(userID)
	if err == nil && provider != nil {
		return provider, nil
	}

	userProvider, err := r.loadUserProvider(userID)
	if err != nil {
		return nil, fmt.Errorf("no AI provider configured. Please configure an AI provider in Settings > Reasoning")
	}

	return userProvider, nil
}

func (r *AIRouter) loadFromReasoningConfig(userID uint) (Provider, error) {
	var config models.ReasoningConfig
	if err := r.db.Where("user_id = ? AND enabled = ?", userID, true).First(&config).Error; err != nil {
		return nil, err
	}

	if config.APIKey == "" && config.Provider != "ollama" {
		return nil, fmt.Errorf("reasoning config has no API key")
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		switch config.Provider {
		case "openai":
			baseURL = "https://api.openai.com/v1"
		case "anthropic":
			baseURL = "https://api.anthropic.com/v1"
		case "openrouter":
			baseURL = "https://openrouter.ai/api/v1"
		case "ollama":
			baseURL = "http://localhost:11434/v1"
		case "deepseek":
			baseURL = "https://api.deepseek.com/v1"
		default:
			baseURL = ""
		}
	}

	model := config.Model
	if model == "" {
		model = defaultModel(config.Provider)
	}

	cfg := ProviderConfig{
		ID:      config.Provider,
		Type:    config.Provider,
		APIKey:  config.APIKey,
		BaseURL: baseURL,
		Model:   model,
	}

	return NewOpenAICompatibleProvider(cfg), nil
}

func (r *AIRouter) loadUserProvider(userID uint) (Provider, error) {
	settingsSvc := services.NewSettingsService(r.db)

	providerID, _ := settingsSvc.GetByKey(userID, "ai_provider_id")
	if providerID == nil {
		return nil, fmt.Errorf("no AI provider configured")
	}

	pid, ok := providerID.(string)
	if !ok || pid == "" {
		return nil, fmt.Errorf("no AI provider configured")
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

func (r *AIRouter) GetCurrentUserConfig(userID uint) map[string]interface{} {
	result := map[string]interface{}{
		"available_providers": AllProviders,
	}

	reasoningProvider, reasoningModel := r.getReasoningConfigInfo(userID)
	if reasoningProvider != "" {
		result["provider_id"] = reasoningProvider
		result["model"] = reasoningModel
		result["provider_source"] = "reasoning"
	}

	settingsSvc := services.NewSettingsService(r.db)

	if v, _ := settingsSvc.GetByKey(userID, "ai_provider_id"); v != nil {
		if s, ok := v.(string); ok && s != "" {
			result["provider_id"] = s
			result["provider_source"] = "pro_config"
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

func (r *AIRouter) getReasoningConfigInfo(userID uint) (provider string, model string) {
	var config models.ReasoningConfig
	if err := r.db.Where("user_id = ? AND enabled = ?", userID, true).First(&config).Error; err != nil {
		return "", ""
	}
	return config.Provider, config.Model
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

func (r *AIRouter) TestConnection(userID uint) (map[string]interface{}, error) {
	provider, err := r.getProvider(userID)
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
	r.usageMu.Lock()
	entry, ok := r.usageBuffer[userID]
	if !ok {
		entry = &usageEntry{}
		r.usageBuffer[userID] = entry
	}
	entry.Calls++
	entry.Tokens += tokensIn + tokensOut
	r.usageMu.Unlock()

	if time.Since(r.lastFlush) > 60*time.Second {
		r.flushUsage()
	}
}

func defaultModel(providerID string) string {
	switch providerID {
	case "deepseek":
		return "deepseek-chat"
	case "openai":
		return "gpt-4o-mini"
	case "ollama":
		return "llama3"
	case "openrouter":
		return "openai/gpt-4o-mini"
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
