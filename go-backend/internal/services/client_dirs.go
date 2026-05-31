package services

import (
	"os"
	"path/filepath"
	"runtime"
)

type ClientInfo struct {
	Name        string
	DisplayName string
	DataDirs    []string
	SkillsDirs  []string
}

func GetSupportedClients() []ClientInfo {
	homeDir, _ := os.UserHomeDir()
	appData := os.Getenv("APPDATA")
	localAppData := os.Getenv("LOCALAPPDATA")
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" && homeDir != "" {
		configDir = filepath.Join(homeDir, ".config")
	}
	exePath, _ := os.Executable()
	exeDir := ""
	if exePath != "" {
		exeDir = filepath.Dir(exePath)
	}
	wd, _ := os.Getwd()

	clients := []ClientInfo{}

	clients = append(clients, ClientInfo{
		Name:        "openclaw",
		DisplayName: "OpenClaw",
		DataDirs: buildDirs(
			filepath.Join(homeDir, ".openclaw"),
			filepath.Join(exeDir, "openclaw"),
			filepath.Join(wd, ".openclaw"),
		),
		SkillsDirs: buildDirs(
			filepath.Join(homeDir, ".openclaw", "skills"),
			filepath.Join(homeDir, ".openclaw", "workspace", "skills"),
		),
	})

	clients = append(clients, ClientInfo{
		Name:        "clawmemory",
		DisplayName: "ClawMemory",
		DataDirs: buildDirs(
			filepath.Join(homeDir, ".clawmemory"),
			filepath.Join(wd, ".clawmemory"),
			filepath.Join(wd, "data"),
			filepath.Join(exeDir, "data"),
		),
		SkillsDirs: buildDirs(
			filepath.Join(homeDir, ".clawmemory", "skills"),
		),
	})

	clients = append(clients, ClientInfo{
		Name:        "trae",
		DisplayName: "Trae / Trae CN",
		DataDirs: buildDirs(
			filepath.Join(homeDir, ".trae"),
			filepath.Join(homeDir, ".trae-cn"),
			dirIf(appData, filepath.Join(appData, "Trae CN", "User", "workspaceStorage")),
			dirIf(appData, filepath.Join(appData, "Trae", "User", "workspaceStorage")),
			dirIf(appData, filepath.Join(appData, "Trae CN", "ModularData", "ai-agent")),
			dirIf(localAppData, filepath.Join(localAppData, "Trae CN")),
			dirIf(localAppData, filepath.Join(localAppData, "Trae")),
		),
		SkillsDirs: buildDirs(
			filepath.Join(homeDir, ".trae", "skills"),
			filepath.Join(homeDir, ".trae-cn", "skills"),
		),
	})

	clients = append(clients, ClientInfo{
		Name:        "codebuddy",
		DisplayName: "CodeBuddy / CodeBuddy CN",
		DataDirs: buildDirs(
			dirIf(appData, filepath.Join(appData, "CodeBuddy CN", "User", "workspaceStorage")),
			dirIf(appData, filepath.Join(appData, "CodeBuddy", "User", "workspaceStorage")),
			dirIf(localAppData, filepath.Join(localAppData, "CodeBuddy CN")),
			dirIf(localAppData, filepath.Join(localAppData, "CodeBuddy")),
		),
		SkillsDirs: []string{},
	})

	clients = append(clients, ClientInfo{
		Name:        "cursor",
		DisplayName: "Cursor",
		DataDirs: buildDirs(
			filepath.Join(homeDir, ".cursor"),
			dirIf(appData, filepath.Join(appData, "Cursor", "User", "workspaceStorage")),
			dirIf(localAppData, filepath.Join(localAppData, "Cursor")),
			dirIf(configDir, filepath.Join(configDir, "Cursor", "User", "workspaceStorage")),
		),
		SkillsDirs: buildDirs(
			filepath.Join(homeDir, ".cursor", "skills"),
		),
	})

	clients = append(clients, ClientInfo{
		Name:        "claude",
		DisplayName: "Claude Code",
		DataDirs: buildDirs(
			filepath.Join(homeDir, ".claude"),
			filepath.Join(homeDir, ".claude", "projects"),
		),
		SkillsDirs: buildDirs(
			filepath.Join(homeDir, ".claude", "skills"),
			filepath.Join(homeDir, ".claude", "commands"),
		),
	})

	clients = append(clients, ClientInfo{
		Name:        "qoder",
		DisplayName: "Qoder",
		DataDirs: buildDirs(
			dirIf(appData, filepath.Join(appData, "Qoder", "User", "workspaceStorage")),
			dirIf(localAppData, filepath.Join(localAppData, "Qoder")),
		),
		SkillsDirs: []string{},
	})

	clients = append(clients, ClientInfo{
		Name:        "codex",
		DisplayName: "OpenAI Codex",
		DataDirs: buildDirs(
			filepath.Join(homeDir, ".codex"),
			dirIf(appData, filepath.Join(appData, "codex")),
		),
		SkillsDirs: []string{},
	})

	clients = append(clients, ClientInfo{
		Name:        "windsurf",
		DisplayName: "Windsurf",
		DataDirs: buildDirs(
			filepath.Join(homeDir, ".codeium", "windsurf"),
			dirIf(appData, filepath.Join(appData, "Windsurf", "User", "workspaceStorage")),
			dirIf(localAppData, filepath.Join(localAppData, "Windsurf")),
			dirIf(configDir, filepath.Join(configDir, "Windsurf", "User", "workspaceStorage")),
		),
		SkillsDirs: buildDirs(
			filepath.Join(homeDir, ".codeium", "windsurf", "skills"),
		),
	})

	clients = append(clients, ClientInfo{
		Name:        "github-copilot",
		DisplayName: "GitHub Copilot",
		DataDirs: buildDirs(
			dirIf(localAppData, filepath.Join(localAppData, "GitHub Copilot")),
		),
		SkillsDirs: []string{},
	})

	clients = append(clients, ClientInfo{
		Name:        "cline",
		DisplayName: "Cline / Continue",
		DataDirs: buildDirs(
			filepath.Join(homeDir, ".cline"),
			filepath.Join(homeDir, ".cline", "data"),
			filepath.Join(homeDir, ".continue"),
			filepath.Join(homeDir, ".continue", "sessions"),
			dirIf(appData, filepath.Join(appData, "Continue")),
		),
		SkillsDirs: buildDirs(
			filepath.Join(homeDir, ".cline", "skills"),
			filepath.Join(homeDir, ".continue", "skills"),
		),
	})

	clients = append(clients, ClientInfo{
		Name:        "augment",
		DisplayName: "Augment Code",
		DataDirs: buildDirs(
			filepath.Join(homeDir, ".augment"),
		),
		SkillsDirs: buildDirs(
			filepath.Join(homeDir, ".augment", "skills"),
		),
	})

	clients = append(clients, ClientInfo{
		Name:        "aider",
		DisplayName: "Aider",
		DataDirs: buildDirs(
			filepath.Join(homeDir, ".aider"),
		),
		SkillsDirs: []string{},
	})

	clients = append(clients, ClientInfo{
		Name:        "hermes",
		DisplayName: "Hermes",
		DataDirs: buildDirs(
			filepath.Join(homeDir, ".hermes"),
		),
		SkillsDirs: buildDirs(
			filepath.Join(homeDir, ".hermes", "skills"),
		),
	})

	return clients
}

