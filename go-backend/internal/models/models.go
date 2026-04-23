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
	ID            uint           `gorm:"primarykey" json:"id"`
	UserID        uint           `gorm:"index;not null;default:1" json:"user_id"`
	Layer         string         `gorm:"size:20;not null" json:"layer"`
	Key           string         `gorm:"size:200;not null" json:"key"`
	Value         string         `gorm:"type:text;not null" json:"value"`
	Importance    float64        `gorm:"default:0.5" json:"importance"`
	AccessCount   int            `gorm:"default:0" json:"access_count"`
	LastAccessedAt *time.Time    `json:"last_accessed_at"`
	IsEncrypted   bool           `gorm:"default:false" json:"is_encrypted"`
	Tags          string         `gorm:"type:text" json:"tags"`
	Source        string         `gorm:"size:50;default:manual" json:"source"`
	Status        string         `gorm:"size:20;default:active;index" json:"status"`
	TrashedAt     *time.Time     `json:"trashed_at"`
	DecayStage    int            `gorm:"default:0" json:"decay_stage"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
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
	ID              uint      `gorm:"primarykey" json:"id"`
	UserID          uint      `gorm:"index;not null;default:1" json:"user_id"`
	SourceID        uint      `gorm:"not null" json:"source_id"`
	TargetID        uint      `gorm:"not null" json:"target_id"`
	RelationType    string    `gorm:"size:100;not null" json:"relation_type"`
	Description     string    `gorm:"type:text" json:"description"`
	Confidence      float64   `gorm:"default:1.0" json:"confidence"`
	DiscoverMethod  string    `gorm:"size:20;default:manual" json:"discover_method"`
	SourceMemoryID  *uint     `json:"source_memory_id"`
	Weight          float64   `gorm:"default:1.0" json:"weight"`
	CreatedAt       time.Time `json:"created_at"`
}

// WikiPage Wiki页面
type WikiPage struct {
	ID         uint           `gorm:"primarykey" json:"id"`
	UserID     uint           `gorm:"index;not null;default:1" json:"user_id"`
	Title      string         `gorm:"size:200;not null" json:"title"`
	Content    string         `gorm:"type:text" json:"content"`
	Category   string         `gorm:"size:100" json:"category"`
	Tags       string         `gorm:"type:text" json:"tags"`
	IsPublic   bool           `gorm:"default:false" json:"is_public"`
	ViewCount  int            `gorm:"default:0" json:"view_count"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

// DailyReport 日报
type DailyReport struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	UserID    uint      `gorm:"index;not null;default:1" json:"user_id"`
	Date      string    `gorm:"size:20;not null" json:"date"`
	Content   string    `gorm:"type:text" json:"content"`
	Summary   string    `gorm:"type:text" json:"summary"`
	Tags      string    `gorm:"type:text" json:"tags"`
	Mood      string    `gorm:"size:20" json:"mood"`
	CreatedAt time.Time `json:"created_at"`
}

// License 授权
type License struct {
	ID               uint      `gorm:"primarykey" json:"id"`
	LicenseKey       string    `gorm:"size:100;not null" json:"license_key"`
	Tier             string    `gorm:"size:20;default:oss" json:"tier"`
	Status           string    `gorm:"size:20;default:inactive" json:"status"`
	DeviceFingerprint string   `gorm:"size:64" json:"device_fingerprint"`
	DeviceName       string    `gorm:"size:200" json:"device_name"`
	ExpiresAt        *time.Time `json:"expires_at"`
	DeviceSlot       string    `gorm:"size:50" json:"device_slot"`
	Features         string    `gorm:"type:text" json:"features"`
	ProDownloadURL   string    `gorm:"size:500" json:"pro_download_url"`
	ProFallbackURLs  string    `gorm:"type:text" json:"pro_fallback_urls"`
	CreatedAt        time.Time `json:"created_at"`
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
