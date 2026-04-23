package services

import (
	"encoding/json"
	"time"

	"clawmemory/internal/models"

	"gorm.io/gorm"
)

type DailyReportService struct {
	db *gorm.DB
}

func NewDailyReportService(db *gorm.DB) *DailyReportService {
	return &DailyReportService{db: db}
}

func (s *DailyReportService) Create(data map[string]interface{}) (*models.DailyReport, error) {
	tags := "[]"
	if t, ok := data["tags"].([]string); ok {
		b, _ := json.Marshal(t)
		tags = string(b)
	}

	report := &models.DailyReport{
		UserID:  1,
		Date:    getString(data, "date", time.Now().Format("2006-01-02")),
		Content: getString(data, "content", ""),
		Summary: getString(data, "summary", ""),
		Tags:    tags,
		Mood:    getString(data, "mood", ""),
	}

	if err := s.db.Create(report).Error; err != nil {
		return nil, err
	}
	return report, nil
}

func (s *DailyReportService) List(page, size int) ([]models.DailyReport, int64, error) {
	var reports []models.DailyReport
	var total int64

	query := s.db.Model(&models.DailyReport{}).Where("user_id = ?", 1)
	query.Count(&total)
	err := query.Order("date DESC").Offset((page - 1) * size).Limit(size).Find(&reports).Error
	return reports, total, err
}

func (s *DailyReportService) GetByDate(date string) (*models.DailyReport, error) {
	var report models.DailyReport
	if err := s.db.Where("user_id = ? AND date = ?", 1, date).First(&report).Error; err != nil {
		return nil, err
	}
	return &report, nil
}
