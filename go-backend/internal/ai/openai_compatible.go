package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type OpenAICompatibleProvider struct {
	providerID string
	name       string
	baseURL    string
	apiKey     string
	model      string
	embedModel string
	free       bool
	proOnly    bool
	httpClient *http.Client
}

func NewOpenAICompatibleProvider(cfg ProviderConfig) *OpenAICompatibleProvider {
	baseURL := strings.TrimRight(cfg.BaseURL, "/")
	if baseURL == "" {
		baseURL = defaultBaseURL(cfg.Type)
	}

	return &OpenAICompatibleProvider{
		providerID: cfg.ID,
		name:       providerName(cfg.ID),
		baseURL:    baseURL,
		apiKey:     cfg.APIKey,
		model:      cfg.Model,
		embedModel: cfg.EmbedModel,
		free:       cfg.ID == "nvidia-nim",
		proOnly:    cfg.ID != "nvidia-nim",
		httpClient: &http.Client{Timeout: 120 * time.Second},
	}
}

func NewNVIDIAFreeProvider() *OpenAICompatibleProvider {
	return &OpenAICompatibleProvider{
		providerID: "nvidia-nim",
		name:       "NVIDIA NIM (Free)",
		baseURL:    "https://integrate.api.nvidia.com/v1",
		apiKey:     "",
		model:      "nvidia/llama-3.1-nemotron-70b-instruct",
		embedModel: "nvidia/llama-3.1-nemotron-70b-instruct",
		free:       true,
		proOnly:    false,
		httpClient: &http.Client{Timeout: 120 * time.Second},
	}
}

func (p *OpenAICompatibleProvider) ID() string { return p.providerID }
func (p *OpenAICompatibleProvider) Info() ProviderInfo {
	return ProviderInfo{
		ID:          p.providerID,
		Name:        p.name,
		Type:        "openai_compatible",
		Free:        p.free,
		ProOnly:     p.proOnly,
		Description: fmt.Sprintf("OpenAI-compatible provider: %s", p.name),
	}
}

func (p *OpenAICompatibleProvider) Validate() error {
	if p.baseURL == "" {
		return fmt.Errorf("base_url is required")
	}
	if !p.free && p.apiKey == "" {
		return fmt.Errorf("api_key is required for %s", p.name)
	}
	return nil
}

func (p *OpenAICompatibleProvider) Chat(ctx context.Context, messages []Message, opts ChatOptions) (*ChatResponse, error) {
	start := time.Now()

	model := p.model
	if opts.Model != "" {
		model = opts.Model
	}

	reqBody := map[string]interface{}{
		"model":      model,
		"messages":   messages,
		"temperature": opts.Temperature,
	}
	if opts.MaxTokens > 0 {
		reqBody["max_tokens"] = opts.MaxTokens
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if p.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.apiKey)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, truncate(string(body), 500))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
		Model string `json:"model"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no response from AI model")
	}

	return &ChatResponse{
		Content:    result.Choices[0].Message.Content,
		TokensIn:   result.Usage.PromptTokens,
		TokensOut:  result.Usage.CompletionTokens,
		Model:      result.Model,
		ProviderID: p.providerID,
		Latency:    time.Since(start),
	}, nil
}

func (p *OpenAICompatibleProvider) Embed(ctx context.Context, texts []string) (*EmbeddingResponse, error) {
	if p.embedModel == "" {
		return nil, fmt.Errorf("embedding model not configured for %s", p.name)
	}

	reqBody := map[string]interface{}{
		"model": p.embedModel,
		"input": texts,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/embeddings", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if p.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.apiKey)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, truncate(string(body), 500))
	}

	var result struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
		Model string `json:"model"`
		Usage struct {
			TotalTokens int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	vectors := make([][]float64, len(result.Data))
	for i, d := range result.Data {
		vectors[i] = d.Embedding
	}

	return &EmbeddingResponse{
		Vectors: vectors,
		Model:   result.Model,
		Tokens:  result.Usage.TotalTokens,
	}, nil
}

func defaultBaseURL(providerType string) string {
	switch providerType {
	case "nvidia-nim":
		return "https://integrate.api.nvidia.com/v1"
	case "deepseek":
		return "https://api.deepseek.com/v1"
	case "openai":
		return "https://api.openai.com/v1"
	default:
		return ""
	}
}

func providerName(id string) string {
	switch id {
	case "nvidia-nim":
		return "NVIDIA NIM (Free)"
	case "deepseek":
		return "DeepSeek"
	case "openai":
		return "OpenAI"
	case "custom":
		return "Custom Provider"
	default:
		return id
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
