<div align="center">

# 🧠 ClawMemory MCP Server

**给你的 AI 编程助手一个持久、可搜索的记忆——跨会话、跨工具。**

[![npm version](https://img.shields.io/npm/v/clawmemory-mcp.svg)](https://www.npmjs.com/package/clawmemory-mcp)
[![npm downloads](https://img.shields.io/npm/dm/clawmemory-mcp.svg)](https://www.npmjs.com/package/clawmemory-mcp)
[![Model Context Protocol](https://img.shields.io/badge/MCP-Protocol-blue)](https://modelcontextprotocol.io)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

[English](./README.md) · [安装](#-安装) · [快速开始](#-快速开始) · [工具](#-工具) · [工作原理](#-工作原理)

</div>

---

## 🤔 为什么需要 ClawMemory？

你的 AI 助手是不是每次重启就失忆？

- 你告诉 Cursor 你偏好 TypeScript——**下次会话，它又建议 JavaScript**
- Claude 帮你调试了一个棘手的架构决策——**3 天后，它完全忘了**
- 你花了好几个小时教 Trae 你的代码库规范——**重启后全没了**

**ClawMemory 就是来解决这个问题的。** 它给 AI 一个持久的记忆层，重启不丢失，跨工具共享，越用越聪明。

### ✨ 和其他方案的区别

| 特性 | ClawMemory | ChatGPT 记忆 | Claude 记忆 |
|------|-----------|-------------|------------|
| 支持任何 MCP 兼容 AI | ✅ | ❌ 仅 OpenAI | ❌ 仅 Anthropic |
| 跨工具记忆共享 | ✅ Cursor ↔ Claude ↔ Trae | ❌ | ❌ |
| 三层记忆模型 | ✅ 情景/语义/程序 | ❌ 扁平结构 | ❌ 扁平结构 |
| 辩证推理 | ✅ 用你自己的 AI 模型 | ❌ | ❌ |
| 自托管，完全控制 | ✅ | ❌ 仅云端 | ❌ 仅云端 |
| 自动治理与衰减 | ✅ 智能清理 | ❌ 手动 | ❌ 手动 |

---

## 📦 安装

### 前提条件

- [ClawMemory 服务端](https://github.com/860016/Clawmemory) 已运行（存储记忆的后端）
- 一个 ClawMemory 实例的 API Key

### 一行安装

```bash
npx -y clawmemory-mcp
```

就这么简单。无需全局安装——`npx` 自动处理一切。

---

## 🚀 快速开始

### 1. 启动 ClawMemory 服务端

```bash
git clone https://github.com/860016/Clawmemory.git
cd Clawmemory/go-backend
go run ./cmd/server
```

服务端默认运行在 `http://localhost:8765`。

### 2. 获取 API Key

打开 ClawMemory Web UI → 设置 → API 密钥 → 创建。

### 3. 配置你的 AI 工具

**Cursor** — 编辑 `~/.cursor/mcp.json`：

```json
{
  "mcpServers": {
    "clawmemory": {
      "command": "npx",
      "args": ["-y", "clawmemory-mcp"],
      "env": {
        "CLAWMEMORY_BASE_URL": "http://localhost:8765",
        "CLAWMEMORY_API_KEY": "cm-你的-api-key"
      }
    }
  }
}
```

**Claude Desktop** — 编辑 `~/AppData/Roaming/Claude/claude_desktop_config.json`（Windows）或 `~/.config/Claude/claude_desktop_config.json`（macOS/Linux）

**Windsurf** — 编辑 `~/.windsurf/mcp.json`

**Trae** — 编辑 `~/.trae/mcp.json`

> 配置格式完全相同，只需替换文件路径。

### 4. 重启 AI 工具

编辑配置后，重启 Cursor/Claude/Windsurf/Trae，你会在 MCP 服务器列表中看到 `clawmemory`。

### 5. 开始使用记忆

自然地和 AI 对话——它会自动使用 ClawMemory 工具：

> 💬 *"我总是用 pnpm，不用 npm"* → AI 自动保存为偏好
> 💬 *"记住我们的 API 用 snake_case"* → AI 自动保存为规范
> 💬 *"上周关于认证我们怎么决定的？"* → AI 搜索记忆回答

---

## 🛠 工具

ClawMemory 提供 6 个 MCP 工具供 AI 使用：

### `memory_save`
保存一条带结构化元数据的记忆。

```
Key: "user-pref-package-manager"
Value: "用户在所有 Node.js 项目中偏好 pnpm 而非 npm"
Layer: semantic
Type: preference
```

### `memory_search`
按关键词搜索记忆，返回排序结果。

```
Query: "包管理器"
→ 找到 3 条记忆：
  - [cursor] user-pref-package-manager: 用户偏好 pnpm 而非 npm...
  - [claude] project-setup-notes: 使用 pnpm workspaces 管理 monorepo...
```

### `memory_context`
获取预格式化的系统提示词，包含相关记忆。完美用于在 AI 回复前注入上下文。

```
Query: "编码风格"
→ [相关记忆自动注入 AI 上下文]
```

### `memory_reason`
**杀手级功能。** 用你自己的 AI 模型对用户进行辩证推理——零额外费用。

```
Query: "这个用户在代码审查中关注什么？"
Depth: 2 (审计轮次)
Level: medium
→ "基于 47 条记忆，该用户优先关注：类型安全、
   最小依赖、显式错误处理..."
```

### `memory_conclude`
保存一条持久结论——跨会话保持稳定的事实或偏好。

```
Content: "用户偏好 TypeScript strict 模式，不允许 any 类型"
Category: preference
```

### `memory_push_conversation`
推送完整对话轮次，用于持久存储和后续引用。

```
Session: "auth-refactor-2024"
Messages: [{role: "user", content: "..."}, {role: "assistant", content: "..."}]
Summary: "决定使用 JWT + refresh token 方案"
```

---

## ⚙️ 配置

### 环境变量

| 变量 | 必填 | 默认值 | 说明 |
|------|------|--------|------|
| `CLAWMEMORY_BASE_URL` | 否 | `http://localhost:8765` | ClawMemory 服务端地址 |
| `CLAWMEMORY_API_KEY` | **是** | — | 你的 API Key |
| `CLAWMEMORY_PLATFORM` | 否 | `mcp` | 平台标识，用于记忆来源追踪 |

### 记忆层级

| 层级 | 用途 | 示例 |
|------|------|------|
| **情景记忆** Episodic | 事件和经历 | "在会话 #42 中修复了认证 bug" |
| **语义记忆** Semantic | 事实和知识 | "项目使用 PostgreSQL 15" |
| **程序记忆** Procedural | 方法和流程 | "部署：push main → CI 构建 → 自动部署" |

### 记忆类型

| 类型 | 说明 |
|------|------|
| `conversation` | 对话摘录 |
| `knowledge` | 事实和信息 |
| `preference` | 用户偏好 |
| `decision` | 架构或设计决策 |

### 可见性级别

| 级别 | 范围 |
|------|------|
| `private` | 仅所有者 |
| `shared` | 跨工具的授权代理 |
| `public` | 所有代理和用户 |

---

## 🔧 工作原理

```
┌─────────────┐     MCP 协议        ┌──────────────────┐     HTTP API     ┌─────────────────┐
│   Cursor     │◄──────────────────►│  clawmemory-mcp  │◄───────────────►│  ClawMemory      │
│   Claude     │  (stdio 传输)      │  (本包)          │  (REST + 认证)  │  服务端 (Go)     │
│   Windsurf   │                    │                  │                 │  ┌─────────────┐ │
│   Trae       │                    │  • 工具路由      │                 │  │  SQLite DB   │ │
│   ...        │                    │  • Zod 校验      │                 │  │  智能加载    │ │
└─────────────┘                    │  • 错误处理      │                 │  │  衰减        │ │
                                   └──────────────────┘                 │  │  治理        │ │
                                                                        │  │  推理        │ │
                                                                        │  └─────────────┘ │
                                                                        └─────────────────┘
```

1. **你的 AI 工具**（Cursor/Claude 等）通过标准 MCP 协议调用工具
2. **clawmemory-mcp**（本包）将工具调用翻译为 ClawMemory API 请求
3. **ClawMemory 服务端**存储、索引、搜索和推理记忆
4. 结果返回给你的 AI，赋予它持久的上下文

---

## 🌟 高级功能

### 自动治理

ClawMemory 服务端自动保持记忆库健康：

- **摘要生成** —— 将冗长记忆压缩为精炼摘要
- **质量修复** —— 修复损坏条目（空值、缺失标签）
- **去重合并** —— 合并相似记忆，保留最佳版本
- **衰减** —— 逐渐降低不活跃记忆的重要性
- **垃圾清理** —— 永久删除已衰减的记忆

### 辩证推理

与简单关键词搜索不同，`memory_reason` 执行多轮分析：

1. **第 1 轮** —— 初步分析相关记忆
2. **第 2 轮（审计）** —— 交叉验证结论
3. **第 3 轮（调和）** —— 产出最终、细致的洞察

使用**你自己的 AI 模型**——零额外 API 费用，无厂商锁定。

### 跨工具记忆共享

在 Cursor 中保存偏好，在 Claude 中调用。ClawMemory 的 `shared` 和 `public` 可见性级别让记忆在工具间流转，同时尊重隐私边界。

---

## 📋 系统要求

- **Node.js** 18+（用于 `npx` 和 ES 模块）
- **ClawMemory 服务端** 2.0+ 已运行且可访问
- 一个 MCP 兼容的 AI 工具（Cursor、Claude Desktop、Windsurf、Trae 等）

---

## 🔄 版本历史

### v2.24.0
- 新增 `memory_conclude` 工具，保存持久结论
- 新增 `memory_push_conversation`，保存完整对话
- 增强 `memory_reason`，支持深度和级别控制
- 改进错误消息和校验

### v2.23.0
- 首次发布 npm
- 6 个 MCP 工具：save / search / context / reason / conclude / push_conversation
- 完整 Zod Schema 校验
- 跨平台支持（Windows、macOS、Linux）

---

## 🤝 参与贡献

欢迎贡献！请随时提交 Pull Request。

1. Fork 本仓库
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 提交 Pull Request

---

## 📄 许可证

MIT License — 详见 [LICENSE](LICENSE)。

---

<div align="center">

**为讨厌重复自己的 AI 开发者而造 ❤️**

[GitHub](https://github.com/860016/Clawmemory) · [npm](https://www.npmjs.com/package/clawmemory-mcp) · [问题反馈](https://github.com/860016/Clawmemory/issues)

</div>
