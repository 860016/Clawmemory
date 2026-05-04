package services

import (
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"clawmemory/internal/config"
	"clawmemory/internal/models"

	"gorm.io/gorm"
)

// LicenseManager 授权管理器
type LicenseManager struct {
	db       *gorm.DB
	cfg      *config.Config
	tier     string
	features []string
}

func NewLicenseManager(db *gorm.DB, cfg *config.Config) *LicenseManager {
	return &LicenseManager{
		db:       db,
		cfg:      cfg,
		tier:     "oss",
		features: []string{},
	}
}

// Activate 激活授权
func (lm *LicenseManager) Activate(licenseKey string) (map[string]interface{}, error) {
	var existingActive models.License
	if err := lm.db.Where("license_key = ? AND status = ?", licenseKey, "active").First(&existingActive).Error; err == nil {
		currentFP := getDeviceFingerprint()
		if existingActive.DeviceFingerprint == currentFP {
			return map[string]interface{}{
				"valid":   true,
				"message": "already activated on this device",
				"tier":    existingActive.Tier,
			}, nil
		}
		return nil, errors.New("this license key is already activated on another device")
	}

	var deactivatedLicense models.License
	if err := lm.db.Where("license_key = ? AND status = ?", licenseKey, "inactive").First(&deactivatedLicense).Error; err == nil {
		currentFP := getDeviceFingerprint()
		if deactivatedLicense.DeviceFingerprint != currentFP {
			return nil, errors.New("this license key has been deactivated and cannot be reactivated on a different device")
		}
	}

	fingerprint := getDeviceFingerprint()
	deviceName := getDeviceName()

	// 调用授权服务器
	resp, err := http.Post(
		lm.cfg.LicenseServerURL+"/api/v1/activate",
		"application/json",
		jsonReader(map[string]interface{}{
			"license_key": licenseKey,
			"fingerprint": fingerprint,
			"device_name": deviceName,
		}),
	)
	if err != nil {
		return nil, errors.New("unable to connect to license server")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		errMsg := "license server returned error"
		var errResp map[string]interface{}
		if json.Unmarshal(body, &errResp) == nil {
			if msg, ok := errResp["message"].(string); ok && msg != "" {
				errMsg = msg
			} else if msg, ok := errResp["error"].(string); ok && msg != "" {
				errMsg = msg
			}
		}
		return nil, errors.New(errMsg)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if valid, ok := result["valid"].(bool); !ok || !valid {
		errMsg := "license activation failed"
		if msg, ok := result["message"].(string); ok && msg != "" {
			errMsg = msg
		} else if msg, ok := result["error"].(string); ok && msg != "" {
			errMsg = msg
		}
		return nil, errors.New(errMsg)
	}

	// RSA 签名验证
	if signature, ok := result["signature"].(string); ok && signature != "" {
		if err := lm.verifySignature(signature); err != nil {
			// 尝试刷新公钥重试
			if err := lm.refreshPublicKey(); err == nil {
				if err := lm.verifySignature(signature); err != nil {
					return nil, errors.New("RSA signature verification failed")
				}
			}
		}
	}

	// 保存到数据库
	features := []string{}
	if f, ok := result["features"].([]interface{}); ok {
		for _, feature := range f {
			if s, ok := feature.(string); ok {
				features = append(features, s)
			}
		}
	}
	featuresJSON, _ := json.Marshal(features)

	fallbackURLs := []string{}
	if f, ok := result["pro_fallback_urls"].([]interface{}); ok {
		for _, url := range f {
			if s, ok := url.(string); ok {
				fallbackURLs = append(fallbackURLs, s)
			}
		}
	}
	fallbackJSON, _ := json.Marshal(fallbackURLs)

	// 停用其他授权（不同 key 的）
	lm.db.Model(&models.License{}).Where("status = ? AND license_key != ?", "active", licenseKey).Update("status", "inactive")

	// 创建或更新授权
	now := time.Now()
	license := &models.License{
		LicenseKey:        licenseKey,
		Tier:              getString(result, "tier", "pro"),
		Status:            "active",
		DeviceFingerprint: fingerprint,
		DeviceName:        deviceName,
		Features:          string(featuresJSON),
		ProDownloadURL:    getString(result, "pro_download_url", ""),
		ProFallbackURLs:   string(fallbackJSON),
		ActivatedAt:       &now,
	}

	if exp, ok := result["expires_at"].(string); ok {
		t, _ := time.Parse(time.RFC3339, exp)
		license.ExpiresAt = &t
	}

	var existingRecord models.License
	if err := lm.db.Where("license_key = ?", licenseKey).First(&existingRecord).Error; err == nil {
		if err := lm.db.Model(&existingRecord).Updates(map[string]interface{}{
			"tier":               license.Tier,
			"status":             "active",
			"device_fingerprint": fingerprint,
			"device_name":        deviceName,
			"features":           string(featuresJSON),
			"pro_download_url":   license.ProDownloadURL,
			"pro_fallback_urls":  string(fallbackJSON),
			"expires_at":         license.ExpiresAt,
			"activated_at":       &now,
		}).Error; err != nil {
			return nil, err
		}
		license = &existingRecord
	} else {
		if err := lm.db.Create(license).Error; err != nil {
			return nil, err
		}
	}

	// 更新内存状态
	lm.tier = license.Tier
	lm.features = features

	return map[string]interface{}{
		"valid":             true,
		"message":           "activated successfully",
		"tier":              license.Tier,
		"expires_at":        license.ExpiresAt,
		"features":          features,
		"pro_download_url":  license.ProDownloadURL,
		"pro_fallback_urls": fallbackURLs,
	}, nil
}

// GetLicenseInfo 获取授权信息
func (lm *LicenseManager) GetLicenseInfo() map[string]interface{} {
	var license models.License
	lm.db.Where("status = ?", "active").First(&license)

	if license.Status == "active" {
		lm.tier = license.Tier
		json.Unmarshal([]byte(license.Features), &lm.features)
	}

	allFeatures := []string{
		"ai_extract", "auto_graph", "unlimited_graph",
		"auto_decay", "decay_report", "prune_suggest", "reinforce",
		"conflict_scan", "conflict_merge",
		"smart_router", "token_stats",
		"wiki", "auto_backup",
	}

	features := []string{}
	if license.Status == "active" {
		json.Unmarshal([]byte(license.Features), &features)
	}

	activeFeatures := []string{}
	for _, f := range allFeatures {
		for _, af := range features {
			if f == af {
				activeFeatures = append(activeFeatures, f)
				break
			}
		}
	}

	fallbackURLs := []string{}
	if license.Status == "active" && license.ProFallbackURLs != "" {
		json.Unmarshal([]byte(license.ProFallbackURLs), &fallbackURLs)
	}

	return map[string]interface{}{
		"active":            license.Status == "active",
		"tier":              lm.tier,
		"type":              license.Tier,
		"features":          activeFeatures,
		"expires_at":        license.ExpiresAt,
		"device_slot":       license.DeviceSlot,
		"license_key":       maskLicenseKey(license.LicenseKey),
		"is_valid":          license.Status == "active",
		"pro_download_url":  license.ProDownloadURL,
		"pro_fallback_urls": fallbackURLs,
	}
}

func (lm *LicenseManager) Deactivate(userID uint) map[string]interface{} {
	var activeLicense models.License
	if err := lm.db.Where("status = ?", "active").First(&activeLicense).Error; err == nil {
		fingerprint := getDeviceFingerprint()
		go func() {
			body := map[string]interface{}{
				"license_key": activeLicense.LicenseKey,
				"fingerprint": fingerprint,
			}
			resp, err := http.Post(
				lm.cfg.LicenseServerURL+"/api/v1/deactivate",
				"application/json",
				jsonReader(body),
			)
			if err == nil {
				resp.Body.Close()
			}
		}()
	}

	result := lm.db.Model(&models.License{}).Where("status = ?", "active").Update("status", "inactive")
	lm.tier = "oss"
	lm.features = []string{}
	return map[string]interface{}{
		"deactivated": true,
		"affected":    result.RowsAffected,
	}
}

// IsFeatureEnabled 检查功能是否启用
func (lm *LicenseManager) IsFeatureEnabled(feature string) bool {
	for _, f := range lm.features {
		if f == feature {
			return true
		}
	}
	return false
}

// GetTier 获取当前层级
func (lm *LicenseManager) GetTier() string {
	return lm.tier
}

// verifySignature 验证 RSA 签名
func (lm *LicenseManager) verifySignature(signatureB64 string) error {
	pubkey := lm.loadPublicKey()
	if pubkey == nil {
		return errors.New("unable to load public key")
	}

	// 解析签名
	parts := splitLast(signatureB64, ".")
	if len(parts) != 2 {
		return errors.New("invalid signature format")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return err
	}

	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return err
	}

	// 验证签名
	hash := sha256.Sum256(payload)
	if err := rsa.VerifyPKCS1v15(pubkey, 0, hash[:], signature); err != nil {
		return err
	}

	return nil
}

