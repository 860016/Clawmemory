package models

import (
	"time"

	"gorm.io/gorm"
)

// User 用户
type User struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	Username  string    `gorm:"uniqueIndex;not null" json:"username"`
	Password  string    `gorm:"not null" json:"-"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Memory 记忆
type Memory struct {
	ID             uint       `gorm:"primarykey" json:"id"`
	UserID         uint       `gorm:"index;not null;default:1" json:"user_id"`
	Layer          string     `gorm:"size:20;not null" json:"layer"`
	Key            string     `gorm:"size:200;not null" json:"key"`
	Value          string     `gorm:"type:text;not null" json:"value"`
	Importance     float64    `gorm:"default:0.5" json:"importance"`
	AccessCount    int        `gorm:"default:0" json:"access_count"`
	LastAccessedAt *time.Time `json:"last_accessed_at"`
	IsEncrypted    bool       `gorm:"default:false" json:"is_encrypted"`
	Tags           string     `gorm:"type:text" json:"tags"`
	Summary        string     `gorm:"size:500" json:"summary"`
	Source         string     `gorm:"size:50;default:manual" json:"source"`
	Platform       string     `gorm:"size:30;default:openclaw;index" json:"platform"`
	Status         string     `gorm:"size:20;default:active;index" json:"status"`
	TrashedAt      *time.Time `json:"trashed_at"`
	DecayStage     int        `gorm:"default:0" json:"decay_stage"`
	ReinforceCount int        `gorm:"default:0" json:"reinforce_count"`
	MemoryType     string     `gorm:"size:20;default:knowledge" json:"memory_type"`
	VerifiedAt     *time.Time `json:"verified_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// Entity 知识实体
