# ClawMemory Go 后端

## 项目结构

```
go-backend/
├── cmd/server/          # 主程序入口
│   └── main.go
├── internal/            # 内部包（不对外暴露）
│   ├── api/            # HTTP API 路由和处理器
│   ├── config/         # 配置管理
│   ├── database/       # 数据库连接和迁移
│   ├── middleware/     # HTTP 中间件（认证、CORS、日志）
│   ├── models/         # 数据模型（GORM）
│   └── services/       # 业务逻辑层
├── pkg/                # 可复用包
│   ├── license/        # 授权验证和加密
│   └── utils/          # 工具函数
├── pro/                # Pro 功能模块（编译时加密）
│   ├── decay/          # 记忆衰减算法
│   ├── conflict/       # 冲突检测与合并
│   └── router/         # 智能模型路由
├── scripts/            # 构建脚本
│   ├── build.ps1       # Windows 构建
│   └── build.sh        # Linux/macOS 构建
└── go.mod
```

## 功能对比

| 功能 | 开源版 | Pro 版 |
|------|--------|--------|
| 记忆管理 | ✓ | ✓ |
| 知识图谱 | ✓ | ✓ |
| Wiki | ✓ | ✓ |
| 日报 | ✓ | ✓ |
| 备份恢复 | ✓ | ✓ |
| 全文搜索 | ✓ | ✓ |
| 向量搜索 | ✓ | ✓ |
| 记忆衰减算法 | 基础 | 高级 |
| 冲突检测 | 基础 | 高级 |
| 智能路由 | 基础 | 高级 |
| 自动备份 | - | ✓ |
| 多设备同步 | - | ✓ |

## 编译

### Windows
```powershell
# 开源版
.\scripts\build.ps1

# Pro 版
go build -tags pro -o clawmemory-pro.exe ./cmd/server
```

### Linux/macOS
```bash
# 开源版
./scripts/build.sh

# Pro 版
go build -tags pro -o clawmemory-pro ./cmd/server
```

### 跨平台编译
```bash
# Windows -> Linux
GOOS=linux GOARCH=amd64 go build -o clawmemory-linux ./cmd/server

# Windows -> macOS
GOOS=darwin GOARCH=amd64 go build -o clawmemory-darwin ./cmd/server
```

## 运行

```bash
# 直接运行
./clawmemory

# 指定端口
PORT=8080 ./clawmemory

# 指定数据目录
DATA_DIR=/path/to/data ./clawmemory
```

## API 文档

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
- `POST /api/v1/memories/:id/restore` - 恢复
- `GET /api/v1/memories/search/keyword` - 关键词搜索
- `GET /api/v1/memories/search/semantic` - 语义搜索

### 知识图谱
- `GET /api/v1/knowledge/entities` - 实体列表
- `POST /api/v1/knowledge/entities` - 创建实体
- `GET /api/v1/knowledge/relations` - 关系列表
- `POST /api/v1/knowledge/relations` - 创建关系
- `GET /api/v1/knowledge/graph` - 获取完整图谱

### Wiki
- `GET /api/v1/wiki` - 列表
- `POST /api/v1/wiki` - 创建
- `GET /api/v1/wiki/:id` - 详情
- `PUT /api/v1/wiki/:id` - 更新
- `DELETE /api/v1/wiki/:id` - 删除

### 日报
- `GET /api/v1/reports` - 列表
- `POST /api/v1/reports` - 创建
- `GET /api/v1/reports/:date` - 按日期获取

### 备份
- `GET /api/v1/backups` - 列表
- `POST /api/v1/backups` - 创建备份
- `GET /api/v1/backups/:filename/download` - 下载备份

### 授权
- `GET /api/v1/license/info` - 授权信息
- `POST /api/v1/license/activate` - 激活
- `POST /api/v1/license/deactivate` - 停用

### 统计
- `GET /api/v1/stats` - 统计数据
- `GET /api/v1/stats/decay` - 衰减统计（Pro）

## 与 Python 版本的区别

1. **单文件可执行**：Go 编译为单个二进制文件，无需 Python 环境
2. **内置 SQLite**：使用 mattn/go-sqlite3，无需额外安装
3. **静态资源嵌入**：前端文件可嵌入二进制（使用 embed）
4. **更好的加密**：Go 的 crypto 包提供更完善的加密支持
5. **跨平台编译**：一次编写，到处编译
6. **Pro 模块保护**：使用 build tags 控制 Pro 功能编译
