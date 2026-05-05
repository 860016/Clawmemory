//go:build !pro

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
	return &ProProxy{db: db, cfg: cfg}
}

func (p *ProProxy) IsPro() bool {
	return false
}

func (p *ProProxy) GetLicenseInfo() map[string]interface{} {
	return map[string]interface{}{
		"active":   false,
		"tier":     "oss",
		"is_valid": false,
	}
}

func (p *ProProxy) InvalidateCache() {}

func (p *ProProxy) getActiveLicense() *models.License {
	return nil
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