func buildDirs(candidates ...string) []string {
	seenDirs := make(map[string]bool)
	var result []string
	for _, d := range candidates {
		if d == "" {
			continue
		}
		abs, err := filepath.Abs(d)
		if err != nil {
			abs = d
		}
		if !seenDirs[abs] {
			seenDirs[abs] = true
			result = append(result, abs)
		}
	}
	return result
}

func dirIf(cond string, path string) string {
	if cond == "" {
		return ""
	}
	return path
}

func GetAllSearchDirs() []string {
	seenDirs := make(map[string]bool)
	var result []string
	for _, client := range GetSupportedClients() {
		for _, d := range client.DataDirs {
			if !seenDirs[d] {
				seenDirs[d] = true
				result = append(result, d)
			}
		}
	}
	return result
}

func GetAllSkillsDirs() []string {
	seenDirs := make(map[string]bool)
	var result []string
	for _, client := range GetSupportedClients() {
		for _, d := range client.SkillsDirs {
			if !seenDirs[d] {
				seenDirs[d] = true
				result = append(result, d)
			}
		}
	}
	return result
}

func DetectInstalledClients() []map[string]interface{} {
	var installed []map[string]interface{}
	for _, client := range GetSupportedClients() {
		found := false
		var foundDirs []string
		for _, d := range client.DataDirs {
			if _, err := os.Stat(d); err == nil {
				found = true
				foundDirs = append(foundDirs, d)
			}
		}
		if found {
			installed = append(installed, map[string]interface{}{
				"name":         client.Name,
				"display_name": client.DisplayName,
				"found_dirs":   foundDirs,
			})
		}
	}
	return installed
}

func init() {
	_ = runtime.GOOS
}
