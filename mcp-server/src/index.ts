#!/usr/bin/env node

import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import { z } from "zod";
import { ClawMemoryClient } from "./client.js";
import { createRequire } from "module";

const require = createRequire(import.meta.url);
const packageJson = require("../package.json");

const CLAWMEMORY_BASE_URL = process.env.CLAWMEMORY_BASE_URL || "http://localhost:8765";
const CLAWMEMORY_API_KEY = process.env.CLAWMEMORY_API_KEY || "";
const CLAWMEMORY_PLATFORM = process.env.CLAWMEMORY_PLATFORM || "mcp";

if (!CLAWMEMORY_API_KEY) {
  console.error("CLAWMEMORY_API_KEY environment variable is required");
  process.exit(1);
}

const client = new ClawMemoryClient({
  baseUrl: CLAWMEMORY_BASE_URL,
  apiKey: CLAWMEMORY_API_KEY,
  platform: CLAWMEMORY_PLATFORM,
});

const server = new McpServer({
  name: "clawmemory",
  version: packageJson.version,
});

server.tool(
  "memory_save",
  "Save a memory to ClawMemory. Use this to persist important information, conversation highlights, user preferences, or key decisions.",
  {
    key: z.string().describe("A short identifier for this memory (e.g., 'user-pref-dark-mode')"),
    value: z.string().describe("The memory content to save"),
    layer: z.enum(["episodic", "semantic", "procedural"]).optional().describe("Memory layer. episodic=events, semantic=facts, procedural=how-to"),
    memory_type: z.enum(["conversation", "knowledge", "preference", "decision"]).optional().describe("Type of memory"),
    visibility: z.enum(["private", "shared", "public"]).optional().describe("Memory visibility. private=only owner, shared=authorized agents, public=all agents"),
    source_agent: z.string().optional().describe("Name of the agent saving this memory (e.g., 'trae', 'cursor', 'claude')"),
  },
  async (params) => {
    try {
      await client.saveMemory({
        key: params.key,
        value: params.value,
        layer: params.layer,
        memory_type: params.memory_type,
        visibility: params.visibility,
        source_agent: params.source_agent,
      });
      return { content: [{ type: "text", text: `Memory saved: ${params.key}` }] };
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err);
      return { content: [{ type: "text", text: `Error: ${msg}` }], isError: true };
    }
  }
);

server.tool(
  "memory_search",
  "Search ClawMemory for relevant memories. Use this when you need to recall past information, user preferences, or context from previous conversations.",
  {
    query: z.string().describe("Search query - what you're looking for"),
    limit: z.number().min(1).max(20).optional().describe("Maximum number of results (default: 5)"),
  },
  async (params) => {
    try {
      const items = await client.searchMemories(params.query, params.limit || 5);
      if (items.length === 0) {
        return { content: [{ type: "text", text: "No matching memories found." }] };
      }
      const lines = items.map(
        (m) => `- [${m.source || "?"}] ${m.key}: ${m.value}`
      );
      return {
        content: [{ type: "text", text: `Found ${items.length} memories:\n${lines.join("\n")}` }],
      };
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err);
      return { content: [{ type: "text", text: `Error: ${msg}` }], isError: true };
    }
  }
);

server.tool(
  "memory_context",
  "Get contextual memories with a pre-formatted system prompt addition. Use this to inject relevant memories into your context before responding.",
  {
    query: z.string().describe("Context query - what the current conversation is about"),
    limit: z.number().min(1).max(20).optional().describe("Maximum number of memories to retrieve (default: 5)"),
  },
  async (params) => {
    try {
      const result = await client.getContext(params.query, params.limit || 5);
      if (result.count === 0) {
        return { content: [{ type: "text", text: "No relevant context found." }] };
      }
      return {
        content: [{ type: "text", text: result.system_prompt_addition || `Found ${result.count} memories but no formatted context.` }],
      };
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err);
      return { content: [{ type: "text", text: `Error: ${msg}` }], isError: true };
    }
  }
);

server.tool(
  "memory_reason",
  "Perform dialectic reasoning about the user using their configured AI model. This analyzes conversation patterns and derives insights about user preferences, goals, and working style. Unlike Honcho's paid Dialectic Reasoning, this uses YOUR own AI model at no extra cost.",
  {
    query: z.string().describe("What to reason about (e.g., 'user preferences for code style', 'what this user cares about')"),
    depth: z.number().min(1).max(3).optional().describe("Reasoning depth: 1=single pass, 2=audit, 3=reconcile (default: 1)"),
    level: z.enum(["minimal", "low", "medium", "high", "max"]).optional().describe("Reasoning intensity - affects token usage (default: medium)"),
  },
  async (params) => {
    try {
      const result = await client.reason({
        query: params.query,
        depth: params.depth || 1,
        level: params.level || "medium",
      }) as Record<string, unknown>;
      return {
        content: [{ type: "text", text: (result.reasoning as string) || "No reasoning result" }],
      };
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err);
      return { content: [{ type: "text", text: `Reasoning error: ${msg}` }], isError: true };
    }
  }
);

server.tool(
  "memory_conclude",
  "Save a durable conclusion about the user. Use this when you've identified a stable preference, pattern, or fact that should persist across sessions.",
  {
    content: z.string().describe("The conclusion to save (e.g., 'User prefers TypeScript over JavaScript')"),
    category: z.enum(["preference", "style", "workflow", "skill", "fact"]).optional().describe("Category of the conclusion (default: fact)"),
    visibility: z.enum(["private", "shared", "public"]).optional().describe("Memory visibility (default: shared)"),
  },
  async (params) => {
    try {
      const category = params.category || "fact";
      await client.saveMemory({
        key: `conclusion-${category}-${Date.now()}`,
        value: params.content,
        layer: "semantic",
        source: "mcp-conclusion",
        memory_type: "knowledge",
        visibility: params.visibility || "shared",
      });
      return {
        content: [{ type: "text", text: `Conclusion saved (${category}): ${params.content.slice(0, 80)}...` }],
      };
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err);
      return { content: [{ type: "text", text: `Error: ${msg}` }], isError: true };
    }
  }
);

server.tool(
  "memory_push_conversation",
  "Push a conversation turn to ClawMemory for persistent storage. Use this to save important conversation exchanges.",
  {
    session_id: z.string().describe("Unique session identifier"),
    messages: z.array(z.object({
      role: z.enum(["user", "assistant", "system"]),
      content: z.string(),
    })).describe("Conversation messages to save"),
    summary: z.string().optional().describe("Optional summary of the conversation"),
    title: z.string().optional().describe("Optional title for the conversation"),
  },
  async (params) => {
    try {
      await client.pushConversation({
        session_id: params.session_id,
        messages: params.messages,
        summary: params.summary,
        title: params.title,
      });
      return {
        content: [{ type: "text", text: `Conversation saved: ${params.messages.length} messages in session ${params.session_id}` }],
      };
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err);
      return { content: [{ type: "text", text: `Error: ${msg}` }], isError: true };
    }
  }
);

async function main() {
  const transport = new StdioServerTransport();
  await server.connect(transport);
}

main().catch((err) => {
  console.error("Fatal error:", err);
  process.exit(1);
});
