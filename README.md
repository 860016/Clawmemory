# ClawMemory v2.0 — AI 记忆管理工具

ClawMemory 是一款现代化的 AI 记忆管理工具，支持知识图谱、智能日报、记忆衰减分析、冲突检测等高级功能。

**支持双后端**：Python (传统) / Go (高性能，推荐)

---

## ✨ 核心特性

| 特性 | 开源版 | Pro 版 |
|------|--------|--------|
| 记忆管理 | ✓ | ✓ |
| 知识图谱 (网格/图谱/列表三视图) | ✓ | ✓ |
| Wiki 知识库 | ✓ | ✓ |
| 智能日报 | ✓ | ✓ |
| 本地数据导出/导入 | ✓ | ✓ |
| 全文搜索 + 语义搜索 | ✓ | ✓ |
| AI 提取/摘要 (多模型) | - | ✓ |
| 记忆衰减算法 | 基础 | 高级 |
| 冲突检测与合并 | 基础 | 高级 |
| Token 智能路由 (单条+批量) | 基础 | 高级 |
| 趋势分析 | - | ✓ |
| 报告生成 | - | ✓ |
| 多设备同步 | - | ✓ |

### 🤖 支持的 AI 模型

- **OpenAI**: GPT-4o / GPT-4o-mini / GPT-4-turbo / GPT-3.5-turbo
- **Anthropic**: Claude 3.5 Sonnet / Claude 3 Opus / Claude 3 Haiku
- **DeepSeek**: DeepSeek-Chat / DeepSeek-Reasoner
- **月之暗面 (Moonshot)**: moonshot-v1-8k/32k/128k
- **智谱 (Zhipu)**: GLM-4 / GLM-4-Flash / GLM-4-Air
- **通义千问 (Qwen)**: Qwen-Max / Qwen-Plus / Qwen-Turbo
- **自定义**: 任意 OpenAI 兼容 API

---

## 🚀 快速开始

### 环境要求

| 组件 | 最低版本 | 说明 |
|------|----------|------|
| **Python** | 3.10+ | 传统后端必需 |
| **Go** | 1.21+ | Go 后端必需 (推荐) |
| Node.js | 18+ | 仅重新构建前端时需要 |
| Docker | 20.10+ | Docker 安装方式 (可选) |

**支持平台**: Windows / macOS / Linux (x86_64 + ARM64)

### 一键安装

#### Linux / macOS

```bash
cd clawmemory
bash install.sh --license-server=https://auth.bestu.top
```

#### Windows

```powershell
cd clawmemory
powershell -ExecutionPolicy Bypass -File install.ps1
```

#### Docker（任何平台）

```bash
cd clawmemory
bash install.sh --docker
```

安装完成后启动：`bash start.sh` 或 `start.bat`

访问：`http://localhost:8765`

---

## 🏗️ 架构

```
clawmemory/
├── backend/                  # Python 后端 (传统)
│   ├── app/                  # 应用代码
│   ├── frontend_dist/        # 前端预构建产物
│   └── requirements.txt
├── go-backend/               # Go 后端 (高性能，推荐)
│   ├── cmd/server/           # 主程序入口
│   ├── internal/             # 内部包
│   ├── pkg/                  # 可复用包
│   └── pro/                  # Pro 功能模块
├── frontend/                 # Vue3 前端源码
│   ├── src/views/            # 页面组件
│   │   ├── KnowledgeViewV2.vue   # 知识图谱 (新版)
│   │   ├── DailyReportViewV2.vue # 日报 (新版)
│   │   └── MainLayout.vue        # 主布局 (新版)
│   └── src/styles/           # 设计系统
├── license-platform-v2/      # PHP 授权平台
│   └── src/
│       ├── AiServiceController.php   # AI 服务
│       └── ProApiController.php      # Pro API
├── install.sh / install.ps1  # 一键安装脚本
└── README.md
```

---

## ⚙️ 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `CLAWMEMORY_SECRET_KEY` | `change-me-in-production` | JWT 密钥 (**务必修改**) |
| `CLAWMEMORY_DB_PATH` | `./data/clawmemory.db` | 数据库路径 |
| `CLAWMEMORY_PORT` | `8765` | 监听端口 |
| `CLAWMEMORY_DEBUG` | `false` | 调试模式 |
| `CLAWMEMORY_LICENSE_SERVER_URL` | `https://auth.bestu.top` | 授权服务器 |
| `CLAWMEMORY_CORS_ORIGINS` | `["*"]` | CORS 来源 |
| `DEVICE_FINGERPRINT` | (自动生成) | **Docker 必须设置**，固定设备指纹 |

---

## 🔨 构建

### 前端构建

```bash
cd frontend
npm install
npm run build
```

构建产物输出到 `backend/frontend_dist/` (Python) 或 `go-backend/frontend_dist/` (Go)。

### Go 后端编译

```bash
cd go-backend

# 当前平台
go build -o clawmemory.exe ./cmd/server

# 跨平台编译 (全平台)
bash scripts/build-all.sh 1.0.0
```

支持 12 个构建目标：
- Windows: amd64, arm64
- Linux: amd64, arm64, 386
- macOS: amd64, arm64

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
- `DELETE /api/v1/memories/:id` - 删除（软删除）

### Pro API (需授权)
- `POST /api/v1/pro/ai/extract` - AI 提取
- `POST /api/v1/pro/ai/summarize` - AI 摘要
- `POST /api/v1/pro/router/batch` - 批量路由
- `POST /api/v1/pro/analyze/trends` - 趋势分析
- `POST /api/v1/pro/analyze/report` - 报告生成
- `POST /api/v1/pro/conflicts/detect` - 冲突检测

---

## 📝 更新日志

### v2.0 (2026-04-24)
- 🎨 全新现代化 UI (知识图谱/日报/主布局)
- 🤖 AI 提取/摘要 (支持国内外 7+ 主流模型)
- 📊 批量路由、趋势分析、报告生成
- 🌍 完善国际化 (中文/英文)
- 🚀 Go 高性能后端 (全平台编译)
- ☁️ Pro 云端 API 方案
- 🔒 移除云端备份，改为本地数据导出/导入

---

## 📄 许可证

- 开源版: MIT License
- Pro 版: 商业授权，联系 [auth.bestu.top](https://auth.bestu.top)

---

## 🤝 贡献

欢迎提交 Issue 和 PR！

GitHub: [https://github.com/860016/Clawmemory](https://github.com/860016/Clawmemory)
