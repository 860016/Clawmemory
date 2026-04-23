#!/bin/bash
# ClawMemory Linux/macOS 安装脚本 v2.0
# 用法: bash install.sh [--license-server=URL] [--port=PORT]

set -e

LICENSE_SERVER="https://auth.bestu.top"
PORT=8765
UPGRADE=false
INSTALL_DIR="$(cd "$(dirname "$0")" && pwd)"

for arg in "$@"; do
  case $arg in
    --license-server=*) LICENSE_SERVER="${arg#*=}" ;;
    --port=*) PORT="${arg#*=}" ;;
    --upgrade) UPGRADE=true ;;
    -h|--help) echo "用法: bash install.sh [--license-server=URL] [--port=PORT] [--upgrade]"; exit 0 ;;
    *) echo "未知参数: $arg"; exit 1 ;;
  esac
done

echo "============================================"
echo "  ClawMemory 安装程序 v2.0 (Linux/macOS)"
echo "============================================"
echo "安装目录: $INSTALL_DIR"
echo "后端端口: $PORT"
echo "授权服务器: $LICENSE_SERVER"
[ "$UPGRADE" = true ] && echo "模式: 升级 (保留配置和数据)"
echo ""

# 检查 Go 环境
echo "[1/5] 检查 Go 环境..."
if ! command -v go &>/dev/null; then
    echo "  未找到 Go，请先安装 Go 1.21+"
    echo "  下载: https://go.dev/dl/"
    exit 1
fi
echo "  Go: $(go version)"

# 检查前端
if [ -f "$INSTALL_DIR/go-backend/frontend_dist/index.html" ]; then
    echo "  前端: 已预构建"
else
    echo "  前端: 未构建"
    if command -v node &>/dev/null; then
        echo "  检测到 Node.js，将自动构建前端"
        REBUILD_FRONTEND=true
    fi
fi

# 前端构建
echo ""
echo "[2/5] 前端构建..."
if [ "$REBUILD_FRONTEND" = true ] || [ ! -f "$INSTALL_DIR/go-backend/frontend_dist/index.html" ]; then
    if command -v node &>/dev/null; then
        echo "  构建前端..."
        cd "$INSTALL_DIR/frontend"
        npm install --silent 2>/dev/null || true
        npm run build
        # 复制到 go-backend
        if [ -d "$INSTALL_DIR/frontend/dist" ]; then
            rm -rf "$INSTALL_DIR/go-backend/frontend_dist" 2>/dev/null || true
            cp -r "$INSTALL_DIR/frontend/dist" "$INSTALL_DIR/go-backend/frontend_dist"
            echo "  前端已复制到 go-backend"
        fi
        cd "$INSTALL_DIR"
    else
        echo "  未找到 Node.js，跳过前端构建"
        echo "  请确保 go-backend/frontend_dist 目录存在"
    fi
else
    echo "  前端已就绪"
fi

# 编译 Go 后端
echo ""
echo "[3/5] 编译 Go 后端..."
cd "$INSTALL_DIR/go-backend"
go build -o clawmemory ./cmd/server
echo "  Go 后端编译成功"
cd "$INSTALL_DIR"

# 配置环境变量
echo ""
echo "[4/5] 配置环境变量..."
if [ ! -f "$INSTALL_DIR/go-backend/.env" ]; then
    SECRET_KEY=$(openssl rand -hex 32 2>/dev/null || cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1)
    cat > "$INSTALL_DIR/go-backend/.env" << EOF
SECRET_KEY=$SECRET_KEY
PORT=$PORT
DATA_DIR=$INSTALL_DIR/data
LICENSE_SERVER_URL=$LICENSE_SERVER
EOF
    echo "  .env 已创建"
elif [ "$UPGRADE" = true ]; then
    echo "  升级模式，保留现有 .env"
else
    echo "  .env 已存在，跳过"
fi

# 创建目录
mkdir -p "$INSTALL_DIR/data"

# 创建启动脚本
echo ""
echo "[5/5] 创建启动脚本..."
cat > "$INSTALL_DIR/start.sh" << 'EOF'
#!/bin/bash
cd "$(dirname "$0")/go-backend"
./clawmemory
EOF
chmod +x "$INSTALL_DIR/start.sh"

cat > "$INSTALL_DIR/stop.sh" << EOF
#!/bin/bash
pkill -f "clawmemory" || echo "ClawMemory 未在运行"
EOF
chmod +x "$INSTALL_DIR/stop.sh"

echo ""
echo "============================================"
echo "  安装完成!"
echo "============================================"
echo ""
echo "启动服务:  $INSTALL_DIR/start.sh"
echo "停止服务:  $INSTALL_DIR/stop.sh"
echo "访问地址:  http://localhost:$PORT"
echo "后端引擎:  Go (高性能)"