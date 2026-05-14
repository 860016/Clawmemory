package proprovider

import (
	"fmt"

	"gorm.io/gorm"
)

type ProProvider interface {
	IsPro() bool
	GetLicenseInfo() map[string]interface{}
	InvalidateCache()
	IsFeatureEnabled(feature string) bool
	SelfCheck() bool
	GetTier() string

	ActivateLicense(licenseKey string) (map[string]interface{}, error)
	DeactivateLicense() error
	Heartbeat() error

	DecayStats(userID uint) (map[string]interface{}, error)
	DecayApply(userID uint) (map[string]interface{}, error)
	PruneSuggest(userID uint) (map[string]interface{}, error)

	ConflictScan(userID uint) (map[string]interface{}, error)
	ConflictResolve(userID uint, conflictIndex int, strategy string) (map[string]interface{}, error)

	CompressPreview(userID uint, level string) (map[string]interface{}, error)
	CompressApply(userID uint, level string) (map[string]interface{}, error)
	CompressConfig(userID uint) (map[string]interface{}, error)
	SetCompressConfig(userID uint, cfg map[string]interface{}) (map[string]interface{}, error)

	TokenRoute(message string, contextLength int) (map[string]interface{}, error)
	TokenStats(userID uint) (map[string]interface{}, error)

	EvolutionInsights(userID uint) (map[string]interface{}, error)
	EvolutionDiscover(userID uint) (map[string]interface{}, error)
	EvolutionInfer(userID uint) (map[string]interface{}, error)
	EvolutionImportance(userID uint) (map[string]interface{}, error)
	EvolutionPrefetch(userID uint, context string) (map[string]interface{}, error)

	ReinforceMemory(userID uint, memoryID uint) (map[string]interface{}, error)
	AutoGraph(userID uint, overwrite bool) (map[string]interface{}, error)
	AIExtract(userID uint) (map[string]interface{}, error)

	BackupSchedule(userID uint) (map[string]interface{}, error)
	SetBackupSchedule(userID uint, enabled bool, intervalHours int) (map[string]interface{}, error)

	SmartLoad(userID uint, context string) (map[string]interface{}, error)
}

type ProError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func (e *ProError) Error() string {
	return e.Message
}

var ErrProRequired = &ProError{
	Message: "This feature requires a Pro license",
	Code:    403,
}

type Config struct {
	LicenseServerURL string
	RSAPublicKeyPath string
	DataDir          string
}

func NewConfig(licenseServerURL, rsaPublicKeyPath, dataDir string) *Config {
	return &Config{
		LicenseServerURL: licenseServerURL,
		RSAPublicKeyPath: rsaPublicKeyPath,
		DataDir:          dataDir,
	}
}

type StubProProvider struct{}

func NewProProvider(db *gorm.DB, cfg *Config) ProProvider {
	return &StubProProvider{}
}

func (s *StubProProvider) IsPro() bool { return false }

func (s *StubProProvider) GetLicenseInfo() map[string]interface{} {
	return map[string]interface{}{
		"active":   false,
		"tier":     "oss",
		"is_valid": false,
		"features": []string{},
	}
}

func (s *StubProProvider) InvalidateCache() {}

func (s *StubProProvider) IsFeatureEnabled(feature string) bool { return false }

func (s *StubProProvider) SelfCheck() bool { return true }

func (s *StubProProvider) GetTier() string { return "oss" }

func (s *StubProProvider) ActivateLicense(licenseKey string) (map[string]interface{}, error) {
	return nil, fmt.Errorf("Pro features are not available in the open-source version")
}

func (s *StubProProvider) DeactivateLicense() error {
	return fmt.Errorf("Pro features are not available in the open-source version")
}

func (s *StubProProvider) Heartbeat() error { return nil }

func (s *StubProProvider) DecayStats(userID uint) (map[string]interface{}, error) {
	return nil, ErrProRequired
}

func (s *StubProProvider) DecayApply(userID uint) (map[string]interface{}, error) {
	return nil, ErrProRequired
}

func (s *StubProProvider) PruneSuggest(userID uint) (map[string]interface{}, error) {
	return nil, ErrProRequired
}

func (s *StubProProvider) ConflictScan(userID uint) (map[string]interface{}, error) {
	return nil, ErrProRequired
}

func (s *StubProProvider) ConflictResolve(userID uint, conflictIndex int, strategy string) (map[string]interface{}, error) {
	return nil, ErrProRequired
}

func (s *StubProProvider) CompressPreview(userID uint, level string) (map[string]interface{}, error) {
	return nil, ErrProRequired
}

func (s *StubProProvider) CompressApply(userID uint, level string) (map[string]interface{}, error) {
	return nil, ErrProRequired
}

func (s *StubProProvider) CompressConfig(userID uint) (map[string]interface{}, error) {
	return nil, ErrProRequired
}

func (s *StubProProvider) SetCompressConfig(userID uint, cfg map[string]interface{}) (map[string]interface{}, error) {
	return nil, ErrProRequired
}

func (s *StubProProvider) TokenRoute(message string, contextLength int) (map[string]interface{}, error) {
	return nil, ErrProRequired
}

func (s *StubProProvider) TokenStats(userID uint) (map[string]interface{}, error) {
	return nil, ErrProRequired
}

func (s *StubProProvider) EvolutionInsights(userID uint) (map[string]interface{}, error) {
	return nil, ErrProRequired
}

func (s *StubProProvider) EvolutionDiscover(userID uint) (map[string]interface{}, error) {
	return nil, ErrProRequired
}

func (s *StubProProvider) EvolutionInfer(userID uint) (map[string]interface{}, error) {
	return nil, ErrProRequired
}

func (s *StubProProvider) EvolutionImportance(userID uint) (map[string]interface{}, error) {
	return nil, ErrProRequired
}

func (s *StubProProvider) EvolutionPrefetch(userID uint, context string) (map[string]interface{}, error) {
	return nil, ErrProRequired
}

func (s *StubProProvider) ReinforceMemory(userID uint, memoryID uint) (map[string]interface{}, error) {
	return nil, ErrProRequired
}

func (s *StubProProvider) AutoGraph(userID uint, overwrite bool) (map[string]interface{}, error) {
	return nil, ErrProRequired
}

func (s *StubProProvider) AIExtract(userID uint) (map[string]interface{}, error) {
	return nil, ErrProRequired
}

func (s *StubProProvider) BackupSchedule(userID uint) (map[string]interface{}, error) {
	return nil, ErrProRequired
}

func (s *StubProProvider) SetBackupSchedule(userID uint, enabled bool, intervalHours int) (map[string]interface{}, error) {
	return nil, ErrProRequired
}

func (s *StubProProvider) SmartLoad(userID uint, context string) (map[string]interface{}, error) {
	return nil, ErrProRequired
}
