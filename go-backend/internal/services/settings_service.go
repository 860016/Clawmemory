package services

import (
	"encoding/json"

	"clawmemory/internal/models"

	"gorm.io/gorm"
)

type SettingsService struct {
	db *gorm.DB
}

func NewSettingsService(db *gorm.DB) *SettingsService {
	return &SettingsService{db: db}
}

func (s *SettingsService) Get(userID uint) (map[string]interface{}, error) {
	var settings []models.Setting
	s.db.Where("user_id = ?", userID).Find(&settings)

	result := map[string]interface{}{
		"language":     "zh-CN",
		"theme":        "light",
		"auto_decay":   false,
		"decay_config": "{}",
	}
	for _, setting := range settings {
		var value interface{}
		if err := json.Unmarshal([]byte(setting.Value), &value); err != nil {
			result[setting.Key] = setting.Value
		} else {
			result[setting.Key] = value
		}
	}
	return result, nil
}

func (s *SettingsService) Update(userID uint, data map[string]interface{}) error {
	for key, val := range data {
		valueBytes, _ := json.Marshal(val)
		var setting models.Setting
		result := s.db.Where("user_id = ? AND key = ?", userID, key).First(&setting)
		if result.Error == gorm.ErrRecordNotFound {
			s.db.Create(&models.Setting{
				UserID: userID,
				Key:    key,
				Value:  string(valueBytes),
			})
		} else {
			s.db.Model(&setting).Update("value", string(valueBytes))
		}
	}
	return nil
}

func (s *SettingsService) GetByKey(userID uint, key string) (interface{}, error) {
	var setting models.Setting
	if err := s.db.Where("user_id = ? AND key = ?", userID, key).First(&setting).Error; err != nil {
		return nil, err
	}
	var value interface{}
	json.Unmarshal([]byte(setting.Value), &value)
	return value, nil
}

func (s *SettingsService) SetByKey(userID uint, key string, value interface{}) error {
	valueBytes, _ := json.Marshal(value)
	var setting models.Setting
	result := s.db.Where("user_id = ? AND key = ?", userID, key).First(&setting)
	if result.Error == gorm.ErrRecordNotFound {
		return s.db.Create(&models.Setting{
			UserID: userID,
			Key:    key,
			Value:  string(valueBytes),
		}).Error
	}
	return s.db.Model(&setting).Update("value", string(valueBytes)).Error
}
