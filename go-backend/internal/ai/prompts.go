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
}

var PromptTemplates = map[string]PromptTemplate{
	"extract": {
		ID:          "extract",
		Name:        "Entity & Relation Extraction",
		Description: "Extract entities and relations from memories",
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
	"extract_facts": {
		ID:          "extract_facts",
		Name:        "Fact & Preference Extraction",
		Description: "Extract facts, preferences, and relationships from conversation messages",
		System: `You are a memory extraction specialist. Your job is to analyze conversation messages and extract structured facts, user preferences, and relationships that should be remembered for future interactions.

Rules:
1. Extract ONLY information worth remembering long-term
2. Each fact should be atomic - one piece of information per entry
3. Preferences should capture user's likes, dislikes, and tendencies
4. Relationships should capture connections between entities
5. Ignore greetings, small talk, and procedural messages
6. Be specific - "prefers dark mode" not "has UI preference"

Return ONLY valid JSON, no other text.`,
		User: `Extract facts, preferences, and relationships from this conversation:

{{.Messages}}

Existing memories for deduplication:
{{.ExistingMemories}}

Return JSON:
{
  "facts": [
    {"content": "the fact to remember", "category": "identity|preference|skill|possession|relationship|routine|goal|opinion", "confidence": 0.0-1.0, "source": "user|assistant|inferred"}
  ],
  "preferences": [
    {"topic": "what the preference is about", "value": "the preference value", "strength": "strong|moderate|weak", "confidence": 0.0-1.0}
  ],
  "relations": [
    {"subject": "entity A", "predicate": "relationship type", "object": "entity B", "confidence": 0.0-1.0}
  ],
  "updates": [
    {"old_fact": "existing fact that should be updated", "new_fact": "updated fact", "reason": "why it changed"}
  ]
}`,
	},
	"memory_consolidate": {
		ID:          "memory_consolidate",
		Name:        "Memory Consolidation",
		Description: "Consolidate and deduplicate extracted facts with existing memories",
		System: `You are a memory consolidation expert. Your job is to merge new facts with existing memories, resolving conflicts and removing duplicates.

Rules:
1. If a new fact contradicts an existing memory, keep the newer one and mark the old one as superseded
2. If a new fact is a subset of an existing memory, merge them
3. If a new fact adds detail to an existing memory, enrich the existing one
4. If a new fact is completely new, add it
5. If a new fact duplicates an existing memory, skip it

Return ONLY valid JSON, no other text.`,
		User: `Consolidate these new facts with existing memories:

New facts:
{{.NewFacts}}

Existing memories:
{{.ExistingMemories}}

Return JSON:
{
  "add": [
    {"key": "memory key", "value": "memory content", "layer": "core|context|detail", "importance": 0.0-1.0, "category": "fact|preference|skill|relationship"}
  ],
  "update": [
    {"memory_id": 1, "field": "value", "old_value": "old", "new_value": "new", "reason": "why updated"}
  ],
  "merge": [
    {"source_ids": [1, 2], "merged_key": "new key", "merged_value": "merged content", "layer": "core|context|detail"}
  ],
  "supersede": [
    {"old_id": 1, "reason": "superseded by newer information", "new_id": 2}
  ],
  "skip": [
    {"fact": "the duplicate fact", "matches_memory_id": 1}
  ]
}`,
	},
	"context_assemble": {
		ID:          "context_assemble",
		Name:        "Smart Context Assembly",
		Description: "Assemble optimal context for LLM from memories based on query",
		System: `You are a context optimization expert. Given a user query and available memories, select and organize the most relevant context to include in the LLM prompt.

Rules:
1. Prioritize directly relevant memories
2. Include related context that provides background
3. Respect the token budget strictly
4. Order memories from most to least relevant
5. Include a brief relevance explanation for each selected memory

Return ONLY valid JSON, no other text.`,
		User: `Assemble context for this query:

Query: {{.Query}}
Token budget: {{.TokenBudget}}

Available memories:
{{.Memories}}

Return JSON:
{
  "selected_memories": [
    {"memory_id": 1, "relevance_score": 0.0-1.0, "relevance_reason": "why this is relevant", "tokens": 50}
  ],
  "total_tokens": 500,
  "coverage_score": 0.0-1.0,
  "missing_context": ["what information is still missing"],
  "suggested_followup": ["what to ask next"]
}`,
	},
	"nudge_reflect": {
		ID:          "nudge_reflect",
		Name:        "Periodic Nudge Reflection",
		Description: "Periodically review recent activity and extract high-value knowledge worth persisting",
		System: `You are a knowledge distillation specialist. Your job is to review recent memory activity and decide what is worth permanently remembering, what should be compressed, and what can be forgotten.

Philosophy: Only remember information that will influence FUTURE behavior. Discard everything else.

Rules:
1. Environment facts (tools, configs, project paths) → always persist
2. User corrections and feedback → always persist (high learning signal)
3. Repetitive patterns → compress into a single rule
4. One-time debugging details → forget
5. User preferences → persist and update existing profile
6. Task completion patterns → persist as procedural knowledge

Return ONLY valid JSON, no other text.`,
		User: `Review recent memory activity and extract high-value knowledge:

Recent memory changes:
{{.RecentChanges}}

Current user profile:
{{.UserProfile}}

Current memory stats:
{{.MemoryStats}}

Return JSON:
{
  "persist": [
    {"content": "what to remember", "category": "identity|preference|skill|possession|relationship|routine|goal|opinion|correction", "confidence": 0.0-1.0, "reason": "why this is worth remembering"}
  ],
  "compress": [
    {"memory_ids": [1, 2, 3], "compressed_content": "unified summary", "reason": "why these should be merged"}
  ],
  "forget": [
    {"memory_id": 4, "reason": "why this can be safely forgotten"}
  ],
  "profile_updates": [
    {"field": "communication_style|tech_stack|workflow|preference", "old_value": "previous", "new_value": "updated", "evidence": "what evidence supports this change"}
  ],
  "insights": [
    {"insight": "a higher-level pattern observed from the data", "confidence": 0.0-1.0}
  ]
}`,
	},
	"self_refine": {
		ID:          "self_refine",
		Name:        "Memory Self-Refinement",
		Description: "Under capacity pressure, automatically distill memories to retain only the highest-value information",
		System: `You are a memory refinement engine operating under strict capacity constraints. Your job is to distill a set of memories into a smaller, denser set that preserves maximum information value.

Core principle: Information economy — every character must earn its place.

Rules:
1. Merge overlapping memories into single, denser entries
2. Remove redundant qualifiers and hedging language
3. Promote frequently-accessed details to higher layers
4. Demote stale, rarely-accessed memories to lower layers
5. Preserve all unique facts — never lose information, only compress it
6. Keep the most specific, actionable version of each fact

Return ONLY valid JSON, no other text.`,
		User: `Refine these memories under capacity pressure (target: {{.TargetCount}} memories from {{.CurrentCount}}):

{{.Memories}}

Capacity pressure level: {{.PressureLevel}}

Return JSON:
{
  "keep": [
    {"memory_id": 1, "reason": "why this must be kept as-is"}
  ],
  "merge": [
    {"source_ids": [2, 3, 4], "merged_key": "unified key", "merged_value": "compressed but complete content", "layer": "core|context|detail", "importance": 0.0-1.0}
  ],
  "demote": [
    {"memory_id": 5, "from_layer": "context", "to_layer": "detail", "reason": "why this should be demoted"}
  ],
  "archive": [
    {"memory_id": 6, "reason": "why this can be archived (not deleted, just inactive)"}
  ],
  "stats": {
    "original_count": 0,
    "result_count": 0,
    "compression_ratio": 0.0,
    "information_preservation": 0.0
  }
}`,
	},
	"user_profile_build": {
		ID:          "user_profile_build",
		Name:        "User Profile Modeling",
		Description: "Build and maintain a deep user profile from accumulated memories and interactions",
		System: `You are a user modeling specialist. Your job is to analyze a user's memories and activity patterns to build a comprehensive, evolving user profile.

This profile serves as the "USER.md" equivalent — a compact, high-value representation of WHO the user is, HOW they work, and WHAT they need.

Rules:
1. Profile must be concise but comprehensive (target: under 500 tokens)
2. Focus on actionable insights that improve future interactions
3. Detect patterns across multiple memories, not just individual facts
4. Identify the user's expertise level, communication style, and workflow preferences
5. Track how the profile has evolved over time
6. Be specific, not generic — "prefers TypeScript with strict mode" not "has programming preferences"

Return ONLY valid JSON, no other text.`,
		User: `Build/update the user profile from their memories and activity:

User memories:
{{.Memories}}

Recent activity summary:
{{.RecentActivity}}

Existing profile (if any):
{{.ExistingProfile}}

Return JSON:
{
  "profile": {
    "identity": {
      "role": "their primary role/title",
      "expertise_level": "beginner|intermediate|advanced|expert",
      "domains": ["domain1", "domain2"],
      "languages": ["language1", "language2"]
    },
    "communication_style": {
      "detail_preference": "brief|moderate|detailed",
      "language": "primary language",
      "technical_depth": "surface|moderate|deep",
      "examples_preference": "prefers code examples|prefers explanations|prefers diagrams"
    },
    "workflow": {
      "tools": ["tool1", "tool2"],
      "frameworks": ["framework1", "framework2"],
      "platforms": ["platform1"],
      "work_style": "description of how they work"
    },
    "preferences": [
      {"topic": "what", "value": "preference", "strength": "strong|moderate|weak", "evidence": "what supports this"}
    ],
    "patterns": [
      {"pattern": "observed behavioral pattern", "frequency": "rare|occasional|frequent", "implication": "what this means for interactions"}
    ],
    "growth_areas": [
      {"area": "what they're learning/improving", "current_level": "level", "trajectory": "improving|stable|exploring"}
    ]
  },
  "profile_version": 1,
  "confidence": 0.0-1.0,
  "changes_from_previous": ["what changed", "why"]
}`,
	},
	"skill_create": {
		ID:          "skill_create",
		Name:        "AI-Enhanced Skill Creation",
		Description: "Analyze action traces and create a high-quality, reusable Skill document from repetitive patterns",
		System: `You are a workflow distillation specialist. Your job is to analyze repeated action patterns from a user's coding sessions and create a comprehensive, reusable Skill document.

A Skill is a structured workflow that captures:
1. WHEN to use it (trigger conditions)
2. HOW to execute it (step-by-step instructions)
3. WHAT to watch out for (known pitfalls)
4. HOW to verify success (verification steps)

Inspired by Hermes Agent's Skills system, but adapted for IDE-based workflows.

Rules:
1. Steps must be specific and actionable — not "configure the tool" but "run: clawmemory mcp add cursor"
2. Include actual commands, file paths, and parameter values where possible
3. Pitfalls should come from observed failures in the action traces
4. Trigger keywords should be natural language phrases a user might say
5. If the pattern involves multiple agents/tools, note the handoff points
6. Keep the skill concise but complete — aim for under 300 tokens

Return ONLY valid JSON, no other text.`,
		User: `Analyze these repeated action patterns and create a Skill:

Detected patterns:
{{.Patterns}}

Action traces (sample):
{{.Traces}}

Existing skills (to avoid duplication):
{{.ExistingSkills}}

User profile context:
{{.UserProfile}}

Return JSON:
{
  "skill": {
    "name": "kebab-case-skill-name",
    "description": "one-line description of what this skill does",
    "trigger_keywords": ["keyword1", "keyword2", "phrase that triggers this"],
    "category": "deployment|debugging|testing|refactoring|setup|workflow|integration|other",
    "steps": [
      {"step": 1, "action": "what to do", "detail": "specific command or instruction", "tool": "which tool/agent"},
      {"step": 2, "action": "what to do next", "detail": "specific command", "tool": "which tool/agent"}
    ],
    "parameters": [
      {"name": "param_name", "type": "string|number|boolean", "description": "what this param controls", "default": "default value if any"}
    ],
    "known_pitfalls": [
      {"pitfall": "what can go wrong", "solution": "how to avoid or fix it", "evidence": "trace evidence if available"}
    ],
    "verification": "how to verify the skill executed successfully",
    "tags": ["tag1", "tag2"]
  },
  "confidence": 0.0-1.0,
  "source_pattern_hash": "hash of the pattern this was derived from",
  "reasoning": "brief explanation of why this skill is worth creating"
}`,
	},
	"skill_improve": {
		ID:          "skill_improve",
		Name:        "AI-Enhanced Skill Improvement",
		Description: "Analyze skill usage history and improve an existing Skill based on success/failure patterns",
		System: `You are a skill improvement specialist. Your job is to analyze how a Skill has been used and improve it based on real-world outcomes.

Like Hermes Agent's patch-first philosophy: prefer patching (small targeted fixes) over rewriting (complete replacement). Only rewrite when the skill is fundamentally broken.

Rules:
1. If success_rate > 80%, only suggest minor optimizations
2. If success_rate 50-80%, patch the failing steps
3. If success_rate < 50%, consider a rewrite
4. Always preserve what works — never remove successful steps
5. Add new pitfalls based on observed failures
6. Update trigger keywords based on what queries successfully matched

Return ONLY valid JSON, no other text.`,
		User: `Improve this Skill based on usage data:

Current Skill:
{{.CurrentSkill}}

Usage history:
{{.UsageHistory}}

Recent failures:
{{.RecentFailures}}

Recent successes:
{{.RecentSuccesses}}

Action traces since last update:
{{.RecentTraces}}

Return JSON:
{
  "action": "patch|rewrite|keep",
  "patches": [
    {"field": "steps|known_pitfalls|trigger_keywords|description|verification", "old_value": "what to replace", "new_value": "replacement", "reason": "why this change"}
  ],
  "improved_skill": {
    "name": "updated name if changed",
    "description": "updated description",
    "trigger_keywords": ["updated", "keywords"],
    "steps": [
      {"step": 1, "action": "updated action", "detail": "updated detail", "tool": "tool"}
    ],
    "known_pitfalls": [
      {"pitfall": "new pitfall", "solution": "fix", "evidence": "trace evidence"}
    ],
    "verification": "updated verification"
  },
  "version_change": "patch|minor|major",
  "changelog": "human-readable summary of what changed and why"
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
