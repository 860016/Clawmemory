package services

type AIConversationItem struct {
	Type      string `json:"type"`
	SessionID string `json:"session_id"`
	AgentID   string `json:"agent_id"`
	Thought   string `json:"thought,omitempty"`
	Content   string `json:"content,omitempty"`
	Name      string `json:"name,omitempty"`
	Status    string `json:"status,omitempty"`
	Platform  string `json:"platform,omitempty"`
}

type ideToolConfig struct {
	ProcessName string
	DisplayName string
	Platform    string
}

var supportedIDETolls = []ideToolConfig{
	{ProcessName: "Trae CN.exe", DisplayName: "Trae CN", Platform: "trae"},
	{ProcessName: "Trae.exe", DisplayName: "Trae", Platform: "trae"},
	{ProcessName: "Cursor.exe", DisplayName: "Cursor", Platform: "cursor"},
	{ProcessName: "CodeBuddy CN.exe", DisplayName: "CodeBuddy CN", Platform: "codebuddy"},
	{ProcessName: "CodeBuddy.exe", DisplayName: "CodeBuddy", Platform: "codebuddy"},
	{ProcessName: "Qoder.exe", DisplayName: "Qoder", Platform: "qoder"},
}

func isValidUTF8Sequence(seq []byte) bool {
	if len(seq) < 2 {
		return false
	}
	for j := 1; j < len(seq); j++ {
		if seq[j]&0xC0 != 0x80 {
			return false
		}
	}
	return true
}
