package services

import (
	"encoding/json"
	"fmt"
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

func (s *DailyReportService) Create(userID uint, data map[string]interface{}) (*models.DailyReport, error) {
	tags := "[]"
	if t, ok := data["tags"].([]string); ok {
		b, _ := json.Marshal(t)
		tags = string(b)
	}

	reportDate := getString(data, "report_date", "")
	date := getString(data, "date", time.Now().Format("2006-01-02"))
	if reportDate == "" {
		reportDate = date
	}

	highlights := toJSONStr(data["highlights"])
	knowledgeGained := toJSONStr(data["knowledge_gained"])
	pendingTasks := toJSONStr(data["pending_tasks"])
	tomorrowSuggestions := toJSONStr(data["tomorrow_suggestions"])
	stats := toJSONStr(data["stats"])

	report := &models.DailyReport{
		UserID:              userID,
		Date:                date,
		ReportDate:          reportDate,
		Content:             getString(data, "content", ""),
		Summary:             getString(data, "summary", ""),
		Highlights:          highlights,
		KnowledgeGained:     knowledgeGained,
		PendingTasks:        pendingTasks,
		TomorrowSuggestions: tomorrowSuggestions,
		Stats:               stats,
		Tags:                tags,
		Mood:                getString(data, "mood", ""),
	}

	if err := s.db.Create(report).Error; err != nil {
		return nil, err
	}
	return report, nil
}

func (s *DailyReportService) List(userID uint, page, size int) ([]models.DailyReport, int64, error) {
	var reports []models.DailyReport
	var total int64

	query := s.db.Model(&models.DailyReport{}).Where("user_id = ?", userID)
	query.Count(&total)
	err := query.Order("date DESC").Offset((page - 1) * size).Limit(size).Find(&reports).Error
	return reports, total, err
}

func (s *DailyReportService) GetByDate(userID uint, date string) (*models.DailyReport, error) {
	var report models.DailyReport
	if err := s.db.Where("user_id = ? AND date = ?", userID, date).First(&report).Error; err != nil {
		return nil, err
	}
	return &report, nil
}

func (s *DailyReportService) Generate(userID uint, date string) (*models.DailyReport, error) {
	existing, err := s.GetByDate(userID, date)
	if err == nil && existing != nil {
		return existing, nil
	}

	var memoryCount, entityCount, wikiCount int64
	s.db.Model(&models.Memory{}).Where("user_id = ? AND DATE(created_at) = ? AND status != ?", userID, date, "trashed").Count(&memoryCount)
	s.db.Model(&models.Entity{}).Where("user_id = ? AND DATE(created_at) = ?", userID, date).Count(&entityCount)
	s.db.Model(&models.WikiPage{}).Where("user_id = ? AND DATE(updated_at) = ?", userID, date).Count(&wikiCount)

	var memories []models.Memory
	s.db.Where("user_id = ? AND DATE(created_at) = ? AND status != ?", userID, date, "trashed").
		Order("importance DESC").Limit(10).Find(&memories)

	highlights := []string{}
	for _, m := range memories {
		if m.Importance >= 0.7 {
			highlights = append(highlights, fmt.Sprintf("%s: %s", m.Key, truncateStr(m.Value, 80)))
		}
	}
	if len(highlights) == 0 && len(memories) > 0 {
		for _, m := range memories[:minInt(3, len(memories))] {
			highlights = append(highlights, fmt.Sprintf("%s: %s", m.Key, truncateStr(m.Value, 80)))
		}
	}

	stats := map[string]interface{}{
		"new_memories":  memoryCount,
		"new_entities":  entityCount,
		"updated_wiki":  wikiCount,
		"active_hours":  0,
	}

	summary := fmt.Sprintf("记录了 %d 条新记忆，创建了 %d 个知识实体，更新了 %d 个 Wiki 页面。", memoryCount, entityCount, wikiCount)

	report := &models.DailyReport{
		UserID:              userID,
		Date:                date,
		ReportDate:          date,
		Summary:             summary,
		Highlights:          toJSONStrFromSlice(highlights),
		KnowledgeGained:     "[]",
		PendingTasks:        "[]",
		TomorrowSuggestions: "[]",
		Stats:               toJSONStr(stats),
		Tags:                "[]",
	}

	if err := s.db.Create(report).Error; err != nil {
		return nil, err
	}
	return report, nil
}

func toJSONStr(v interface{}) string {
	if v == nil {
		return "[]"
	}
	b, err := json.Marshal(v)
	if err != nil {
		return "[]"
	}
	return string(b)
}

func toJSONStrFromSlice(v []string) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "[]"
	}
	return string(b)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
