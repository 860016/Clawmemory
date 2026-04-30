package services

import (
	"regexp"
	"strings"
)

type SecretMatch struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
}

type SecretScanResult struct {
	Found   bool          `json:"found"`
	Matches []SecretMatch `json:"matches"`
}

var secretPatterns = []struct {
	pattern     *regexp.Regexp
	secretType  string
	description string
	severity    string
}{
	{
		regexp.MustCompile(`(?i)(?:aws_access_key_id|aws_secret_access_key)\s*[=:]\s*['"]?[A-Za-z0-9/+=]{16,}['"]?`),
		"aws_key", "AWS Access Key", "high",
	},
	{
		regexp.MustCompile(`(?i)AKIA[0-9A-Z]{16}`),
		"aws_key_id", "AWS Key ID", "high",
	},
	{
		regexp.MustCompile(`(?i)(?:ghp|gho|ghu|ghs|ghr)_[A-Za-z0-9_]{36,}`),
		"github_token", "GitHub Token", "high",
	},
	{
		regexp.MustCompile(`(?i)sk-[A-Za-z0-9]{20,}`),
		"openai_key", "OpenAI API Key", "high",
	},
	{
		regexp.MustCompile(`(?i)sk-ant-api[A-Za-z0-9_-]{20,}`),
		"anthropic_key", "Anthropic API Key", "high",
	},
	{
		regexp.MustCompile(`(?i)(?:xai-|xai_)[A-Za-z0-9_-]{20,}`),
		"xai_key", "xAI API Key", "high",
	},
	{
		regexp.MustCompile(`(?i)-----BEGIN (?:RSA |EC |DSA |OPENSSH )?PRIVATE KEY-----`),
		"private_key", "Private Key", "high",
	},
	{
		regexp.MustCompile(`(?i)(?:password|passwd|pwd)\s*[=:]\s*['"][^'"]{4,}['"]`),
		"password", "Password in Config", "medium",
	},
	{
		regexp.MustCompile(`(?i)(?:api_key|apikey|api-key)\s*[=:]\s*['"][^'"]{8,}['"]`),
		"api_key", "API Key", "medium",
	},
	{
		regexp.MustCompile(`(?i)(?:token|auth_token|access_token|bearer)\s*[=:]\s*['"][A-Za-z0-9_.-]{20,}['"]`),
		"auth_token", "Authentication Token", "medium",
	},
	{
		regexp.MustCompile(`(?i)(?:mongodb(?:\+srv)?|postgres(?:ql)?|mysql|redis)://[^\s'"]{10,}`),
		"db_connection", "Database Connection String", "high",
	},
	{
		regexp.MustCompile(`(?i)(?:jwt_secret|secret_key|encryption_key)\s*[=:]\s*['"][^'"]{8,}['"]`),
		"secret_key", "Secret Key", "medium",
	},
	{
		regexp.MustCompile(`(?i)(?:slack_token|slack_webhook)\s*[=:]\s*['"]xox[bpras]-[A-Za-z0-9-]{10,}['"]`),
		"slack_token", "Slack Token", "medium",
	},
	{
		regexp.MustCompile(`(?i)(?:stripe_(?:api_key|secret_key|publishable_key))\s*[=:]\s*['"][sr]k_(?:live|test)_[A-Za-z0-9]{20,}['"]`),
		"stripe_key", "Stripe API Key", "high",
	},
	{
		regexp.MustCompile(`(?i)(?:sendgrid_api_key)\s*[=:]\s*['"]SG\.[A-Za-z0-9_-]{20,}['"]`),
		"sendgrid_key", "SendGrid API Key", "medium",
	},
}

func ScanSecrets(content string) *SecretScanResult {
	result := &SecretScanResult{Found: false, Matches: []SecretMatch{}}

	if content == "" {
		return result
	}

	seen := make(map[string]bool)
	for _, sp := range secretPatterns {
		if sp.pattern.MatchString(content) {
			key := sp.secretType
			if seen[key] {
				continue
			}
			seen[key] = true
			result.Found = true
			result.Matches = append(result.Matches, SecretMatch{
				Type:        sp.secretType,
				Description: sp.description,
				Severity:    sp.severity,
			})
		}
	}

	return result
}

func IsSensitiveContent(content string) bool {
	lower := strings.ToLower(content)
	sensitivePhrases := []string{
		"do not share",
		"confidential",
		"top secret",
		"internal only",
		"do not distribute",
	}
	for _, phrase := range sensitivePhrases {
		if strings.Contains(lower, phrase) {
			return true
		}
	}
	return false
}
