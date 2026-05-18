package models

import (
	"time"

	"gorm.io/gorm"
)

// User 用户
type User struct {
	ID               uint       `gorm:"primarykey" json:"id"`
	Username         string     `gorm:"uniqueIndex;not null" json:"username"`
	Password         string     `gorm:"not null" json:"-"`
	Email            string     `json:"email"`
	Role             string     `gorm:"size:20;default:user" json:"role"`
	IsFounder        bool       `gorm:"default:false" json:"is_founder"`
	TokenVersion     int        `gorm:"default:1" json:"token_version"`
	InvitationCode   string     `gorm:"size:50" json:"-"`
	FailedAttempts   int        `gorm:"default:0" json:"-"`
	LockedUntil      *time.Time `json:"-"`
	RefreshToken     string     `gorm:"size:500" json:"-"`
	RefreshTokenExp  *time.Time `json:"-"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
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
	Platform       string     `gorm:"size:30;default:clawmemory;index" json:"platform"`
	Status         string     `gorm:"size:20;default:active;index" json:"status"`
	TrashedAt      *time.Time `json:"trashed_at"`
	DecayStage     int        `gorm:"default:0" json:"decay_stage"`
	ReinforceCount int        `gorm:"default:0" json:"reinforce_count"`
	MemoryType     string     `gorm:"size:20;default:knowledge" json:"memory_type"`
	VerifiedAt     *time.Time `json:"verified_at"`
	SourceAgent    string     `gorm:"size:50;index" json:"source_agent"`
	Visibility     string     `gorm:"size:20;default:private;index" json:"visibility"`
	OriginChain    string     `gorm:"size:200" json:"origin_chain"`
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
	ID          uint       `gorm:"primarykey" json:"id"`
	UserID      uint       `gorm:"index;not null" json:"user_id"`
	Name        string     `gorm:"size:100;not null" json:"name"`
	KeyHash     string     `gorm:"size:64;not null;uniqueIndex" json:"-"`
	KeyPrefix   string     `gorm:"size:8;not null" json:"key_prefix"`
	Permissions string     `gorm:"size:200;default:memories:read,memories:write,conversations:write,sessions:write,reason:execute,ai:execute" json:"permissions"`
	AgentName   string     `gorm:"size:50;index" json:"agent_name"`
	LastUsedAt  *time.Time `json:"last_used_at"`
	ExpiresAt   *time.Time `json:"expires_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

type AuditLog struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	UserID    uint      `gorm:"index" json:"user_id"`
	Action    string    `gorm:"size:50;not null;index" json:"action"`
	Target    string    `gorm:"size:100" json:"target"`
	Detail    string    `gorm:"size:500" json:"detail"`
	IP        string    `gorm:"size:45" json:"ip"`
	UserAgent string    `gorm:"size:500" json:"user_agent"`
	AgentName string    `gorm:"size:50;index" json:"agent_name"`
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

type MemoryShare struct {
	ID         uint       `gorm:"primarykey" json:"id"`
	MemoryID   uint       `gorm:"index;not null" json:"memory_id"`
	FromUserID uint       `gorm:"index;not null" json:"from_user_id"`
	ToUserID   uint       `gorm:"index" json:"to_user_id"`
	ToAgent    string     `gorm:"size:50;index" json:"to_agent"`
	ShareType  string     `gorm:"size:20;not null;default:manual" json:"share_type"`
	Status     string     `gorm:"size:20;default:pending;index" json:"status"`
	ApprovedBy *uint      `json:"approved_by"`
	ApprovedAt *time.Time `json:"approved_at"`
	ExpiresAt  *time.Time `json:"expires_at"`
	RuleID     *uint      `json:"rule_id"`
	Memory     Memory     `gorm:"foreignKey:MemoryID" json:"memory,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type ShareRule struct {
	ID            uint      `gorm:"primarykey" json:"id"`
	UserID        uint      `gorm:"index;not null" json:"user_id"`
	Name          string    `gorm:"size:100;not null" json:"name"`
	SourceAgent   string    `gorm:"size:50" json:"source_agent"`
	TargetAgent   string    `gorm:"size:50" json:"target_agent"`
	Layer         string    `gorm:"size:20" json:"layer"`
	MinImportance float64   `gorm:"default:0.5" json:"min_importance"`
	AutoApprove   bool      `gorm:"default:false" json:"auto_approve"`
	Enabled       bool      `gorm:"default:true" json:"enabled"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Invitation struct {
	ID        uint       `gorm:"primarykey" json:"id"`
	Code      string     `gorm:"uniqueIndex;size:50;not null" json:"code"`
	CreatedBy uint       `gorm:"index;not null" json:"created_by"`
	UsedBy    *uint      `json:"used_by"`
	UsedAt    *time.Time `json:"used_at"`
	MaxUses   int        `gorm:"default:1" json:"max_uses"`
	UsedCount int        `gorm:"default:0" json:"used_count"`
	ExpiresAt *time.Time `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`
}

type SystemLog struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	Timestamp string    `gorm:"size:30;index" json:"timestamp"`
	Level     string    `gorm:"size:10;index" json:"level"`
	Service   string    `gorm:"size:50" json:"service"`
	Message   string    `gorm:"size:500" json:"message"`
	Caller    string    `gorm:"size:100" json:"caller"`
	Data      string    `gorm:"type:text" json:"data"`
	CreatedAt time.Time `json:"created_at"`
}

type SchemaMigration struct {
	ID      uint   `gorm:"primarykey" json:"id"`
	Version int    `gorm:"default:0" json:"version"`
	Dirty   bool   `gorm:"default:false" json:"dirty"`
	Key     string `gorm:"size:100;uniqueIndex" json:"key"`
}

type MemoryHistory struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	UserID      uint      `gorm:"index;not null" json:"user_id"`
	MemoryID    uint      `gorm:"index;not null" json:"memory_id"`
	OldKey      string    `gorm:"size:200" json:"old_key"`
	OldValue    string    `gorm:"type:text" json:"old_value"`
	NewKey      string    `gorm:"size:200" json:"new_key"`
	NewValue    string    `gorm:"type:text" json:"new_value"`
	ChangeType  string    `gorm:"size:30;not null" json:"change_type"`
	Reason      string    `gorm:"size:500" json:"reason"`
	SourceAgent string    `gorm:"size:50" json:"source_agent"`
	SessionID   string    `gorm:"size:100;index" json:"session_id"`
	CreatedAt   time.Time `json:"created_at"`
}

