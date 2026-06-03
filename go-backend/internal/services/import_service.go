package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"clawmemory/internal/models"

	"gorm.io/gorm"
)

type ImportService struct {
	db *gorm.DB
}

func NewImportService(db *gorm.DB) *ImportService {
	return &ImportService{db: db}
}

type ImportResult struct {
	Imported       int      `json:"imported"`
	Skipped        int      `json:"skipped"`
	EntitiesCreated int     `json:"entities_created"`
	FoundFiles     []string `json:"files_found"`
}

// AutoImport scans known directories and imports memories from files.
func (s *ImportService) AutoImport(userID uint) *ImportResult {
	searchDirs := GetAllSearchDirs()
	result := &ImportResult{FoundFiles: []string{}}
	seenKeys := s.loadExistingKeys(userID)

	for _, dir := range searchDirs {
		memFile := filepath.Join(dir, "MEMORY.md")
		if data, err := os.ReadFile(memFile); err == nil && len(data) > 0 {
			result.FoundFiles = append(result.FoundFiles, memFile)
			s.importFromMarkdown(userID, memFile, string(data), seenKeys, result)
		}

		memoryDir := filepath.Join(dir, "memory")
		if files, err := os.ReadDir(memoryDir); err == nil {
			for _, f := range files {
				if f.IsDir() {
					continue
				}
				ext := strings.ToLower(filepath.Ext(f.Name()))
				if ext != ".md" && ext != ".txt" {
					continue
				}
				path := filepath.Join(memoryDir, f.Name())
				data, err := os.ReadFile(path)
				if err != nil || len(data) == 0 {
					continue
				}
				result.FoundFiles = append(result.FoundFiles, path)
				content := string(data)
				if ext == ".md" {
					s.importFromMarkdown(userID, path, content, seenKeys, result)
				} else if ext == ".txt" {
					s.importFromText(userID, path, content, seenKeys, result)
				}
			}
		}
	}

	return result
}

func (s *ImportService) loadExistingKeys(userID uint) map[string]bool {
	seenKeys := make(map[string]bool)
	var existingKeys []string
	s.db.Table("memories").Where("user_id = ? AND status != ?", userID, "trashed").Pluck("key", &existingKeys)
	for _, k := range existingKeys {
		seenKeys[k] = true
	}
	return seenKeys
}

func (s *ImportService) importFromJSON(userID uint, content string, seenKeys map[string]bool, result *ImportResult) {
	var memories []map[string]interface{}
	if json.Unmarshal([]byte(content), &memories) != nil {
		var single map[string]interface{}
		if json.Unmarshal([]byte(content), &single) != nil {
			return
		}
		memories = []map[string]interface{}{single}
	}

	for _, m := range memories {
		key, _ := m["key"].(string)
		contentStr, _ := m["content"].(string)
		if key == "" {
			if name, ok := m["name"].(string); ok {
				key = name
			} else if title, ok := m["title"].(string); ok {
				key = title
			}
		}
		if contentStr == "" {
			contentStr, _ = m["value"].(string)
			if contentStr == "" {
				contentStr, _ = m["text"].(string)
				if contentStr == "" {
					contentStr, _ = m["description"].(string)
				}
			}
		}
		if key == "" || contentStr == "" {
			result.Skipped++
			continue
		}

		if seenKeys[key] {
			result.Skipped++
			continue
		}

		layer := classifyLayer(key, contentStr)
		importance := 0.5
		if imp, ok := m["importance"].(float64); ok {
			importance = imp
		}

		tags := extractTags(m)

		source := "auto_import"
		if src, ok := m["source"].(string); ok && src != "" {
			source = src
		}

		memSvc := NewMemoryService(s.db)
		_, err := memSvc.Create(userID, map[string]interface{}{
			"key":        key,
			"value":      contentStr,
			"layer":      layer,
			"importance": importance,
			"tags":       tags,
			"source":     source,
		})
		if err != nil {
			result.Skipped++
			continue
		}
		seenKeys[key] = true
		result.Imported++

		s.tryCreateEntity(userID, key, contentStr, result)
	}
}

