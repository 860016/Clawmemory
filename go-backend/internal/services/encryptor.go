package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

type Encryptor struct {
	gcm cipher.AEAD
}

func NewEncryptor(secretKey string) (*Encryptor, error) {
	key := deriveKey(secretKey)
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

func deriveKey(secret string) []byte {
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
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (e *Encryptor) Decrypt(encoded string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	nonceSize := e.gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := e.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func IsEncrypted(value string) bool {
	if len(value) < 4 {
		return false
	}
	prefix := value[:3]
	return prefix == "ENC"
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
