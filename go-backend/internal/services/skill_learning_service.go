package services

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"clawmemory/internal/models"

	"gorm.io/gorm"
)

type SkillLearningService struct {
	db *gorm.DB
}

func NewSkillLearningService(db *gorm.DB) *SkillLearningService {
	return &SkillLearningService{db: db}
}

func (s *SkillLearningService) RecordAction(userID uint, sessionID, agentName, platform, actionType, actionName, parameters, result string, duration int) error {
	patternHash := computePatternHash(actionType, actionName, parameters)

	trace := models.ActionTrace{
		UserID:      userID,
		SessionID:   sessionID,
		AgentName:   agentName,
		Platform:    platform,
		ActionType:  actionType,
		ActionName:  actionName,
		Parameters:  parameters,
		Result:      result,
		Duration:    duration,
		PatternHash: patternHash,
	}
	return s.db.Create(&trace).Error
}

func (s *SkillLearningService) RecordActionBatch(userID uint, sessionID, agentName, platform string, actions []map[string]interface{}) (int, error) {
	created := 0
	for _, a := range actions {
		actionType, _ := a["action_type"].(string)
		actionName, _ := a["action_name"].(string)
		paramsJSON, _ := json.Marshal(a["parameters"])
		result, _ := a["result"].(string)
		duration := 0
		if d, ok := a["duration"].(float64); ok {
			duration = int(d)
		}
		if actionType == "" || actionName == "" {
			continue
		}
		if err := s.RecordAction(userID, sessionID, agentName, platform, actionType, actionName, string(paramsJSON), result, duration); err == nil {
			created++
		}
	}
	return created, nil
}

func (s *SkillLearningService) DetectPatterns(userID uint) ([]map[string]interface{}, error) {
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	var traces []models.ActionTrace
	if err := s.db.Where("user_id = ? AND created_at > ?", userID, sevenDaysAgo).
		Order("created_at ASC").Limit(5000).Find(&traces).Error; err != nil {
		return nil, fmt.Errorf("failed to load action traces: %w", err)
	}

	sessionActions := make(map[string][]models.ActionTrace)
	for _, t := range traces {
		if t.SessionID == "" {
			t.SessionID = "session_" + t.CreatedAt.Format("20060102_150405") + "_" + fmt.Sprintf("%d", t.ID)
		}
		sessionActions[t.SessionID] = append(sessionActions[t.SessionID], t)
	}

	patterns := make([]map[string]interface{}, 0)
	patternCount := make(map[string]int)
	patternSessions := make(map[string][]string)
	patternExamples := make(map[string][]models.ActionTrace)

	for sessionID, actions := range sessionActions {
		if len(actions) < 3 {
			continue
		}
		for i := 0; i <= len(actions)-3; i++ {
			seq := actions[i : i+3]
			hash := computeSequenceHash(seq)
			patternCount[hash]++
			if _, exists := patternSessions[hash]; !exists {
				patternExamples[hash] = seq
			}
			found := false
			for _, sid := range patternSessions[hash] {
				if sid == sessionID {
					found = true
					break
				}
			}
			if !found {
				patternSessions[hash] = append(patternSessions[hash], sessionID)
			}
		}
	}

	for hash, count := range patternCount {
		sessionCount := len(patternSessions[hash])
		if sessionCount < 2 || count < 4 {
			continue
		}
		example := patternExamples[hash]
		actionNames := make([]string, len(example))
		actionTypes := make([]string, len(example))
		for i, e := range example {
			actionNames[i] = e.ActionName
			actionTypes[i] = e.ActionType
		}

		patterns = append(patterns, map[string]interface{}{
			"pattern_hash":     hash,
			"action_sequence":  actionNames,
			"action_types":     actionTypes,
			"occurrence_count": count,
			"session_count":    sessionCount,
			"agents":           getUniqueAgents(example),
			"platforms":        getUniquePlatforms(example),
			"worth_skill":      sessionCount >= 2 && count >= 4,
		})
	}

	return patterns, nil
}

