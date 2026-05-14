export class ClawMemoryClient {
    baseUrl;
    apiKey;
    platform;
    constructor(config) {
        this.baseUrl = config.baseUrl.replace(/\/$/, "");
        this.apiKey = config.apiKey;
        this.platform = config.platform || "mcp";
    }
    async request(method, path, body) {
        const url = `${this.baseUrl}/api/v1/external${path}`;
        const headers = {
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
    async saveMemory(params) {
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
    async batchSaveMemories(memories) {
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
    async searchMemories(query, limit = 10) {
        const result = (await this.request("GET", `/memories/search?q=${encodeURIComponent(query)}&limit=${limit}`));
        return result.items || [];
    }
    async getContext(query, limit = 5) {
        const result = (await this.request("GET", `/memories/context?q=${encodeURIComponent(query)}&limit=${limit}`));
        return result;
    }
    async reason(params) {
        return this.request("POST", "/reason", {
            query: params.query,
            depth: params.depth || 1,
            level: params.level || "medium",
            session_id: params.session_id || "",
        });
    }
    async pushConversation(params) {
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
    async trackSession(sessionId, metadata) {
        return this.request("POST", "/sessions/track", {
            session_id: sessionId,
            metadata: metadata || "",
        });
    }
}
