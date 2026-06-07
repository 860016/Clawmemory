// Core domain types for ClawMemory API

// --- Memory ---

export type MemoryLayer = 'core' | 'context' | 'detail'
export type MemoryStatus = 'active' | 'archived' | 'trashed'
export type MemoryVisibility = 'private' | 'shared' | 'public'

export interface Memory {
  id: number
  user_id: number
  key: string
  value: string
  layer: MemoryLayer
  importance: number
  status: MemoryStatus
  source: string
  tags: string
  visibility: MemoryVisibility
  reinforce_count: number
  decay_stage: number
  memory_type: string
  source_agent: string
  is_encrypted: boolean
  summary: string
  access_count: number
  platform: string
  verify_count: number
  validation_status: string
  score?: number
  score_detail?: Record<string, number>
  load_level?: string
  verified_at?: string
  created_at: string
  updated_at: string
  trashed_at?: string
  accessed_at?: string
}

export interface MemoryCreateParams {
  key: string
  value: string
  layer?: MemoryLayer
  importance?: number
  tags?: string[] | string
  source?: string
  source_agent?: string
  memory_type?: string
  visibility?: MemoryVisibility
  is_encrypted?: boolean
}

export interface MemoryUpdateParams {
  key?: string
  value?: string
  layer?: MemoryLayer
  importance?: number
  tags?: string[] | string
  source?: string
  visibility?: MemoryVisibility
}

export interface MemorySearchParams {
  q: string
  mode?: 'keyword' | 'semantic' | 'graphrag' | 'graph-rag' | 'smart'
  limit?: number
}

export interface SmartLoadParams {
  q?: string
  token_budget?: number
  load_level?: string
  source_agent?: string
  workspace?: string
}

export interface MemoryListParams {
  layer?: string
  page?: number
  size?: number
  status?: string
  memory_type?: string
  source_agent?: string
  visibility?: string
}

// --- Entity ---

export type EntityType = 'person' | 'organization' | 'technology' | 'location' | 'event' | 'concept'

export interface Entity {
  id: number
  user_id: number
  name: string
  entity_type: EntityType
  description: string
  confidence: number
  extract_method: string
  source_memory_id?: number
  properties?: string
  created_at: string
  updated_at: string
}

export interface EntityCreateParams {
  name: string
  entity_type: string
  description?: string
  confidence?: number
  properties?: Record<string, unknown>
}

export interface EntityUpdateParams {
  name?: string
  entity_type?: string
  description?: string
  confidence?: number
  properties?: Record<string, unknown>
}

// --- Relation ---

export interface Relation {
  id: number
  user_id: number
  source_id: number
  target_id: number
  relation_type: string
  weight: number
  properties?: string
  created_at: string
}

export interface RelationCreateParams {
  source_id: number
  target_id: number
  relation_type: string
  weight?: number
}

// --- Wiki ---

export type WikiStatus = 'draft' | 'in_progress' | 'completed'

export interface Wiki {
  id: number
  user_id: number
  title: string
  content: string
  category: string
  tags: string
  status: WikiStatus
  summary: string
  is_public: boolean
  is_pinned: boolean
  parent_id?: number
  ai_generated: boolean
  ai_confidence: number
  key_decisions: string
  action_items: string
  view_count: number
  created_at: string
  updated_at: string
}

export interface WikiCreateParams {
  title: string
  content?: string
  category?: string
  tags?: string
  status?: WikiStatus
  parent_id?: number
}

export interface WikiUpdateParams {
  title?: string
  content?: string
  category?: string
  tags?: string
  status?: WikiStatus
  parent_id?: number
}

// --- Project ---

export type ProjectStatus = 'active' | 'completed' | 'archived'

export interface Project {
  id: number
  user_id: number
  name: string
  description: string
  status: ProjectStatus
  category: string
  tags: string
  key_decisions: string
  action_items: string
  progress: number
  source_session_id: string
  source_agent: string
  is_pinned: boolean
  last_discussed_at?: string
  created_at: string
  updated_at: string
}

export interface ProjectCreateParams {
  name: string
  description?: string
  category?: string
  status?: ProjectStatus
  tags?: string
  key_decisions?: string | unknown[]
  action_items?: string | unknown[]
  progress?: number
  source_agent?: string
}

// --- Knowledge Graph ---

export interface GraphNode {
  id: string
  label: string
  type: string
  data?: Record<string, unknown>
}

export interface GraphEdge {
  id: string
  source: string
  target: string
  label: string
  weight: number
}

export interface GraphData {
  nodes: GraphNode[]
  edges: GraphEdge[]
}

// --- Decay ---

export interface DecayStats {
  processed: number
  archived: number
  trashed: number
  adjusted: number
  reinforced: number
  locked: number
  algorithm: string
}

export interface DecaySettings {
  enabled: boolean
}

// --- Governance ---

export interface GovernanceStepResult {
  step: string
  status: 'success' | 'error' | 'skipped'
  duration_ms: number
  error?: string
  details?: Record<string, unknown>
}

export interface GovernanceStatus {
  last_run?: string
  steps: GovernanceStepResult[]
  total_duration_ms: number
}

// --- Search ---

export interface SearchResult {
  id: number
  key: string
  value: string
  layer: string
  importance: number
  score?: number
  score_detail?: {
    keyword: number
    semantic: number
    combined: number
  }
}

// --- Dedup ---

export interface DuplicateGroup {
  key: string
  similarity: number
  memories: {
    id: number
    key: string
    value: string
    layer: string
    importance: number
    source: string
    created_at: string
  }[]
}

// --- Skill ---

export interface Skill {
  id: number
  user_id: number
  name: string
  description: string
  pattern: string
  category: string
  usage_count: number
  confidence: number
  auto_created: boolean
  created_at: string
}

// --- Share ---

export interface ShareRule {
  id: number
  user_id: number
  name: string
  source_agent: string
  target_agent: string
  layer: string
  min_importance: number
  auto_approve: boolean
  enabled: boolean
  created_at: string
}

// --- Session Memory ---

export interface SessionMemory {
  id: number
  user_id: number
  session_id: string
  title: string
  current_state: string
  task_spec: string
  files_and_funcs: string
  worklog: string
  learnings: string
  key_results: string
  docs: string
  errors: string
  workflow: string
  status: string
  key: string
  value: string
  layer: string
  token_count: number
  compressed_from: string
  created_at: string
  updated_at: string
}

// --- Settings ---

export interface AppSettings {
  [key: string]: string | number | boolean | null
}

// --- Report ---

export interface DailyReport {
  id: number
  user_id: number
  date: string
  content: string
  summary: string
  created_at: string
}
