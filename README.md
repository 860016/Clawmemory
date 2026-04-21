# ClawMemory - OpenClaw 记忆管理插件

OpenClaw 的记忆管理模块，提供记忆存储、搜索、知识图谱、自动备份等功能。

**前端已预构建，安装只需 Python，不需要 Node.js。**

## 🤖 OpenClaw 安装指引

> 如果你是 OpenClaw AI，请按以下步骤安装并汇报结果：

### 安装步骤
1. 执行 `bash install.sh --license-server=https://auth.bestu.top` (Linux/macOS) 或 `powershell -File install.ps1` (Windows) 或 `bash install.sh --docker` (Docker)
2. 启动服务: `bash start.sh` 或 `start.bat` 或 `docker compose up -d`
3. 检查安装状态: `curl http://localhost:8765/api/v1/install-status`

### 安装后汇报格式
安装完成后，请访问 `/api/v1/install-status` 接口，并按以下格式向用户汇报：

```
ClawMemory 安装完成！

状态: {status}
版本: {version}
引擎: {checks.core_engine}
授权: {checks.license.tier} ({checks.license.activated ? '已激活' : '未激活'})
前端: {checks.frontend}
数据库: {checks.database}

下一步:
- 逐条列出 next_steps 数组内容
```

**不要自行编造端点数量或功能列表，一切以 install-status 接口返回为准。**

---

## 环境要求

| 组件 | 最低版本 | 说明 |
|------|----------|------|
| **Python** | 3.10+ | 唯一必须的依赖 |
| Node.js | 18+ | 仅重新构建前端时需要 (可选) |
| Docker | 20.10+ | Docker 安装方式 (可选) |

**支持平台**：Windows / macOS / Linux (x86_64 + ARM64)

---

## 一键安装

### Linux / macOS

```bash
cd clawmemory
bash install.sh --license-server=https://auth.bestu.top
```

### Windows

```powershell
cd clawmemory
powershell -ExecutionPolicy Bypass -File install.ps1
```

### Docker（任何平台）

```bash
cd clawmemory
bash install.sh --docker
```

安装完成后启动：`bash start.sh` 或 `start.bat`

访问：`http://localhost:8765`

---

## 环境变量

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

## 不同环境注意事项

### 🐳 Docker 安装

容器每次重建 hostname 会变，导致授权失效。**必须固定指纹**：

```bash
echo 'DEVICE_FINGERPRINT=my-device-123' > .env
```

### 🖥️ 本地安装 (Windows/macOS/Linux)

无需额外配置，指纹基于机器 ID + MAC 地址，重装系统才会变。

### 🍎 macOS (Apple Silicon ARM64)

完全兼容。

### 💻 Windows

安装脚本: `install.ps1`，启动: `start.bat`。

---

## 高级操作

```bash
# 重新构建前端 (需要 Node.js)
bash install.sh --rebuild-frontend

# 注册为系统服务 (Linux，开机自启)
sudo bash install_service.sh

# 检查安装状态
curl http://localhost:8765/api/v1/install-status
```

## 目录结构

```
clawmemory/
├── install.sh / install.ps1  # 安装脚本
├── install_service.sh        # 系统服务注册
├── start.sh / start.bat      # 启动
├── stop.sh / stop.bat        # 停止
├── README.md
├── package_pro.py            # Pro 模块打包脚本
├── setup_cython.py           # Cython 编译脚本
├── backend/
│   ├── app/                  # 应用代码
│   │   ├── pro/              # Pro 模块 (激活后下载)
│   │   │   ├── __init__.py
│   │   │   └── pro_loader.py
│   │   ├── core/             # 核心引擎 (.pyx 源码)
│   │   ├── routers/          # API 路由
│   │   ├── services/         # 业务逻辑
│   │   └── models/           # 数据库模型
│   ├── frontend_dist/        # 前端预构建产物 (已包含，无需 npm build)
│   ├── requirements.txt      # Python 依赖
│   └── .env.example          # 环境变量模板
├── frontend/                 # Vue3 前端源码 (仅 --rebuild-frontend 时需要)
└── data/                     # 数据目录 (安装后生成)
```