func (s *SkillLearningService) AutoCreateSkillFromPattern(userID uint, pattern map[string]interface{}) (*models.Skill, error) {
	actionSeq := toStringSlice(pattern["action_sequence"])
	if len(actionSeq) == 0 {
		return nil, fmt.Errorf("empty action sequence")
	}

	name := generateSkillName(actionSeq)

	var existing models.Skill
	if err := s.db.Where("user_id = ? AND name = ?", userID, name).First(&existing).Error; err == nil {
		return &existing, nil
	}

	agents := toStringSlice(pattern["agents"])
	platforms := toStringSlice(pattern["platforms"])
	sourceAgent := "auto"
	if len(agents) > 0 {
		sourceAgent = agents[0]
	}

	stepsJSON, _ := json.Marshal(actionSeq)
	triggerKW := strings.Join(actionSeq, ",")
	tags := strings.Join(platforms, ",")

	skill := models.Skill{
		UserID:          userID,
		Name:            name,
		Description:     fmt.Sprintf("Auto-detected workflow: %s (used %v times across %v sessions)", strings.Join(actionSeq, " → "), pattern["occurrence_count"], pattern["session_count"]),
		TriggerKeywords: triggerKW,
		Steps:           string(stepsJSON),
		SourceAgent:     sourceAgent,
		Category:        "auto_detected",
		Tags:            tags,
		AutoCreated:     true,
		Status:          "active",
	}

	if err := s.db.Create(&skill).Error; err != nil {
		return nil, err
	}

	return &skill, nil
}

func (s *SkillLearningService) DetectAndCreateSkills(userID uint) (map[string]interface{}, error) {
	patterns, err := s.DetectPatterns(userID)
	if err != nil {
		return nil, err
	}

	created := 0
	skipped := 0
	var newSkills []models.Skill

	for _, p := range patterns {
		worth, _ := p["worth_skill"].(bool)
		if !worth {
			skipped++
			continue
		}
		skill, err := s.AutoCreateSkillFromPattern(userID, p)
		if err != nil {
			skipped++
			continue
		}
		created++
		newSkills = append(newSkills, *skill)
	}

	return map[string]interface{}{
		"patterns_found": len(patterns),
		"skills_created": created,
		"skills_skipped": skipped,
		"new_skills":     newSkills,
	}, nil
}

func (s *SkillLearningService) MatchSkill(userID uint, query string) ([]models.Skill, error) {
	var skills []models.Skill
	keywords := strings.ToLower(query)

	skillsQuery := s.db.Where("user_id = ? AND status = ?", userID, "active")

	if err := skillsQuery.Find(&skills).Error; err != nil {
		return nil, err
	}

	var matched []models.Skill
	for _, skill := range skills {
		triggerKW := strings.ToLower(skill.TriggerKeywords)
		desc := strings.ToLower(skill.Description)
		name := strings.ToLower(skill.Name)
		tags := strings.ToLower(skill.Tags)

		if strings.Contains(triggerKW, keywords) ||
			strings.Contains(desc, keywords) ||
			strings.Contains(name, keywords) ||
			strings.Contains(tags, keywords) ||
			containsAnyWord(keywords, triggerKW) {
			matched = append(matched, skill)
		}
	}

	return matched, nil
}

func (s *SkillLearningService) RecordSkillUsage(userID uint, skillID uint, success bool) error {
	var skill models.Skill
	if err := s.db.Where("id = ? AND user_id = ?", skillID, userID).First(&skill).Error; err != nil {
		return err
	}

	now := time.Now()
	updates := map[string]interface{}{
		"usage_count":  skill.UsageCount + 1,
		"last_used_at": &now,
	}
	if success {
		updates["success_count"] = skill.SuccessCount + 1
	} else {
		updates["fail_count"] = skill.FailCount + 1
	}

	return s.db.Model(&skill).Updates(updates).Error
}

