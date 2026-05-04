# ClawMemory v2.18.0 — AI 记忆管理中枢

**ClawMemory** 是一款为 AI 助手设计的**长期记忆管理系统**。它让 AI 能够"记住"过去的对话、知识和上下文，实现跨会话的智能记忆检索与关联。

## 🎯 ClawMemory 能做什么？

| 场景 | 功能 |
|------|------|
| **跨会话记忆** | AI 记住你之前说过的偏好、项目细节、重要事件，无需重复说明 |
| **知识沉淀** | 将对话中的知识点自动提取、分类、建立关联，形成个人知识库 |
| **智能检索** | 支持关键词搜索、语义搜索、向量搜索，快速找到相关记忆 |
| **记忆衰减** | 模拟人类遗忘曲线，重要记忆自动强化，过时信息逐渐淡化 |
| **冲突检测** | 自动发现记忆矛盾，提示用户确认正确信息 |
| **日报生成** | 自动汇总每日活动，生成结构化工作报告 |
| **AI 增强** | 内置 NVIDIA NIM 免费模型，智能提取实体、生成摘要、分析记忆 |
| **外部集成** | 提供 API 接口，支持 OpenClaw 等外部应用自动写入记忆 |

**典型用例**：
- 🤖 AI 助手记住你的编程习惯、项目架构、技术偏好
- 📚 构建个人知识图谱，实体关系可视化
- 📊 自动生成项目进度报告、学习总结
- 🔗 与 OpenClaw 联动，对话自动转化为持久记忆

**技术栈**: Go 后端（高性能、跨平台） + Vue3 前端 + SQLite + ChromaDB 向量引擎

---

## ✨ 核心特性

### 📝 记忆管理
- 多层级记忆分类（情景记忆/语义记忆/程序记忆）
- 记忆重要性评分与自动衰减算法
- 记忆强化机制（手动强化 + 自动强化）
- 全文搜索 + 语义搜索 + ChromaDB 向量搜索

### 🕸️ 知识图谱
- 实体自动提取与关系建立
- 三种视图：网格视图 / 图谱视图 / 列表视图
- 实体类型分类（人物/地点/事件/概念/项目）
- 关系可视化与交互式编辑

### 📖 Wiki 知识库
- Markdown 格式知识页面
- 目录结构与标签分类
- 与记忆实体双向关联
- 版本历史与编辑记录

### 📊 智能日报
- 自动汇总每日记忆活动
- 项目进度追踪
- 关键事件提取
- 趋势分析与统计图表

### 🧠 AI 增强功能
- **内置 NVIDIA NIM 免费模型**：无需 API Key，开箱即用
- **智能实体提取**：自动从记忆中提取人物、地点、事件、概念
- **AI 摘要生成**：长文本自动生成精炼摘要
- **智能记忆加载**：Token 预算控制，按重要性动态加载相关记忆
- **记忆冲突检测**：发现矛盾记忆并提示确认

### 🔌 外部集成
- **API Key 认证**：安全的第三方应用接入
- **外部 API**：`/api/v1/external/memories` 供 OpenClaw 等写入记忆
- **批量导入**：支持批量写入记忆（最多 100 条/次）
- **OpenClaw 自动同步**：安装后自动监控 `~/.openclaw/` 目录，每 60 秒增量同步新对话

### 🔐 安全特性
- JWT 认证 + API Key 双重认证机制
- CORS 白名单限制（仅 localhost + 局域网）
- 速率限制（防暴力破解）
- 审计日志记录
- 输入验证与长度限制
- **敏感内容加密存储**：默认跳过含 API Key/密码/Token 的内容；开启后以 AES-GCM 加密存储，查看时解密

### 💾 数据管理
- 本地 SQLite 数据库
- 数据导出（JSON 格式）
- 数据导入与恢复
- 自动备份机制

---

## 🚀 快速开始

### 环境要求

| 组件 | 最低版本 | 说明 |
|------|----------|------|
| **Go** | 1.21+ | 后端必需 |

**支持平台**: Windows / macOS / Linux (x86_64 + ARM64)