type Entity struct {
	ID             uint      `gorm:"primarykey" json:"id"`
	UserID         uint      `gorm:"index;not null;default:1" json:"user_id"`
	Name           string    `gorm:"size:200;not null" json:"name"`
	EntityType     string    `gorm:"size:50;not null" json:"entity_type"`
	Description    string    `gorm:"type:text" json:"description"`
	Properties     string    `gorm:"type:text" json:"properties"`
	SourceMemoryID *uint     `json:"source_memory_id"`
	Confidence     float64   `gorm:"default:1.0" json:"confidence"`
	ExtractMethod  string    `gorm:"size:20;default:manual" json:"extract_method"`
	CanonicalName  string    `gorm:"size:200" json:"canonical_name"`
	Aliases        string    `gorm:"type:text" json:"aliases"`
	EmbeddingID    string    `gorm:"size:100" json:"embedding_id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Relation 知识关系
type Relation struct {
	ID             uint      `gorm:"primarykey" json:"id"`
	UserID         uint      `gorm:"index;not null;default:1" json:"user_id"`
	SourceID       uint      `gorm:"not null" json:"source_id"`
	TargetID       uint      `gorm:"not null" json:"target_id"`
	RelationType   string    `gorm:"size:100;not null" json:"relation_type"`
	Description    string    `gorm:"type:text" json:"description"`
	Confidence     float64   `gorm:"default:1.0" json:"confidence"`
	DiscoverMethod string    `gorm:"size:20;default:manual" json:"discover_method"`
	SourceMemoryID *uint     `json:"source_memory_id"`
	Weight         float64   `gorm:"default:1.0" json:"weight"`
	CreatedAt      time.Time `json:"created_at"`
}

// Project 项目知识
type Project struct {
	ID              uint           `gorm:"primarykey" json:"id"`
	UserID          uint           `gorm:"index;not null;default:1" json:"user_id"`
	Name            string         `gorm:"size:200;not null" json:"name"`
	Description     string         `gorm:"type:text" json:"description"`
	Status          string         `gorm:"size:20;default:active" json:"status"`
	Category        string         `gorm:"size:100" json:"category"`
	Tags            string         `gorm:"type:text" json:"tags"`
	KeyDecisions    string         `gorm:"type:text" json:"key_decisions"`
	ActionItems     string         `gorm:"type:text" json:"action_items"`
	Progress        int            `gorm:"default:0" json:"progress"`
	SourceSessionID string         `gorm:"size:100" json:"source_session_id"`
	SourceAgent     string         `gorm:"size:100" json:"source_agent"`
	IsPinned        bool           `gorm:"default:false" json:"is_pinned"`
	LastDiscussedAt *time.Time     `json:"last_discussed_at"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

// ProjectNote 项目笔记
type ProjectNote struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	UserID     uint      `gorm:"index;not null;default:1" json:"user_id"`
	ProjectID  uint      `gorm:"index;not null" json:"project_id"`
	Content    string    `gorm:"type:text;not null" json:"content"`
	NoteType   string    `gorm:"size:30;default:note" json:"note_type"`
	Source     string    `gorm:"size:50;default:manual" json:"source"`
	IsKeyPoint bool      `gorm:"default:false" json:"is_key_point"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// WikiPage Wiki页面 (legacy, kept for migration)
type WikiPage struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	UserID       uint           `gorm:"index;not null;default:1" json:"user_id"`
	Title        string         `gorm:"size:200;not null" json:"title"`
	Content      string         `gorm:"type:text" json:"content"`
	Category     string         `gorm:"size:100" json:"category"`
	Tags         string         `gorm:"type:text" json:"tags"`
	Status       string         `gorm:"size:20;default:draft" json:"status"`
	Summary      string         `gorm:"type:text" json:"summary"`
	IsPublic     bool           `gorm:"default:false" json:"is_public"`
	IsPinned     bool           `gorm:"default:false" json:"is_pinned"`
	ParentID     *uint          `json:"parent_id"`
	AIGenerated  bool           `gorm:"default:false" json:"ai_generated"`
	AIConfidence float64        `gorm:"default:0" json:"ai_confidence"`
	KeyDecisions string         `gorm:"type:text" json:"key_decisions"`
	ActionItems  string         `gorm:"type:text" json:"action_items"`
	ViewCount    int            `gorm:"default:0" json:"view_count"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// DailyReport 日报
type DailyReport struct {
	ID                  uint      `gorm:"primarykey" json:"id"`
	UserID              uint      `gorm:"index;not null;default:1" json:"user_id"`
	Date                string    `gorm:"size:20;not null" json:"date"`
	ReportDate          string    `gorm:"size:20" json:"report_date"`
	Content             string    `gorm:"type:text" json:"content"`
	Summary             string    `gorm:"type:text" json:"summary"`
	Highlights          string    `gorm:"type:text" json:"highlights"`
	KnowledgeGained     string    `gorm:"type:text" json:"knowledge_gained"`
	PendingTasks        string    `gorm:"type:text" json:"pending_tasks"`
	TomorrowSuggestions string    `gorm:"type:text" json:"tomorrow_suggestions"`
	Stats               string    `gorm:"type:text" json:"stats"`
	Tags                string    `gorm:"type:text" json:"tags"`
	Mood                string    `gorm:"size:20" json:"mood"`
	CreatedAt           time.Time `json:"created_at"`
}

// License 授权
type License struct {
	ID                uint       `gorm:"primarykey" json:"id"`
	LicenseKey        string     `gorm:"size:100;not null;uniqueIndex" json:"license_key"`
	Tier              string     `gorm:"size:20;default:oss" json:"tier"`
	Status            string     `gorm:"size:20;default:inactive" json:"status"`
	DeviceFingerprint string     `gorm:"size:64" json:"device_fingerprint"`
	DeviceName        string     `gorm:"size:200" json:"device_name"`
	ExpiresAt         *time.Time `json:"expires_at"`
	DeviceSlot        string     `gorm:"size:50" json:"device_slot"`
	Features          string     `gorm:"type:text" json:"features"`
	ProDownloadURL    string     `gorm:"size:500" json:"pro_download_url"`
	ProFallbackURLs   string     `gorm:"type:text" json:"pro_fallback_urls"`
	ActivatedAt       *time.Time `json:"activated_at"`
	CreatedAt         time.Time  `json:"created_at"`
}

// Setting 设置
type Setting struct {
	ID     uint   `gorm:"primarykey" json:"id"`
	UserID uint   `gorm:"index;not null;default:1" json:"user_id"`
	Key    string `gorm:"size:100;not null;uniqueIndex:idx_user_key" json:"key"`
	Value  string `gorm:"type:text" json:"value"`
}

// Backup 备份
type Backup struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	UserID    uint      `gorm:"index;not null;default:1" json:"user_id"`
	Filename  string    `gorm:"size:200;not null" json:"filename"`
	Size      int64     `json:"size"`
	Type      string    `gorm:"size:20" json:"type"`
	CreatedAt time.Time `json:"created_at"`
}

