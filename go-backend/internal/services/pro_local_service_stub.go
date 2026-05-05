//go:build !pro

package services

import (
	"gorm.io/gorm"
)

type ProLocalService struct {
	db *gorm.DB
}

func NewProLocalService(db *gorm.DB) *ProLocalService {
	return &ProLocalService{db: db}
}

func (s *ProLocalService) DecayStats(userID uint) (map[string]interface{}, error) {
	return nil, errProNotAvailable
}

func (s *ProLocalService) DecayApply(userID uint) (map[string]interface{}, error) {
	return nil, errProNotAvailable
}

func (s *ProLocalService) PruneSuggest(userID uint) (map[string]interface{}, error) {
	return nil, errProNotAvailable
}

func (s *ProLocalService) ConflictScan(userID uint) (map[string]interface{}, error) {
	return nil, errProNotAvailable
}

func (s *ProLocalService) CompressPreview(userID uint, level string) (map[string]interface{}, error) {
	return nil, errProNotAvailable
}

func (s *ProLocalService) CompressApply(userID uint, level string) (map[string]interface{}, error) {
	return nil, errProNotAvailable
}

func (s *ProLocalService) EvolutionDiscover(userID uint) (map[string]interface{}, error) {
	return nil, errProNotAvailable
}

func (s *ProLocalService) EvolutionPrefetch(userID uint, context string) (map[string]interface{}, error) {
	return nil, errProNotAvailable
}

func (s *ProLocalService) EvolutionImportance(userID uint) (map[string]interface{}, error) {
	return nil, errProNotAvailable
}

func (s *ProLocalService) ReinforceMemory(userID uint, memoryID uint) (map[string]interface{}, error) {
	return nil, errProNotAvailable
}

func (s *ProLocalService) ConflictResolve(userID uint, conflictIndex int, strategy string) (map[string]interface{}, error) {
	return nil, errProNotAvailable
}

func (s *ProLocalService) TokenRoute(message string, contextLength int) (map[string]interface{}, error) {
	return nil, errProNotAvailable
}

func (s *ProLocalService) TokenStats(userID uint) (map[string]interface{}, error) {
	return nil, errProNotAvailable
}

func (s *ProLocalService) AIExtract(userID uint) (map[string]interface{}, error) {
	return nil, errProNotAvailable
}

func (s *ProLocalService) AutoGraph(userID uint, overwrite bool) (map[string]interface{}, error) {
	return nil, errProNotAvailable
}

func (s *ProLocalService) BackupSchedule(userID uint) (map[string]interface{}, error) {
	return nil, errProNotAvailable
}

func (s *ProLocalService) SetBackupSchedule(userID uint, enabled bool, intervalHours int) (map[string]interface{}, error) {
	return nil, errProNotAvailable
}

func (s *ProLocalService) CompressConfig(userID uint) (map[string]interface{}, error) {
	return nil, errProNotAvailable
}

func (s *ProLocalService) SetCompressConfig(userID uint, config map[string]interface{}) (map[string]interface{}, error) {
	return nil, errProNotAvailable
}

func (s *ProLocalService) EvolutionInsights(userID uint) (map[string]interface{}, error) {
	return nil, errProNotAvailable
}

func (s *ProLocalService) EvolutionInfer(userID uint) (map[string]interface{}, error) {
	return nil, errProNotAvailable
}
