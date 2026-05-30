<div align="center">

# 🧠 ClawMemory MCP Server

**Give your AI coding assistant a persistent, searchable memory — across sessions, across tools.**

[![npm version](https://img.shields.io/npm/v/clawmemory-mcp.svg)](https://www.npmjs.com/package/clawmemory-mcp)
[![npm downloads](https://img.shields.io/npm/dm/clawmemory-mcp.svg)](https://www.npmjs.com/package/clawmemory-mcp)
[![Model Context Protocol](https://img.shields.io/badge/MCP-Protocol-blue)](https://modelcontextprotocol.io)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

[Install](#-install) · [Quick Start](#-quick-start) · [Tools](#-tools) · [Configuration](#-configuration) · [How It Works](#-how-it-works)

</div>

---

## 🤔 Why ClawMemory?

Ever noticed your AI assistant forgets everything between sessions?

- You told Cursor you prefer TypeScript — **next session, it suggests JavaScript again**
- Claude helped you debug a tricky architecture decision — **3 days later, it has no idea**
- You spent hours teaching Trae your codebase conventions — **gone after restart**

**ClawMemory fixes this.** It gives your AI a persistent memory layer that survives restarts, works across tools, and gets smarter over time.

### ✨ What Makes It Different

| Feature | ClawMemory | ChatGPT Memory | Claude Memory |
|---------|-----------|----------------|---------------|
| Works with any MCP-compatible AI | ✅ | ❌ OpenAI only | ❌ Anthropic only |
| Cross-tool memory sharing | ✅ Cursor ↔ Claude ↔ Trae | ❌ | ❌ |
| 3-layer memory model | ✅ Episodic / Semantic / Procedural | ❌ Flat | ❌ Flat |
| Dialectic reasoning | ✅ Your own AI model | ❌ | ❌ |
| Self-hosted, full control | ✅ | ❌ Cloud-only | ❌ Cloud-only |
| Auto-governance & decay | ✅ Smart cleanup | ❌ Manual | ❌ Manual |

---

## 📦 Install

### Prerequisites

- [ClawMemory Server](https://github.com/860016/Clawmemory) running (the backend that stores memories)
- An API key from your ClawMemory instance

### One-Line Install

```bash
npx -y clawmemory-mcp
```

That's it. No global install needed — `npx` handles everything.

---

## 🚀 Quick Start

### 1. Start ClawMemory Server

```bash
# Clone and run the backend
git clone https://github.com/860016/Clawmemory.git
cd Clawmemory/go-backend
go run ./cmd/server
```

The server runs at `http://localhost:8765` by default.

### 2. Get Your API Key

Open ClawMemory Web UI → Settings → API Keys → Create one.

### 3. Configure Your AI Tool

**Cursor** — edit `~/.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "clawmemory": {
      "command": "npx",
      "args": ["-y", "clawmemory-mcp"],
      "env": {
        "CLAWMEMORY_BASE_URL": "http://localhost:8765",
        "CLAWMEMORY_API_KEY": "cm-your-api-key-here"
      }
    }
  }
}
```

**Claude Desktop** — edit `~/AppData/Roaming/Claude/claude_desktop_config.json` (Windows) or `~/.config/Claude/claude_desktop_config.json` (macOS/Linux):

```json
{
  "mcpServers": {
    "clawmemory": {
      "command": "npx",
      "args": ["-y", "clawmemory-mcp"],
      "env": {
        "CLAWMEMORY_BASE_URL": "http://localhost:8765",
        "CLAWMEMORY_API_KEY": "cm-your-api-key-here"
      }
    }
  }
}
```

**Windsurf** — edit `~/.windsurf/mcp.json`:

```json
{
  "mcpServers": {
    "clawmemory": {
      "command": "npx",
      "args": ["-y", "clawmemory-mcp"],
      "env": {
        "CLAWMEMORY_BASE_URL": "http://localhost:8765",
        "CLAWMEMORY_API_KEY": "cm-your-api-key-here"
      }
    }
  }
}
```

**Trae** — edit `~/.trae/mcp.json`:

```json
{
  "mcpServers": {
    "clawmemory": {
      "command": "npx",
      "args": ["-y", "clawmemory-mcp"],
      "env": {
        "CLAWMEMORY_BASE_URL": "http://localhost:8765",
        "CLAWMEMORY_API_KEY": "cm-your-api-key-here"
      }
    }
  }
}
```

### 4. Restart Your AI Tool

After editing the config, restart Cursor/Claude/Windsurf/Trae. You'll see `clawmemory` in the MCP servers list.

### 5. Start Using Memory

Just talk to your AI naturally — it will use ClawMemory tools automatically:

> 💬 *"I always use pnpm, not npm"* → AI saves this as a preference
> 💬 *"Remember that our API uses snake_case"* → AI saves this as a convention
> 💬 *"What did we decide about auth last week?"* → AI searches memories

---

## 🛠 Tools

ClawMemory exposes 6 MCP tools that your AI can use:

### `memory_save`
Save a memory with structured metadata.

```
Key: "user-pref-package-manager"
Value: "User prefers pnpm over npm for all Node.js projects"
Layer: semantic
Type: preference
```

### `memory_search`
Search memories by keyword. Returns ranked results.

```
Query: "package manager"
→ Found 3 memories:
  - [cursor] user-pref-package-manager: User prefers pnpm over npm...
  - [claude] project-setup-notes: Using pnpm workspaces for monorepo...
```

### `memory_context`
Get a pre-formatted system prompt with relevant memories. Perfect for injecting context before responding.

```
Query: "coding style"
→ [Relevant memories injected into AI context automatically]
```

### `memory_reason`
**The killer feature.** Perform dialectic reasoning about the user using your own AI model — no extra cost.

```
Query: "What does this user care about in code reviews?"
Depth: 2 (audit pass)
Level: medium
→ "Based on 47 memories, this user prioritizes: type safety, 
   minimal dependencies, explicit error handling..."
```

### `memory_conclude`
Save a durable conclusion — a stable fact or preference that persists across sessions.

```
Content: "User prefers TypeScript strict mode with no any types"
Category: preference
```

### `memory_push_conversation`
Push an entire conversation turn for persistent storage and future reference.

```
Session: "auth-refactor-2024"
Messages: [{role: "user", content: "..."}, {role: "assistant", content: "..."}]
Summary: "Decided to use JWT with refresh tokens"
```

---

## ⚙️ Configuration

### Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `CLAWMEMORY_BASE_URL` | No | `http://localhost:8765` | ClawMemory server URL |
| `CLAWMEMORY_API_KEY` | **Yes** | — | Your API key |
| `CLAWMEMORY_PLATFORM` | No | `mcp` | Platform identifier for memory source tracking |

### Memory Layers

ClawMemory uses a 3-layer memory model inspired by cognitive science:

| Layer | Purpose | Example |
|-------|---------|---------|
| **Episodic** | Events and experiences | "Fixed auth bug in session #42" |
| **Semantic** | Facts and knowledge | "Project uses PostgreSQL 15" |
| **Procedural** | How-to and processes | "Deploy: push to main → CI builds → auto-deploy" |

### Memory Types

| Type | Description |
|------|-------------|
| `conversation` | Dialogue excerpts |
| `knowledge` | Facts and information |
| `preference` | User preferences |
| `decision` | Architectural or design decisions |

### Visibility Levels

| Level | Scope |
|-------|-------|
| `private` | Only the owner |
| `shared` | Authorized agents across tools |
| `public` | All agents and users |

---

## 🔧 How It Works

```
┌─────────────┐     MCP Protocol      ┌──────────────────┐     HTTP API      ┌─────────────────┐
│   Cursor     │◄────────────────────►│  clawmemory-mcp  │◄────────────────►│  ClawMemory      │
│   Claude     │   (stdio transport)  │  (this package)  │   (REST + Auth)  │  Server (Go)     │
│   Windsurf   │                      │                  │                  │  ┌─────────────┐ │
│   Trae       │                      │  • Tool routing  │                  │  │  SQLite DB   │ │
│   ...        │                      │  • Zod validation│                  │  │  Smart Load  │ │
└─────────────┘                      │  • Error handling│                  │  │  Decay       │ │
                                     └──────────────────┘                  │  │  Governance  │ │
                                                                           │  │  Reasoning   │ │
                                                                           │  └─────────────┘ │
                                                                           └─────────────────┘
```

1. **Your AI tool** (Cursor/Claude/etc.) calls MCP tools via the standard protocol
2. **clawmemory-mcp** (this package) translates tool calls into ClawMemory API requests
3. **ClawMemory Server** stores, indexes, searches, and reasons over memories
4. Results flow back to your AI, giving it persistent context

---

## 🌟 Advanced Features

### Auto Governance

ClawMemory Server automatically keeps your memory base healthy:

- **Summary Generation** — condenses verbose memories into concise summaries
- **Quality Fix** — repairs broken entries (empty values, missing tags)
- **Dedup Merge** — merges similar memories, keeping the best version
- **Decay** — gradually reduces importance of stale memories
- **Trash Cleanup** — permanently removes decayed memories

### Dialectic Reasoning

Unlike simple keyword search, `memory_reason` performs multi-pass analysis:

1. **Pass 1** — Initial analysis of relevant memories
2. **Pass 2 (Audit)** — Cross-references and validates conclusions
3. **Pass 3 (Reconcile)** — Produces final, nuanced insight

Uses **your own AI model** — no extra API costs, no vendor lock-in.

### Cross-Tool Memory Sharing

Save a preference in Cursor, recall it in Claude. ClawMemory's `shared` and `public` visibility levels let memories flow between tools while respecting privacy boundaries.

---

## 📋 Requirements

- **Node.js** 18+ (for `npx` and ES modules)
- **ClawMemory Server** 2.0+ running and accessible
- An MCP-compatible AI tool (Cursor, Claude Desktop, Windsurf, Trae, etc.)

---

## 🔄 Version History

### v2.24.0
- Added `memory_conclude` tool for durable conclusions
- Added `memory_push_conversation` for conversation persistence
- Enhanced `memory_reason` with depth and level controls
- Improved error messages and validation

### v2.23.0
- Initial npm publication
- 6 MCP tools: save, search, context, reason, conclude, push_conversation
- Full Zod schema validation
- Cross-platform support (Windows, macOS, Linux)

---

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## 📄 License

MIT License — see [LICENSE](LICENSE) for details.

---

<div align="center">

**Made with ❤️ for AI developers who hate repeating themselves**

[GitHub](https://github.com/860016/Clawmemory) · [npm](https://www.npmjs.com/package/clawmemory-mcp) · [Report Bug](https://github.com/860016/Clawmemory/issues)

</div>