> 💡 **提示**: 前端已预编译，无需安装 Node.js，直接运行安装脚本即可。

### 一键安装

#### Linux / macOS

```bash
cd clawmemory
bash install.sh
```

#### Windows

```powershell
cd clawmemory
powershell -ExecutionPolicy Bypass -File install.ps1
```

安装完成后启动：`bash start.sh` 或 `start.bat`

访问：`http://localhost:8765`

---

## 🏗️ 架构

```
clawmemory/
├── go-backend/               # Go 后端 (开源)
│   ├── cmd/server/           # 主程序入口
│   ├── internal/             # 内部包
│   │   ├── api/              # HTTP API
│   │   ├── services/         # 业务逻辑
│   │   ├── ai/               # AI 增强功能
│   │   └── models/           # 数据模型
│   ├── frontend_dist/        # 前端构建产物
│   └── go.mod
├── frontend/                 # Vue3 前端源码
│   ├── src/views/            # 页面组件
│   │   ├── KnowledgeViewV2.vue
│   │   ├── DailyReportViewV2.vue
│   │   └── MainLayout.vue
│   └── src/styles/           # 设计系统
├── install.sh / install.ps1  # 一键安装脚本
└── README.md
```

---

## ⚙️ 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `SECRET_KEY` | (自动生成) | JWT 密钥 + 敏感内容加密密钥（开启敏感内容记录时必须手动设置，否则敏感内容将被跳过） |
| `PORT` | `8765` | 监听端口 |
| `DATA_DIR` | `./data` | 数据目录 |

---

## 🔨 构建

### 前端构建

```bash
cd frontend
npm install
npm run build
# 构建产物自动复制到 go-backend/frontend_dist/
```

### Go 后端编译

