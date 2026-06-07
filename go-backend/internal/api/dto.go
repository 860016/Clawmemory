package api

// MemoryCreateRequest is the DTO for creating a memory.
type MemoryCreateRequest struct {
	Key          string   `json:"key" binding:"required"`
	Value        string   `json:"value" binding:"required"`
	Layer        string   `json:"layer"`
	Importance   *float64 `json:"importance"`
	Tags         []string `json:"tags"`
	MemoryType   string   `json:"memory_type"`
	Visibility   string   `json:"visibility"`
	SourceAgent  string   `json:"source_agent"`
	Source       string   `json:"source"`
	IsEncrypted  bool     `json:"is_encrypted"`
	EncryptionIV string   `json:"encryption_iv"`
}

// ToMap converts the DTO to a map for service layer compatibility.
func (r *MemoryCreateRequest) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"key":   r.Key,
		"value": r.Value,
	}
	if r.Layer != "" {
		m["layer"] = r.Layer
	}
	if r.Importance != nil {
		m["importance"] = *r.Importance
	}
	if len(r.Tags) > 0 {
		tags := make([]interface{}, len(r.Tags))
		for i, t := range r.Tags {
			tags[i] = t
		}
		m["tags"] = tags
	}
	if r.MemoryType != "" {
		m["memory_type"] = r.MemoryType
	}
	if r.Visibility != "" {
		m["visibility"] = r.Visibility
	}
	if r.SourceAgent != "" {
		m["source_agent"] = r.SourceAgent
	}
	if r.Source != "" {
		m["source"] = r.Source
	}
	if r.IsEncrypted {
		m["is_encrypted"] = true
	}
	if r.EncryptionIV != "" {
		m["encryption_iv"] = r.EncryptionIV
	}
	return m
}

// MemoryUpdateRequest is the DTO for updating a memory.
type MemoryUpdateRequest struct {
	Key          *string   `json:"key"`
	Value        *string   `json:"value"`
	Layer        *string   `json:"layer"`
	Importance   *float64  `json:"importance"`
	Tags         *[]string `json:"tags"`
	MemoryType   *string   `json:"memory_type"`
	Visibility   *string   `json:"visibility"`
	SourceAgent  *string   `json:"source_agent"`
	Source       *string   `json:"source"`
	IsEncrypted  *bool     `json:"is_encrypted"`
	EncryptionIV *string   `json:"encryption_iv"`
}

// ToMap converts the DTO to a map for service layer compatibility.
// Only non-nil fields are included (partial update semantics).
func (r *MemoryUpdateRequest) ToMap() map[string]interface{} {
	m := map[string]interface{}{}
	if r.Key != nil {
		m["key"] = *r.Key
	}
	if r.Value != nil {
		m["value"] = *r.Value
	}
	if r.Layer != nil {
		m["layer"] = *r.Layer
	}
	if r.Importance != nil {
		m["importance"] = *r.Importance
	}
	if r.Tags != nil {
		tags := make([]interface{}, len(*r.Tags))
		for i, t := range *r.Tags {
			tags[i] = t
		}
		m["tags"] = tags
	}
	if r.MemoryType != nil {
		m["memory_type"] = *r.MemoryType
	}
	if r.Visibility != nil {
		m["visibility"] = *r.Visibility
	}
	if r.SourceAgent != nil {
		m["source_agent"] = *r.SourceAgent
	}
	if r.Source != nil {
		m["source"] = *r.Source
	}
	if r.IsEncrypted != nil {
		m["is_encrypted"] = *r.IsEncrypted
	}
	if r.EncryptionIV != nil {
		m["encryption_iv"] = *r.EncryptionIV
	}
	return m
}

// ReasoningConfigRequest is the DTO for setting reasoning configuration.
type ReasoningConfigRequest struct {
	Provider       string `json:"provider"`
	Model          string `json:"model"`
	BaseURL        string `json:"base_url"`
	APIKey         string `json:"api_key"`
	DialecticDepth int    `json:"dialectic_depth"`
	ReasoningLevel string `json:"reasoning_level"`
	MaxTokens      int    `json:"max_tokens"`
	Enabled        *bool  `json:"enabled"`
}

// SessionCreateRequest is the DTO for creating a session memory via MCP.
type SessionCreateRequest struct {
	SessionID      string `json:"session_id"`
	Title          string `json:"title"`
	CurrentState   string `json:"current_state"`
	TaskSpec       string `json:"task_spec"`
	Worklog        string `json:"worklog"`
	Learnings      string `json:"learnings"`
	KeyResults     string `json:"key_results"`
	Docs           string `json:"docs"`
	Errors         string `json:"errors"`
	Workflow       string `json:"workflow"`
	Status         string `json:"status"`
	TokenCount     int    `json:"token_count"`
	CompressedFrom string `json:"compressed_from"`
}

// EntityCreateRequest is the DTO for creating a knowledge entity.
type EntityCreateRequest struct {
	Name       string                 `json:"name" binding:"required"`
	EntityType string                 `json:"entity_type" binding:"required"`
	Properties map[string]interface{} `json:"properties"`
}
