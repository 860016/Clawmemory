# ClawMemory Python → Go 迁移指南

## 已完成的工作

### 1. 项目架构设计
- 采用标准 Go 项目结构：`cmd/`, `internal/`, `pkg/`, `pro/`
- 开源部分和 Pro 部分分离，使用 build tags 控制编译

### 2. 核心功能实现

#### 开源部分
| 模块 | 文件 | 功能 |
|------|------|------|
| 主入口 | `cmd/server/main.go` | HTTP 服务器启动 |
| 配置 | `internal/config/config.go` | 环境变量和配置文件 |
| 数据库 | `internal/database/database.go` | SQLite + GORM |
| 模型 | `internal/models/models.go` | 所有数据模型 |
| 中间件 | `internal/middleware/middleware.go` | JWT、CORS、日志 |
| 认证服务 | `internal/services/auth_service.go` | 注册、登录、密码修改 |
| 记忆服务 | `internal/services/memory_service.go` | CRUD、搜索 |
| 知识服务 | `internal/services/knowledge_service.go` | 实体、关系、图谱 |
| Wiki 服务 | `internal/services/wiki_service.go` | Wiki 页面管理 |
| 日报服务 | `internal/services/daily_report_service.go` | 日报管理 |
| 备份服务 | `internal/services/backup_service.go` | 备份创建和下载 |
| 授权服务 | `internal/services/license_service.go` | RSA 验证、激活 |
| API 路由 | `internal/api/routes.go` | 路由注册 |
| API 处理器 | `internal/api/handlers.go` | HTTP 处理器 |

#### Pro 部分（编译时加密）
| 模块 | 文件 | 功能 |
|------|------|------|
| 记忆衰减 | `pro/decay/decay.go` | 高级衰减算法 |
| 冲突检测 | `pro/conflict/conflict.go` | 冲突检测与合并 |
| 智能路由 | `pro/router/router.go` | 模型路由和 Token 统计 |

### 3. 授权验证系统
- RSA 公钥验证签名
- 支持从服务器刷新公钥
- 设备指纹绑定
- 功能开关控制

### 4. 跨平台编译脚本
- `scripts/build.ps1` - Windows PowerShell 脚本
- `scripts/build.sh` - Linux/macOS Bash 脚本
- 支持：Windows/Linux/macOS × AMD64/ARM64

### 5. 前端适配
- `frontend/src/api/go-client.ts` - Go 后端 HTTP 客户端
- `frontend/src/api/go-*.ts` - 各模块 API 适配
- 与 Python 后端 API 格式兼容

## 与 Python 版本的功能对比

### 开源版功能（一致）
- ✅ 记忆管理（CRUD、软删除、恢复）
- ✅ 关键词搜索
- ✅ 知识图谱（实体、关系）
- ✅ Wiki 管理
- ✅ 日报管理
- ✅ 备份恢复
- ✅ 用户认证（JWT）
- ✅ 授权验证

### Pro 版功能
| 功能 | Python | Go |
|------|--------|-----|
| 记忆衰减算法 | Python 实现 | Go 实现（编译加密） |
| 冲突检测 | Python 实现 | Go 实现（编译加密） |
| 智能路由 | Python 实现 | Go 实现（编译加密） |
| 向量搜索 | ChromaDB | 简化实现 |
| OpenClaw 集成 | 完整支持 | 暂不支持 |

### 新增功能
- ✅ 单文件可执行（无需 Python 环境）
- ✅ 更好的 RSA 加密支持
- ✅ 跨平台编译更简单
- ✅ Pro 模块编译时保护（非运行时加载）

## 本地编译步骤

### 1. 安装 Go
```bash
# Windows: 下载 MSI 安装包 https://golang.org/dl/
# macOS: brew install go
# Linux: sudo apt-get install golang-go
```

### 2. 克隆项目
```bash
cd go-backend
```

### 3. 下载依赖
```bash
go mod tidy
```

### 4. 编译开源版
```bash
# Windows
go build -o clawmemory.exe ./cmd/server

# Linux/macOS
go build -o clawmemory ./cmd/server
```

### 5. 编译 Pro 版
```bash
# Windows
go build -tags pro -o clawmemory-pro.exe ./cmd/server

# Linux/macOS
go build -tags pro -o clawmemory-pro ./cmd/server
```

### 6. 跨平台编译
```bash
# Windows -> Linux
set GOOS=linux
set GOARCH=amd64
go build -o clawmemory-linux ./cmd/server

# Windows -> macOS
set GOOS=darwin
set GOARCH=amd64
go build -o clawmemory-darwin ./cmd/server
```

## 运行

```bash
# 直接运行
./clawmemory

# 指定端口
set PORT=8080
./clawmemory

# 前端文件放在 frontend_dist/ 目录
```

## 前端构建

```bash
cd frontend

# 修改 API 导入（临时方案）
# 在 main.ts 或相关文件中切换 API 客户端

npm install
npm run build

# 将 dist/ 复制到 go-backend/frontend_dist/
```

## 数据迁移

Go 版使用相同的 SQLite 数据库结构，可以直接使用 Python 版的数据库文件：

```bash
# 复制数据库
cp python-backend/data/clawmemory.db go-backend/data/clawmemory.db
```

## 下一步建议

1. **安装 Go 环境** 并测试编译
2. **完善向量搜索** - 集成 ChromaDB 或 Milvus
3. **添加 OpenClaw 支持** - 重新实现技能扫描和记忆导入
4. **前端切换** - 修改前端 API 导入路径
5. **测试授权流程** - 验证 RSA 签名和 Pro 功能开关
6. **性能优化** - Go 的并发特性可以大幅提升性能
