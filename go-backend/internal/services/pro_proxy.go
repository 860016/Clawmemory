package services

import (
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"clawmemory/internal/config"
	"clawmemory/internal/models"

	"gorm.io/gorm"
)

type ProProxy struct {
	db       *gorm.DB
	cfg      *config.Config
	cached   *models.License
	cacheAt  time.Time
}

func NewProProxy(db *gorm.DB, cfg *config.Config) *ProProxy {
	return &ProProxy{
		db:  db,
		cfg: cfg,
	}
}

func (p *ProProxy) getActiveLicense() *models.License {
	if p.cached != nil && time.Since(p.cacheAt) < 5*time.Minute {
		if p.validateLicense(p.cached) {
			return p.cached
		}
		p.cached = nil
		p.cacheAt = time.Time{}
	}

	var license models.License
	if err := p.db.Where("status = ?", "active").First(&license).Error; err != nil {
		return nil
	}

	p.cached = &license
	p.cacheAt = time.Now()
	return &license
}

func (p *ProProxy) validateLicense(license *models.License) bool {
	if license == nil {
		return false
	}

	if license.LicenseKey == "" {
		return false
	}

	if license.Tier != "pro" && license.Tier != "enterprise" {
		return false
	}

	if license.ExpiresAt != nil && license.ExpiresAt.Before(time.Now()) {
		return false
	}

	if !p.validateLicenseKeyFormat(license.LicenseKey) {
		return false
	}

	if license.DeviceFingerprint != "" {
		currentFP := getDeviceFingerprint()
		if license.DeviceFingerprint != currentFP {
			storedFP := strings.TrimPrefix(license.DeviceFingerprint, "fp_")
			currentFP = strings.TrimPrefix(currentFP, "fp_")
			if storedFP != currentFP {
				return false
			}
		}
	}

	return true
}

func (p *ProProxy) validateLicenseKeyFormat(key string) bool {
	if len(key) < 16 {
		return false
	}

	prefixes := []string{"cm-pro-", "cm-ent-", "claw-pro-", "claw-ent-"}
	hasPrefix := false
	for _, prefix := range prefixes {
		if strings.HasPrefix(key, prefix) {
			hasPrefix = true
			break
		}
	}

	if !hasPrefix {
		if strings.Contains(key, "-") && len(key) >= 20 {
			hasPrefix = true
		}
	}

	if !hasPrefix {
		return false
	}

	if strings.Contains(strings.ToLower(key), "fake") ||
		strings.Contains(strings.ToLower(key), "test") ||
		strings.Contains(strings.ToLower(key), "hack") ||
		strings.Contains(strings.ToLower(key), "crack") ||
		strings.Contains(strings.ToLower(key), "bypass") {
		return false
	}

	return true
}

func (p *ProProxy) IsPro() bool {
	license := p.getActiveLicense()
	return p.validateLicense(license)
}

func (p *ProProxy) verifyOfflineSignature(license *models.License) bool {
	if license == nil {
		return false
	}

	pubkey := p.loadPublicKey()
	if pubkey == nil {
		return true
	}

	payload := fmt.Sprintf("%s:%s:%s", license.LicenseKey, license.Tier, license.DeviceFingerprint)
	hash := sha256.Sum256([]byte(payload))

	sigFile := filepath.Join(filepath.Dir(p.cfg.RSAPublicKeyPath), "license.sig")
	sigData, err := os.ReadFile(sigFile)
	if err != nil {
		return true
	}

	sig, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(sigData)))
	if err != nil {
		return true
	}

	if err := rsa.VerifyPKCS1v15(pubkey, 0, hash[:], sig); err != nil {
		return false
	}

	return true
}

func (p *ProProxy) loadPublicKey() *rsa.PublicKey {
	if p.cfg.RSAPublicKeyPath == "" {
		return nil
	}
	if data, err := os.ReadFile(p.cfg.RSAPublicKeyPath); err == nil {
		return parsePublicKey(data)
	}
	return nil
}

func (p *ProProxy) GetLicenseInfo() map[string]interface{} {
	license := p.getActiveLicense()
	if license == nil {
		return map[string]interface{}{
			"active":   false,
			"tier":     "oss",
			"is_valid": false,
		}
	}

	features := []string{}
	json.Unmarshal([]byte(license.Features), &features)

	return map[string]interface{}{
		"active":       p.validateLicense(license),
		"tier":         license.Tier,
		"type":         license.Tier,
		"features":     features,
		"expires_at":   license.ExpiresAt,
		"device_slot":  license.DeviceSlot,
		"license_key":  maskLicenseKey(license.LicenseKey),
		"is_valid":     p.validateLicense(license),
	}
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
