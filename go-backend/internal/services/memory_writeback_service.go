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

type MemoryWritebackService struct {
	db *gorm.DB
}

func NewMemoryWritebackService(db *gorm.DB) *MemoryWritebackService {
	return &MemoryWritebackService{db: db}
}

type WritebackTarget struct {
	AgentName  string `json:"agent_name"`
	TargetPath string `json:"target_path"`
	Format     string `json:"format"`
}

type WritebackResult struct {
	AgentName  string `json:"agent_name"`
	TargetPath string `json:"target_path"`
	Count      int    `json:"count"`
	Success    bool   `json:"success"`
	Error      string `json:"error,omitempty"`
}

func (s *MemoryWritebackService) GetWritebackTargets() []WritebackTarget {
	return []WritebackTarget{
		{
			AgentName:  "clawmemory",
			TargetPath: ".clawmemory/memory.md",
			Format:     "markdown",
		},
		{
			AgentName:  "openclaw",
			TargetPath: ".openclaw/memory.md",
			Format:     "markdown",
		},
		{
			AgentName:  "trae",
			TargetPath: ".trae/memory.md",
			Format:     "markdown",
		},
		{
			AgentName:  "cursor",
			TargetPath: ".cursor/memories.md",
			Format:     "markdown",
		},
		{
			AgentName:  "cline",
			TargetPath: ".clinerules/memory.md",
			Format:     "markdown",
		},
		{
			AgentName:  "windsurf",
			TargetPath: ".windsurf/memories.md",
			Format:     "markdown",
		},
		{
			AgentName:  "continue",
			TargetPath: ".continue/memory.json",
			Format:     "json",
		},
	}
}

func (s *MemoryWritebackService) Writeback(userID uint, agentName string, projectPath string) (*WritebackResult, error) {
	riskSvc := GetRiskSwitchService()
	if riskSvc != nil && riskSvc.IsDisabled(RiskCrossAgentWrite) {
		return nil, fmt.Errorf("cross-agent write is disabled by risk switch")
	}

	var memories []models.Memory
	query := s.db.Where("user_id = ? AND status = ?", userID, "active")
	if agentName != "" {
		query = query.Where("source_agent = ? OR visibility IN ?", agentName, []string{"shared", "public"})
	}
	logDBErr("load memories for writeback", query.Find(&memories).Error)

	if len(memories) == 0 {
		return &WritebackResult{
			AgentName: agentName,
			Count:     0,
			Success:   true,
		}, nil
	}

	target := s.findTarget(agentName)
	if target == nil {
		return nil, fmt.Errorf("no writeback target found for agent: %s", agentName)
	}

	var targetPath string
	if projectPath != "" {
		targetPath = filepath.Join(projectPath, target.TargetPath)
	} else {
		homeDir, _ := os.UserHomeDir()
		targetPath = filepath.Join(homeDir, target.TargetPath)
	}

	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	var err error
	switch target.Format {
	case "markdown":
		err = s.writeMarkdown(memories, targetPath)
	case "json":
		err = s.writeJSON(memories, targetPath)
	default:
		err = s.writeMarkdown(memories, targetPath)
	}

	if err != nil {
		return &WritebackResult{
			AgentName:  agentName,
			TargetPath: targetPath,
			Count:      len(memories),
			Success:    false,
			Error:      err.Error(),
		}, err
	}

	sec := GetSecurity()
	if sec != nil {
		sec.LogAudit(AuditEntry{
			UserID:    userID,
			Action:    "memory.writeback",
			Target:    targetPath,
			Detail:    fmt.Sprintf("wrote %d memories for agent %s", len(memories), agentName),
			AgentName: agentName,
		})
	}

	return &WritebackResult{
		AgentName:  agentName,
		TargetPath: targetPath,
		Count:      len(memories),
		Success:    true,
	}, nil
}

func (s *MemoryWritebackService) PreviewWriteback(userID uint, agentName string) ([]models.Memory, error) {
	var memories []models.Memory
	query := s.db.Where("user_id = ? AND status = ?", userID, "active")
	if agentName != "" {
		query = query.Where("source_agent = ? OR visibility IN ?", agentName, []string{"shared", "public"})
	}
	logDBErr("load memories for writeback sync", query.Find(&memories).Error)
	return memories, nil
}

func (s *MemoryWritebackService) findTarget(agentName string) *WritebackTarget {
	targets := s.GetWritebackTargets()
	for i := range targets {
		if targets[i].AgentName == agentName {
			return &targets[i]
		}
	}
	return nil
}

func (s *MemoryWritebackService) writeMarkdown(memories []models.Memory, path string) error {
	var sb strings.Builder
	sb.WriteString("# ClawMemory Export\n")
	sb.WriteString(fmt.Sprintf("# Generated at %s\n\n", time.Now().Format(time.RFC3339)))

	for _, m := range memories {
		sb.WriteString(fmt.Sprintf("## %s\n", m.Key))
		sb.WriteString(fmt.Sprintf("%s\n\n", m.Value))
		if m.Tags != "" {
			sb.WriteString(fmt.Sprintf("Tags: %s\n\n", m.Tags))
		}
	}

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, []byte(sb.String()), 0644); err != nil {
		return err
	}

	if _, err := os.Stat(path); err == nil {
		backupPath := path + ".bak." + time.Now().Format("20060102-150405")
		os.Rename(path, backupPath)
	}

	return os.Rename(tmpPath, path)
}

func (s *MemoryWritebackService) writeJSON(memories []models.Memory, path string) error {
	data := make([]map[string]interface{}, 0, len(memories))
	for _, m := range memories {
		data = append(data, map[string]interface{}{
			"key":          m.Key,
			"value":        m.Value,
			"layer":        m.Layer,
			"importance":   m.Importance,
			"tags":         m.Tags,
			"source":       m.Source,
			"source_agent": m.SourceAgent,
			"visibility":   m.Visibility,
		})
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, jsonBytes, 0644); err != nil {
		return err
	}

	if _, err := os.Stat(path); err == nil {
		backupPath := path + ".bak." + time.Now().Format("20060102-150405")
		os.Rename(path, backupPath)
	}

	return os.Rename(tmpPath, path)
}