type UserProfile struct {
	ID               uint       `gorm:"primarykey" json:"id"`
	UserID           uint       `gorm:"uniqueIndex;not null" json:"user_id"`
	Identity         string     `gorm:"type:text" json:"identity"`
	Communication    string     `gorm:"type:text" json:"communication"`
	Workflow         string     `gorm:"type:text" json:"workflow"`
	Preferences      string     `gorm:"type:text" json:"preferences"`
	Patterns         string     `gorm:"type:text" json:"patterns"`
	GrowthAreas      string     `gorm:"type:text" json:"growth_areas"`
	ProfileVersion   int        `gorm:"default:1" json:"profile_version"`
	Confidence       float64    `gorm:"default:0" json:"confidence"`
	LastNudgeAt      *time.Time `json:"last_nudge_at"`
	NudgeCount       int        `gorm:"default:0" json:"nudge_count"`
	TotalRefinements int        `gorm:"default:0" json:"total_refinements"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type ActionTrace struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	UserID      uint      `gorm:"index;not null" json:"user_id"`
	SessionID   string    `gorm:"size:100;index" json:"session_id"`
	AgentName   string    `gorm:"size:50;index" json:"agent_name"`
	Platform    string    `gorm:"size:30;index" json:"platform"`
	ActionType  string    `gorm:"size:30;not null;index" json:"action_type"`
	ActionName  string    `gorm:"size:100;not null" json:"action_name"`
	Parameters  string    `gorm:"type:text" json:"parameters"`
	Result      string    `gorm:"size:20;default:success" json:"result"`
	Duration    int       `gorm:"default:0" json:"duration"`
	PatternHash string    `gorm:"size:64;index" json:"pattern_hash"`
	CreatedAt   time.Time `json:"created_at"`
}

type Skill struct {
	ID              uint       `gorm:"primarykey" json:"id"`
	UserID          uint       `gorm:"not null;uniqueIndex:idx_user_skill" json:"user_id"`
	Name            string     `gorm:"size:100;not null;uniqueIndex:idx_user_skill" json:"name"`
	Description     string     `gorm:"size:500" json:"description"`
	TriggerKeywords string     `gorm:"type:text" json:"trigger_keywords"`
	Steps           string     `gorm:"type:text" json:"steps"`
	Parameters      string     `gorm:"type:text" json:"parameters"`
	KnownPitfalls   string     `gorm:"type:text" json:"known_pitfalls"`
	Verification    string     `gorm:"size:500" json:"verification"`
	SourceAgent     string     `gorm:"size:50;index" json:"source_agent"`
	SourceSession   string     `gorm:"size:100" json:"source_session"`
	Category        string     `gorm:"size:30;index" json:"category"`
	Tags            string     `gorm:"type:text" json:"tags"`
	UsageCount      int        `gorm:"default:0" json:"usage_count"`
	SuccessCount    int        `gorm:"default:0" json:"success_count"`
	FailCount       int        `gorm:"default:0" json:"fail_count"`
	Version         int        `gorm:"default:1" json:"version"`
	AutoCreated     bool       `gorm:"default:true" json:"auto_created"`
	Status          string     `gorm:"size:20;default:active;index" json:"status"`
	LastUsedAt      *time.Time `json:"last_used_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type SkillSuggestion struct {
	ID          uint       `gorm:"primarykey" json:"id"`
	UserID      uint       `gorm:"index;not null" json:"user_id"`
	AgentName   string     `gorm:"size:50;index" json:"agent_name"`
	SuggestType string     `gorm:"size:30;not null" json:"suggest_type"`
	Title       string     `gorm:"size:200;not null" json:"title"`
	Description string     `gorm:"type:text" json:"description"`
	ImportURL   string     `gorm:"size:500" json:"import_url"`
	ImportGuide string     `gorm:"type:text" json:"import_guide"`
	Status      string     `gorm:"size:20;default:pending;index" json:"status"`
	DismissedAt *time.Time `json:"dismissed_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

type FileSyncIndex struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	FilePath  string    `gorm:"size:500;uniqueIndex" json:"file_path"`
	FileHash  string    `gorm:"size:64;not null" json:"file_hash"`
	FileSize  int64     `json:"file_size"`
	ModTime   int64     `json:"mod_time"`
	Source    string    `gorm:"size:50;index" json:"source"`
	ChunkCount int      `gorm:"default:0" json:"chunk_count"`
	Platform  string    `gorm:"size:30;index" json:"platform"`
	SyncedAt  time.Time `json:"synced_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
