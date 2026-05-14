package services

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"clawmemory/internal/models"

	"gorm.io/gorm"
)

type SecurityService struct {
	db        *gorm.DB
	csrfKey   []byte
	noncePool sync.Pool
}

var (
	globalSecurity *SecurityService
	securityOnce   sync.Once
)

func InitSecurity(db *gorm.DB, jwtSecret string) *SecurityService {
	securityOnce.Do(func() {
		h := hmac.New(sha256.New, []byte("clawmemory-csrf-"+jwtSecret))
		h.Write([]byte(time.Now().Format("2006-01-02")))
		csrfKey := h.Sum(nil)

		globalSecurity = &SecurityService{
			db:      db,
			csrfKey: csrfKey,
			noncePool: sync.Pool{
				New: func() interface{} {
					b := make([]byte, 16)
					return b
				},
			},
		}
	})
	return globalSecurity
}

func GetSecurity() *SecurityService {
	return globalSecurity
}

func (s *SecurityService) GenerateCSRFToken(userID uint) string {
	timestamp := time.Now().Unix()
	payload := fmt.Sprintf("%d:%d", userID, timestamp)

	h := hmac.New(sha256.New, s.csrfKey)
	h.Write([]byte(payload))
	sig := h.Sum(nil)

	token := fmt.Sprintf("%s:%s", payload, hex.EncodeToString(sig))
	return token
}

func (s *SecurityService) ValidateCSRFToken(token string, userID uint) bool {
	var payloadStr, sigHex string
	n, _ := fmt.Sscanf(token, "%s:%s", &payloadStr, &sigHex)
	if n != 2 {
		return false
	}

	expectedPayload := fmt.Sprintf("%d:", userID)
	if len(payloadStr) < len(expectedPayload) || payloadStr[:len(expectedPayload)] != expectedPayload {
		return false
	}

	var timestamp int64
	fmt.Sscanf(payloadStr[len(expectedPayload):], "%d", &timestamp)
	if timestamp == 0 || time.Now().Unix()-timestamp > 3600 {
		return false
	}

	h := hmac.New(sha256.New, s.csrfKey)
	h.Write([]byte(payloadStr))
	expectedSig := h.Sum(nil)

	sig, err := hex.DecodeString(sigHex)
	if err != nil {
		return false
	}

	return hmac.Equal(sig, expectedSig)
}

func DeriveKey(secret string, purpose string, length int) []byte {
	prk := hmac.New(sha256.New, []byte("clawmemory-salt-"+purpose))
	prk.Write([]byte(secret))
	extracted := prk.Sum(nil)

	okm := make([]byte, 0, length)
	counter := byte(1)
	prev := []byte{}
	for len(okm) < length {
		h := hmac.New(sha256.New, extracted)
		h.Write(prev)
		h.Write([]byte(purpose))
		h.Write([]byte{counter})
		prev = h.Sum(nil)
		okm = append(okm, prev...)
		counter++
		if counter == 0 {
			break
		}
	}

	if length > len(okm) {
		length = len(okm)
	}
	return okm[:length]
}

func GenerateSecureToken(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s *SecurityService) GenerateNonce() string {
	b := s.noncePool.Get().([]byte)
	rand.Read(b)
	nonce := hex.EncodeToString(b)
	s.noncePool.Put(b)
	return nonce
}

type AuditEntry struct {
	UserID    uint   `json:"user_id"`
	Action    string `json:"action"`
	Target    string `json:"target"`
	Detail    string `json:"detail"`
	IP        string `json:"ip"`
	UserAgent string `json:"user_agent"`
	AgentName string `json:"agent_name"`
}

func (s *SecurityService) LogAudit(entry AuditEntry) {
	if s.db == nil {
		return
	}

	auditLog := &models.AuditLog{
		UserID:    entry.UserID,
		Action:    entry.Action,
		Target:    entry.Target,
		Detail:    entry.Detail,
		IP:        entry.IP,
		UserAgent: entry.UserAgent,
		AgentName: entry.AgentName,
	}
	s.db.Create(auditLog)
}

func (s *SecurityService) GetAuditLog(userID uint, limit, offset int) ([]map[string]interface{}, int64) {
	var total int64
	s.db.Table("audit_logs").Where("user_id = ?", userID).Count(&total)

	var logs []map[string]interface{}
	s.db.Table("audit_logs").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs)

	return logs, total
}

func ConstantTimeEqual(a, b string) bool {
	return hmac.Equal([]byte(a), []byte(b))
}
