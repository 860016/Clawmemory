package services

import (
	"encoding/json"
	"fmt"
	"strings"

	"clawmemory/internal/models"

	"gorm.io/gorm"
)

type TemplateService struct {
	db *gorm.DB
}

func NewTemplateService(db *gorm.DB) *TemplateService {
	return &TemplateService{db: db}
}

type MemoryTemplate struct {
	ID          uint            `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Layer       string          `json:"layer"`
	KeyPattern  string          `json:"key_pattern"`
	ValueFields []TemplateField `json:"value_fields"`
	Category    string          `json:"category"`
	IsBuiltin   bool            `json:"is_builtin"`
}

type TemplateField struct {
	Name     string `json:"name"`
	Label    string `json:"label"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
	Default  string `json:"default,omitempty"`
	Hint     string `json:"hint,omitempty"`
}

var builtinTemplates = []MemoryTemplate{
	{
		Name:        "API Endpoint",
		Description: "Record an API endpoint with its method, path, and parameters",
		Layer:       "procedural",
		KeyPattern:  "api:{{method}}:{{path}}",
		Category:    "development",
		IsBuiltin:   true,
		ValueFields: []TemplateField{
			{Name: "method", Label: "HTTP Method", Type: "select", Required: true, Default: "GET", Hint: "GET/POST/PUT/DELETE"},
			{Name: "path", Label: "API Path", Type: "text", Required: true, Hint: "/api/v1/resource"},
			{Name: "description", Label: "Description", Type: "textarea", Required: true},
			{Name: "parameters", Label: "Parameters", Type: "textarea", Required: false, Hint: "JSON format"},
			{Name: "response", Label: "Response Format", Type: "textarea", Required: false},
		},
	},
	{
		Name:        "Bug Fix",
		Description: "Document a bug and its fix for future reference",
		Layer:       "episodic",
		KeyPattern:  "bugfix:{{component}}:{{summary}}",
		Category:    "development",
		IsBuiltin:   true,
		ValueFields: []TemplateField{
			{Name: "component", Label: "Component", Type: "text", Required: true},
			{Name: "summary", Label: "Bug Summary", Type: "text", Required: true},
			{Name: "symptom", Label: "Symptom", Type: "textarea", Required: true},
			{Name: "root_cause", Label: "Root Cause", Type: "textarea", Required: true},
			{Name: "fix", Label: "Fix Applied", Type: "textarea", Required: true},
			{Name: "lesson", Label: "Lesson Learned", Type: "textarea", Required: false},
		},
	},
	{
		Name:        "Design Decision",
		Description: "Record an architectural or design decision with rationale",
		Layer:       "semantic",
		KeyPattern:  "decision:{{domain}}:{{title}}",
		Category:    "architecture",
		IsBuiltin:   true,
		ValueFields: []TemplateField{
			{Name: "domain", Label: "Domain", Type: "text", Required: true},
			{Name: "title", Label: "Decision Title", Type: "text", Required: true},
			{Name: "context", Label: "Context", Type: "textarea", Required: true},
			{Name: "options", Label: "Options Considered", Type: "textarea", Required: true},
			{Name: "decision", Label: "Decision", Type: "textarea", Required: true},
			{Name: "rationale", Label: "Rationale", Type: "textarea", Required: true},
		},
	},
	{
		Name:        "Environment Config",
		Description: "Record environment configuration and setup details",
		Layer:       "procedural",
		KeyPattern:  "env:{{project}}:{{name}}",
		Category:    "operations",
		IsBuiltin:   true,
		ValueFields: []TemplateField{
			{Name: "project", Label: "Project Name", Type: "text", Required: true},
			{Name: "name", Label: "Config Name", Type: "text", Required: true},
			{Name: "environment", Label: "Environment", Type: "select", Required: true, Default: "development", Hint: "development/staging/production"},
			{Name: "details", Label: "Configuration Details", Type: "textarea", Required: true},
			{Name: "notes", Label: "Notes", Type: "textarea", Required: false},
		},
	},
	{
		Name:        "Learning Note",
		Description: "Capture a learning insight or knowledge gained",
		Layer:       "semantic",
		KeyPattern:  "learning:{{topic}}:{{title}}",
		Category:    "knowledge",
		IsBuiltin:   true,
		ValueFields: []TemplateField{
			{Name: "topic", Label: "Topic", Type: "text", Required: true},
			{Name: "title", Label: "Title", Type: "text", Required: true},
			{Name: "insight", Label: "Key Insight", Type: "textarea", Required: true},
			{Name: "example", Label: "Example", Type: "textarea", Required: false},
			{Name: "references", Label: "References", Type: "textarea", Required: false},
		},
	},
	{
		Name:        "Command Reference",
		Description: "Record a useful command or script for quick reference",
		Layer:       "procedural",
		KeyPattern:  "cmd:{{category}}:{{name}}",
		Category:    "operations",
		IsBuiltin:   true,
		ValueFields: []TemplateField{
			{Name: "category", Label: "Category", Type: "text", Required: true, Hint: "git/docker/k8s/etc."},
			{Name: "name", Label: "Command Name", Type: "text", Required: true},
			{Name: "command", Label: "Command", Type: "textarea", Required: true},
			{Name: "description", Label: "Description", Type: "textarea", Required: true},
			{Name: "flags", Label: "Important Flags", Type: "textarea", Required: false},
		},
	},
}

