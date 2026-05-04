package ai

import (
	"fmt"
	"strings"
)

type PromptTemplate struct {
	ID          string
	Name        string
	Description string
	System      string
	User        string
	ProOnly     bool
}

var PromptTemplates = map[string]PromptTemplate{
	"extract": {
		ID:          "extract",
		Name:        "Entity & Relation Extraction",
		Description: "Extract entities and relations from memories",
		ProOnly:     false,
		System: `You are a knowledge graph expert. Extract entities and their relationships from the given memory entries.
Return ONLY valid JSON, no other text.`,
		User: `Extract entities and relations from these memories:

{{.Memories}}

Return JSON format:
{
  "entities": [
    {"name": "entity name", "type": "person|organization|technology|location|event|concept", "description": "brief description", "confidence": 0.0-1.0}
  ],
  "relations": [
    {"source": "entity A", "target": "entity B", "type": "relation type", "description": "relation description", "confidence": 0.0-1.0}
  ]
}

Relation types: works_at, manages, depends_on, located_in, member_of, uses, created_by, related_to, part_of, owns, knows, reports_to`,
	},
	"conflict_scan": {
		ID:          "conflict_scan",
		Name:        "Semantic Conflict Detection",
		Description: "Detect semantic conflicts between memories",
		ProOnly:     true,
		System: `You are a data consistency expert. Identify semantic conflicts between memory entries.
Return ONLY valid JSON, no other text.`,
		User: `Check these memories for semantic conflicts:

{{.Memories}}

Conflict criteria:
1. Same entity with contradictory attributes (e.g., project uses React vs Vue)
2. Timeline contradictions (e.g., started 2023 vs completed 2022)
3. Logical contradictions (e.g., team size 3 vs 20)

Return JSON:
{
  "conflicts": [
    {"memory_a_id": 1, "memory_b_id": 2, "type": "attribute|timeline|logic", "description": "conflict description", "severity": "high|medium|low", "suggestion": "merge suggestion"}
  ]
}`,
	},
	"decay_evaluate": {
		ID:          "decay_evaluate",
		Name:        "Smart Decay Evaluation",
		Description: "AI-powered memory decay assessment",
		ProOnly:     true,
		System: `You are a memory management expert. Evaluate which memories should decay, be archived, or kept.
Return ONLY valid JSON, no other text.`,
		User: `Evaluate these memories for decay:

{{.Memories}}

Evaluation dimensions:
1. Timeliness: Is the information outdated?
2. Reuse value: Will it likely be needed again?
3. Connection density: How many other memories reference this?
4. Uniqueness: Is this the only record of this information?

Return JSON:
{
  "evaluations": [
    {"id": 1, "action": "keep|archive|merge|delete", "reason": "explanation", "new_importance": 0.0-1.0, "merge_with": []}
  ]
}`,
	},
	"daily_report": {
		ID:          "daily_report",
		Name:        "Daily Report Generation",
		Description: "Generate intelligent daily report from memories",
		ProOnly:     false,
		System: `You are a knowledge management assistant. Generate a concise daily report based on the user's memory activity.
Return ONLY valid JSON, no other text.`,
		User: `Generate a daily report based on today's activity:

Date: {{.Date}}
New memories: {{.MemoryCount}}
Key memories:
{{.Highlights}}

Return JSON:
{
  "summary": "2-3 sentence summary of today",
  "highlights": ["key finding 1", "key finding 2", "key finding 3"],
  "knowledge_gained": ["what was learned 1", "what was learned 2"],
  "pending_tasks": ["task 1", "task 2"],
  "tomorrow_suggestions": ["suggestion 1", "suggestion 2"],
  "mood": "productive|exploratory|routine|intensive"
}`,
	},
	"wiki_generate": {
		ID:          "wiki_generate",
		Name:        "Wiki Auto-Generation",
		Description: "Auto-generate wiki pages from memories",
		ProOnly:     true,
		System: `You are a technical documentation expert. Generate structured wiki documentation from memory entries.
Return ONLY valid JSON, no other text.`,
		User: `Generate a wiki document from these memories:

Topic: {{.Topic}}
Related memories:
{{.Memories}}

Return JSON:
{
  "title": "document title",
  "category": "category name",
  "content": "full wiki content in markdown",
  "summary": "2-3 sentence summary",
  "tags": ["tag1", "tag2"],
  "key_decisions": ["decision 1", "decision 2"],
  "action_items": ["action 1", "action 2"]
}`,
	},
	"compress": {
		ID:          "compress",
		Name:        "Memory Compression",
		Description: "Compress and refine multiple memories into one",
		ProOnly:     true,
		System: `You are an information refinement expert. Merge and compress multiple related memories into a single high-quality entry.
Return ONLY valid JSON, no other text.`,
		User: `Compress these related memories into one refined entry:

{{.Memories}}

Requirements:
1. Preserve all key information, remove redundancy
2. Unify expressions, resolve contradictions
3. Increase information density
4. Note information sources

Return JSON:
{
  "key": "compressed memory key",
  "value": "compressed memory content",
  "importance": 0.0-1.0,
  "layer": "core|context|detail",
  "tags": ["tag1", "tag2"],
  "merged_count": 3,
  "notes": "any notes about the compression"
}`,
	},
	"evolution_discover": {
		ID:          "evolution_discover",
		Name:        "Deep Relation Discovery",
		Description: "Discover hidden relationships between memories",
		ProOnly:     true,
		System: `You are a knowledge discovery expert. Find hidden, non-obvious relationships between memories.
Return ONLY valid JSON, no other text.`,
		User: `Discover deep relationships between these memories:

{{.Memories}}

Look for:
1. Causal relationships (A caused B)
2. Temporal patterns (A always happens before B)
3. Conceptual links (A and B share underlying principles)
4. Dependency chains (A depends on B which depends on C)

Return JSON:
{
  "relations": [
    {"source_id": 1, "target_id": 2, "type": "causes|precedes|depends_on|enables|contradicts|refines", "description": "why this relation exists", "confidence": 0.0-1.0}
  ]
}`,
	},
	"smart_route": {
		ID:          "smart_route",
		Name:        "Smart Token Router",
		Description: "Analyze query complexity for optimal model routing",
		ProOnly:     true,
		System: `You are a model routing expert. Analyze the complexity of the given text and recommend the appropriate AI model.
Return ONLY valid JSON, no other text.`,
		User: `Analyze this text for model routing:

{{.Text}}

Return JSON:
{
  "complexity_score": 1-10,
  "complexity": "simple|moderate|complex",
  "recommended_model": "small|medium|large",
  "reason": "why this model",
  "estimated_tokens": 100,
  "technical_terms": 3,
  "sentence_count": 2
}`,
	},
}

func GetPromptTemplate(id string) (*PromptTemplate, bool) {
	t, ok := PromptTemplates[id]
	return &t, ok
}

func RenderPrompt(templateStr string, data map[string]string) string {
	result := templateStr
	for key, value := range data {
		result = strings.ReplaceAll(result, "{{."+key+"}}", value)
	}
	return result
}

func FormatMemoriesForPrompt(memories []map[string]interface{}) string {
	var sb strings.Builder
	for i, m := range memories {
		id := fmt.Sprintf("%v", m["id"])
		key := fmt.Sprintf("%v", m["key"])
		value := fmt.Sprintf("%v", m["value"])
		if len(value) > 300 {
			value = value[:300] + "..."
		}
		layer := fmt.Sprintf("%v", m["layer"])
		importance := fmt.Sprintf("%v", m["importance"])
		sb.WriteString(fmt.Sprintf("[%s] key=%s layer=%s importance=%s\n  %s\n", id, key, layer, importance, value))
		if i >= 49 {
			sb.WriteString("... (truncated, showing first 50)\n")
			break
		}
	}
	return sb.String()
}