```bash
cd go-backend

# 当前平台
go build -o clawmemory ./cmd/server

# 跨平台编译
GOOS=linux GOARCH=amd64 go build -o clawmemory-linux ./cmd/server
GOOS=darwin GOARCH=arm64 go build -o clawmemory-macos ./cmd/server
GOOS=windows GOARCH=amd64 go build -o clawmemory.exe ./cmd/server
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

### 授权
- `GET /api/v1/license/info` - 授权信息
- `POST /api/v1/license/activate` - 激活授权
- `POST /api/v1/license/deactivate` - 停用授权

### API 密钥管理
- `GET /api/v1/api-keys` - 列表
- `POST /api/v1/api-keys` - 创建（返回完整密钥，仅一次）
- `DELETE /api/v1/api-keys/:id` - 删除

### 外部 API（供 OpenClaw 等外部应用调用，需 X-API-Key 请求头）
- `POST /api/v1/external/memories` - 写入单条记忆
- `POST /api/v1/external/memories/batch` - 批量写入记忆
- `GET /api/v1/external/memories/search?q=keyword` - 搜索记忆

### OpenClaw 自动同步
- `GET /api/v1/openclaw-sync/status` - 同步状态
- `POST /api/v1/openclaw-sync/force` - 手动触发同步
- `POST /api/v1/openclaw-sync/toggle` - 开关自动同步
- `GET /api/v1/openclaw/agents-md` - 获取 AGENTS.md 指令内容

---

## 📝 更新日志

### v2.18.0 (2026-05-04)
- 🔗 新增：AGENTS.md 指令集成，安装时自动写入 OpenClaw AGENTS.md
- 🔗 新增：AGENTS.md 内容生成 API（`GET /api/v1/openclaw/agents-md`）
- 🔗 新增：前端设置页 AGENTS.md 复制/预览功能
- 🧹 移除：旧版 hook 安装方式，改用 AGENTS.md 指令方式
- 📦 新增：openclaw-hook 安装脚本（setup.sh / setup.ps1），自动处理 AGENTS.md 模板

### v2.17.0 (2026-05-04)
- 🧠 新增：AI 大模型集成系统，内置 NVIDIA NIM 免费模型
- 🧠 新增：AI 实体提取、摘要生成、日报生成等智能功能
- 🧠 新增：AI Provider 抽象层，支持 OpenAI/DeepSeek/自定义端点
- 🧠 新增：Prompt 模板系统，8 种智能分析模板
- 🔧 优化：AI 功能分级，免费用户可使用 NVIDIA NIM 免费模型

### v2.16.0 (2026-05-04)
- 🐛 修复：备份功能路径硬编码问题
- 🐛 修复：前端单元测试 mock 路径错误
- 🔧 优化：配置管理统一化

### v2.15.0 (2026-05-03)
- 🛡️ 新增：敏感内容记录开关（默认关闭），开启后敏感内容以 AES-GCM 加密存储
- 🔐 新增：加密记忆解密 API（`POST /api/v1/memories/:id/decrypt`）
- 🎛️ 新增：设置页面「记录敏感内容」开关 + 安全警告提示
- 📋 新增：同步状态显示跳过数量（skipped_count）

### v2.14.0 (2026-05-03)
- 🔒 安全：OpenClaw 同步增加敏感信息过滤（API Key、密码、token 等）
- 🔒 安全：跳过敏感文件（.env、credentials、secrets 等）
- 🔒 安全：内容长度限制（最大 50000 字符）
- 📊 新增：同步状态显示跳过数量（skipped_count）

### v2.13.0 (2026-05-03)
- 🔄 新增：OpenClaw 自动同步服务，安装后自动监控 `~/.openclaw/` 目录
- 🔄 新增：后台定时扫描（每 60 秒），自动导入新对话和记忆
- 🔄 新增：增量同步机制，避免重复导入
- 📋 新增：同步状态 API（`/api/v1/openclaw-sync/status`）
- 🎛️ 新增：手动触发同步和开关自动同步 API
- 🚀 改进：启动时自动检测 OpenClaw 目录并开始同步

### v2.12.0 (2026-05-03)
- 🔑 新增：API Key 认证机制，支持外部应用安全调用 ClawMemory API
- 🤖 新增：外部 API 端点（`/api/v1/external/memories`），供 OpenClaw 自动写入记忆
- 📦 新增：批量写入记忆 API（`/api/v1/external/memories/batch`）
- 🔍 新增：外部记忆搜索 API（`/api/v1/external/memories/search`）
- 🎨 新增：设置页面 API 密钥管理界面（生成/复制/删除）
- 🌍 新增：API Key 相关中英文翻译

### v2.11.0 (2026-04-30)
- 🐛 修复：SQLite 驱动注册冲突（modernc.org/sqlite 与 glebarez/go-sqlite 同名注册）
- 🔐 新增：终端命令重置密码 `./clawmemory --reset-password NEW_PASSWORD`
- 🔐 新增：`./clawmemory --version` 查看版本号
- 📋 新增：设置页面显示版本号、检查更新、更新说明
- 🔄 新增：GitHub Releases 版本更新检查 API
- 🔒 改进：忘记密码 API 增加用户名验证
- 📦 统一版本号管理（config.AppVersion）

### v2.10.0 (2026-04-30)
- 🔒 安全加固：Auth 中间件拆分，API Key 只能访问 /api/v1/external 端点
- 🔒 安全加固：CORS 从 `*` 改为 Origin 白名单（localhost + 局域网）
- 🔒 安全加固：添加速率限制（API Key 60次/分，JWT 120次/分，登录 10次/分）
- 🔒 安全加固：API Key 数量上限（每用户最多 5 个）
- 🔒 安全加固：批量写入上限（100条/次）+ 输入长度限制
- 📋 新增：审计日志（API Key 创建/删除、外部记忆写入自动记录）

### v2.0 (2026-04-29)
- 🎨 全新现代化 UI (知识图谱/日报/主布局)
- 🤖 AI 提取/摘要 (支持国内外 7+ 主流模型)
- 📊 批量路由、趋势分析、报告生成
- 🧠 智能记忆加载 (Token 预算控制)
- 📌 记忆强化机制 (防止重要记忆衰减)
- 🔍 ChromaDB 向量搜索支持
- 🌍 完善国际化 (中文/英文)
- 🚀 Go 高性能后端 (全平台编译)
- 🔒 移除 Python 后端和 Docker，简化架构

---

## 📄 许可证

MIT License

---

## 🤝 贡献

欢迎提交 Issue 和 PR！

GitHub: [https://github.com/860016/Clawmemory](https://github.com/860016/Clawmemory)

---

## 🤖 OpenClaw 自动安装

ClawMemory 支持 **OpenClaw 自动安装**，只需一条命令即可完成完整安装：

### 📦 安装命令

#### Windows (PowerShell)
```powershell
# 克隆项目
git clone https://github.com/860016/Clawmemory.git

