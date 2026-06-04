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
IMPORTANT: You MUST output entity names, descriptions, and relation types in the SAME LANGUAGE as the input content. If the input is in Chinese, all output must be in Chinese. If the input is in English, output in English.
Return ONLY valid JSON, no other text.`,
		User: `Extract entities and relations from these memories:

{{.Memories}}

Return JSON format:
{
  "entities": [
    {"name": "实体名称（必须与输入语言一致）", "type": "person|organization|technology|location|event|concept", "description": "简要描述", "confidence": 0.0-1.0}
  ],
  "relations": [
    {"source": "实体A", "target": "实体B", "type": "关系类型", "description": "关系描述", "confidence": 0.0-1.0}
  ]
}

Relation types: works_at, manages, depends_on, located_in, member_of, uses, created_by, related_to, part_of, owns, knows, reports_to, 属于, 管理, 依赖, 位于, 使用, 创建, 相关, 包含

CRITICAL: All entity names, descriptions, and relation types MUST be in the same language as the input memories. Never translate Chinese content to English.`,
	},
	"discover_projects": {
		ID:          "discover_projects",
		Name:        "Project Discovery",
		Description: "Discover projects from memories using AI",
		System: `You are a project analysis expert. Analyze the user's memory entries to discover distinct projects they are working on.
IMPORTANT: You MUST output all text in the SAME LANGUAGE as the input content. If the input is in Chinese, all output must be in Chinese.
Return ONLY valid JSON, no other text.`,
		User: `Analyze these memories and discover distinct projects the user is working on:

{{.Memories}}

Existing projects (do NOT duplicate these):
{{.ExistingProjects}}

Return JSON format:
{
  "projects": [
    {
      "name": "项目名称",
      "description": "项目描述，包括目标和技术栈",
      "category": "分类（如：前端开发、后端开发、数据分析等）",
      "status": "active|paused|completed",
      "progress": 0-100,
      "key_decisions": ["关键决策1", "关键决策2"],
      "action_items": ["待办事项1", "待办事项2"],
      "confidence": 0.0-1.0
    }
  ]
}

Rules:
1. Only include REAL projects with clear evidence in the memories
2. Do NOT duplicate existing projects
3. Estimate progress based on completed vs pending items
4. Extract key decisions and action items from the memories
5. Set confidence based on how clearly the project is defined`,
	},
	"conflict_scan": {
		ID:          "conflict_scan",
		Name:        "Semantic Conflict Detection",
		Description: "Detect semantic conflicts between memories",
		System: `You are a data consistency expert. Identify semantic conflicts between memory entries.
IMPORTANT: You MUST output descriptions and suggestions in the SAME LANGUAGE as the input content. If the input is in Chinese, all output must be in Chinese.
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
    {"memory_a_id": 1, "memory_b_id": 2, "type": "attribute|timeline|logic", "description": "矛盾描述（用输入内容的语言）", "severity": "high|medium|low", "suggestion": "合并建议（用输入内容的语言）"}
  ]
}

CRITICAL: All descriptions and suggestions MUST be in the same language as the input memories.`,
	},
	"decay_evaluate": {
		ID:          "decay_evaluate",
		Name:        "Smart Decay Evaluation",
		Description: "AI-powered memory decay assessment",
		System: `You are a memory management expert. Evaluate which memories should decay, be archived, or kept.
IMPORTANT: You MUST output reasons in the SAME LANGUAGE as the input content. If the input is in Chinese, all output must be in Chinese.
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
    {"id": 1, "action": "keep|archive|merge|delete", "reason": "原因说明（用输入内容的语言）", "new_importance": 0.0-1.0, "merge_with": []}
  ]
}`,
	},
	"daily_report": {
		ID:          "daily_report",
		Name:        "Daily Report Generation",
		Description: "Generate intelligent daily report from memories",
		System: `You are a knowledge management assistant. Generate a concise daily report based on the user's memory activity.
IMPORTANT: You MUST output the report in the SAME LANGUAGE as the input content. If the input is in Chinese, all output must be in Chinese.
Return ONLY valid JSON, no other text.`,
		User: `Generate a daily report based on today's activity:

Date: {{.Date}}
New memories: {{.MemoryCount}}
Key memories:
{{.Highlights}}

Return JSON:
{
  "summary": "今日总结（用输入内容的语言）",
  "highlights": ["关键发现1", "关键发现2", "关键发现3"],
  "knowledge_gained": ["学到了什么1", "学到了什么2"],
  "pending_tasks": ["待办1", "待办2"],
  "tomorrow_suggestions": ["建议1", "建议2"],
  "mood": "productive|exploratory|routine|intensive"
}`,
	},
	"wiki_generate": {
		ID:          "wiki_generate",
		Name:        "Wiki Auto-Generation",
		Description: "Auto-generate wiki pages from memories",
		System: `You are a technical documentation expert. Generate structured wiki documentation from memory entries.
