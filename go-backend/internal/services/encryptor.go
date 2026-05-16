package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

const (
	pbkdf2Iterations = 100000
	keyVersionV1     = "v1"
	keyVersionV2     = "v2"
)

type Encryptor struct {
	gcm cipher.AEAD
}

func NewEncryptor(secretKey string) (*Encryptor, error) {
	key := deriveKeyV2(secretKey)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &Encryptor{gcm: gcm}, nil
}

func deriveKeyV2(secret string) []byte {
	salt := getEncryptionSalt()
	return pbkdf2.Key([]byte(secret), []byte(salt), pbkdf2Iterations, 32, sha256.New)
}

func getEncryptionSalt() string {
	if salt := os.Getenv("ENCRYPTION_SALT"); salt != "" {
		return salt
	}
	secretKey := os.Getenv("SECRET_KEY")
	if secretKey != "" && secretKey != "clawmemory-default-secret-change-me" {
		h := sha256.Sum256([]byte("clawmemory-salt-" + secretKey))
		return hex.EncodeToString(h[:16])
	}
	return "clawmemory-v2-enc-salt"
}

func deriveKeyV1(secret string) []byte {
	key := make([]byte, 32)
	src := []byte(secret)
	for i := 0; i < 32; i++ {
		key[i] = src[i%len(src)] ^ byte(i)
	}
	return key
}

func (e *Encryptor) Encrypt(plaintext string) (string, error) {
	nonce := make([]byte, e.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := e.gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return keyVersionV2 + ":" + base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (e *Encryptor) Decrypt(encoded string) (string, error) {
	version := ""
	payload := encoded

	if idx := strings.Index(encoded, ":"); idx == 2 {
		version = encoded[:idx]
		payload = encoded[idx+1:]
	}

	data, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return "", err
	}

	if version == keyVersionV2 {
		return decryptWithGCM(e.gcm, data)
	}

	if version == keyVersionV1 {
		return decryptV1(data)
	}

	plaintext, err := decryptWithGCM(e.gcm, data)
	if err == nil {
		return plaintext, nil
	}

	v1Plaintext, v1Err := decryptV1(data)
	if v1Err == nil {
		return v1Plaintext, nil
	}

	return "", fmt.Errorf("decrypt failed (v2: %v, v1: %v)", err, v1Err)
}

func decryptWithGCM(gcm cipher.AEAD, data []byte) (string, error) {
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func decryptV1(data []byte) (string, error) {
	key := deriveKeyV1(os.Getenv("SECRET_KEY"))
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	return decryptWithGCM(gcm, data)
}

func IsEncrypted(value string) bool {
	return strings.HasPrefix(value, "ENC:")
}

func EncryptValue(encryptor *Encryptor, value string) (string, error) {
	encrypted, err := encryptor.Encrypt(value)
	if err != nil {
		return "", err
	}
	return "ENC:" + encrypted, nil
}

func DecryptValue(encryptor *Encryptor, value string) (string, error) {
	if !IsEncrypted(value) {
		return value, nil
	}
	encoded := value[4:]
	return encryptor.Decrypt(encoded)
}

func GetEncryptionKey() string {
	if key := os.Getenv("SECRET_KEY"); key != "" && key != "clawmemory-default-secret-change-me" {
		return key
	}
	return ""
}