func (s *SkillLearningService) PatchSkill(userID uint, skillID uint, field, oldValue, newValue string) error {
	var skill models.Skill
	if err := s.db.Where("id = ? AND user_id = ?", skillID, userID).First(&skill).Error; err != nil {
		return err
	}

	switch field {
	case "steps":
		skill.Steps = strings.Replace(skill.Steps, oldValue, newValue, 1)
	case "known_pitfalls":
		if skill.KnownPitfalls == "" {
			skill.KnownPitfalls = newValue
		} else {
			skill.KnownPitfalls += "\n" + newValue
		}
	case "trigger_keywords":
		skill.TriggerKeywords += "," + newValue
	case "description":
		skill.Description = newValue
	case "status":
		validStatuses := map[string]bool{"active": true, "inactive": true, "archived": true}
		if !validStatuses[newValue] {
			return fmt.Errorf("invalid status value: %s", newValue)
		}
		skill.Status = newValue
	default:
		return fmt.Errorf("unsupported field: %s", field)
	}

	skill.Version++
	return s.db.Save(&skill).Error
}

func (s *SkillLearningService) GenerateAgentSuggestions(userID uint) ([]map[string]interface{}, error) {
	installed := DetectInstalledClients()
	installedMap := make(map[string]bool)
	for _, c := range installed {
		name, _ := c["name"].(string)
		installedMap[name] = true
	}

	allClients := GetSupportedClients()
	suggestions := make([]map[string]interface{}, 0)

	var existingSuggestions []models.SkillSuggestion
	_ = s.db.Where("user_id = ? AND status = ?", userID, "pending").Find(&existingSuggestions).Error
	existingTypes := make(map[string]bool)
	for _, es := range existingSuggestions {
		key := es.AgentName + ":" + es.SuggestType
		existingTypes[key] = true
	}

	for _, client := range allClients {
		if installedMap[client.Name] {
			continue
		}

		suggestType := "install_agent"
		key := client.Name + ":" + suggestType
		if existingTypes[key] {
			continue
		}

		importGuide := generateImportGuide(client.Name)
		importURL := generateImportURL(client.Name)

		suggestions = append(suggestions, map[string]interface{}{
			"agent_name":   client.Name,
			"display_name": client.DisplayName,
			"suggest_type": suggestType,
			"title":        fmt.Sprintf("Connect %s to unlock more skills", client.DisplayName),
			"description":  fmt.Sprintf("%s is not connected. Importing it will allow ClawMemory to learn from your %s workflows and auto-create skills.", client.DisplayName, client.DisplayName),
			"import_url":   importURL,
			"import_guide": importGuide,
		})
	}

	var traces []models.ActionTrace
	_ = s.db.Where("user_id = ?", userID).Order("created_at DESC").Limit(100).Find(&traces).Error
	agentActionCount := make(map[string]int)
	for _, t := range traces {
		agentActionCount[t.AgentName]++
	}

	for agent, count := range agentActionCount {
		if count < 5 {
			continue
		}
		key := agent + ":enable_skill_learning"
		if existingTypes[key] {
			continue
		}
		suggestions = append(suggestions, map[string]interface{}{
			"agent_name":   agent,
			"display_name": agent,
			"suggest_type": "enable_skill_learning",
			"title":        fmt.Sprintf("Enable Skill Learning for %s (%d actions tracked)", agent, count),
			"description":  fmt.Sprintf("You've used %s %d times. Enabling skill learning will auto-detect reusable workflows and create skills.", agent, count),
			"import_url":   "",
			"import_guide": "Call POST /api/v1/ai/skills/detect to auto-detect patterns from your activity.",
		})
	}

	for _, sug := range suggestions {
		s.SaveSuggestion(userID, sug)
	}

	return suggestions, nil
}

