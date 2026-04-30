# ClawMemory v2.11.0 — AI 记忆管理工具

ClawMemory 是一款现代化的 AI 记忆管理工具，支持知识图谱、智能日报、记忆衰减分析、冲突检测等高级功能。

**后端**: Go (高性能，全平台支持)

---

## ✨ 核心特性

- 记忆管理
- 知识图谱 (网格/图谱/列表三视图)
- Wiki 知识库
- 智能日报
- 本地数据导出/导入
- 全文搜索 + 语义搜索
- ChromaDB 向量搜索
- 智能记忆加载 (Token 预算控制)
- 记忆强化机制
- OpenClaw 记忆导入
- 技能系统
- 记忆衰减算法
- 冲突检测
- Token 智能路由
- AI 提取/摘要
- 趋势分析
- 报告生成
- Pro 高级功能

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
| `SECRET_KEY` | (自动生成) | JWT 密钥 |
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
- `GET /api/v1/memories/smart-load` - 智能加载
- `POST /api/v1/memories/:id/reinforce` - 强化记忆
- `POST /api/v1/memories/generate-summaries` - 生成摘要

### 知识图谱
- `GET /api/v1/knowledge/entities` - 实体列表
- `POST /api/v1/knowledge/entities` - 创建实体
- `GET /api/v1/knowledge/relations` - 关系列表
- `GET /api/v1/knowledge/graph` - 图谱数据

### 授权
- `GET /api/v1/license/info` - 授权信息
- `POST /api/v1/license/activate` - 激活授权
- `POST /api/v1/license/deactivate` - 停用授权

---

## 📝 更新日志

### v2.11.0 (2026-04-30)
- 🐛 修复：SQLite 驱动注册冲突（modernc.org/sqlite 与 glebarez/go-sqlite 同名注册）
- 🔐 新增：终端命令重置密码 `./clawmemory --reset-password NEW_PASSWORD`
- 🔐 新增：`./clawmemory --version` 查看版本号
- 📋 新增：设置页面显示版本号、检查更新、更新说明
- 🔄 新增：GitHub Releases 版本更新检查 API
- 🔒 改进：忘记密码 API 增加用户名验证
- 📦 统一版本号管理（config.AppVersion）

### v2.9.1 (2026-04-30)
- 🔒 安全加固：密码重置验证、备份路径遍历防护
- 🐛 修复：ChromaDB 搜索结果字段补全、导出数据格式修正
- 🔗 前后端一致性：错误响应字段统一为 `error`

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

### 第三步：配置 OpenClaw 连接 ClawMemory

安装 ClawMemory 后，需要在 OpenClaw 的 `AGENTS.md` 中添加配置，让 OpenClaw 自动将对话记忆存入 ClawMemory。

**OpenClaw AGENTS.md 位置**：
- Windows: `C:\Users\<用户名>\.openclaw\AGENTS.md`
- Linux/macOS: `~/.openclaw/AGENTS.md`

**添加以下内容到 AGENTS.md**：

```markdown
## ClawMemory - 主要记忆管理工具

- **类型**: memory
- **地址**: http://localhost:8765
- **功能**: 记忆管理、知识图谱、智能日报、向量搜索
- **API 端点**: 
  - 创建记忆: POST /api/v1/memories
  - 记忆列表: GET /api/v1/memories
  - 智能加载: GET /api/v1/memories/smart-load
  - 知识图谱: GET /api/v1/knowledge/graph
```

配置完成后，OpenClaw 会通过 API 自动将对话记忆存入 ClawMemory。

### 第四步：导入 OpenClaw 已有记忆

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
