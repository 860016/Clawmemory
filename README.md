<div align="center">

# 🧠 ClawMemory

### 给 AI 一个不会遗忘的大脑

**你的 AI 助手每次重启都失忆？ClawMemory 让它永远记住你。**

[![Version](https://img.shields.io/badge/v2.30.0-blue.svg)](https://github.com/860016/Clawmemory)
[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8.svg)](https://go.dev)
[![Vue3](https://img.shields.io/badge/Vue-3.x-4FC08D.svg)](https://vuejs.org)
[![MCP](https://img.shields.io/badge/MCP-Protocol-6366F1.svg)](https://modelcontextprotocol.io)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](https://opensource.org/licenses/MIT)

[English](./README_EN.md) · [快速开始](#-快速开始) · [功能展示](#-核心功能) · [MCP 接入](#-mcp-一键接入) · [API 文档](#-api-文档)

</div>

---

## 😤 你一定遇到过这些

> 💬 *"我上周跟你说过用 pnpm，你怎么又用 npm 了？"*
>
> 💬 *"上次讨论的认证方案，你完全忘了？"*
>
> 💬 *"每次新对话都要重新解释项目架构，烦不烦？"*

**AI 助手没有长期记忆，这是最大的痛点。** 每次重启、每次新对话，一切从零开始。

**ClawMemory 就是来解决这个问题的。**

> ### 🔒 你的记忆，只属于你
>
> 所有数据存储在**你自己的机器上**——SQLite 本地数据库，不传云端，不经第三方。
>
> - **100% 本地**：记忆、偏好、对话记录，全部存在你电脑上
> - **完全可控**：随时导出、删除、备份，你拥有绝对控制权
> - **隐私安全**：敏感内容 AES-GCM 加密，即使数据库被拷走也无法读取
> - **无需注册云账号**：没有厂商可以查看、分析、出售你的记忆
>
> ChatGPT Memory 和 Claude Memory 的数据在云端——**你只是租用，不是拥有。**
> ClawMemory 的数据在你硬盘上——**你拥有，你控制。**

---

## ⚡ 30 秒了解 ClawMemory

```
  你 → AI 助手 → ClawMemory → 永久记忆
                    ↑                    ↓
              下次对话自动加载 ← 智能检索相关记忆
```

| 你做的事 | ClawMemory 做的事 |
|----------|-------------------|
| 告诉 AI "我偏好 TypeScript" | 自动保存为语义记忆，标记为偏好 |
| 下次打开 Cursor | AI 自动加载你的偏好，直接用 TypeScript |
| 说 "上次怎么决定的？" | 搜索历史记忆，精确回忆 |
| 一周没提某个项目 | 自动衰减不活跃记忆，保持记忆库清爽 |

---

## 🏆 为什么选 ClawMemory？

| 特性 | ClawMemory | ChatGPT Memory | Claude Memory | Mem0 | Zep |
|------|-----------|---------------|--------------|------|-----|
| **开源** | ✅ MIT | ❌ 闭源 | ❌ 闭源 | ✅ | ✅ |
| **自托管，数据本地** | ✅ 完全本地 | ❌ 云端 | ❌ 云端 | ⚠️ 需自建 | ⚠️ 需自建 |
| **跨工具共享** | ✅ Cursor ↔ Claude ↔ Trae ↔ Windsurf | ❌ 仅 ChatGPT | ❌ 仅 Claude | ❌ | ❌ |
| **MCP 协议** | ✅ 原生支持 | ❌ | ❌ | ❌ | ❌ |
| **三层记忆模型** | ✅ 情景/语义/程序 | ❌ 扁平 | ❌ 扁平 | ❌ 扁平 | ⚠️ 简单分类 |
| **记忆自动治理** | ✅ 5 步流水线 | ❌ | ❌ | ⚠️ 基础 | ⚠️ 基础 |
| **记忆衰减** | ✅ 模拟遗忘曲线 | ❌ | ❌ | ❌ | ⚠️ TTL |
| **辩证推理** | ✅ 用自己的 AI 模型 | ❌ | ❌ | ❌ | ❌ |
| **知识图谱** | ✅ 实体+关系可视化 | ❌ | ❌ | ❌ | ❌ |
| **质量评估+修复** | ✅ 一键扫描修复 | ❌ | ❌ | ❌ | ❌ |
| **内置免费 AI** | ✅ NVIDIA NIM | ❌ | ❌ | ❌ | ❌ |
| **一行安装** | ✅ `npx -y clawmemory-mcp` | N/A | N/A | ⚠️ 需配置 | ⚠️ 需配置 |

> 💡 **简单说**：其他方案要么闭源、要么只绑定自家产品、要么功能单一。ClawMemory 是唯一一个**开源 + 自托管 + 跨工具 + MCP 原生 + 完整治理**的一站式方案。

---

## 🎯 核心功能

### 🧠 三层记忆系统

受认知科学启发，ClawMemory 用三层模型组织记忆：

| 层级 | 存什么 | 举例 |
|------|--------|------|
| **情景记忆** Episodic | 事件和经历 | "今天修复了 auth 模块的 JWT bug" |
| **语义记忆** Semantic | 事实和知识 | "项目使用 PostgreSQL 15" |
| **程序记忆** Procedural | 方法和流程 | "部署流程：push main → CI 构建 → 自动部署" |

### 🔍 三种检索方式

| 方式 | 说明 | 何时用 |
|------|------|--------|
| **关键词搜索** | SQLite FTS5 全文检索 | 精确查找 |
| **语义搜索** | AI 理解含义，匹配相关内容 | 模糊查找 |
| **向量搜索** | ChromaDB 向量引擎，语义相似度 | 深度关联 |

### 🔌 MCP 一键接入

**一行命令，让任何 MCP 兼容的 AI 工具拥有记忆：**

```bash
npx -y clawmemory-mcp
```

支持 Cursor / Claude Desktop / Windsurf / Trae，设置页一键生成配置 JSON，粘贴即用。

### 🏥 记忆自动治理

记忆库也需要"健康管理"——自动执行 5 步治理流水线：

```
摘要生成 → 质量修复 → 去重合并 → 衰减应用 → 垃圾清理
```

每步独立开关，支持每天/每周自动执行，也可以一键手动触发。

### 🩺 质量评估

一键扫描记忆库健康度：空值、过短、缺失标签、重复 Key……自动分级标记严重度，可修复的一键搞定。

### 🕸️ 知识图谱

自动从记忆中提取实体和关系，三种视图（网格/图谱/列表）可视化你的知识网络。

### 📖 Wiki 知识库

Markdown 格式知识页面，与记忆实体双向关联，版本历史可追溯。

### 📊 智能日报

自动汇总每日活动，项目进度追踪，趋势分析图表——你的 AI 工作日志。

### 🧠 AI 增强

- **内置 NVIDIA NIM 免费模型**——无需 API Key，开箱即用
- 智能实体提取、AI 摘要生成、记忆冲突检测
- **Dialectic Reasoning**——用你自己的 AI 模型做多轮推理，零额外费用

### 🔐 安全

- JWT + API Key 双重认证
- CORS 白名单、速率限制、审计日志
- 敏感内容 AES-GCM 加密存储
- 数据完全本地，你拥有绝对控制权

---

## 🚀 快速开始

### 一键安装

```bash
# Linux / macOS
cd clawmemory && bash install.sh

# Windows
cd clawmemory && powershell -ExecutionPolicy Bypass -File install.ps1
```

启动：`bash start.sh` 或 `start.bat`

访问：**http://localhost:8765**

### Docker 部署

```bash
docker compose up -d
```

### 环境要求

| 组件 | 版本 | 说明 |
|------|------|------|
| Go | 1.21+ | 后端必需 |
| 平台 | Windows / macOS / Linux | x86_64 + ARM64 |

> 💡 前端已预编译，无需安装 Node.js

---

## 🔌 MCP 一键接入

### 1. 启动 ClawMemory

```bash
./clawmemory
```

### 2. 获取 API Key

打开 Web UI → 设置 → API 密钥 → 创建

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
        "CLAWMEMORY_API_KEY": "cm-your-key"
      }
    }
  }
}
```

**Claude Desktop** — 编辑 `~/AppData/Roaming/Claude/claude_desktop_config.json`

**Windsurf** — 编辑 `~/.windsurf/mcp.json`

**Trae** — 编辑 `~/.trae/mcp.json`

> 配置格式完全相同，只需替换文件路径。

### 4. 重启 AI 工具，开始使用

> 💬 *"我总是用 pnpm"* → AI 自动保存
> 💬 *"上次怎么决定的？"* → AI 搜索记忆回答

### 6 个 MCP 工具

| 工具 | 用途 |
|------|------|
| `memory_save` | 保存记忆 |
| `memory_search` | 搜索记忆 |
| `memory_context` | 获取上下文（注入 AI 提示词） |
| `memory_reason` | 辩证推理（用你自己的 AI 模型） |
| `memory_conclude` | 保存持久结论 |
| `memory_push_conversation` | 保存完整对话 |

---

## 🏗️ 架构

```
clawmemory/
├── go-backend/                        # Go 后端
│   ├── cmd/server/                    # 主程序入口
│   ├── internal/
│   │   ├── api/                       # HTTP API
│   │   ├── services/
│   │   │   ├── governance_service.go  # 记忆治理编排层
│   │   │   ├── decay_service.go       # 记忆衰减
│   │   │   ├── health_service.go      # 质量评估与修复
│   │   │   └── smart_load_service.go  # 智能加载与摘要
│   │   ├── ai/                        # AI 增强功能
│   │   └── models/                    # 数据模型
│   └── frontend_dist/                 # 前端构建产物
├── frontend/                          # Vue3 前端源码
├── mcp-server/                        # MCP Server (TypeScript, npm)
├── openclaw-plugin/                   # OpenClaw 插件
├── hermes-plugin/                     # Hermes Agent Memory Provider
└── install.sh / install.ps1           # 一键安装脚本
```

---

## ⚙️ 配置

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `SECRET_KEY` | 自动生成 | JWT 密钥 + 敏感内容加密密钥 |
| `PORT` | `8765` | 监听端口 |
| `DATA_DIR` | `./data` | 数据目录 |

---

## 🔨 构建

```bash
# 前端
cd frontend && npm install && npm run build

# 后端
cd go-backend && go build -o clawmemory ./cmd/server

# 跨平台
GOOS=linux GOARCH=amd64 go build -o clawmemory-linux ./cmd/server
GOOS=darwin GOARCH=arm64 go build -o clawmemory-macos ./cmd/server
```

---

## 🌐 API 文档

### 认证
- `POST /api/v1/auth/register` - 注册
- `POST /api/v1/auth/login` - 登录
- `POST /api/v1/auth/reset-password` - 重置密码

### 记忆
- `GET /api/v1/memories` - 列表
- `POST /api/v1/memories` - 创建
- `GET /api/v1/memories/:id` - 详情
- `PUT /api/v1/memories/:id` - 更新
- `DELETE /api/v1/memories/:id` - 删除
- `POST /api/v1/memories/:id/decrypt` - 解密加密记忆
- `GET /api/v1/memories/smart-load` - 智能加载
- `POST /api/v1/memories/:id/reinforce` - 强化记忆
- `POST /api/v1/memories/generate-summaries` - 生成摘要

### 知识图谱
- `GET /api/v1/knowledge/entities` - 实体列表
- `POST /api/v1/knowledge/entities` - 创建实体
- `GET /api/v1/knowledge/relations` - 关系列表
- `GET /api/v1/knowledge/graph` - 图谱数据

### AI 增强
- `GET /api/v1/ai/config` - 获取 AI 配置
- `GET /api/v1/ai/providers` - 列出可用 AI 模型
- `POST /api/v1/ai/test` - 测试 AI 连接
- `POST /api/v1/ai/extract` - AI 实体提取
- `GET /api/v1/ai/daily-report` - AI 日报生成

### API 密钥管理
- `GET /api/v1/api-keys` - 列表
- `POST /api/v1/api-keys` - 创建（返回完整密钥，仅一次）
- `DELETE /api/v1/api-keys/:id` - 删除

### MCP 配置
- `GET /api/v1/mcp/config` - 获取 MCP Server 配置（自动检测 baseURL、自动创建 API Key）

### 记忆治理
- `GET /api/v1/memories/governance/status` - 治理状态
- `POST /api/v1/memories/governance/run` - 立即执行治理
- `PUT /api/v1/memories/governance/config` - 更新治理配置

### 记忆质量
- `GET /api/v1/memories/quality` - 质量评估
- `POST /api/v1/memories/auto-fix` - 自动修复质量问题

### 外部 API（需 X-API-Key 请求头）
- `POST /api/v1/external/memories` - 写入单条记忆
- `POST /api/v1/external/memories/batch` - 批量写入
- `GET /api/v1/external/memories/search?q=keyword` - 搜索
- `GET /api/v1/external/memories/context?q=keyword` - 获取上下文
- `POST /api/v1/external/conversations` - 推送对话
- `POST /api/v1/external/conversations/batch` - 批量推送
- `POST /api/v1/external/sessions/track` - 跟踪会话
- `POST /api/v1/external/reason` - Dialectic Reasoning 推理

### 推理配置
- `GET /api/v1/reasoning/config` - 获取推理配置
- `PUT /api/v1/reasoning/config` - 更新推理配置
- `POST /api/v1/reasoning/test` - 测试推理模型
- `POST /api/v1/reasoning/execute` - 执行推理

### OpenClaw 同步
- `GET /api/v1/openclaw-sync/status` - 同步状态
- `POST /api/v1/openclaw-sync/force` - 手动触发同步
- `POST /api/v1/openclaw-sync/toggle` - 开关自动同步
- `GET /api/v1/openclaw/agents-md` - 获取 AGENTS.md

---

## 📝 更新日志

### v2.30.0 (2026-05-30)
- 🔌 新增：MCP 一键配置，设置页生成 Cursor/Claude/Windsurf/Trae 配置 JSON
- 🔌 新增：MCP 配置 API，自动检测 baseURL、自动创建 API Key
- 🏥 新增：记忆自动治理系统，5 步治理流水线，每步独立开关
- 🏥 新增：治理 API（status/run/config），支持每天/每周自动执行
- 🩺 新增：质量评估 + 一键自动修复
- 📦 MCP Server 发布 npm（`clawmemory-mcp@2.24.0`）
- 🌍 新增中英文翻译 33 条

<details>
<summary>📜 历史版本</summary>

### v2.29.0 (2026-05-19)
- 🔑 登录锁定阈值 3→5 次
- 🏗️ 统一双存储体系
- 🐛 修复 Validator 指针导致所有记忆写入失败的致命 Bug
- 📊 44 处静默 DB 错误添加日志

### v2.28.0 (2026-05-19)
- 🧠 记忆扫描全面升级：增量索引、路径感知分类、多级标题分块
- 🧠 自进化：NudgeReflect 技能审查闭环
- 🔓 所有高级功能完全开源

### v2.27.0 (2026-05-17)
- 🔓 Pro 功能全部免费开源，删除约 1200 行授权代码
- 🏗️ ToolboxService 整合 21 个核心方法
- 🐳 Docker 一键部署

### v2.21.0 (2026-05-14)
- 🔒 安全加固：JWT 自动密钥、SQL 注入防护、CORS 白名单
- ⚡ 性能优化：FTS5 预过滤、GraphRAG N² 优化、LRU 缓存
- 🏗️ 全局单例改为依赖注入

### v2.20.0 (2026-05-11)
- ⚡ OpenClaw 同步从轮询改为 fsnotify 实时监听
- ⚡ 多 IDE 目录实时监控

### v2.19.0 (2026-05-08)
- 🔮 Dialectic Reasoning 多轮推理引擎
- 🔌 MCP Server + 6 个 MCP Tools
- 🐍 Hermes Agent Memory Provider Plugin

</details>

---

## 🤝 参与贡献

1. Fork 本仓库
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 提交 Pull Request

## 📄 许可证

MIT License

---

<div align="center">

**让 AI 不再失忆 🧠**

[GitHub](https://github.com/860016/Clawmemory) · [npm](https://www.npmjs.com/package/clawmemory-mcp) · [问题反馈](https://github.com/860016/Clawmemory/issues)

</div>