func (s *SkillLearningService) SaveSuggestion(userID uint, suggestion map[string]interface{}) error {
	agentName, _ := suggestion["agent_name"].(string)
	suggestType, _ := suggestion["suggest_type"].(string)
	title, _ := suggestion["title"].(string)
	desc, _ := suggestion["description"].(string)
	importURL, _ := suggestion["import_url"].(string)
	importGuide, _ := suggestion["import_guide"].(string)

	sug := models.SkillSuggestion{
		UserID:      userID,
		AgentName:   agentName,
		SuggestType: suggestType,
		Title:       title,
		Description: desc,
		ImportURL:   importURL,
		ImportGuide: importGuide,
		Status:      "pending",
	}
	return s.db.Create(&sug).Error
}

func (s *SkillLearningService) GetPendingSuggestions(userID uint) ([]models.SkillSuggestion, error) {
	var suggestions []models.SkillSuggestion
	err := s.db.Where("user_id = ? AND status = ?", userID, "pending").Order("created_at DESC").Find(&suggestions).Error
	return suggestions, err
}

func (s *SkillLearningService) DismissSuggestion(userID uint, suggestionID uint) error {
	now := time.Now()
	return s.db.Model(&models.SkillSuggestion{}).
		Where("id = ? AND user_id = ?", suggestionID, userID).
		Updates(map[string]interface{}{"status": "dismissed", "dismissed_at": &now}).Error
}

func (s *SkillLearningService) GetSkillStats(userID uint) map[string]interface{} {
	var totalSkills int64
	var activeSkills int64
	var autoCreated int64
	var totalUsage int64

	s.db.Model(&models.Skill{}).Where("user_id = ?", userID).Count(&totalSkills)
	s.db.Model(&models.Skill{}).Where("user_id = ? AND status = ?", userID, "active").Count(&activeSkills)
	s.db.Model(&models.Skill{}).Where("user_id = ? AND auto_created = ?", userID, true).Count(&autoCreated)
	s.db.Model(&models.ActionTrace{}).Where("user_id = ?", userID).Count(&totalUsage)

	var traces []models.ActionTrace
	_ = s.db.Where("user_id = ?", userID).Order("created_at DESC").Limit(1).Find(&traces).Error
	lastActivity := ""
	if len(traces) > 0 {
		lastActivity = traces[0].CreatedAt.Format("2006-01-02 15:04:05")
	}

	return map[string]interface{}{
		"total_skills":  totalSkills,
		"active_skills": activeSkills,
		"auto_created":  autoCreated,
		"total_actions": totalUsage,
		"last_activity": lastActivity,
	}
}

func (s *SkillLearningService) CleanupOldTraces(retentionDays int) (int64, error) {
	if retentionDays < 7 {
		retentionDays = 7
	}
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	result := s.db.Where("created_at < ?", cutoff).Delete(&models.ActionTrace{})
	return result.RowsAffected, result.Error
}

func (s *SkillLearningService) CleanupDismissedSuggestions() (int64, error) {
	result := s.db.Where("status = ? AND dismissed_at IS NOT NULL AND dismissed_at < ?",
		"dismissed", time.Now().AddDate(0, 0, -7)).Delete(&models.SkillSuggestion{})
	return result.RowsAffected, result.Error
}

func computePatternHash(actionType, actionName, parameters string) string {
	h := sha256.New()
	h.Write([]byte(actionType + ":" + actionName))
	return fmt.Sprintf("%x", h.Sum(nil))[:16]
}

