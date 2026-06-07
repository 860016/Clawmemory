package services

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"clawmemory/internal/models"

	"gorm.io/gorm"
)

type ExtractedMemory struct {
	Key        string   `json:"key"`
	Value      string   `json:"value"`
	Layer      string   `json:"layer"`
	MemoryType string   `json:"memory_type"`
	Importance float64  `json:"importance"`
	Tags       []string `json:"tags"`
	Source     string   `json:"source"`
	Reason     string   `json:"reason"`
}

type ExtractionResult struct {
	Memories []ExtractedMemory `json:"memories"`
	Count    int               `json:"count"`
	Warnings []string          `json:"warnings,omitempty"`
}

type ExtractionService struct {
	db     *gorm.DB
	userID uint
}

func NewExtractionService(db *gorm.DB, userID uint) *ExtractionService {
	return &ExtractionService{db: db, userID: userID}
}

var (
	preferencePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(?:i prefer|my preference|我喜欢|我偏好|我习惯)`),
		regexp.MustCompile(`(?i)(?:always|never|make sure|务必|一定|永远不要|总是)`),
		regexp.MustCompile(`(?i)(?:i like it|i don't like|我喜欢|我不喜欢)`),
		regexp.MustCompile(`(?i)(?:use\s+\w+\s+instead|用\w+代替)`),
		regexp.MustCompile(`(?i)(?:please\s+(?:always|never)|请(?:务必|永远不要))`),
	}

	feedbackPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(?:don't|stop|avoid|不要|避免|停止)`),
		regexp.MustCompile(`(?i)(?:that's wrong|incorrect|不对|错误|不是这样)`),
		regexp.MustCompile(`(?i)(?:please fix|fix this|修复|改正)`),
		regexp.MustCompile(`(?i)(?:not what i wanted|不是我要的|不是这样)`),
	}

	projectPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(?:project|repo|repository|项目|仓库)`),
		regexp.MustCompile(`(?i)(?:deadline|milestone|截止|里程碑)`),
		regexp.MustCompile(`(?i)(?:sprint|release|deploy|迭代|发布|部署)`),
		regexp.MustCompile(`(?i)(?:tech stack|技术栈|框架|framework)`),
	}

	referencePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(?:url|link|http|https|链接|地址)`),
		regexp.MustCompile(`(?i)(?:documentation|docs|文档|手册)`),
		regexp.MustCompile(`(?i)(?:api endpoint|endpoint|接口|端点)`),
		regexp.MustCompile(`(?i)(?:config|configuration|配置)`),
	}

	knowledgePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(?:note that|important|注意|重要|关键)`),
		regexp.MustCompile(`(?i)(?:the system|the app|the code|系统|应用|代码)`),
		regexp.MustCompile(`(?i)(?:architecture|design pattern|架构|设计模式)`),
		regexp.MustCompile(`(?i)(?:database|schema|数据库|表结构)`),
		regexp.MustCompile(`(?i)(?:algorithm|approach|算法|方案)`),
	}
)

func (s *ExtractionService) ExtractFromConversation(content string) *ExtractionResult {
	result := &ExtractionResult{
		Memories: []ExtractedMemory{},
		Warnings: []string{},
	}

	if content == "" {
		return result
	}

	secretResult := ScanSecrets(content)
	if secretResult.Found {
		for _, m := range secretResult.Matches {
			result.Warnings = append(result.Warnings, "Detected "+m.Description+" — sensitive content will be flagged")
		}
	}

	paragraphs := splitIntoParagraphs(content)
	seen := make(map[string]bool)

	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" || utf8.RuneCountInString(para) < 10 {
			continue
		}

		mem := s.classifyParagraph(para)
		if mem == nil {
			continue
		}

		dedupKey := mem.Key + "|" + truncateStr(mem.Value, 50)
		if seen[dedupKey] {
			continue
		}
		seen[dedupKey] = true

		if s.isDuplicate(mem) {
			continue
		}

		result.Memories = append(result.Memories, *mem)
	}

	result.Count = len(result.Memories)
	return result
}

func (s *ExtractionService) classifyParagraph(para string) *ExtractedMemory {
	for _, p := range preferencePatterns {
		if p.MatchString(para) {
			return &ExtractedMemory{
				Key:        extractKey(para),
				Value:      para,
				Layer:      "core",
				MemoryType: "preference",
				Importance: 0.7,
				Tags:       []string{"auto-extracted", "preference"},
				Source:     "extraction",
				Reason:     "Detected user preference pattern",
			}
		}
	}

	for _, p := range feedbackPatterns {
		if p.MatchString(para) {
			return &ExtractedMemory{
				Key:        extractKey(para),
				Value:      para,
				Layer:      "core",
				MemoryType: "feedback",
				Importance: 0.8,
				Tags:       []string{"auto-extracted", "feedback"},
				Source:     "extraction",
				Reason:     "Detected user feedback/correction pattern",
			}
		}
	}

	for _, p := range projectPatterns {
		if p.MatchString(para) {
			return &ExtractedMemory{
				Key:        extractKey(para),
				Value:      para,
				Layer:      "knowledge",
				MemoryType: "project",
				Importance: 0.6,
				Tags:       []string{"auto-extracted", "project"},
				Source:     "extraction",
				Reason:     "Detected project-related information",
			}
		}
	}

	for _, p := range referencePatterns {
		if p.MatchString(para) {
			return &ExtractedMemory{
				Key:        extractKey(para),
				Value:      para,
				Layer:      "knowledge",
				MemoryType: "reference",
				Importance: 0.5,
				Tags:       []string{"auto-extracted", "reference"},
				Source:     "extraction",
				Reason:     "Detected reference/URL information",
			}
		}
	}

	for _, p := range knowledgePatterns {
		if p.MatchString(para) {
			return &ExtractedMemory{
				Key:        extractKey(para),
				Value:      para,
				Layer:      "knowledge",
				MemoryType: "knowledge",
				Importance: 0.5,
				Tags:       []string{"auto-extracted", "knowledge"},
				Source:     "extraction",
				Reason:     "Detected factual knowledge pattern",
			}
		}
	}

	return nil
}

func (s *ExtractionService) isDuplicate(mem *ExtractedMemory) bool {
	var count int64
	s.db.Model(&models.Memory{}).
		Where("user_id = ? AND key = ? AND status != ?", s.userID, mem.Key, "trashed").
		Count(&count)
	return count > 0
}

func splitIntoParagraphs(content string) []string {
	var paragraphs []string
	var current strings.Builder

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if current.Len() > 0 {
				paragraphs = append(paragraphs, current.String())
				current.Reset()
			}
			continue
		}
		if current.Len() > 0 {
			current.WriteString(" ")
		}
		current.WriteString(trimmed)
	}

	if current.Len() > 0 {
		paragraphs = append(paragraphs, current.String())
	}

	return paragraphs
}

func extractKey(para string) string {
	if utf8.RuneCountInString(para) <= 60 {
		return para
	}

	runes := []rune(para)
	firstSentence := ""
	for i, r := range runes {
		if r == '.' || r == '。' || r == '!' || r == '！' || r == '?' || r == '？' || r == ':' || r == '：' {
			firstSentence = string(runes[:i])
			break
		}
	}

	if firstSentence == "" {
		firstSentence = string(runes[:min(50, len(runes))])
	}

	if utf8.RuneCountInString(firstSentence) > 60 {
		return string([]rune(firstSentence)[:57]) + "..."
	}

	return strings.TrimSpace(firstSentence)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
