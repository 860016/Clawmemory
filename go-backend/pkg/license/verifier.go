package license

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"time"
)

// LicenseData 授权数据结构
type LicenseData struct {
	LicenseKey string    `json:"license_key"`
	Tier       string    `json:"tier"`
	DeviceFP   string    `json:"device_fp"`
	ExpiresAt  time.Time `json:"expires_at"`
	Features   []string  `json:"features"`
	IssuedAt   time.Time `json:"issued_at"`
}

// Verifier 授权验证器
type Verifier struct {
	publicKey *rsa.PublicKey
}

// NewVerifier 创建验证器
func NewVerifier(publicKeyPEM string) (*Verifier, error) {
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return nil, errors.New("invalid PEM block")
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	pubKey, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}

	return &Verifier{publicKey: pubKey}, nil
}

// Verify 验证授权签名
func (v *Verifier) Verify(signedData string) (*LicenseData, error) {
	// 解析数据格式: base64(payload).base64(signature)
	parts := splitLast(signedData, ".")
	if len(parts) != 2 {
		return nil, errors.New("invalid signed data format")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("decode payload: %w", err)
	}

	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decode signature: %w", err)
	}

	// 验证签名
	hash := sha256.Sum256(payloadBytes)
	if err := rsa.VerifyPKCS1v15(v.publicKey, 0, hash[:], signature); err != nil {
		return nil, fmt.Errorf("signature verification failed: %w", err)
	}

	// 解析授权数据
	var data LicenseData
	if err := json.Unmarshal(payloadBytes, &data); err != nil {
		return nil, fmt.Errorf("unmarshal license data: %w", err)
	}

	// 检查是否过期
	if time.Now().After(data.ExpiresAt) {
		return nil, errors.New("license expired")
	}

	return &data, nil
}

// Encrypt 加密数据（用于 Pro 模块加密）
func Encrypt(plaintext []byte, key []byte) ([]byte, error) {
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

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// Decrypt 解密数据
func Decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// GenerateLicenseKey 生成授权密钥
func GenerateLicenseKey() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("CM-%s-%s-%s-%s",
		base64.RawURLEncoding.EncodeToString(b[:4]),
		base64.RawURLEncoding.EncodeToString(b[4:8]),
		base64.RawURLEncoding.EncodeToString(b[8:12]),
		base64.RawURLEncoding.EncodeToString(b[12:16]),
	)
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
