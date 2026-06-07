package api

import (
	"clawmemory/internal/config"
	"clawmemory/internal/services"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

func handleScanSkills(c *gin.Context) {
	dataDirs := []string{}
	seenDirs := make(map[string]bool)

	addDir := func(d string) {
		abs, err := filepath.Abs(d)
		if err != nil {
			abs = d
		}
		if !seenDirs[abs] {
			seenDirs[abs] = true
			dataDirs = append(dataDirs, abs)
		}
	}

	cfg := config.Load()
	addDir(cfg.SkillsDir)

	exe, _ := os.Executable()
	if exe != "" {
		addDir(filepath.Join(filepath.Dir(exe), "skills"))
	}

	for _, d := range services.GetAllSkillsDirs() {
		addDir(d)
	}

	homeDir, _ := os.UserHomeDir()
	if homeDir != "" {
		addDir(filepath.Join(homeDir, ".agents", "skills"))
	}

	wd, _ := os.Getwd()
	if wd != "" {
		addDir(filepath.Join(wd, "skills"))
		addDir(filepath.Join(wd, ".agents", "skills"))
	}

	globalSkills := make([]map[string]interface{}, 0)
	workspaceSkills := make([]map[string]interface{}, 0)

	globalDir, _ := filepath.Abs(cfg.SkillsDir)

	for _, dir := range dataDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			skillFiles := []string{
				filepath.Join(dir, entry.Name(), "skill.json"),
				filepath.Join(dir, entry.Name(), "skill.yaml"),
				filepath.Join(dir, entry.Name(), "skill.yml"),
				filepath.Join(dir, entry.Name(), "SKILL.md"),
			}
			var content []byte
			var skillFile string
			for _, sf := range skillFiles {
				if data, err := os.ReadFile(sf); err == nil {
					content = data
					skillFile = sf
					break
				}
			}
			if content == nil {
				continue
			}
			var skill map[string]interface{}
			ext := filepath.Ext(skillFile)
			baseName := filepath.Base(skillFile)
			if ext == ".json" {
				json.Unmarshal(content, &skill)
			} else if ext == ".yaml" || ext == ".yml" {
				if parsed, err := parseYAML(content); err == nil {
					skill = parsed
				}
			} else if baseName == "SKILL.md" {
				if parsed, err := parseSKILLMd(content); err == nil {
					skill = parsed
				}
			}
			if skill == nil {
				continue
			}
			skill["skill_dir"] = entry.Name()
			absDir, _ := filepath.Abs(dir)
			if absDir == globalDir {
				skill["scope"] = "global"
				globalSkills = append(globalSkills, skill)
			} else {
				skill["scope"] = "workspace"
				workspaceSkills = append(workspaceSkills, skill)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"global_skills":    globalSkills,
		"workspace_skills": workspaceSkills,
		"clients":          services.DetectInstalledClients(),
	})
}

func handleInstallSkill(c *gin.Context) {
	var req struct {
		RepoURL string `json:"repo_url" binding:"required"`
		Scope   string `json:"scope"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "repo_url is required"})
		return
	}
	if req.Scope == "" {
		req.Scope = "global"
	}

	cfg := config.Load()
	targetDir := cfg.SkillsDir

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create skills directory"})
		return
	}

	repoURL := req.RepoURL
	if !strings.HasPrefix(repoURL, "https://") && !strings.HasPrefix(repoURL, "git@") {
		repoURL = "https://github.com/" + repoURL
	}

	repoName := repoURL
	if idx := strings.LastIndex(repoName, "/"); idx >= 0 {
		repoName = repoName[idx+1:]
	}
	if strings.HasSuffix(repoName, ".git") {
		repoName = repoName[:len(repoName)-4]
	}

	destPath := filepath.Join(targetDir, repoName)
	if _, err := os.Stat(destPath); err == nil {
		c.JSON(http.StatusOK, gin.H{
			"message":   "skill already installed",
			"skill_dir": repoName,
			"path":      destPath,
		})
		return
	}

	cmd := exec.Command("git", "clone", "--depth", "1", repoURL, destPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "failed to clone repository: " + string(output),
			"detail": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "skill installed successfully",
		"skill_dir": repoName,
		"path":      destPath,
	})
}

func parseYAML(data []byte) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			if strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"") {
				val = val[1 : len(val)-1]
			}
			result[key] = val
		}
	}
	return result, nil
}

func parseSKILLMd(data []byte) (map[string]interface{}, error) {
	content := string(data)
	result := make(map[string]interface{})

	if strings.HasPrefix(content, "---") {
		endIdx := strings.Index(content[3:], "---")
		if endIdx >= 0 {
			frontmatter := content[3 : endIdx+3]
			lines := strings.Split(frontmatter, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					val := strings.TrimSpace(parts[1])
					if strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"") {
						val = val[1 : len(val)-1]
					}
					result[key] = val
				}
			}
			bodyStart := endIdx + 6
			if bodyStart < len(content) {
				result["body_full"] = strings.TrimSpace(content[bodyStart:])
			}
		}
	} else {
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "# ") {
				result["name"] = strings.TrimSpace(line[2:])
				break
			}
		}
		result["body_full"] = content
	}

	if _, ok := result["name"]; !ok {
		result["name"] = "unknown"
	}

	return result, nil
}

func handleSkillDetail(c *gin.Context) {
	skillDir := c.Query("skill_dir")
	scope := c.Query("scope")

	searchDirs := []string{}
	if scope == "global" {
		searchDirs = append(searchDirs, services.GetAllSkillsDirs()...)
	}

	cfg := config.Load()
	if cfg.SkillsDir != "" {
		searchDirs = append(searchDirs, cfg.SkillsDir)
	}
	if cfg.DataDir != "" {
		searchDirs = append(searchDirs, filepath.Join(cfg.DataDir, "skills"))
	}

	exe, _ := os.Executable()
	if exe != "" {
		searchDirs = append(searchDirs, filepath.Join(filepath.Dir(exe), "skills"))
	}

	wd, _ := os.Getwd()
	if wd != "" {
		searchDirs = append(searchDirs, filepath.Join(wd, "skills"))
	}

	for _, baseDir := range searchDirs {
		skillFiles := []string{
			filepath.Join(baseDir, skillDir, "skill.json"),
			filepath.Join(baseDir, skillDir, "skill.yaml"),
			filepath.Join(baseDir, skillDir, "skill.yml"),
			filepath.Join(baseDir, skillDir, "SKILL.md"),
		}
		for _, skillFile := range skillFiles {
			content, err := os.ReadFile(skillFile)
			if err != nil {
				continue
			}
			var skill map[string]interface{}
			ext := filepath.Ext(skillFile)
			baseName := filepath.Base(skillFile)
			if ext == ".json" {
				if err := json.Unmarshal(content, &skill); err != nil {
					continue
				}
			} else if ext == ".yaml" || ext == ".yml" {
				if parsed, err := parseYAML(content); err != nil {
					continue
				} else {
					skill = parsed
				}
			} else if baseName == "SKILL.md" {
				if parsed, err := parseSKILLMd(content); err != nil {
					continue
				} else {
					skill = parsed
				}
			}
			if skill != nil {
				skill["skill_dir"] = skillDir
				skill["scope"] = scope
				c.JSON(http.StatusOK, skill)
				return
			}
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "skill not found"})
}
