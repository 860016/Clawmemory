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
		return nil, errors.New("无法连接授权服务器")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("授权服务器返回错误")
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result["valid"].(bool) {
		return result, nil
	}

	// RSA 签名验证
	if signature, ok := result["signature"].(string); ok && signature != "" {
		if err := lm.verifySignature(signature); err != nil {
			// 尝试刷新公钥重试
			if err := lm.refreshPublicKey(); err == nil {
				if err := lm.verifySignature(signature); err != nil {
					return nil, errors.New("RSA 签名验证失败")
				}
			}
		}
	}

	// 保存到数据库
	features := []string{}
	if f, ok := result["features"].([]interface{}); ok {
		for _, feature := range f {
			features = append(features, feature.(string))
		}
	}
	featuresJSON, _ := json.Marshal(features)

	fallbackURLs := []string{}
	if f, ok := result["pro_fallback_urls"].([]interface{}); ok {
		for _, url := range f {
			fallbackURLs = append(fallbackURLs, url.(string))
		}
	}
	fallbackJSON, _ := json.Marshal(fallbackURLs)

	// 停用旧授权
	lm.db.Model(&models.License{}).Where("status = ?", "active").Update("status", "inactive")

	// 创建新授权
	license := &models.License{
		LicenseKey:        licenseKey,
		Tier:              getString(result, "tier", "pro"),
		Status:            "active",
		DeviceFingerprint: fingerprint,
		DeviceName:        deviceName,
		Features:          string(featuresJSON),
		ProDownloadURL:    getString(result, "pro_download_url", ""),
		ProFallbackURLs:   string(fallbackJSON),
	}

	if exp, ok := result["expires_at"].(string); ok {
		t, _ := time.Parse(time.RFC3339, exp)
		license.ExpiresAt = &t
	}

	if err := lm.db.Create(license).Error; err != nil {
		return nil, err
	}

	// 更新内存状态
	lm.tier = license.Tier
	lm.features = features

	return map[string]interface{}{
		"valid":             true,
		"message":           "激活成功",
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

	return map[string]interface{}{
		"active":            license.Status == "active",
		"tier":              lm.tier,
		"type":              license.Tier,
		"features":          activeFeatures,
		"expires_at":        license.ExpiresAt,
		"device_slot":       license.DeviceSlot,
		"license_key":       maskLicenseKey(license.LicenseKey),
		"is_valid":          true,
		"pro_download_url":  license.ProDownloadURL,
		"pro_fallback_urls": []string{},
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
		return errors.New("无法加载公钥")
	}

	// 解析签名
	parts := splitLast(signatureB64, ".")
	if len(parts) != 2 {
		return errors.New("签名格式错误")
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
		return errors.New("服务器未返回公钥")
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
	// 简化实现，实际应该使用机器码等
	hostname, _ := os.Hostname()
	return fmt.Sprintf("fp_%s_%d", hostname, time.Now().Unix())
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
