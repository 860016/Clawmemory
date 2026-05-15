export = {
  id: "clawmemory",
  name: "ClawMemory Context Engine",
  description:
    "Use ClawMemory as OpenClaw's memory backend. Automatically records conversations and injects relevant memories into context.",
  kind: "context-engine",
  register: function (api: {
    registerContextEngine: (
      id: string,
      factory: (ctx: {
        config: Record<string, unknown>;
        agentDir: string;
        workspaceDir: string;
      }) => unknown
    ) => void;
  }): void {
    api.registerContextEngine("clawmemory", (ctx) => {
      const config: Record<string, unknown> = ctx.config || {};
      const baseUrl = (config.baseUrl as string) || "http://localhost:8765";
      const apiKey = (config.apiKey as string) || process.env.CLAWMEMORY_API_KEY || "";
      const maxContextMemories = (config.maxContextMemories as number) || 5;
      const enableAutoIngest = (config.enableAutoIngest as boolean) !== false;

      if (!apiKey) {
        console.warn("[ClawMemory] No API key configured. Set clawmemory.apiKey in plugin config or CLAWMEMORY_API_KEY env var.");
      }

      const NOISE_PATTERNS = [
        /^(ok|好的|嗯|yes|no|sure|got it|明白了|收到|thanks|谢谢|done|完成)$/i,
        /^(继续|continue|next|下一步|go ahead)$/i,
        /^\.{1,3}$/,
        /^\W*$/,
      ];
      const MIN_CONTENT_LENGTH = 5;

      function isNoisyContent(text: string): boolean {
        const trimmed = text.trim();
        if (trimmed.length < MIN_CONTENT_LENGTH) return true;
        for (const pattern of NOISE_PATTERNS) {
          if (pattern.test(trimmed)) return true;
        }
        return false;
      }

      function estimateTokens(text: string): number {
        return Math.ceil(text.length / 4);
      }

      function extractContent(content: unknown): string {
        if (typeof content === "string") return content;
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

      function extractLastUserMessage(messages: Array<{ role: string; content: unknown }>): string {
        for (let i = messages.length - 1; i >= 0; i--) {
          if (messages[i].role === "user") {
            return extractContent(messages[i].content);
          }
        }
        return "";
      }

      async function clawMemoryRequest(method: string, path: string, body?: unknown): Promise<unknown> {
        const url = `${baseUrl}/api/v1/external${path}`;
        const headers: Record<string, string> = {
          "Content-Type": "application/json",
          "X-API-Key": apiKey,
          "X-Platform": "openclaw",
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

      async function writeMemory(key: string, value: string, layer?: string, memoryType?: string): Promise<void> {
        try {
          await clawMemoryRequest("POST", "/memories", {
            key,
            value,
            layer: layer || "episodic",
            source: "openclaw",
            memory_type: memoryType || "conversation",
          });
        } catch (err) {
          console.error("[ClawMemory] Write memory failed:", err);
        }
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

      let turnCount = 0;

      return {
        info: {
          id: "clawmemory",
          name: "ClawMemory Context Engine",
          ownsCompaction: true,
        },

        async ingest({ sessionId, message, isHeartbeat }: { sessionId: string; message: { role: string; content: unknown }; isHeartbeat?: boolean }) {
          if (isHeartbeat || !enableAutoIngest) {
            return { ingested: false };
          }
          const content = extractContent(message.content);
          if (isNoisyContent(content)) {
            return { ingested: false };
          }
          await writeMemory(
            `${sessionId}_${message.role}_${Date.now()}`,
            `[${message.role}] ${content}`
          );
          return { ingested: true };
        },

        async ingestBatch({ sessionId, messages }: { sessionId: string; messages: Array<{ role: string; content: unknown }> }) {
          if (!enableAutoIngest) {
            return { ingested: false };
          }
          const filtered = messages.filter(m => !isNoisyContent(extractContent(m.content)));
          if (filtered.length === 0) {
            return { ingested: false };
          }
          const combined = filtered
            .map(m => `[${m.role}] ${extractContent(m.content)}`)
            .join("\n");
          await writeMemory(`${sessionId}_batch_${Date.now()}`, combined);
          return { ingested: true };
        },

        async afterTurn({ sessionId }: { sessionId: string }) {
          turnCount++;
          try {
            await clawMemoryRequest("POST", "/sessions/track", {
              session_id: sessionId,
              metadata: `last_active:${new Date().toISOString()}`,
            });
          } catch {}
          return { ok: true };
        },

        async assemble({ sessionId, messages, tokenBudget }: { sessionId: string; messages: Array<{ role: string; content: unknown }>; tokenBudget: number }) {
          if (enableAutoIngest && messages.length > 0) {
            const lastMsg = messages[messages.length - 1];
            if (lastMsg) {
              const content = extractContent(lastMsg.content);
              if (!isNoisyContent(content)) {
                writeMemory(
                  `${sessionId}_${lastMsg.role}_${Date.now()}`,
                  `[${lastMsg.role}] ${content}`
                ).catch(() => {});
              }
            }
          }

          const lastUserMsg = extractLastUserMessage(messages);
          let systemPromptAddition = "";

          if (lastUserMsg) {
            const queryWords = lastUserMsg.slice(0, 200);
            const contextMemories = await searchMemories(queryWords, maxContextMemories);

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

        async compact() {
          return { ok: true, compacted: true };
        },

        async dispose() {},
      };
    });
  },
};