// APIKey API 密钥 - 供外部应用（如 OpenClaw）调用
type APIKey struct {
	ID         uint       `gorm:"primarykey" json:"id"`
	UserID     uint       `gorm:"index;not null" json:"user_id"`
	Name       string     `gorm:"size:100;not null" json:"name"`
	KeyHash    string     `gorm:"size:64;not null;uniqueIndex" json:"-"`
	KeyPrefix  string     `gorm:"size:8;not null" json:"key_prefix"`
	LastUsedAt *time.Time `json:"last_used_at"`
	ExpiresAt  *time.Time `json:"expires_at"`
	CreatedAt  time.Time  `json:"created_at"`
}

type AuditLog struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	UserID    uint      `gorm:"index" json:"user_id"`
	Action    string    `gorm:"size:50;not null;index" json:"action"`
	Target    string    `gorm:"size:100" json:"target"`
	Detail    string    `gorm:"size:500" json:"detail"`
	IP        string    `gorm:"size:45" json:"ip"`
	CreatedAt time.Time `json:"created_at"`
}

type SessionMemory struct {
	ID             uint      `gorm:"primarykey" json:"id"`
	UserID         uint      `gorm:"index;not null;default:1" json:"user_id"`
	SessionID      string    `gorm:"size:100;index" json:"session_id"`
	Title          string    `gorm:"size:200" json:"title"`
	CurrentState   string    `gorm:"type:text" json:"current_state"`
	TaskSpec       string    `gorm:"type:text" json:"task_spec"`
	FilesAndFuncs  string    `gorm:"type:text" json:"files_and_funcs"`
	Workflow       string    `gorm:"type:text" json:"workflow"`
	Errors         string    `gorm:"type:text" json:"errors"`
	Docs           string    `gorm:"type:text" json:"docs"`
	Learnings      string    `gorm:"type:text" json:"learnings"`
	KeyResults     string    `gorm:"type:text" json:"key_results"`
	Worklog        string    `gorm:"type:text" json:"worklog"`
	TokenCount     int       `gorm:"default:0" json:"token_count"`
	CompressedFrom string    `gorm:"size:100" json:"compressed_from"`
	Status         string    `gorm:"size:20;default:active" json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type ReasoningConfig struct {
	ID             uint      `gorm:"primarykey" json:"id"`
	UserID         uint      `gorm:"index;not null" json:"user_id"`
	Provider       string    `gorm:"size:30;not null" json:"provider"`
	Model          string    `gorm:"size:50" json:"model"`
	APIKey         string    `gorm:"size:200" json:"api_key"`
	BaseURL        string    `gorm:"size:200" json:"base_url"`
	DialecticDepth int       `gorm:"default:1" json:"dialectic_depth"`
	ReasoningLevel string    `gorm:"size:20;default:medium" json:"reasoning_level"`
	MaxTokens      int       `gorm:"default:1000" json:"max_tokens"`
	Enabled        bool      `gorm:"default:false" json:"enabled"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