// loadPublicKey 加载公钥
func (lm *LicenseManager) loadPublicKey() *rsa.PublicKey {
	// 尝试从文件加载
	if data, err := os.ReadFile(lm.cfg.RSAPublicKeyPath); err == nil {
		if key := parsePublicKey(data); key != nil {
			return key
		}
	}

	// 尝试从服务器获取
	if err := lm.refreshPublicKey(); err == nil {
		if data, err := os.ReadFile(lm.cfg.RSAPublicKeyPath); err == nil {
			return parsePublicKey(data)
		}
	}

	return nil
}

// refreshPublicKey 从服务器刷新公钥
func (lm *LicenseManager) refreshPublicKey() error {
	resp, err := http.Get(lm.cfg.LicenseServerURL + "/api/v1/public-key")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	pubkey, ok := result["public_key"].(string)
	if !ok || pubkey == "" {
		return errors.New("server did not return public key")
	}

	// 保存到文件
	dir := filepath.Dir(lm.cfg.RSAPublicKeyPath)
	os.MkdirAll(dir, 0755)
	return os.WriteFile(lm.cfg.RSAPublicKeyPath, []byte(pubkey), 0644)
}

// Helper functions
func parsePublicKey(data []byte) *rsa.PublicKey {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil
	}

	if pubkey, ok := key.(*rsa.PublicKey); ok {
		return pubkey
	}
	return nil
}

func getDeviceFingerprint() string {
	hostname, _ := os.Hostname()
	username := os.Getenv("USERNAME")
	if username == "" {
		username = os.Getenv("USER")
	}
	homedir, _ := os.UserHomeDir()
	raw := fmt.Sprintf("%s|%s|%s", hostname, username, homedir)
	hash := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("fp_%x", hash[:8])
}

func getDeviceName() string {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "Unknown"
	}
	return hostname
}

func maskLicenseKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}

func splitLast(s, sep string) []string {
	idx := -1
	for i := len(s) - len(sep); i >= 0; i-- {
		if s[i:i+len(sep)] == sep {
			idx = i
			break
		}
	}
	if idx == -1 {
		return []string{s}
	}
	return []string{s[:idx], s[idx+len(sep):]}
}

func jsonReader(v interface{}) io.Reader {
	b, _ := json.Marshal(v)
	return strings.NewReader(string(b))
}
