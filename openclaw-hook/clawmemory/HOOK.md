---
name: clawmemory
description: "Auto-capture conversations to ClawMemory - your AI memory hub"
homepage: https://github.com/860016/Clawmemory
metadata:
  openclaw:
    emoji: "🧠"
    events:
      - command:new
      - command:reset
      - session:compact:before
    requires:
      bins:
        - node
      env:
        - CLAWMEMORY_URL
---

# ClawMemory Hook

Automatically captures conversation context and pushes it to your ClawMemory instance.

## What It Does

- On `/new` or `/reset`: saves the session that just ended to ClawMemory
- Before compaction: flushes important context before it gets summarized away
- Zero configuration needed if ClawMemory is running locally on port 8765

## Setup

### Quick Start (Local)

If ClawMemory is running on the same machine (default: `http://localhost:8765`):

1. Create an API key in ClawMemory Settings page
2. Set environment variable:

```bash
export CLAWMEMORY_URL=http://localhost:8765
export CLAWMEMORY_API_KEY=cm_your_api_key_here
```

3. Install and enable the hook:

```bash
openclaw plugins install /path/to/clawmemory-openclaw-hook
openclaw hooks enable clawmemory
```

### Remote Setup

If ClawMemory is on a different machine:

```bash
export CLAWMEMORY_URL=http://your-server:8765
export CLAWMEMORY_API_KEY=cm_your_api_key_here
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `CLAWMEMORY_URL` | `http://localhost:8765` | ClawMemory server URL |
| `CLAWMEMORY_API_KEY` | (required) | API key from ClawMemory settings |

## What Gets Captured

- User messages and AI responses
- Session metadata (agent name, timestamps)
- Project/workspace path

Sensitive content (passwords, API keys, tokens) is automatically filtered out.
