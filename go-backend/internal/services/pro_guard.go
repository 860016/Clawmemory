package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"clawmemory/internal/config"
	"clawmemory/internal/models"

	"gorm.io/gorm"
)

const proGuardVersion = "v1"

type ProGuard struct {
	proxy      *ProProxy
	db         *gorm.DB
	cfg        *config.Config
	mu         sync.RWMutex
	derivedKey []byte
	keyAt      time.Time
}

var (
	globalGuard *ProGuard
	guardOnce   sync.Once
)

func InitProGuard(proxy *ProProxy, db *gorm.DB, cfg *config.Config) *ProGuard {
	guardOnce.Do(func() {
		globalGuard = &ProGuard{
			proxy: proxy,
			db:    db,
			cfg:   cfg,
		}
	})
	return globalGuard
}

func GetProGuard() *ProGuard {
	return globalGuard
}

func (g *ProGuard) deriveProKey(license *models.License) []byte {
	g.mu.RLock()
	if g.derivedKey != nil && time.Since(g.keyAt) < 5*time.Minute {
		key := make([]byte, len(g.derivedKey))
		copy(key, g.derivedKey)
		g.mu.RUnlock()
		return key
	}
	g.mu.RUnlock()

	if license == nil {
		license = g.proxy.getActiveLicense()
	}

	if license == nil || license.LicenseKey == "" {
		return nil
	}

	salt := []byte("clawmemory-pro-guard-" + proGuardVersion)
	h := hmac.New(sha512.New384, []byte(license.LicenseKey))
	h.Write(salt)
	fp := getDeviceFingerprint()
	h.Write([]byte(fp))
	h.Write([]byte(license.Tier))
	derived := h.Sum(nil)

	aesKey := make([]byte, 32)
	copy(aesKey, derived[:32])

	g.mu.Lock()
	g.derivedKey = aesKey
	g.keyAt = time.Now()
	g.mu.Unlock()

	key := make([]byte, 32)
	copy(key, aesKey)
	return key
}

func (g *ProGuard) EncryptProData(data []byte) ([]byte, error) {
	license := g.proxy.getActiveLicense()
	if license == nil {
		return nil, fmt.Errorf("no active license")
	}

	key := g.deriveProKey(license)
	if key == nil {
		return nil, fmt.Errorf("failed to derive encryption key")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	sealed := gcm.Seal(nonce, nonce, data, nil)
	return sealed, nil
}

func (g *ProGuard) DecryptProData(data []byte) ([]byte, error) {
	license := g.proxy.getActiveLicense()
	if license == nil {
		return nil, fmt.Errorf("no active license")
	}

	key := g.deriveProKey(license)
	if key == nil {
		return nil, fmt.Errorf("failed to derive decryption key")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: license key mismatch")
	}

	return plaintext, nil
}

func (g *ProGuard) VerifyProToken(token string) bool {
	if token == "" {
		return false
	}

	license := g.proxy.getActiveLicense()
	if license == nil {
		return false
	}

	key := g.deriveProKey(license)
	if key == nil {
		return false
	}

	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(license.LicenseKey))
	mac.Write([]byte(license.Tier))
	expected := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(token), []byte(expected))
}

func (g *ProGuard) GenerateProToken() (string, error) {
	license := g.proxy.getActiveLicense()
	if license == nil {
		return "", fmt.Errorf("no active license")
	}

	key := g.deriveProKey(license)
	if key == nil {
		return "", fmt.Errorf("failed to derive key")
	}

	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(license.LicenseKey))
	mac.Write([]byte(license.Tier))
	token := hex.EncodeToString(mac.Sum(nil))

	return token, nil
}

func (g *ProGuard) IsProFeatureEnabled(feature string) bool {
	if !g.proxy.IsPro() {
		return false
	}

	license := g.proxy.getActiveLicense()
	if license == nil {
		return false
	}

	key := g.deriveProKey(license)
	if key == nil {
		return false
	}

	featureMAC := hmac.New(sha256.New, key)
	featureMAC.Write([]byte(feature))
	featureMAC.Write([]byte(license.LicenseKey))
	featureHash := featureMAC.Sum(nil)

	checkMAC := hmac.New(sha256.New, key)
	checkMAC.Write([]byte(feature))
	checkMAC.Write([]byte(license.LicenseKey))
	checkHash := checkMAC.Sum(nil)

	return hmac.Equal(featureHash, checkHash)
}

func (g *ProGuard) GetProFeatureToken(feature string) (string, error) {
	if !g.proxy.IsPro() {
		return "", fmt.Errorf("Pro license required")
	}

	license := g.proxy.getActiveLicense()
	if license == nil {
		return "", fmt.Errorf("no active license")
	}

	key := g.deriveProKey(license)
	if key == nil {
		return "", fmt.Errorf("failed to derive key")
	}

	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(feature))
	mac.Write([]byte(license.LicenseKey))
	mac.Write([]byte(license.Tier))
	token := fmt.Sprintf("%x", mac.Sum(nil))

	return token, nil
}

func (g *ProGuard) VerifyProFeatureToken(feature, token string) bool {
	if token == "" {
		return false
	}

	license := g.proxy.getActiveLicense()
	if license == nil {
		return false
	}

	key := g.deriveProKey(license)
	if key == nil {
		return false
	}

	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(feature))
	mac.Write([]byte(license.LicenseKey))
	mac.Write([]byte(license.Tier))
	expected := fmt.Sprintf("%x", mac.Sum(nil))

	return hmac.Equal([]byte(token), []byte(expected))
}

var integritySeed = []byte{0x43, 0x4c, 0x41, 0x57, 0x4d, 0x45, 0x4d}

func (g *ProGuard) SelfCheck() bool {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return true
	}

	var vcsHash string
	for _, s := range buildInfo.Settings {
		if s.Key == "vcs.revision" {
			vcsHash = s.Value
			break
		}
	}

	if vcsHash == "" {
		return true
	}

	exePath, err := os.Executable()
	if err != nil {
		return true
	}

	exeInfo, err := os.Stat(exePath)
	if err != nil {
		return true
	}

	sizeBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(sizeBytes, uint64(exeInfo.Size()))

	modBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(modBytes, uint64(exeInfo.ModTime().Unix()))

	h := sha256.New()
	h.Write(integritySeed)
	h.Write([]byte(vcsHash))
	h.Write(sizeBytes)
	h.Write(modBytes)
	h.Write([]byte(proGuardVersion))
	checksum := fmt.Sprintf("%x", h.Sum(nil))[:16]

	checksumPath := g.getChecksumPath()
	if _, err := os.Stat(checksumPath); os.IsNotExist(err) {
		os.WriteFile(checksumPath, []byte(checksum), 0600)
		return true
	}

	stored, err := os.ReadFile(checksumPath)
	if err != nil {
		return true
	}

	storedStr := strings.TrimSpace(string(stored))
	if storedStr == "" {
		os.WriteFile(checksumPath, []byte(checksum), 0600)
		return true
	}

	if storedStr != checksum {
		return false
	}

	return true
}

func (g *ProGuard) getChecksumPath() string {
	dataDir := filepath.Dir(g.cfg.RSAPublicKeyPath)
	if dataDir == "" || dataDir == "." {
		dataDir = "data"
	}
	keysDir := filepath.Join(dataDir, "keys")
	os.MkdirAll(keysDir, 0700)
	return filepath.Join(keysDir, ".integrity")
}

func (g *ProGuard) InvalidateDerivedKey() {
	g.mu.Lock()
	g.derivedKey = nil
	g.keyAt = time.Time{}
	g.mu.Unlock()
}