func (s *ImportService) importFromMarkdown(userID uint, filePath, content string, seenKeys map[string]bool, result *ImportResult) {
	sections := strings.Split(content, "\n## ")
	for i, section := range sections {
		var key, body string
		if i == 0 {
			lines := strings.SplitN(section, "\n", 2)
			key = strings.TrimPrefix(strings.TrimSpace(lines[0]), "# ")
			if len(lines) > 1 {
				body = strings.TrimSpace(lines[1])
			}
		} else {
			lines := strings.SplitN(section, "\n", 2)
			key = strings.TrimSpace(lines[0])
			if len(lines) > 1 {
				body = strings.TrimSpace(lines[1])
			}
		}

		if key == "" || body == "" {
			continue
		}

		key = fmt.Sprintf("md:%s", key)
		if seenKeys[key] {
			result.Skipped++
			continue
		}

		layer := classifyLayer(key, body)
		importance := 0.6

		source := "auto_import_md"
		relPath := filePath
		if len(relPath) > 100 {
			relPath = "..." + relPath[len(relPath)-97:]
		}

		memSvc := NewMemoryService(s.db)
		_, err := memSvc.Create(userID, map[string]interface{}{
			"key":        key,
			"value":      body,
			"layer":      layer,
			"importance": importance,
			"tags":       "markdown",
			"source":     source,
		})
		if err != nil {
			result.Skipped++
			continue
		}
		seenKeys[key] = true
		result.Imported++

		s.tryCreateEntity(userID, key, body, result)
	}
}

func (s *ImportService) importFromText(userID uint, filePath, content string, seenKeys map[string]bool, result *ImportResult) {
	lines := strings.Split(content, "\n")
	var buffer []string
	var currentKey string

	flushBuffer := func() {
		if currentKey != "" && len(buffer) > 0 {
			body := strings.Join(buffer, "\n")
			key := fmt.Sprintf("txt:%s", currentKey)

			if !seenKeys[key] {
				layer := classifyLayer(key, body)
				memSvc := NewMemoryService(s.db)
				_, err := memSvc.Create(userID, map[string]interface{}{
					"key":        key,
					"value":      body,
					"layer":      layer,
					"importance": 0.4,
					"tags":       "text",
					"source":     "auto_import_txt",
				})
				if err == nil {
					seenKeys[key] = true
					result.Imported++
					s.tryCreateEntity(userID, key, body, result)
				} else {
					result.Skipped++
				}
			} else {
				result.Skipped++
			}
		}
		buffer = nil
		currentKey = ""
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if len(trimmed) < 80 && !strings.HasPrefix(trimmed, "-") && !strings.HasPrefix(trimmed, "*") && strings.HasSuffix(trimmed, ":") {
			flushBuffer()
			currentKey = strings.TrimSuffix(trimmed, ":")
		} else if len(trimmed) < 60 && !strings.HasPrefix(trimmed, "-") && !strings.HasPrefix(trimmed, "*") && currentKey == "" {
			flushBuffer()
			currentKey = trimmed
		} else {
			buffer = append(buffer, trimmed)
		}
	}
	flushBuffer()

	if result.Imported == 0 && len(strings.Fields(content)) > 5 {
		key := fmt.Sprintf("txt:%s", filepath.Base(filePath))
		key = strings.TrimSuffix(key, filepath.Ext(key))
		if !seenKeys[key] && len(content) > 10 {
			layer := classifyLayer(key, content)
			memSvc := NewMemoryService(s.db)
			_, err := memSvc.Create(userID, map[string]interface{}{
				"key":        key,
				"value":      content,
				"layer":      layer,
				"importance": 0.3,
				"tags":       "text",
				"source":     "auto_import_txt",
			})
			if err == nil {
				seenKeys[key] = true
				result.Imported++
				s.tryCreateEntity(userID, key, content, result)
			}
		}
	}
}