# 进入目录
cd Clawmemory

# 一键安装（自动检测依赖、构建前端、编译后端）
powershell -ExecutionPolicy Bypass -File install.ps1

# 启动服务
start.bat
```

#### Linux / macOS (Bash)
```bash
# 克隆项目
git clone https://github.com/860016/Clawmemory.git

# 进入目录
cd Clawmemory

# 一键安装（自动检测依赖、构建前端、编译后端）
bash install.sh

# 启动服务
./start.sh
```

### 🎯 安装流程（自动完成）

安装脚本会自动执行以下 5 个步骤：

```
[1/5] 检查环境依赖...
      ✓ Go 环境 (必需)
      ✓ Git (可选，用于技能安装)

[2/5] 检查前端文件...
      ✓ 前端已预编译，跳过构建

[3/5] 编译 Go 后端...
      ✓ go build -o clawmemory.exe ./cmd/server

[4/5] 配置环境...
      ✓ 创建 .env 配置文件
      ✓ 创建统一目录结构 (data/skills, data/backups)

[5/5] 验证安装...
      ✓ 检查所有关键文件
      ✓ 安装完成!
```

### ⚡ 快速安装（单行命令）

#### Windows
```powershell
git clone https://github.com/860016/Clawmemory.git; cd Clawmemory; powershell -ExecutionPolicy Bypass -File install.ps1; start.bat
```

#### Linux / macOS
```bash
git clone https://github.com/860016/Clawmemory.git && cd Clawmemory && bash install.sh && ./start.sh
```

### 🔧 安装参数

| 参数 | 说明 | 示例 |
|------|------|------|
| `--port=PORT` | 自定义端口 | `--port=3000` |
| `--install-path=PATH` | 自定义安装路径 | `--install-path=/opt/clawmemory` |
| `--auto-start` | 安装后自动启动 | `--auto-start` |
| `--upgrade` | 升级模式（保留配置） | `--upgrade` |

#### Windows 示例
```powershell
# 自定义端口 + 自动启动
powershell -ExecutionPolicy Bypass -File install.ps1 -Port 3000 -AutoStart

# 自定义安装路径
powershell -ExecutionPolicy Bypass -File install.ps1 -InstallPath "D:\Apps\ClawMemory"

# 升级模式（保留配置和数据）
powershell -ExecutionPolicy Bypass -File install.ps1 -Upgrade
```

#### Linux / macOS 示例
```bash
# 自定义端口 + 自动启动
bash install.sh --port=3000 --auto-start

# 自定义安装路径
bash install.sh --install-path=/opt/clawmemory

# 升级模式（保留配置和数据）
bash install.sh --upgrade

# 查看帮助
bash install.sh --help
```

### 📁 安装后目录结构

```
ClawMemory/                      # 安装根目录
├── go-backend/                  # 后端程序
│   ├── clawmemory.exe          # 可执行文件
│   ├── .env                    # 配置文件
│   └── frontend_dist/          # 前端构建产物
├── data/                        # 📌 统一数据目录
│   ├── clawmemory.db           # 数据库
│   ├── skills/                 # 技能插件目录
│   ├── backups/                # 备份文件目录
│   ├── keys/                   # 密钥文件
│   └── uploads/                # 上传文件
├── start.bat / start.sh        # 启动脚本
├── stop.bat / stop.sh          # 停止脚本
├── install.ps1                 # Windows 安装脚本
└── install.sh                  # Linux/macOS 安装脚本
```

### 🚀 启动与访问

安装完成后：

1. **启动服务**: 双击 `start.bat` (Windows) 或运行 `./start.sh` (Linux/macOS)
2. **访问地址**: 打开浏览器访问 `http://localhost:8765`
3. **首次设置**: 首次访问需设置管理员密码