IMPORTANT: You MUST output the wiki content in the SAME LANGUAGE as the input content. If the input is in Chinese, all output must be in Chinese.
Return ONLY valid JSON, no other text.`,
		User: `Generate a wiki document from these memories:

Topic: {{.Topic}}
Related memories:
{{.Memories}}

Return JSON:
{
  "title": "文档标题（用输入内容的语言）",
  "category": "分类名称",
  "content": "完整wiki内容（markdown格式，用输入内容的语言）",
  "summary": "2-3句总结",
  "tags": ["标签1", "标签2"],
  "key_decisions": ["关键决策1", "关键决策2"],
  "action_items": ["行动项1", "行动项2"]
}`,
	},
	"compress": {
		ID:          "compress",
		Name:        "Memory Compression",
		Description: "Compress and refine multiple memories into one",
		System: `You are an information refinement expert. Merge and compress multiple related memories into a single high-quality entry.
IMPORTANT: You MUST output the compressed content in the SAME LANGUAGE as the input content. If the input is in Chinese, all output must be in Chinese.
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
  "key": "压缩后的记忆键名（用输入内容的语言）",
  "value": "压缩后的记忆内容（用输入内容的语言）",
  "importance": 0.0-1.0,
  "layer": "core|context|detail",
  "tags": ["标签1", "标签2"],
  "merged_count": 3,
  "notes": "压缩说明"
}`,
	},
	"evolution_discover": {
		ID:          "evolution_discover",
		Name:        "Deep Relation Discovery",
		Description: "Discover hidden relationships between memories",
		System: `You are a knowledge discovery expert. Find hidden, non-obvious relationships between memories.
IMPORTANT: You MUST output descriptions in the SAME LANGUAGE as the input content. If the input is in Chinese, all output must be in Chinese.
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
    {"source_id": 1, "target_id": 2, "type": "causes|precedes|depends_on|enables|contradicts|refines", "description": "关系说明（用输入内容的语言）", "confidence": 0.0-1.0}
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

CRITICAL LANGUAGE RULE: You MUST extract and output ALL content in the SAME LANGUAGE as the conversation. If the conversation is in Chinese, ALL extracted facts, preferences, and relations MUST be in Chinese. NEVER translate Chinese content to English.

Rules:
1. Extract ONLY information worth remembering long-term
2. Each fact should be atomic - one piece of information per entry
3. Preferences should capture user's likes, dislikes, and tendencies
4. Relationships should capture connections between entities
5. Ignore greetings, small talk, and procedural messages
6. Be specific - "prefers dark mode" not "has UI preference"
7. NEVER translate - if user says "我喜欢Python", extract "我喜欢Python" NOT "I like Python"

Return ONLY valid JSON, no other text.`,
		User: `Extract facts, preferences, and relationships from this conversation:

{{.Messages}}

Existing memories for deduplication:
{{.ExistingMemories}}

Return JSON:
{
  "facts": [
    {"content": "提取的事实（必须与对话语言一致）", "category": "identity|preference|skill|possession|relationship|routine|goal|opinion", "confidence": 0.0-1.0, "source": "user|assistant|inferred"}
  ],
  "preferences": [
    {"topic": "偏好主题（用对话语言）", "value": "偏好值（用对话语言）", "strength": "strong|moderate|weak", "confidence": 0.0-1.0}
  ],
  "relations": [
    {"subject": "实体A（用对话语言）", "predicate": "关系类型", "object": "实体B（用对话语言）", "confidence": 0.0-1.0}
  ],
  "updates": [
    {"old_fact": "旧事实（用对话语言）", "new_fact": "新事实（用对话语言）", "reason": "变更原因"}
  ]
}

CRITICAL: If the conversation is in Chinese, ALL output fields MUST be in Chinese. Never translate Chinese to English.`,
	},
	"memory_consolidate": {
		ID:          "memory_consolidate",
		Name:        "Memory Consolidation",
		Description: "Consolidate and deduplicate extracted facts with existing memories",
		System: `You are a memory consolidation expert. Your job is to merge new facts with existing memories, resolving conflicts and removing duplicates.
IMPORTANT: You MUST output all content in the SAME LANGUAGE as the input. If the input is in Chinese, all output must be in Chinese.

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
    {"key": "记忆键名（用输入语言）", "value": "记忆内容（用输入语言）", "layer": "core|context|detail", "importance": 0.0-1.0, "category": "fact|preference|skill|relationship"}
  ],
  "update": [
    {"memory_id": 1, "field": "value", "old_value": "旧值", "new_value": "新值", "reason": "更新原因"}
  ],
  "merge": [
    {"source_ids": [1, 2], "merged_key": "合并后键名", "merged_value": "合并后内容", "layer": "core|context|detail"}
  ],
  "supersede": [
    {"old_id": 1, "reason": "被新信息替代的原因", "new_id": 2}
  ],
  "skip": [
    {"fact": "重复的事实", "matches_memory_id": 1}
  ]
}`,
	},
	"context_assemble": {
		ID:          "context_assemble",
		Name:        "Smart Context Assembly",
		Description: "Assemble optimal context for LLM from memories based on query",
		System: `You are a context optimization expert. Given a user query and available memories, select and organize the most relevant context to include in the LLM prompt.
