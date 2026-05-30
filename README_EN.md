<div align="center">

# 🧠 ClawMemory

### Give Your AI a Brain That Never Forgets

**Your AI assistant forgets everything after restart? ClawMemory makes it remember forever.**

[![Version](https://img.shields.io/badge/v2.30.0-blue.svg)](https://github.com/860016/Clawmemory)
[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8.svg)](https://go.dev)
[![Vue3](https://img.shields.io/badge/Vue-3.x-4FC08D.svg)](https://vuejs.org)
[![MCP](https://img.shields.io/badge/MCP-Protocol-6366F1.svg)](https://modelcontextprotocol.io)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](https://opensource.org/licenses/MIT)

[中文](./README.md) · [Quick Start](#-quick-start) · [Features](#-core-features) · [MCP Integration](#-mcp-one-click-integration) · [API Docs](#-api-documentation)

</div>

---

## 😤 Sound Familiar?

> 💬 *"I told you last week I use pnpm — why are you suggesting npm again?"*
>
> 💬 *"We discussed the auth architecture already. You forgot?"*
>
> 💬 *"Every new chat I have to re-explain the entire project..."*

**AI assistants have no long-term memory. That's the #1 pain point.** Every restart, every new conversation — back to square one.

**ClawMemory fixes this.**

---

## ⚡ Understand ClawMemory in 30 Seconds

```
  You → AI Assistant → ClawMemory → Permanent Memory
                         ↑                     ↓
                   Next session auto-loads ← Smart retrieval of relevant memories
```

| What You Do | What ClawMemory Does |
|-------------|---------------------|
| Tell AI "I prefer TypeScript" | Auto-saves as semantic memory, tagged as preference |
| Open Cursor next time | AI auto-loads your preference, uses TypeScript directly |
| Ask "What did we decide last time?" | Searches historical memories, recalls precisely |
| Don't mention a project for a week | Auto-decays inactive memories, keeps memory store clean |

---

## 🏆 Why ClawMemory?

| Feature | ClawMemory | ChatGPT Memory | Claude Memory | Mem0 | Zep |
|---------|-----------|----------------|---------------|------|-----|
| **Open Source** | ✅ MIT | ❌ Closed | ❌ Closed | ✅ | ✅ |
| **Self-hosted, local data** | ✅ Fully local | ❌ Cloud | ❌ Cloud | ⚠️ Self-host needed | ⚠️ Self-host needed |
| **Cross-tool sharing** | ✅ Cursor ↔ Claude ↔ Trae ↔ Windsurf | ❌ ChatGPT only | ❌ Claude only | ❌ | ❌ |
| **MCP Protocol** | ✅ Native support | ❌ | ❌ | ❌ | ❌ |
| **3-layer memory model** | ✅ Episodic / Semantic / Procedural | ❌ Flat | ❌ Flat | ❌ Flat | ⚠️ Simple |
| **Auto governance** | ✅ 5-step pipeline | ❌ | ❌ | ⚠️ Basic | ⚠️ Basic |
| **Memory decay** | ✅ Forgetting curve | ❌ | ❌ | ❌ | ⚠️ TTL only |
| **Dialectic reasoning** | ✅ Your own AI model | ❌ | ❌ | ❌ | ❌ |
| **Knowledge graph** | ✅ Entity + relation visualization | ❌ | ❌ | ❌ | ❌ |
| **Quality assessment + fix** | ✅ One-click scan & repair | ❌ | ❌ | ❌ | ❌ |
| **Built-in free AI** | ✅ NVIDIA NIM | ❌ | ❌ | ❌ | ❌ |
| **One-line install** | ✅ `npx -y clawmemory-mcp` | N/A | N/A | ⚠️ Config needed | ⚠️ Config needed |

> 💡 **TL;DR**: Other solutions are either closed-source, locked to one vendor, or feature-limited. ClawMemory is the only **open-source + self-hosted + cross-tool + MCP-native + full governance** all-in-one solution.

---

## 🎯 Core Features

### 🧠 Three-Layer Memory System

Inspired by cognitive science, ClawMemory organizes memories in three layers:

| Layer | What It Stores | Example |
|-------|---------------|---------|
| **Episodic** | Events and experiences | "Fixed the JWT bug in auth module today" |
| **Semantic** | Facts and knowledge | "The project uses PostgreSQL 15" |
| **Procedural** | Methods and processes | "Deploy flow: push main → CI build → auto-deploy" |

### 🔍 Three Retrieval Methods

| Method | Description | When to Use |
|--------|-------------|-------------|
| **Keyword Search** | SQLite FTS5 full-text search | Exact lookup |
| **Semantic Search** | AI understands meaning, matches related content | Fuzzy lookup |
| **Vector Search** | ChromaDB vector engine, semantic similarity | Deep association |

### 🔌 MCP One-Click Integration

**One command to give any MCP-compatible AI tool a memory:**

```bash
npx -y clawmemory-mcp
```

Supports Cursor / Claude Desktop / Windsurf / Trae. Generate config JSON from the settings page — just paste and go.

### 🏥 Auto Memory Governance

Your memory store needs "health management" too — auto-executes a 5-step governance pipeline:

```
Summary Generation → Quality Fix → Dedup & Merge → Decay Apply → Garbage Cleanup
```

Each step can be toggled independently. Supports daily/weekly auto-execution or one-click manual trigger.

### 🩺 Quality Assessment

One-click memory store health scan: empty values, too-short content, missing tags, duplicate keys... Auto-graded severity levels, fixable issues resolved with one click.

### 🕸️ Knowledge Graph

Auto-extract entities and relationships from memories. Three views (grid/graph/list) to visualize your knowledge network.

### 📖 Wiki Knowledge Base

Markdown knowledge pages with bidirectional entity linking and version history.

### 📊 Smart Daily Report

Auto-summarize daily activities, track project progress, trend analysis charts — your AI work log.

### 🧠 AI Enhancement

- **Built-in NVIDIA NIM free models** — no API key needed, works out of the box
- Smart entity extraction, AI summary generation, memory conflict detection
- **Dialectic Reasoning** — multi-round reasoning with your own AI model, zero extra cost

### 🔐 Security

- JWT + API Key dual authentication
- CORS whitelist, rate limiting, audit logging
- Sensitive content encrypted with AES-GCM
- All data stored locally — you have full control

---

## 🚀 Quick Start

### One-Click Install

```bash
# Linux / macOS
cd clawmemory && bash install.sh

# Windows
cd clawmemory && powershell -ExecutionPolicy Bypass -File install.ps1
```

Start: `bash start.sh` or `start.bat`

Visit: **http://localhost:8765**

### Docker Deploy

```bash
docker compose up -d
```

### Requirements

| Component | Version | Notes |
|-----------|---------|-------|
| Go | 1.21+ | Required for backend |
| Platform | Windows / macOS / Linux | x86_64 + ARM64 |

> 💡 Frontend is pre-built — no Node.js needed

---

## 🔌 MCP One-Click Integration

### 1. Start ClawMemory

```bash
./clawmemory
```

### 2. Get Your API Key

Open Web UI → Settings → API Keys → Create

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
        "CLAWMEMORY_API_KEY": "cm-your-key"
      }
    }
  }
}
```

**Claude Desktop** — edit `~/AppData/Roaming/Claude/claude_desktop_config.json`

**Windsurf** — edit `~/.windsurf/mcp.json`

**Trae** — edit `~/.trae/mcp.json`

> Same config format — just change the file path.

### 4. Restart Your AI Tool and Start Using

> 💬 *"I always use pnpm"* → AI auto-saves
> 💬 *"What did we decide last time?"* → AI searches memories and answers

### 6 MCP Tools

| Tool | Purpose |
|------|---------|
| `memory_save` | Save a memory |
| `memory_search` | Search memories |
| `memory_context` | Get context (inject into AI prompt) |
| `memory_reason` | Dialectic reasoning (uses your own AI model) |
| `memory_conclude` | Save a persistent conclusion |
| `memory_push_conversation` | Save a full conversation |

---

## 🏗️ Architecture

```
clawmemory/
├── go-backend/                        # Go Backend
│   ├── cmd/server/                    # Main entry point
│   ├── internal/
│   │   ├── api/                       # HTTP API
│   │   ├── services/
│   │   │   ├── governance_service.go  # Memory governance orchestrator
│   │   │   ├── decay_service.go       # Memory decay
│   │   │   ├── health_service.go      # Quality assessment & fix
│   │   │   └── smart_load_service.go  # Smart loading & summarization
│   │   ├── ai/                        # AI enhancement
│   │   └── models/                    # Data models
│   └── frontend_dist/                 # Frontend build output
├── frontend/                          # Vue3 frontend source
├── mcp-server/                        # MCP Server (TypeScript, npm)
├── openclaw-plugin/                   # OpenClaw plugin
├── hermes-plugin/                     # Hermes Agent Memory Provider
└── install.sh / install.ps1           # One-click install scripts
```

---

## ⚙️ Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `SECRET_KEY` | Auto-generated | JWT secret + sensitive content encryption key |
| `PORT` | `8765` | Listen port |
| `DATA_DIR` | `./data` | Data directory |

---

## 🔨 Build

```bash
# Frontend
cd frontend && npm install && npm run build

# Backend
cd go-backend && go build -o clawmemory ./cmd/server

# Cross-platform
GOOS=linux GOARCH=amd64 go build -o clawmemory-linux ./cmd/server
GOOS=darwin GOARCH=arm64 go build -o clawmemory-macos ./cmd/server
```

---

## 🌐 API Documentation

### Auth
- `POST /api/v1/auth/register` - Register
- `POST /api/v1/auth/login` - Login
- `POST /api/v1/auth/reset-password` - Reset password

### Memories
- `GET /api/v1/memories` - List
- `POST /api/v1/memories` - Create
- `GET /api/v1/memories/:id` - Detail
- `PUT /api/v1/memories/:id` - Update
- `DELETE /api/v1/memories/:id` - Delete
- `POST /api/v1/memories/:id/decrypt` - Decrypt encrypted memory
- `GET /api/v1/memories/smart-load` - Smart load
- `POST /api/v1/memories/:id/reinforce` - Reinforce memory
- `POST /api/v1/memories/generate-summaries` - Generate summaries

### Knowledge Graph
- `GET /api/v1/knowledge/entities` - Entity list
- `POST /api/v1/knowledge/entities` - Create entity
- `GET /api/v1/knowledge/relations` - Relation list
- `GET /api/v1/knowledge/graph` - Graph data

### AI Enhancement
- `GET /api/v1/ai/config` - Get AI config
- `GET /api/v1/ai/providers` - List available AI models
- `POST /api/v1/ai/test` - Test AI connection
- `POST /api/v1/ai/extract` - AI entity extraction
- `GET /api/v1/ai/daily-report` - AI daily report

### API Key Management
- `GET /api/v1/api-keys` - List
- `POST /api/v1/api-keys` - Create (full key returned once only)
- `DELETE /api/v1/api-keys/:id` - Delete

### MCP Config
- `GET /api/v1/mcp/config` - Get MCP Server config (auto-detect baseURL, auto-create API Key)

### Memory Governance
- `GET /api/v1/memories/governance/status` - Governance status
- `POST /api/v1/memories/governance/run` - Run governance now
- `PUT /api/v1/memories/governance/config` - Update governance config

### Memory Quality
- `GET /api/v1/memories/quality` - Quality assessment
- `POST /api/v1/memories/auto-fix` - Auto-fix quality issues

### External API (requires X-API-Key header)
- `POST /api/v1/external/memories` - Write single memory
- `POST /api/v1/external/memories/batch` - Batch write
- `GET /api/v1/external/memories/search?q=keyword` - Search
- `GET /api/v1/external/memories/context?q=keyword` - Get context
- `POST /api/v1/external/conversations` - Push conversation
- `POST /api/v1/external/conversations/batch` - Batch push
- `POST /api/v1/external/sessions/track` - Track session
- `POST /api/v1/external/reason` - Dialectic Reasoning

### Reasoning Config
- `GET /api/v1/reasoning/config` - Get reasoning config
- `PUT /api/v1/reasoning/config` - Update reasoning config
- `POST /api/v1/reasoning/test` - Test reasoning model
- `POST /api/v1/reasoning/execute` - Execute reasoning

### OpenClaw Sync
- `GET /api/v1/openclaw-sync/status` - Sync status
- `POST /api/v1/openclaw-sync/force` - Force sync
- `POST /api/v1/openclaw-sync/toggle` - Toggle auto-sync
- `GET /api/v1/openclaw/agents-md` - Get AGENTS.md

---

## 📝 Changelog

### v2.30.0 (2026-05-30)
- 🔌 New: MCP one-click config — generate Cursor/Claude/Windsurf/Trae config JSON from settings
- 🔌 New: MCP config API — auto-detect baseURL, auto-create API Key
- 🏥 New: Auto memory governance system — 5-step pipeline, each step independently toggleable
- 🏥 New: Governance API (status/run/config) — supports daily/weekly auto-execution
- 🩺 New: Quality assessment + one-click auto-fix
- 📦 MCP Server published to npm (`clawmemory-mcp@2.24.0`)
- 🌍 Added 33 i18n translations (zh/en)

<details>
<summary>📜 Previous Versions</summary>

### v2.29.0 (2026-05-19)
- 🔑 Login lockout threshold 3→5 attempts
- 🏗️ Unified dual storage system
- 🐛 Fixed critical bug: Validator pointer caused all memory writes to fail
- 📊 Added logging to 44 silently-ignored DB errors

### v2.28.0 (2026-05-19)
- 🧠 Memory scan overhaul: incremental indexing, path-aware classification, multi-level heading chunking
- 🧠 Self-evolution: NudgeReflect skill review loop
- 🔓 All premium features fully open-sourced

### v2.27.0 (2026-05-17)
- 🔓 Pro features made free & open-source, ~1200 lines of auth code removed
- 🏗️ ToolboxService consolidates 21 core methods
- 🐳 Docker one-click deploy

### v2.21.0 (2026-05-14)
- 🔒 Security hardening: JWT auto-key, SQL injection protection, CORS whitelist
- ⚡ Performance: FTS5 pre-filtering, GraphRAG N² optimization, LRU cache
- 🏗️ Global singletons replaced with dependency injection

### v2.20.0 (2026-05-11)
- ⚡ OpenClaw sync switched from polling to fsnotify real-time file watching
- ⚡ Multi-IDE directory real-time monitoring

### v2.19.0 (2026-05-08)
- 🔮 Dialectic Reasoning multi-round inference engine
- 🔌 MCP Server + 6 MCP Tools
- 🐍 Hermes Agent Memory Provider Plugin

</details>

---

## 🤝 Contributing

1. Fork this repo
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

MIT License

---

<div align="center">

**Make AI Never Forget Again 🧠**

[GitHub](https://github.com/860016/Clawmemory) · [npm](https://www.npmjs.com/package/clawmemory-mcp) · [Issues](https://github.com/860016/Clawmemory/issues)

</div>
