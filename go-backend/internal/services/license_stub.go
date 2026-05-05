//go:build !pro

package services

import (
	"clawmemory/internal/config"

	"gorm.io/gorm"
)

type LicenseManager struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewLicenseManager(db *gorm.DB, cfg *config.Config) *LicenseManager {
	return &LicenseManager{db: db, cfg: cfg}
}

func (lm *LicenseManager) Activate(licenseKey string) (map[string]interface{}, error) {
	return nil, errProNotAvailable
}

func (lm *LicenseManager) GetLicenseInfo() map[string]interface{} {
	return map[string]interface{}{
		"active":   false,
		"tier":     "oss",
		"is_valid": false,
	}
}

func (lm *LicenseManager) Deactivate(userID uint) map[string]interface{} {
	return map[string]interface{}{
		"deactivated": false,
		"message":     "Pro features not available in OSS edition",
	}
}

func (lm *LicenseManager) IsFeatureEnabled(feature string) bool {
	return false
}

func (lm *LicenseManager) GetTier() string {
	return "oss"
}

var errProNotAvailable = &ProError{Message: "Pro features require ClawMemory Pro edition. Visit https://github.com/860016/Clawmemory for more information.", Code: 403}