IMPORTANT: You MUST output relevance reasons and suggestions in the SAME LANGUAGE as the query. If the query is in Chinese, all output must be in Chinese.

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
    {"memory_id": 1, "relevance_score": 0.0-1.0, "relevance_reason": "相关性说明（用查询语言）", "tokens": 50}
  ],
  "total_tokens": 500,
  "coverage_score": 0.0-1.0,
  "missing_context": ["缺失的信息"],
  "suggested_followup": ["建议的后续问题"]
}`,
	},
	"nudge_reflect": {
		ID:          "nudge_reflect",
		Name:        "Periodic Nudge Reflection",
		Description: "Periodically review recent activity and extract high-value knowledge worth persisting",
		System: `You are a knowledge distillation specialist. Your job is to review recent memory activity and decide what is worth permanently remembering, what should be compressed, and what can be forgotten.

Philosophy: Only remember information that will influence FUTURE behavior. Discard everything else.

IMPORTANT: You MUST output all content in the SAME LANGUAGE as the input. If the input is in Chinese, all output must be in Chinese.

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
    {"content": "值得记住的内容（用输入语言）", "category": "identity|preference|skill|possession|relationship|routine|goal|opinion|correction", "confidence": 0.0-1.0, "reason": "为什么值得记住"}
  ],
  "compress": [
    {"memory_ids": [1, 2, 3], "compressed_content": "统一摘要（用输入语言）", "reason": "为什么应该合并"}
  ],
  "forget": [
    {"memory_id": 4, "reason": "为什么可以安全遗忘"}
  ],
  "profile_updates": [
    {"field": "communication_style|tech_stack|workflow|preference", "old_value": "旧值", "new_value": "新值", "evidence": "支持此变更的证据"}
  ],
  "insights": [
    {"insight": "从数据中观察到的高层模式（用输入语言）", "confidence": 0.0-1.0}
  ]
}`,
	},
	"self_refine": {
		ID:          "self_refine",
		Name:        "Memory Self-Refinement",
		Description: "Under capacity pressure, automatically distill memories to retain only the highest-value information",
		System: `You are a memory refinement engine operating under strict capacity constraints. Your job is to distill a set of memories into a smaller, denser set that preserves maximum information value.

Core principle: Information economy — every character must earn its place.

IMPORTANT: You MUST output all content in the SAME LANGUAGE as the input memories. If the input is in Chinese, all output must be in Chinese.

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
    {"memory_id": 1, "reason": "为什么必须原样保留"}
  ],
  "merge": [
    {"source_ids": [2, 3, 4], "merged_key": "统一键名（用输入语言）", "merged_value": "压缩但完整的内容（用输入语言）", "layer": "core|context|detail", "importance": 0.0-1.0}
  ],
  "demote": [
    {"memory_id": 5, "from_layer": "context", "to_layer": "detail", "reason": "为什么应该降级"}
  ],
  "archive": [
    {"memory_id": 6, "reason": "为什么可以归档（不删除，仅设为不活跃）"}
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

IMPORTANT: You MUST output all content in the SAME LANGUAGE as the input. If the input is in Chinese, all output must be in Chinese.

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
      "role": "用户的主要角色/头衔",
      "expertise_level": "beginner|intermediate|advanced|expert",
      "domains": ["领域1", "领域2"],
      "languages": ["语言1", "语言2"]
    },
    "communication_style": {
      "detail_preference": "brief|moderate|detailed",
      "language": "主要语言",
      "technical_depth": "surface|moderate|deep",
      "examples_preference": "偏好代码示例|偏好解释|偏好图表"
    },
    "workflow": {
      "tools": ["工具1", "工具2"],
      "frameworks": ["框架1", "框架2"],
      "platforms": ["平台1"],
      "work_style": "工作方式描述"
    },
    "preferences": [
      {"topic": "什么方面", "value": "偏好", "strength": "strong|moderate|weak", "evidence": "支持此偏好的证据"}
    ],
    "patterns": [
      {"pattern": "观察到的行为模式", "frequency": "rare|occasional|frequent", "implication": "这对交互意味着什么"}
    ],
    "growth_areas": [
      {"area": "正在学习/改进的领域", "current_level": "当前水平", "trajectory": "improving|stable|exploring"}
    ]
  },
  "profile_version": 1,
  "confidence": 0.0-1.0,
  "changes_from_previous": ["变更内容", "原因"]
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

IMPORTANT: You MUST output all descriptions, steps, pitfalls, and reasoning in the SAME LANGUAGE as the input. If the input is in Chinese, all output must be in Chinese.

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

IMPORTANT: You MUST output all content in the SAME LANGUAGE as the input. If the input is in Chinese, all output must be in Chinese.

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
