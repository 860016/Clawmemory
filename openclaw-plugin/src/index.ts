interface IngestParams {
  sessionId: string;
  message: Message;
  isHeartbeat?: boolean;
}

interface IngestBatchParams {
  sessionId: string;
  messages: Message[];
}

interface AfterTurnParams {
  sessionId: string;
}

interface AssembleParams {
  sessionId: string;
  messages: Message[];
  tokenBudget: number;
  availableTools?: Set<string>;
  citationsMode?: string;
}

interface CompactParams {
  sessionId: string;
  force: boolean;
}

interface Message {
  role: "user" | "assistant" | "system";
  content: string | Record<string, unknown> | unknown[];
  timestamp?: string;
  metadata?: Record<string, unknown>;
}

interface ContextEngineInfo {
  id: string;
  name: string;
  ownsCompaction: boolean;
}

interface ContextEngine {
  info: ContextEngineInfo;
  ingest(params: IngestParams): Promise<{ ingested: boolean }>;
  ingestBatch?(params: IngestBatchParams): Promise<{ ingested: boolean }>;
  afterTurn?(params: AfterTurnParams): Promise<{ ok: boolean }>;
  assemble(params: AssembleParams): Promise<AssembleResult>;
  compact(params: CompactParams): Promise<{ ok: boolean; compacted: boolean }>;
  dispose?(): Promise<void>;
}

interface AssembleResult {
  messages: Message[];
  estimatedTokens: number;
  systemPromptAddition?: string;
}

interface OpenClawPluginApi {
  registerContextEngine(id: string, factory: (ctx: PluginContext) => ContextEngine): void;
}

interface PluginContext {
  config: Record<string, unknown>;
  agentDir: string;
  workspaceDir: string;
}

interface ClawMemoryConfig {
  baseUrl: string;
  apiKey: string;
  maxContextMemories?: number;
  enableAutoIngest?: boolean;
}

interface ConversationTurn {
  session_id: string;
  agent_name: string;
  messages: Array<{ role: string; content: string }>;
  summary?: string;
}

export = function register(api: OpenClawPluginApi): void {
  api.registerContextEngine("clawmemory", (ctx: PluginContext) => {
    const config: ClawMemoryConfig = {
      baseUrl: (ctx.config.baseUrl as string) || "http://localhost:8765",
      apiKey: (ctx.config.apiKey as string) || "",
      maxContextMemories: (ctx.config.maxContextMemories as number) || 5,
      enableAutoIngest: (ctx.config.enableAutoIngest as boolean) !== false,
    };

    if (!config.apiKey) {
      console.warn("[ClawMemory] No API key configured. Set clawmemory.apiKey in plugin config.");
    }

    async function clawMemoryRequest(
      method: string,
      path: string,
      body?: unknown
    ): Promise<unknown> {
      const url = `${config.baseUrl}/api/v1/external${path}`;
      const headers: Record<string, string> = {
        "Content-Type": "application/json",
        "X-API-Key": config.apiKey,
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

    async function searchMemories(query: string, limit: number): Promise<unknown[]> {
      try {
        const result = (await clawMemoryRequest(
          "GET",
          `/memories/context?q=${encodeURIComponent(query)}&limit=${limit}`
        )) as Record<string, unknown>;
        if (result && Array.isArray(result.memories)) {
          return result.memories;
        }
        return [];
      } catch (err) {
        console.error("[ClawMemory] Search failed:", err);
        return [];
      }
    }

    async function pushConversation(turn: ConversationTurn): Promise<void> {
      try {
        await clawMemoryRequest("POST", "/conversations", turn);
      } catch (err) {
        console.error("[ClawMemory] Push conversation failed:", err);
      }
    }

    async function trackSession(sessionId: string, metadata?: string): Promise<void> {
      try {
        await clawMemoryRequest("POST", "/sessions/track", {
          session_id: sessionId,
          metadata: metadata || "",
        });
      } catch (err) {
        console.error("[ClawMemory] Session track failed:", err);
      }
    }

    function estimateTokens(text: string): number {
      return Math.ceil(text.length / 4);
    }

    function extractContent(content: unknown): string {
      if (typeof content === "string") {
        return content;
      }
      if (Array.isArray(content)) {
        return content
          .map((part: unknown) => {
            if (typeof part === "string") return part;
            if (part && typeof part === "object" && "text" in (part as Record<string, unknown>)) {
              return String((part as Record<string, unknown>).text);
            }
            return "";
          })
          .filter(Boolean)
          .join(" ");
      }
      if (content && typeof content === "object" && "text" in (content as Record<string, unknown>)) {
        return String((content as Record<string, unknown>).text);
      }
      return "";
    }

    function extractLastUserMessage(messages: Message[]): string {
      for (let i = messages.length - 1; i >= 0; i--) {
        if (messages[i].role === "user") {
          return extractContent(messages[i].content);
        }
      }
      return "";
    }

    return {
      info: {
        id: "clawmemory",
        name: "ClawMemory Context Engine",
        ownsCompaction: true,
      },

      async ingest({ sessionId, message, isHeartbeat }: IngestParams) {
        if (isHeartbeat || !config.enableAutoIngest) {
          return { ingested: false };
        }

        await pushConversation({
          session_id: sessionId,
          agent_name: "openclaw",
          messages: [{
            role: message.role,
            content: extractContent(message.content),
          }],
        });

        return { ingested: true };
      },

      async ingestBatch({ sessionId, messages }: IngestBatchParams) {
        if (!config.enableAutoIngest) {
          return { ingested: false };
        }

        await pushConversation({
          session_id: sessionId,
          agent_name: "openclaw",
          messages: messages.map(m => ({
            role: m.role,
            content: extractContent(m.content),
          })),
        });

        return { ingested: true };
      },

      async afterTurn({ sessionId }: AfterTurnParams) {
        await trackSession(sessionId, `last_active:${new Date().toISOString()}`);
        return { ok: true };
      },

      async assemble({ sessionId, messages, tokenBudget }: AssembleParams) {
        const lastUserMsg = extractLastUserMessage(messages);

        let systemPromptAddition = "";
        let contextMemories: unknown[] = [];

        if (lastUserMsg) {
          const queryWords = lastUserMsg.slice(0, 200);
          contextMemories = await searchMemories(queryWords, config.maxContextMemories!);

          if (Array.isArray(contextMemories) && contextMemories.length > 0) {
            const memoryLines: string[] = [];
            for (const m of contextMemories as Array<{ key?: string; value?: string }>) {
              if (m.key && m.value) {
                memoryLines.push(`- ${m.key}: ${m.value}`);
              }
            }
            if (memoryLines.length > 0) {
              systemPromptAddition =
                "\n\n[ClawMemory Relevant Memories]\nYou have access to the following relevant memories from past conversations. Use them naturally when responding:\n" +
                memoryLines.join("\n");
            }
          }
        }

        const totalTokens =
          estimateTokens(systemPromptAddition) +
          messages.reduce((sum, m) => sum + estimateTokens(extractContent(m.content)), 0);

        return {
          messages,
          estimatedTokens: totalTokens,
          ...(systemPromptAddition ? { systemPromptAddition } : {}),
        };
      },

      async compact({ sessionId, force }: CompactParams) {
        try {
          await pushConversation({
            session_id: sessionId,
            agent_name: "openclaw",
            messages: [],
            summary: `[compact] Session compacted at ${new Date().toISOString()} (force: ${force})`,
          });
        } catch {
          // ignore
        }

        return { ok: true, compacted: true };
      },

      async dispose() {
        // no pending state to flush
      },
    };
  });
};
