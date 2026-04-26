package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

	openclawHighlights := s.scanOpenClawDataForDate(date)
	if memoryCount == 0 && len(openclawHighlights) > 0 {
		highlights = openclawHighlights
		memoryCount = int64(len(openclawHighlights))
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

func (s *DailyReportService) scanOpenClawDataForDate(date string) []string {
	var highlights []string

	homeDir, _ := os.UserHomeDir()
	if homeDir == "" {
		return highlights
	}

	searchDirs := []string{
		filepath.Join(homeDir, ".openclaw"),
		filepath.Join(homeDir, ".trae"),
		filepath.Join(homeDir, ".trae-cn"),
	}

	for _, dir := range searchDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}

		sessionHighlights := s.scanSessionsForDate(dir, date)
		highlights = append(highlights, sessionHighlights...)

		memoryHighlights := s.scanWorkspaceMemoryForDate(dir, date)
		highlights = append(highlights, memoryHighlights...)
	}

	if len(highlights) > 10 {
		highlights = highlights[:10]
	}
	return highlights
}

func (s *DailyReportService) scanSessionsForDate(baseDir string, date string) []string {
	var highlights []string

	agentsDir := filepath.Join(baseDir, "agents")
	if info, err := os.Stat(agentsDir); err != nil || !info.IsDir() {
		return highlights
	}

	agentDirs, _ := os.ReadDir(agentsDir)
	for _, ad := range agentDirs {
		if !ad.IsDir() {
			continue
		}
		sessDir := filepath.Join(agentsDir, ad.Name(), "sessions")
		files, _ := os.ReadDir(sessDir)

		for _, f := range files {
			if f.IsDir() || !strings.HasSuffix(strings.ToLower(f.Name()), ".jsonl") {
				continue
			}
			path := filepath.Join(sessDir, f.Name())
			data, err := os.ReadFile(path)
			if err != nil || len(data) == 0 {
				continue
			}

			lines := strings.Split(string(data), "\n")
			userMessages := []string{}
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				var msg map[string]interface{}
				if json.Unmarshal([]byte(line), &msg) != nil {
					continue
				}
				msgType, _ := msg["type"].(string)
				if msgType != "user" && msgType != "human" {
					continue
				}
				text, _ := msg["text"].(string)
				if text == "" {
					text, _ = msg["content"].(string)
				}
				if text == "" {
					continue
				}
				ts, _ := msg["timestamp"].(string)
				if ts != "" && strings.HasPrefix(ts, date) {
					userMessages = append(userMessages, text)
				} else if ts == "" && len(text) > 5 {
					userMessages = append(userMessages, text)
				}
			}

			for i, msg := range userMessages {
				if i >= 5 {
					break
				}
				preview := msg
				if len(preview) > 80 {
					preview = preview[:80] + "..."
				}
				highlights = append(highlights, fmt.Sprintf("会话: %s", preview))
			}
		}
	}
	return highlights
}

func (s *DailyReportService) scanWorkspaceMemoryForDate(baseDir string, date string) []string {
	var highlights []string

	memFile := filepath.Join(baseDir, "workspace", "MEMORY.md")
	if data, err := os.ReadFile(memFile); err == nil && len(data) > 0 {
		content := string(data)
		lines := strings.Split(content, "\n")
		currentSection := ""
		for _, line := range lines {
			if strings.HasPrefix(line, "# ") {
				currentSection = strings.TrimSpace(line[2:])
			} else if currentSection != "" && strings.TrimSpace(line) != "" {
				preview := strings.TrimSpace(line)
				if len(preview) > 80 {
					preview = preview[:80] + "..."
				}
				highlights = append(highlights, fmt.Sprintf("[%s] %s", currentSection, preview))
				if len(highlights) >= 3 {
					return highlights
				}
			}
		}
	}

	memoryDir := filepath.Join(baseDir, "workspace", "memory")
	if files, _ := os.ReadDir(memoryDir); len(files) > 0 {
		for _, f := range files {
			if f.IsDir() || !strings.HasSuffix(strings.ToLower(f.Name()), ".md") {
				continue
			}
			if strings.HasPrefix(f.Name(), date) {
				path := filepath.Join(memoryDir, f.Name())
				data, err := os.ReadFile(path)
				if err != nil || len(data) == 0 {
					continue
				}
				content := string(data)
				lines := strings.Split(content, "\n")
				for _, line := range lines {
					line = strings.TrimSpace(line)
					if line != "" && !strings.HasPrefix(line, "# ") {
						if len(line) > 80 {
							line = line[:80] + "..."
						}
						highlights = append(highlights, fmt.Sprintf("记忆日志: %s", line))
						if len(highlights) >= 3 {
							return highlights
						}
					}
				}
			}
		}
	}

	return highlights
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
