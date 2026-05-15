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
        const maxRetries = 2;
        let lastErr: unknown;
        for (let attempt = 0; attempt <= maxRetries; attempt++) {
          try {
            await clawMemoryRequest("POST", "/memories", {
              key,
              value,
              layer: layer || "episodic",
              source: "openclaw",
              memory_type: memoryType || "conversation",
            });
            return;
          } catch (err) {
            lastErr = err;
            console.error(`[ClawMemory] Write memory failed (attempt ${attempt + 1}/${maxRetries + 1}):`, err);
            if (attempt < maxRetries) {
              await new Promise(r => setTimeout(r, (attempt + 1) * 1000));
            }
          }
        }
        console.error("[ClawMemory] Write memory failed after all retries:", lastErr);
      }

      async function writeMemoryBatch(items: Array<{ key: string; value: string; layer?: string; memoryType?: string }>): Promise<void> {
        if (items.length === 0) return;
        if (items.length === 1) {
          await writeMemory(items[0].key, items[0].value, items[0].layer, items[0].memoryType);
          return;
        }
        const maxRetries = 2;
        let lastErr: unknown;
        for (let attempt = 0; attempt <= maxRetries; attempt++) {
          try {
            await clawMemoryRequest("POST", "/memories/batch", {
              memories: items.map(item => ({
                key: item.key,
                value: item.value,
                layer: item.layer || "episodic",
                source: "openclaw",
                memory_type: item.memoryType || "conversation",
              })),
            });
            return;
          } catch (err) {
            lastErr = err;
            console.error(`[ClawMemory] Batch write failed (attempt ${attempt + 1}/${maxRetries + 1}):`, err);
            if (attempt < maxRetries) {
              await new Promise(r => setTimeout(r, (attempt + 1) * 1000));
            }
          }
        }
        console.error("[ClawMemory] Batch write failed after all retries, falling back to individual writes:", lastErr);
        for (const item of items) {
          await writeMemory(item.key, item.value, item.layer, item.memoryType);
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
      let lastAssembleSessionId = "";
      let lastAssembleMessages: Array<{ role: string; content: unknown }> = [];

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
          console.log(`[ClawMemory] ingest called (role=${message.role}) — but relying on assemble→afterTurn bridge for actual write`);
          return { ingested: true };
        },

        async ingestBatch({ sessionId, messages }: { sessionId: string; messages: Array<{ role: string; content: unknown }> }) {
          if (!enableAutoIngest) {
            return { ingested: false };
          }
          console.log(`[ClawMemory] ingestBatch called (${messages.length} messages) — but relying on assemble→afterTurn bridge for actual write`);
          return { ingested: true };
        },

        async afterTurn({ sessionId, messages }: { sessionId: string; messages?: Array<{ role: string; content: unknown }> }) {
          turnCount++;
          try {
            await clawMemoryRequest("POST", "/sessions/track", {
              session_id: sessionId,
              metadata: `last_active:${new Date().toISOString()}`,
            });
          } catch {}

          if (enableAutoIngest) {
            const sourceMessages = (messages && messages.length > 0)
              ? messages
              : (lastAssembleSessionId === sessionId ? lastAssembleMessages : []);

            const filtered = sourceMessages.filter(m => !isNoisyContent(extractContent(m.content)));

            if (filtered.length > 0) {
              const items = filtered.map(m => ({
                key: `${sessionId}_${m.role}_${Date.now()}_${Math.random().toString(36).slice(2, 6)}`,
                value: `[${m.role}] ${extractContent(m.content)}`,
              }));

              try {
                await writeMemoryBatch(items);
                console.log(`[ClawMemory] afterTurn: wrote ${items.length} messages for session ${sessionId}`);
              } catch (err) {
                console.error("[ClawMemory] afterTurn write failed:", err);
              }
            } else {
              console.log(`[ClawMemory] afterTurn: no non-noisy messages to write for session ${sessionId}`);
            }

            lastAssembleMessages = [];
          }

          return { ok: true };
        },

        async assemble({ sessionId, messages, tokenBudget }: { sessionId: string; messages: Array<{ role: string; content: unknown }>; tokenBudget: number }) {
          lastAssembleSessionId = sessionId;
          lastAssembleMessages = messages.slice();

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

        async compact({ sessionId }: { sessionId: string; force: boolean }) {
          if (lastAssembleMessages.length > 0 && lastAssembleSessionId === sessionId && enableAutoIngest) {
            const filtered = lastAssembleMessages.filter(m => !isNoisyContent(extractContent(m.content)));
            if (filtered.length > 0) {
              const items = filtered.map(m => ({
                key: `${sessionId}_${m.role}_${Date.now()}_${Math.random().toString(36).slice(2, 6)}`,
                value: `[${m.role}] ${extractContent(m.content)}`,
              }));
              try {
                await writeMemoryBatch(items);
                console.log(`[ClawMemory] compact: flushed ${items.length} messages for session ${sessionId}`);
              } catch (err) {
                console.error("[ClawMemory] compact flush failed:", err);
              }
            }
            lastAssembleMessages = [];
          }
          return { ok: true, compacted: true };
        },

        async dispose() {
          if (lastAssembleMessages.length > 0) {
            console.warn(`[ClawMemory] dispose: ${lastAssembleMessages.length} cached messages from last assemble will be lost`);
          }
        },
      };
    });
  },
};
