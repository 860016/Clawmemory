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
│   └── utils/          # 工具函数
├── scripts/            # 构建脚本
│   ├── build.ps1       # Windows 构建
│   └── build.sh        # Linux/macOS 构建
└── go.mod
```

## 功能列表

所有功能完全开源（MIT License）：

| 功能 | 说明 |
|------|------|
| 记忆管理 | 多层级记忆分类、衰减算法、强化机制 |
| 知识图谱 | 实体提取、关系建立、图谱可视化 |
| Wiki | Markdown 知识页面、版本历史 |
| 日报 | 自动汇总、趋势分析 |
| 备份恢复 | 自动备份、数据导入导出 |
| 全文搜索 | SQLite FTS5 关键词搜索 |
| 向量搜索 | ChromaDB 语义搜索 |
| AI 增强 | 冲突扫描、Wiki 生成、记忆压缩、关系发现、智能路由、衰减评估 |
| 自进化系统 | NudgeReflect 技能审查、自动创建和改进技能 |
| OpenClaw 集成 | 实时文件监听、增量索引、显著性过滤 |
| Dialectic Reasoning | 多轮推理引擎 |

## 编译

### Windows
```powershell
.\scripts\build.ps1
```

### Linux/macOS
```bash
./scripts/build.sh
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

### AI 增强
- `GET /api/v1/ai/config` - 获取 AI 配置
- `GET /api/v1/ai/providers` - 列出可用 AI 模型
- `POST /api/v1/ai/test` - 测试 AI 连接
- `POST /api/v1/ai/extract` - AI 实体提取
- `GET /api/v1/ai/daily-report` - AI 日报生成

### 统计
- `GET /api/v1/stats` - 统计数据
- `GET /api/v1/stats/decay` - 衰减统计

## 与 Python 版本的区别

1. **单文件可执行**：Go 编译为单个二进制文件，无需 Python 环境
2. **内置 SQLite**：使用纯 Go SQLite 驱动，无需 CGO
3. **静态资源嵌入**：前端文件可嵌入二进制（使用 embed）
4. **更好的加密**：Go 的 crypto 包提供更完善的加密支持
5. **跨平台编译**：一次编写，到处编译
