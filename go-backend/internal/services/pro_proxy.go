package services

import (
	"clawmemory/internal/config"
	"clawmemory/internal/models"

	"gorm.io/gorm"
)

type ProProxy struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewProProxy(db *gorm.DB, cfg *config.Config) *ProProxy {
	return &ProProxy{
		db:  db,
		cfg: cfg,
	}
}

func (p *ProProxy) getLicenseKey() string {
	var license models.License
	if err := p.db.Where("status = ?", "active").First(&license).Error; err != nil {
		return ""
	}
	return license.LicenseKey
}

func (p *ProProxy) IsPro() bool {
	return p.getLicenseKey() != ""
}

var (
	ErrProRequired = &ProError{Message: "Pro license required", Code: 403}
)

type ProError struct {
	Message string
	Code    int
}

func (e *ProError) Error() string {
	return e.Message
}