func (s *ImportService) tryCreateEntity(userID uint, key, content string, result *ImportResult) {
	if len(content) < 10 || len(content) > 2000 {
		return
	}

	entityType := "concept"
	lowerContent := strings.ToLower(content)
	if strings.Contains(lowerContent, "项目") || strings.Contains(lowerContent, "project") {
		entityType = "organization"
	} else if strings.Contains(lowerContent, "工具") || strings.Contains(lowerContent, "tool") || strings.Contains(lowerContent, "软件") || strings.Contains(lowerContent, "software") {
		entityType = "technology"
	} else if strings.Contains(lowerContent, "人") || strings.Contains(lowerContent, "person") || strings.Contains(lowerContent, "用户") {
		entityType = "person"
	} else if strings.Contains(lowerContent, "地点") || strings.Contains(lowerContent, "location") || strings.Contains(lowerContent, "城市") {
		entityType = "location"
	} else if strings.Contains(lowerContent, "事件") || strings.Contains(lowerContent, "event") {
		entityType = "event"
	}

	name := key
	if strings.HasPrefix(name, "md:") || strings.HasPrefix(name, "txt:") {
		name = name[3:]
	}
	if len(name) > 50 {
		name = name[:50]
	}

	var entityCount int64
	logDBErr("count entities by name", s.db.Table("entities").Where("user_id = ? AND name = ?", userID, name).Count(&entityCount).Error)
	if entityCount == 0 {
		entity := models.Entity{
			UserID:        userID,
			Name:          name,
			EntityType:    entityType,
			Description:   content,
			Confidence:    0.7,
			ExtractMethod: "auto_import",
		}
		if s.db.Create(&entity).Error == nil {
			result.EntitiesCreated++
		}
	}
}

// classifyLayer determines the memory layer based on key and content keywords.
func classifyLayer(key, content string) string {
	lowerKey := strings.ToLower(key)
	lowerContent := strings.ToLower(content)

	if strings.Contains(lowerKey, "偏好") || strings.Contains(lowerKey, "preference") ||
		strings.Contains(lowerContent, "我喜欢") || strings.Contains(lowerContent, "i prefer") ||
		strings.Contains(lowerContent, "偏好") || strings.Contains(lowerContent, "preference") {
		return "preference"
	}
	if strings.Contains(lowerKey, "临时") || strings.Contains(lowerKey, "temporary") ||
		strings.Contains(lowerKey, "todo") || strings.Contains(lowerKey, "待办") ||
		strings.Contains(lowerContent, "临时") || strings.Contains(lowerContent, "temporary") {
		return "short_term"
	}
	if strings.Contains(lowerKey, "私密") || strings.Contains(lowerKey, "private") ||
		strings.Contains(lowerKey, "密码") || strings.Contains(lowerKey, "password") ||
		strings.Contains(lowerContent, "私密") || strings.Contains(lowerContent, "private") {
		return "private"
	}
	if strings.Contains(lowerKey, "项目") || strings.Contains(lowerKey, "project") ||
		strings.Contains(lowerContent, "项目") || strings.Contains(lowerContent, "project") {
		return "knowledge"
	}
	if strings.Contains(lowerKey, "工具") || strings.Contains(lowerKey, "tool") ||
		strings.Contains(lowerContent, "工具") || strings.Contains(lowerContent, "software") {
		return "knowledge"
	}
	return "knowledge"
}

// extractTags extracts tags from a raw memory map.
func extractTags(m map[string]interface{}) string {
	if t, ok := m["tags"].([]interface{}); ok && len(t) > 0 {
		tagStrs := make([]string, 0, len(t))
		for _, tag := range t {
			if s, ok := tag.(string); ok {
				tagStrs = append(tagStrs, s)
			}
		}
		return strings.Join(tagStrs, ",")
	}
	if t, ok := m["tags"].(string); ok {
		return t
	}
	if cat, ok := m["category"].(string); ok {
		return cat
	}
	return ""
}
