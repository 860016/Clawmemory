package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ProAPIClient 云端 Pro API 客户端
type ProAPIClient struct {
	baseURL    string
	licenseKey string
	httpClient *http.Client
}

// NewProAPIClient 创建 Pro API 客户端
func NewProAPIClient(baseURL, licenseKey string) *ProAPIClient {
	return &ProAPIClient{
		baseURL:    baseURL,
		licenseKey: licenseKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// DecayRequest 衰减计算请求
type DecayRequest struct {
	Memories []struct {
		ID          string  `json:"id"`
		AgeHours    float64 `json:"age_hours"`
		Importance  float64 `json:"importance"`
		AccessCount int     `json:"access_count"`
	} `json:"memories"`
}

// DecayResponse 衰减计算响应
type DecayResponse struct {
	Results []struct {
		ID             string  `json:"id"`
		DecayScore     float64 `json:"decay_score"`
		Stage          string  `json:"stage"`
		Recommendation string  `json:"recommendation"`
	} `json:"results"`
	AlgorithmVersion string `json:"algorithm_version"`
}

// CalculateDecay 调用云端衰减计算
func (c *ProAPIClient) CalculateDecay(req DecayRequest) (*DecayResponse, error) {
	resp, err := c.post("/api/v1/pro/decay", req)
	if err != nil {
		return nil, err
	}

	var result DecayResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

// ConflictRequest 冲突检测请求
type ConflictRequest struct {
	Content1 string `json:"content1"`
	Content2 string `json:"content2"`
}

// ConflictResponse 冲突检测响应
type ConflictResponse struct {
	HasConflict bool     `json:"has_conflict"`
	Similarity  float64  `json:"similarity"`
	CommonWords []string `json:"common_words"`
	Suggestion  string   `json:"suggestion"`
}

// DetectConflict 调用云端冲突检测
func (c *ProAPIClient) DetectConflict(req ConflictRequest) (*ConflictResponse, error) {
	resp, err := c.post("/api/v1/pro/conflict", req)
	if err != nil {
		return nil, err
	}

	var result ConflictResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

// ConflictBatchRequest 批量冲突检测请求
type ConflictBatchRequest struct {
	Items     []struct {
		ID      string `json:"id"`
		Content string `json:"content"`
	} `json:"items"`
	Threshold float64 `json:"threshold,omitempty"`
}

// ConflictBatchResponse 批量冲突检测响应
type ConflictBatchResponse struct {
	Conflicts    []struct {
		Item1ID    string  `json:"item1_id"`
		Item2ID    string  `json:"item2_id"`
		Similarity float64 `json:"similarity"`
		HasConflict bool   `json:"has_conflict"`
	} `json:"conflicts"`
	TotalChecked int `json:"total_checked"`
}

// DetectConflictBatch 批量冲突检测
func (c *ProAPIClient) DetectConflictBatch(req ConflictBatchRequest) (*ConflictBatchResponse, error) {
	resp, err := c.post("/api/v1/pro/conflict/batch", req)
	if err != nil {
		return nil, err
	}

	var result ConflictBatchResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

// RouteRequest Token 路由请求
type RouteRequest struct {
	Text   string   `json:"text"`
	Models []string `json:"models,omitempty"`
}

// RouteResponse Token 路由响应
type RouteResponse struct {
	Model      string `json:"model"`
	Complexity struct {
		Score          int    `json:"score"`
		Complexity     string `json:"complexity"`
		Length         int    `json:"length"`
		SentenceCount  int    `json:"sentence_count"`
		TechnicalTerms int    `json:"technical_terms"`
	} `json:"complexity"`
	Reason string `json:"reason"`
}

// RouteModel 调用云端模型路由
func (c *ProAPIClient) RouteModel(req RouteRequest) (*RouteResponse, error) {
	resp, err := c.post("/api/v1/pro/route", req)
	if err != nil {
		return nil, err
	}

	var result RouteResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

// ExtractRequest 实体提取请求
type ExtractRequest struct {
	Text string `json:"text"`
}

// ExtractResponse 实体提取响应
type ExtractResponse struct {
	Entities []struct {
		Name       string  `json:"name"`
		Type       string  `json:"type"`
		Confidence float64 `json:"confidence"`
	} `json:"entities"`
	TextLength  int    `json:"text_length"`
	ExtractedAt string `json:"extracted_at"`
}

// ExtractEntities 调用云端实体提取
func (c *ProAPIClient) ExtractEntities(req ExtractRequest) (*ExtractResponse, error) {
	resp, err := c.post("/api/v1/pro/extract", req)
	if err != nil {
		return nil, err
	}

	var result ExtractResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

// GraphAnalyzeRequest 图谱分析请求
type GraphAnalyzeRequest struct {
	Nodes []map[string]interface{} `json:"nodes"`
	Edges []map[string]interface{} `json:"edges"`
}

// GraphAnalyzeResponse 图谱分析响应
type GraphAnalyzeResponse struct {
	Analysis struct {
		Density        float64 `json:"density"`
		IsolatedNodes  []string `json:"isolated_nodes"`
		IsolatedCount  int     `json:"isolated_count"`
		AvgDegree      float64 `json:"avg_degree"`
		Suggestion     string  `json:"suggestion"`
	} `json:"analysis"`
	NodeCount int `json:"node_count"`
	EdgeCount int `json:"edge_count"`
}

// AnalyzeGraph 调用云端图谱分析
func (c *ProAPIClient) AnalyzeGraph(req GraphAnalyzeRequest) (*GraphAnalyzeResponse, error) {
	resp, err := c.post("/api/v1/pro/graph/analyze", req)
	if err != nil {
		return nil, err
	}

	var result GraphAnalyzeResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

// UsageResponse 使用情况响应
type UsageResponse struct {
	LicenseKey       string `json:"license_key"`
	MonthlyRequests  int    `json:"monthly_requests"`
	MonthlyTokens    int    `json:"monthly_tokens"`
	Quota            int    `json:"quota"`
	ResetDate        string `json:"reset_date"`
}

// GetUsage 获取 Pro API 使用情况
func (c *ProAPIClient) GetUsage() (*UsageResponse, error) {
	resp, err := c.get("/api/v1/pro/usage")
	if err != nil {
		return nil, err
	}

	var result UsageResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

// HTTP 辅助方法

func (c *ProAPIClient) post(path string, body interface{}) ([]byte, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("encode request: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+path, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.licenseKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func (c *ProAPIClient) get(path string) ([]byte, error) {
	req, err := http.NewRequest("GET", c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.licenseKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}