func (s *TemplateService) ListTemplates(userID uint) ([]map[string]interface{}, error) {
	var settings []models.Setting
	s.db.Where("user_id = ? AND key LIKE ?", userID, "template_%").Find(&settings)

	results := make([]map[string]interface{}, 0, len(builtinTemplates)+len(settings))

	for _, t := range builtinTemplates {
		fieldsJSON, _ := json.Marshal(t.ValueFields)
		results = append(results, map[string]interface{}{
			"name":         t.Name,
			"description":  t.Description,
			"layer":        t.Layer,
			"key_pattern":  t.KeyPattern,
			"value_fields": string(fieldsJSON),
			"category":     t.Category,
			"is_builtin":   true,
		})
	}

	for _, st := range settings {
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(st.Value), &data); err == nil {
			data["is_builtin"] = false
			data["setting_id"] = st.ID
			results = append(results, data)
		}
	}

	return results, nil
}

func (s *TemplateService) CreateTemplate(userID uint, template MemoryTemplate) error {
	fieldsJSON, _ := json.Marshal(template.ValueFields)
	data := map[string]interface{}{
		"name":         template.Name,
		"description":  template.Description,
		"layer":        template.Layer,
		"key_pattern":  template.KeyPattern,
		"value_fields": string(fieldsJSON),
		"category":     template.Category,
	}
	dataJSON, _ := json.Marshal(data)

	setting := models.Setting{
		UserID: userID,
		Key:    "template_" + template.Name,
		Value:  string(dataJSON),
	}
	return s.db.Create(&setting).Error
}

func (s *TemplateService) DeleteTemplate(userID uint, name string) error {
	return s.db.Where("user_id = ? AND key = ?", userID, "template_"+name).Delete(&models.Setting{}).Error
}

func (s *TemplateService) ApplyTemplate(userID uint, templateName string, values map[string]string) (string, string, string, error) {
	var template *MemoryTemplate
	for i := range builtinTemplates {
		if builtinTemplates[i].Name == templateName {
			template = &builtinTemplates[i]
			break
		}
	}

	if template == nil {
		var settings []models.Setting
		s.db.Where("user_id = ? AND key = ?", userID, "template_"+templateName).Find(&settings)
		if len(settings) > 0 {
			var data map[string]interface{}
			if err := json.Unmarshal([]byte(settings[0].Value), &data); err == nil {
				template = &MemoryTemplate{
					Name:       templateName,
					Layer:      data["layer"].(string),
					KeyPattern: data["key_pattern"].(string),
				}
			}
		}
	}

	if template == nil {
		return "", "", "", ErrTemplateNotFound
	}

	key := template.KeyPattern
	for k, v := range values {
		key = replaceAll(key, "{{"+k+"}}", v)
	}

	var valueParts []string
	for _, field := range template.ValueFields {
		if v, ok := values[field.Name]; ok && v != "" {
			valueParts = append(valueParts, field.Label+": "+v)
		}
	}
	value := joinStrings(valueParts, "\n")

	return key, value, template.Layer, nil
}

var ErrTemplateNotFound = fmt.Errorf("template not found")

func replaceAll(s, old, new string) string {
	return strings.ReplaceAll(s, old, new)
}

func joinStrings(parts []string, sep string) string {
	return strings.Join(parts, sep)
}
