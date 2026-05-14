export interface ClawMemoryConfig {
  baseUrl: string;
  apiKey: string;
  platform: string;
}

export interface MemoryItem {
  id: number;
  key: string;
  value: string;
  layer: string;
  source: string;
  platform: string;
  memory_type: string;
  importance: number;
  created_at: string;
  updated_at: string;
}

export interface ContextResult {
  memories: MemoryItem[];
  count: number;
  system_prompt_addition: string;
}

export class ClawMemoryClient {
  private baseUrl: string;
  private apiKey: string;
  private platform: string;

  constructor(config: ClawMemoryConfig) {
    this.baseUrl = config.baseUrl.replace(/\/$/, "");
    this.apiKey = config.apiKey;
    this.platform = config.platform || "mcp";
  }

  private async request(
    method: string,
    path: string,
    body?: unknown
  ): Promise<unknown> {
    const url = `${this.baseUrl}/api/v1/external${path}`;
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
      "X-API-Key": this.apiKey,
      "X-Platform": this.platform,
    };

    const response = await fetch(url, {
      method,
      headers,
      body: body ? JSON.stringify(body) : undefined,
    });

    if (!response.ok) {
      const text = await response.text();
      throw new Error(`ClawMemory API error ${response.status}: ${text}`);
    }

    return response.json();
  }

  async saveMemory(params: {
    key: string;
    value: string;
    layer?: string;
    source?: string;
    memory_type?: string;
    visibility?: string;
    source_agent?: string;
  }): Promise<unknown> {
    return this.request("POST", "/memories", {
      key: params.key,
      value: params.value,
      layer: params.layer || "episodic",
      source: params.source || this.platform,
      memory_type: params.memory_type || "knowledge",
      visibility: params.visibility || "private",
      source_agent: params.source_agent || this.platform,
    });
  }

  async batchSaveMemories(
    memories: Array<{
      key: string;
      value: string;
      layer?: string;
      source?: string;
      memory_type?: string;
      visibility?: string;
      source_agent?: string;
    }>
  ): Promise<unknown> {
    return this.request("POST", "/memories/batch", {
      memories: memories.map((m) => ({
        key: m.key,
        value: m.value,
        layer: m.layer || "episodic",
        source: m.source || this.platform,
        memory_type: m.memory_type || "knowledge",
        visibility: m.visibility || "private",
        source_agent: m.source_agent || this.platform,
      })),
    });
  }

  async searchMemories(query: string, limit: number = 10): Promise<MemoryItem[]> {
    const result = (await this.request(
      "GET",
      `/memories/search?q=${encodeURIComponent(query)}&limit=${limit}`
    )) as Record<string, unknown>;
    return (result.items as MemoryItem[]) || [];
  }

  async getContext(query: string, limit: number = 5): Promise<ContextResult> {
    const result = (await this.request(
      "GET",
      `/memories/context?q=${encodeURIComponent(query)}&limit=${limit}`
    )) as ContextResult;
    return result;
  }

  async reason(params: {
    query: string;
    depth?: number;
    level?: string;
    session_id?: string;
  }): Promise<unknown> {
    return this.request("POST", "/reason", {
      query: params.query,
      depth: params.depth || 1,
      level: params.level || "medium",
      session_id: params.session_id || "",
    });
  }

  async pushConversation(params: {
    session_id: string;
    agent_name?: string;
    messages?: Array<{ role: string; content: string }>;
    summary?: string;
    title?: string;
    project_path?: string;
    visibility?: string;
  }): Promise<unknown> {
    return this.request("POST", "/conversations", {
      session_id: params.session_id,
      agent_name: params.agent_name || this.platform,
      platform: this.platform,
      messages: params.messages || [],
      summary: params.summary || "",
      title: params.title || "",
      project_path: params.project_path || "",
      visibility: params.visibility || "private",
    });
  }

  async trackSession(
    sessionId: string,
    metadata?: string
  ): Promise<unknown> {
    return this.request("POST", "/sessions/track", {
      session_id: sessionId,
      metadata: metadata || "",
    });
  }
}
