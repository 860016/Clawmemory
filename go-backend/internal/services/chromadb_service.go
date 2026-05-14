package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"clawmemory/internal/models"

	"gorm.io/gorm"
)

type ChromaDBService struct {
	baseURL    string
	db         *gorm.DB
	httpClient *http.Client
}

func NewChromaDBService(db *gorm.DB) *ChromaDBService {
	return &ChromaDBService{
		baseURL: "http://localhost:8000/api/v1",
		db:      db,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *ChromaDBService) getEmbeddingService() *EmbeddingService {
	return GetEmbeddingService()
}

func (s *ChromaDBService) IsAvailable() bool {
	resp, err := s.httpClient.Get(s.baseURL + "/heartbeat")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func (s *ChromaDBService) GetStatus() map[string]interface{} {
	if !s.IsAvailable() {
		return map[string]interface{}{
			"available": false,
			"reason":    "ChromaDB server is not running. Start it with: chroma run --host 0.0.0.0 --port 8000",
		}
	}

	collections, err := s.listCollections()
	if err != nil {
		return map[string]interface{}{
			"available": true,
			"version":   "unknown",
			"error":     err.Error(),
		}
	}

	memoryCount := 0
	for _, c := range collections {
		if c["name"] == "clawmemory_memories" {
			if count, ok := c["count"].(float64); ok {
				memoryCount = int(count)
			}
		}
	}

	return map[string]interface{}{
		"available":   true,
		"version":     "running",
		"collections": len(collections),
		"memoryCount": memoryCount,
	}
}

func (s *ChromaDBService) Install() map[string]interface{} {
	return map[string]interface{}{
		"success": false,
		"message": "ChromaDB needs to be installed and started manually. Please run:\n1. pip install chromadb\n2. chroma run --host 0.0.0.0 --port 8000",
		"steps": []string{
			"pip install chromadb",
			"chroma run --host 0.0.0.0 --port 8000",
		},
	}
}

type collection struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Tenant string `json:"tenant"`
}

func (s *ChromaDBService) listCollections() ([]map[string]interface{}, error) {
	resp, err := s.httpClient.Get(s.baseURL + "/collections")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *ChromaDBService) ensureCollection() (string, error) {
	collections, err := s.listCollections()
	if err != nil {
		return "", err
	}

	for _, c := range collections {
		if c["name"] == "clawmemory_memories" {
			if id, ok := c["id"].(string); ok {
				return id, nil
			}
		}
	}

	payload := map[string]interface{}{
		"name": "clawmemory_memories",
		"metadata": map[string]interface{}{
			"description": "ClawMemory vector embeddings",
		},
	}
	body, _ := json.Marshal(payload)

	resp, err := s.httpClient.Post(s.baseURL+"/collections", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create collection: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("failed to parse collection response: %w", err)
	}

	if id, ok := result["id"].(string); ok {
		return id, nil
	}
	return "", fmt.Errorf("collection created but ID not found")
}

func (s *ChromaDBService) SyncMemories(userID uint) (int, error) {
	collectionID, err := s.ensureCollection()
	if err != nil {
		return 0, err
	}

	var memories []models.Memory
	_ = s.db.Where("user_id = ? AND status != ?", userID, "trashed").Limit(5000).Find(&memories).Error

	if len(memories) == 0 {
		return 0, nil
	}

	ids := make([]string, len(memories))
	documents := make([]string, len(memories))
	metadatas := make([]map[string]interface{}, len(memories))

	for i, m := range memories {
		ids[i] = fmt.Sprintf("mem_%d", m.ID)
		documents[i] = m.Key + ": " + m.Value
		metadatas[i] = map[string]interface{}{
			"user_id":    fmt.Sprintf("%d", m.UserID),
			"layer":      m.Layer,
			"importance": m.Importance,
			"source":     m.Source,
			"memory_id":  fmt.Sprintf("%d", m.ID),
		}
	}

	var embeddings [][]float64
	embSvc := s.getEmbeddingService()
	if embSvc != nil {
		embeddings = embSvc.GetEmbeddings(documents)
	} else {
		embeddings = make([][]float64, len(memories))
		for i, doc := range documents {
			embeddings[i] = textToVector(doc)
		}
	}

	payload := map[string]interface{}{
		"ids":        ids,
		"embeddings": embeddings,
		"documents":  documents,
		"metadatas":  metadatas,
	}

	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s/collections/%s/add", s.baseURL, collectionID)

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to sync to ChromaDB: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("ChromaDB add failed (status %d): %s", resp.StatusCode, string(respBody))
	}

	return len(memories), nil
}

