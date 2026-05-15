package ai

import (
	"context"
	"time"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatOptions struct {
	Temperature float64
	MaxTokens   int
	Model       string
}

type ChatResponse struct {
	Content    string
	TokensIn   int
	TokensOut  int
	Model      string
	ProviderID string
	Latency    time.Duration
}

type EmbeddingResponse struct {
	Vectors [][]float64
	Model   string
	Tokens  int
}

type ProviderInfo struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Models      []string `json:"models"`
	Free        bool     `json:"free"`
	ProOnly     bool     `json:"pro_only"`
	Description string   `json:"description"`
}

type Provider interface {
	ID() string
	Chat(ctx context.Context, messages []Message, opts ChatOptions) (*ChatResponse, error)
	Embed(ctx context.Context, texts []string) (*EmbeddingResponse, error)
	Info() ProviderInfo
	Validate() error
}

type ProviderConfig struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	APIKey     string `json:"api_key,omitempty"`
	BaseURL    string `json:"base_url,omitempty"`
	Model      string `json:"model"`
	EmbedModel string `json:"embed_model,omitempty"`
}

var AllProviders = []ProviderInfo{
	{
		ID:          "deepseek",
		Name:        "DeepSeek",
		Type:        "openai_compatible",
		Models:      []string{"deepseek-chat", "deepseek-reasoner"},
		Free:        false,
		ProOnly:     true,
		Description: "DeepSeek V3 / R1, excellent Chinese support",
	},
	{
		ID:          "openai",
		Name:        "OpenAI",
		Type:        "openai_compatible",
		Models:      []string{"gpt-4o-mini", "gpt-4o", "gpt-4-turbo"},
		Free:        false,
		ProOnly:     true,
		Description: "OpenAI GPT models",
	},
	{
		ID:          "ollama",
		Name:        "Ollama (Local)",
		Type:        "openai_compatible",
		Models:      []string{"llama3", "mistral", "qwen2"},
		Free:        true,
		ProOnly:     false,
		Description: "Local Ollama models, free and private",
	},
	{
		ID:          "custom",
		Name:        "Custom (OpenAI Compatible)",
		Type:        "openai_compatible",
		Models:      []string{},
		Free:        false,
		ProOnly:     true,
		Description: "Any OpenAI-compatible API endpoint",
	},
}

func GetProviderInfo(id string) *ProviderInfo {
	for _, p := range AllProviders {
		if p.ID == id {
			return &p
		}
	}
	return nil
}

func FreeProviderIDs() []string {
	var ids []string
	for _, p := range AllProviders {
		if p.Free {
			ids = append(ids, p.ID)
		}
	}
	return ids
}

func ProProviderIDs() []string {
	var ids []string
	for _, p := range AllProviders {
		if p.ProOnly {
			ids = append(ids, p.ID)
		}
	}
	return ids
}

func DefaultChatOptions() ChatOptions {
	return ChatOptions{
		Temperature: 0.7,
		MaxTokens:   4096,
	}
}