### 💡 提示

- ✅ **统一目录**: 所有数据、技能、备份统一存储在 `data/` 目录下
- ✅ **自动检测**: 自动检测 Go/Git 环境
- ✅ **完整验证**: 安装完成后自动验证所有关键文件
- ✅ **跨平台**: 支持 Windows / Linux / macOS (x86_64 + ARM64)
- ✅ **升级友好**: 使用 `--upgrade` 参数可保留配置和数据
- ✅ **前端预编译**: 前端已预编译，无需安装 Node.js

---

## 🤖 OpenClaw 集成说明

ClawMemory 可以作为 OpenClaw 的主要记忆管理工具，通过 API 联动实现对话自动记忆。

### ⚠️ 安装目录要求

**必须安装在持久化目录中，禁止安装在临时目录（如 `/tmp`）！**

| 平台 | 推荐安装目录 | ❌ 禁止目录 |
|------|-------------|------------|
| **Windows** | `C:\Users\<用户名>\ClawMemory\` | `%TEMP%%`、`C:\Windows\Temp\` |
| **Linux/macOS** | `~/clawmemory/` 或 `/opt/clawmemory/` | `/tmp/`、`/var/tmp/` |

> 💡 **原因**: 临时目录会在系统重启或清理时被删除，导致数据丢失。ClawMemory 的数据库、配置、备份都存储在安装目录下。

### 第一步：安装并启动 ClawMemory

```bash
# 1. 克隆到持久化目录
git clone https://github.com/860016/Clawmemory.git ~/clawmemory
cd ~/clawmemory

# 2. 一键安装
bash install.sh   # Linux/macOS
# 或
powershell -ExecutionPolicy Bypass -File install.ps1  # Windows

# 3. 启动服务
./start.sh   # Linux/macOS
# 或
start.bat    # Windows
```

### 第二步：注册管理员账号

1. 打开浏览器访问 `http://localhost:8765`
2. 首次访问需注册管理员账号
3. 登录后即可使用 ClawMemory

### 第三步：创建 API 密钥（用于 OpenClaw 自动记录）

1. 登录 ClawMemory 网页 → 设置 → API 密钥
2. 点击「创建密钥」，输入名称如 `OpenClaw 自动记录`
3. **立即复制并保存密钥**（关闭后无法再次查看！）

### 第四步：配置 OpenClaw 连接 ClawMemory

安装 ClawMemory 后，需要在 OpenClaw 的 `AGENTS.md` 中添加 ClawMemory 指令，让 OpenClaw 自动将对话记忆存入 ClawMemory。

**方式一：使用安装脚本（推荐）**

```bash
# 进入 openclaw-hook 目录
cd openclaw-hook

# 设置 API Key 环境变量
export CLAWMEMORY_API_KEY='cm你的密钥'   # Linux/macOS
$env:CLAWMEMORY_API_KEY = 'cm你的密钥'   # Windows

# 运行安装脚本
bash setup.sh       # Linux/macOS
pwsh -File setup.ps1  # Windows
```

脚本会自动将 ClawMemory 指令写入 OpenClaw 的 AGENTS.md 文件，包括：
- 自动保存对话到 ClawMemory
- 搜索记忆时优先查询 ClawMemory
- 心跳时从文件提炼精华写入 ClawMemory

**方式二：手动复制**

1. 登录 ClawMemory 网页 → 设置 → OpenClaw 连接
2. 点击「复制 AGENTS.md」按钮
3. 将内容粘贴到 OpenClaw 的 AGENTS.md 文件中

**OpenClaw AGENTS.md 位置**：
- Windows: `C:\Users\<用户名>\.openclaw\AGENTS.md`
- Linux/macOS: `~/.openclaw/AGENTS.md`

**AGENTS.md 指令内容（示例）**：