func (s *ChromaDBService) Search(userID uint, query string, limit int) ([]map[string]interface{}, error) {
	collectionID, err := s.ensureCollection()
	if err != nil {
		return nil, err
	}

	var queryEmbedding []float64
	embSvc := s.getEmbeddingService()
	if embSvc != nil {
		queryEmbedding = embSvc.GetEmbedding(query)
	} else {
		queryEmbedding = textToVector(query)
	}

	payload := map[string]interface{}{
		"query_embeddings": [][][]float64{{queryEmbedding}},
		"n_results":        limit,
		"where": map[string]interface{}{
			"user_id": fmt.Sprintf("%d", userID),
		},
	}

	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s/collections/%s/query", s.baseURL, collectionID)

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ChromaDB query failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var chromaResult map[string]interface{}
	if err := json.Unmarshal(respBody, &chromaResult); err != nil {
		return nil, fmt.Errorf("failed to parse ChromaDB response: %w", err)
	}

	return parseChromaResults(chromaResult), nil
}

func parseChromaResults(result map[string]interface{}) []map[string]interface{} {
	idsRaw, ok := result["ids"]
	if !ok {
		return []map[string]interface{}{}
	}

	idsList, ok := idsRaw.([]interface{})
	if !ok || len(idsList) == 0 {
		return []map[string]interface{}{}
	}

	firstBatch, ok := idsList[0].([]interface{})
	if !ok || len(firstBatch) == 0 {
		return []map[string]interface{}{}
	}

	distancesList := extractBatch(result, "distances")
	documentsList := extractBatch(result, "documents")
	metadatasList := extractBatch(result, "metadatas")

	results := make([]map[string]interface{}, 0, len(firstBatch))
	for i := range firstBatch {
		item := map[string]interface{}{
			"id": firstBatch[i],
		}
		if distancesList != nil && i < len(distancesList) {
			item["distance"] = distancesList[i]
			if dist, ok := distancesList[i].(float64); ok {
				item["score"] = 1.0 - dist
			}
		}
		if documentsList != nil && i < len(documentsList) {
			item["document"] = documentsList[i]
		}
		if metadatasList != nil && i < len(metadatasList) {
			if meta, ok := metadatasList[i].(map[string]interface{}); ok {
				for k, v := range meta {
					item[k] = v
				}
			}
		}
		results = append(results, item)
	}

	return results
}

func extractBatch(result map[string]interface{}, key string) []interface{} {
	raw, ok := result[key]
	if !ok {
		return nil
	}
	list, ok := raw.([]interface{})
	if !ok || len(list) == 0 {
		return nil
	}
	batch, ok := list[0].([]interface{})
	if !ok {
		return nil
	}
	return batch
}

func textToVector(text string) []float64 {
	return enhancedTextToVector(text)
}

func enhancedTextToVector(text string) []float64 {
	text = strings.ToLower(text)
	tokens := tokenize(text)

	vectorSize := 256
	vector := make([]float64, vectorSize)

	if len(tokens) == 0 {
		return vector
	}

	tokenFreq := make(map[string]float64)
	for _, t := range tokens {
		tokenFreq[t]++
	}

	totalTokens := float64(len(tokens))

	ngrams := make(map[string]float64)
	for i := 0; i < len(tokens)-1; i++ {
		bigram := tokens[i] + "_" + tokens[i+1]
		ngrams[bigram]++
	}
	for i := 0; i < len(tokens)-2; i++ {
		trigram := tokens[i] + "_" + tokens[i+1] + "_" + tokens[i+2]
		ngrams[trigram] += 0.5
	}

	for token, freq := range tokenFreq {
		tf := freq / totalTokens

		h1 := fnvHash(token)
		h2 := simpleHash(token)

		idx1 := h1 % uint32(vectorSize)
		idx2 := h2 % uint32(vectorSize)

		weight := tf * (1.0 + float64(len(token))/8.0)

		vector[idx1] += weight
		vector[idx2] -= weight * 0.3

		if len(token) > 3 {
			prefix := token[:len(token)/2]
			suffix := token[len(token)/2:]
			prefixIdx := fnvHash(prefix) % uint32(vectorSize)
			suffixIdx := fnvHash(suffix) % uint32(vectorSize)
			vector[prefixIdx] += weight * 0.4
			vector[suffixIdx] += weight * 0.3
		}
	}

	for ngram, freq := range ngrams {
		tf := freq / float64(len(ngrams)+1)
		idx := fnvHash(ngram) % uint32(vectorSize)
		vector[idx] += tf * 0.6
	}

	words := strings.Fields(text)
	for _, w := range words {
		if len(w) > 0 {
			charIdx := fnvHash("char_"+w[:1]) % uint32(vectorSize)
			vector[charIdx] += 0.05
		}
	}

	norm := 0.0
	for _, v := range vector {
		norm += v * v
	}
	norm = sqrt(norm)

	if norm > 0 {
		for i := range vector {
			vector[i] /= norm
		}
	}

	return vector
}

func fnvHash(s string) uint32 {
	h := uint32(2166136261)
	for _, c := range s {
		h ^= uint32(c)
		h *= 16777619
	}
	return h
}

func simpleHash(s string) uint32 {
	h := uint32(2166136261)
	for _, c := range s {
		h ^= uint32(c)
		h *= 16777619
	}
	return h
}

func sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	z := x
	for i := 0; i < 20; i++ {
		z = (z + x/z) / 2
	}
	return z
}
