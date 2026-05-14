package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ProAPIClient struct {
	baseURL    string
	licenseKey string
	httpClient *http.Client
}

func NewProAPIClient(baseURL, licenseKey string) *ProAPIClient {
	return &ProAPIClient{
		baseURL:    baseURL,
		licenseKey: licenseKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (c *ProAPIClient) doRequest(method, endpoint string, body interface{}) (map[string]interface{}, error) {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, c.baseURL+endpoint, reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.licenseKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("pro api request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("invalid pro api response")
	}

	if resp.StatusCode >= 400 {
		if errMsg, ok := result["error"].(string); ok {
			return nil, fmt.Errorf("%s", errMsg)
		}
		return nil, fmt.Errorf("pro api error: status %d", resp.StatusCode)
	}

	return result, nil
}

func (c *ProAPIClient) Decay(memories []map[string]interface{}) (map[string]interface{}, error) {
	return c.doRequest("POST", "/api/v1/pro/decay", map[string]interface{}{
		"memories": memories,
	})
}

func (c *ProAPIClient) Conflict(content1, content2 string) (map[string]interface{}, error) {
	return c.doRequest("POST", "/api/v1/pro/conflict", map[string]interface{}{
		"content1": content1,
		"content2": content2,
	})
}

func (c *ProAPIClient) TokenRoute(message string, contextLength int) (map[string]interface{}, error) {
	return c.doRequest("POST", "/api/v1/pro/token-route", map[string]interface{}{
		"message":        message,
		"context_length": contextLength,
	})
}

func (c *ProAPIClient) AIExtract(content string) (map[string]interface{}, error) {
	return c.doRequest("POST", "/api/v1/pro/ai/extract", map[string]interface{}{
		"content": content,
	})
}

func (c *ProAPIClient) Compress(content string, level string) (map[string]interface{}, error) {
	return c.doRequest("POST", "/api/v1/pro/compress", map[string]interface{}{
		"content": content,
		"level":   level,
	})
}

func (c *ProAPIClient) EvolutionDiscover(memories []map[string]interface{}) (map[string]interface{}, error) {
	return c.doRequest("POST", "/api/v1/pro/evolution/discover", map[string]interface{}{
		"memories": memories,
	})
}

func (c *ProAPIClient) EvolutionInfer(premise, conclusion string) (map[string]interface{}, error) {
	return c.doRequest("POST", "/api/v1/pro/evolution/infer", map[string]interface{}{
		"premise":    premise,
		"conclusion": conclusion,
	})
}

func (c *ProAPIClient) EvolutionImportance(memory map[string]interface{}) (map[string]interface{}, error) {
	return c.doRequest("POST", "/api/v1/pro/evolution/importance", memory)
}

func (c *ProAPIClient) EvolutionPrefetch(context string) (map[string]interface{}, error) {
	return c.doRequest("POST", "/api/v1/pro/evolution/prefetch", map[string]interface{}{
		"context": context,
	})
}
