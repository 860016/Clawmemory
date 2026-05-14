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
export declare class ClawMemoryClient {
    private baseUrl;
    private apiKey;
    private platform;
    constructor(config: ClawMemoryConfig);
    private request;
    saveMemory(params: {
        key: string;
        value: string;
        layer?: string;
        source?: string;
        memory_type?: string;
        visibility?: string;
        source_agent?: string;
    }): Promise<unknown>;
    batchSaveMemories(memories: Array<{
        key: string;
        value: string;
        layer?: string;
        source?: string;
        memory_type?: string;
        visibility?: string;
        source_agent?: string;
    }>): Promise<unknown>;
    searchMemories(query: string, limit?: number): Promise<MemoryItem[]>;
    getContext(query: string, limit?: number): Promise<ContextResult>;
    reason(params: {
        query: string;
        depth?: number;
        level?: string;
        session_id?: string;
    }): Promise<unknown>;
    pushConversation(params: {
        session_id: string;
        agent_name?: string;
        messages?: Array<{
            role: string;
            content: string;
        }>;
        summary?: string;
        title?: string;
        project_path?: string;
        visibility?: string;
    }): Promise<unknown>;
    trackSession(sessionId: string, metadata?: string): Promise<unknown>;
}