```markdown
## 🧠 ClawMemory Auto-Record

每次回复结束时，自动将当次对话写入 ClawMemory。步骤如下：

1. 读取当前 session 的 jsonl 文件：`~/.openclaw/agents/main/sessions/` 下的 `.jsonl` 文件（找最新修改的）
2. 提取用户消息和我的回复
3. 追加写到 `~/.openclaw/workspace/memory/auto-conversation.yml`（按时间追加，用 `---` 分隔）
4. ClawMemory 的 auto-sync 服务每60秒自动扫描并导入

## 🧠 直接读写 ClawMemory（记忆后端）

ClawMemory 是我的长期记忆后端，我可以直接通过 API 读写它。

### 连接信息
- **地址**: http://localhost:8765
- **API Key**: <你的 API Key>
- **管理端账号**: `admin` / `admin123`

### 何时读记忆
- 收到需要回忆的任务时，先搜索 ClawMemory
- 替代原有 memory_search 工具，因为 ClawMemory 支持语义搜索，更强

### 何时写记忆
- 每次回复末尾，自动将本轮对话写入 ClawMemory
- 使用 API：`POST /api/v1/external/conversations`
- 不需要等文件同步，直接写，即刻生效
```

配置完成后，OpenClaw 会按照 AGENTS.md 中的指令自动将对话记忆存入 ClawMemory。

**自动记录请求示例**：

```bash
# 写入单条记忆
curl -X POST http://localhost:8765/api/v1/external/memories \
  -H "X-API-Key: cm你的密钥" \
  -H "Content-Type: application/json" \
  -d '{"key":"用户偏好","value":"喜欢深色主题","source":"openclaw"}'

# 批量写入
curl -X POST http://localhost:8765/api/v1/external/memories/batch \
  -H "X-API-Key: cm你的密钥" \
  -H "Content-Type: application/json" \
  -d '{"memories":[{"key":"topic1","value":"content1"},{"key":"topic2","value":"content2"}]}'

# 搜索记忆
curl "http://localhost:8765/api/v1/external/memories/search?q=关键词" \
  -H "X-API-Key: cm你的密钥"
```

### 第五步：导入 OpenClaw 已有记忆

如果你之前使用 OpenClaw 的本地记忆，可以在 ClawMemory 中一键导入。

**扫描范围（仅此目录）**：

| 平台 | 扫描目录 |
|------|----------|
| **Windows** | `C:\Users\<用户名>\.openclaw\workspace\` |
| **Linux/macOS** | `~/.openclaw/workspace/` |

> ⚠️ **注意**: 扫描功能**只扫描** `~/.openclaw/workspace/` 目录，不会扫描其他位置。

该目录下的文件结构：

```
~/.openclaw/workspace/
├── MEMORY.md          # 长期记忆（会被导入）
└── memory/            # 日常记忆文件目录（会被导入）
```

**导入方式**：
1. 登录 ClawMemory 网页
2. 进入设置页面
3. 点击「扫描 / 导入 OpenClaw 记忆」
4. 系统会自动扫描 `~/.openclaw/workspace/` 目录下的 `MEMORY.md` 和 `memory/` 文件
5. 确认扫描结果中的路径指向 `~/.openclaw/workspace/` 后完成导入

> 💡 **提示**: 如果扫描结果为空或路径不对，请确认 `~/.openclaw/workspace/` 目录下存在 `MEMORY.md` 或 `memory/` 文件。

### 局域网访问

ClawMemory 默认监听 `0.0.0.0`，支持局域网访问：

- 本机访问: `http://localhost:8765`
- 局域网访问: `http://<你的IP>:8765`

> 💡 **提示**: 如需修改端口，可在 `.env` 文件中设置 `PORT=端口号`

### 数据目录说明

| 目录 | 说明 |
|------|------|
| `~/.openclaw/` | OpenClaw 默认数据目录 |
| `~/.openclaw/workspace/` | OpenClaw 记忆文件（MEMORY.md + memory/） |
| `./data/` | ClawMemory 安装目录下的数据目录 |

ClawMemory 会自动扫描以下目录中的技能：
- `~/.openclaw/skills/` (OpenClaw 默认目录)
- `./data/skills/` (ClawMemory 安装目录)