func toStringSlice(v interface{}) []string {
	if v == nil {
		return nil
	}
	if s, ok := v.([]string); ok {
		return s
	}
	if arr, ok := v.([]interface{}); ok {
		var result []string
		for _, item := range arr {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}
	return nil
}

func computeSequenceHash(actions []models.ActionTrace) string {
	h := sha256.New()
	for _, a := range actions {
		h.Write([]byte(a.ActionType + ":" + a.ActionName + "|"))
	}
	return fmt.Sprintf("%x", h.Sum(nil))[:16]
}

func generateSkillName(actions []string) string {
	if len(actions) == 0 {
		return "unnamed-skill"
	}
	if len(actions) <= 3 {
		return strings.Join(actions, "-")
	}
	return actions[0] + "-to-" + actions[len(actions)-1]
}

func getUniqueAgents(actions []models.ActionTrace) []string {
	seen := make(map[string]bool)
	var result []string
	for _, a := range actions {
		if !seen[a.AgentName] && a.AgentName != "" {
			seen[a.AgentName] = true
			result = append(result, a.AgentName)
		}
	}
	return result
}

func getUniquePlatforms(actions []models.ActionTrace) []string {
	seen := make(map[string]bool)
	var result []string
	for _, a := range actions {
		if !seen[a.Platform] && a.Platform != "" {
			seen[a.Platform] = true
			result = append(result, a.Platform)
		}
	}
	return result
}

func containsAnyWord(query, text string) bool {
	words := strings.Fields(query)
	for _, w := range words {
		if len(w) >= 3 && strings.Contains(text, w) {
			return true
		}
	}
	return false
}

func generateImportGuide(agentName string) string {
	guides := map[string]string{
		"cursor":         "1. Open Cursor Settings → Extensions\n2. Search 'ClawMemory' and install\n3. Open Command Palette → 'ClawMemory: Connect'\n4. Enter your ClawMemory server URL",
		"claude":         "1. Install Claude Code CLI\n2. Run: clawmemory mcp add claude\n3. Restart Claude Code\n4. Your memories will auto-sync",
		"trae":           "1. Open Trae Settings → Plugins\n2. Search 'ClawMemory' and install\n3. Click 'Connect' and enter server URL\n4. Start chatting — actions will be tracked",
		"windsurf":       "1. Open Windsurf Settings → Extensions\n2. Search 'ClawMemory' and install\n3. Configure server URL in settings\n4. Memories will sync automatically",
		"cline":          "1. Open VS Code with Cline extension\n2. Install ClawMemory MCP server\n3. Add to .cline/config.json: {\"mcpServers\": {\"clawmemory\": {...}}}\n4. Restart Cline",
		"augment":        "1. Open Augment Code settings\n2. Add ClawMemory as MCP server\n3. Configure connection URL\n4. Actions will be tracked for skill learning",
		"aider":          "1. Install aider-chat Python package\n2. Run: clawmemory mcp add aider\n3. Configure in .aider.conf.yml\n4. Start coding session",
		"hermes":         "1. Install Hermes Agent\n2. Configure ClawMemory as external memory provider\n3. Run: hermes memory setup → select ClawMemory\n4. Skills will be shared bi-directionally",
		"codebuddy":      "1. Open CodeBuddy settings\n2. Search 'ClawMemory' and install\n3. Configure server URL\n4. Start tracking",
		"github-copilot": "1. Open VS Code Settings → Extensions\n2. Install ClawMemory extension\n3. Configure MCP server in settings.json\n4. Copilot chats will be tracked",
	}
	if guide, ok := guides[agentName]; ok {
		return guide
	}
	return "1. Install the ClawMemory MCP server\n2. Configure your agent's MCP settings\n3. Add the server URL and API key\n4. Restart your agent"
}

func generateImportURL(agentName string) string {
	urls := map[string]string{
		"cursor":         "https://marketplace.visualstudio.com/items?itemName=clawmemory",
		"claude":         "https://docs.claude.ai/mcp-setup",
		"trae":           "https://www.trae.ai/plugins",
		"windsurf":       "https://codeium.com/windsurf/plugins",
		"cline":          "https://github.com/cline/cline",
		"augment":        "https://www.augmentcode.com/docs",
		"aider":          "https://github.com/paul-gauthier/aider",
		"hermes":         "https://github.com/NousResearch/hermes-agent",
		"codebuddy":      "https://www.codebuddy.ai/plugins",
		"github-copilot": "https://github.com/features/copilot",
	}
	if url, ok := urls[agentName]; ok {
		return url
	}
	return "https://clawmemory.com/docs/mcp-setup"
}
