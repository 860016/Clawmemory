#!/bin/bash
# ============================================
# ClawMemory 一键安装脚本 (Linux/macOS)
# 用法: bash install.sh [选项]
#   --license-server=URL   授权服务器地址
#   --port=PORT            后端端口 (默认 8765)
#   --docker               使用 Docker 安装
#   --rebuild-frontend     重新构建前端 (需要 Node.js)
# ============================================

set -e

# 默认配置
LICENSE_SERVER="https://auth.bestu.top"
BACKEND_PORT=8765
USE_DOCKER=false
REBUILD_FRONTEND=false
INSTALL_DIR="$(cd "$(dirname "$0")" && pwd)"

# 解析参数
for arg in "$@"; do
  case $arg in
    --license-server=*) LICENSE_SERVER="${arg#*=}" ;;
    --port=*) BACKEND_PORT="${arg#*=}" ;;
    --docker) USE_DOCKER=true ;;
    --rebuild-frontend) REBUILD_FRONTEND=true ;;
    *) echo "未知参数: $arg"; exit 1 ;;
  esac
done

echo "============================================"
echo "  ClawMemory 安装程序"
echo "============================================"
echo "安装目录: $INSTALL_DIR"
echo "后端端口: $BACKEND_PORT"
echo "授权服务器: $LICENSE_SERVER"
echo "安装方式: $([ "$USE_DOCKER" = true ] && echo 'Docker' || echo '本地')"
echo "============================================"

# ==================== Docker 安装 ====================
if [ "$USE_DOCKER" = true ]; then
    echo ""
    echo "使用 Docker 安装..."

    if ! command -v docker &>/dev/null; then
        echo "❌ 未找到 Docker，请先安装"
        exit 1
    fi

    if ! command -v docker-compose &>/dev/null && ! docker compose version &>/dev/null; then
        echo "❌ 未找到 docker-compose，请先安装"
        exit 1
    fi

    # 创建 docker-compose.yml
    cat > "$INSTALL_DIR/docker-compose.yml" << EOF
version: '3.8'

services:
  backend:
    build: ./backend
    ports:
      - "${BACKEND_PORT}:8765"
    volumes:
      - ./data:/app/data
      - ./backend/keys:/app/keys
    environment:
      - CLAWMEMORY_SECRET_KEY=\${SECRET_KEY:-change-me-in-production}
      - CLAWMEMORY_DATA_DIR=/app/data
      - CLAWMEMORY_LICENSE_SERVER_URL=${LICENSE_SERVER}
      - CLAWMEMORY_RSA_PUBLIC_KEY_PATH=/app/keys/public.pem
      - DEVICE_FINGERPRINT=\${DEVICE_FINGERPRINT:-}
    restart: unless-stopped
EOF

    # 获取公钥
    mkdir -p "$INSTALL_DIR/backend/keys"
    if [ ! -f "$INSTALL_DIR/backend/keys/public.pem" ]; then
        echo "正在获取 RSA 公钥..."
        curl -sf "$LICENSE_SERVER/api/v1/public-key" -o "$INSTALL_DIR/backend/keys/public.pem" 2>/dev/null && \
            echo "✅ 公钥已获取" || \
            echo "⚠️  无法获取公钥，请手动放置到 backend/keys/public.pem"
    fi

    # 启动
    cd "$INSTALL_DIR"
    if docker compose version &>/dev/null; then
        docker compose up --build -d
    else
        docker-compose up --build -d
    fi

    echo ""
    echo "============================================"
    echo "  ✅ Docker 安装完成！"
    echo "============================================"
    echo "访问地址: http://localhost:${BACKEND_PORT}"
    echo ""
    echo "⚠️  Docker 用户注意："
    echo "  容器重建后设备指纹可能变化，导致授权失效。"
    echo "  解决：创建 .env 文件设置固定指纹："
    echo "  echo 'DEVICE_FINGERPRINT=my-device-123' > $INSTALL_DIR/.env"
    exit 0
fi

# ==================== 本地安装 ====================

# 检查 Python
echo ""
echo "[1/5] 检查环境..."
if command -v python3 &>/dev/null; then
    PYVER=$(python3 -c 'import sys; print(f"{sys.version_info.major}.{sys.version_info.minor}")')
    echo "  Python: $(python3 --version)"

    MAJOR=$(echo $PYVER | cut -d. -f1)
    MINOR=$(echo $PYVER | cut -d. -f2)
    if [ "$MAJOR" -lt 3 ] || ([ "$MAJOR" -eq 3 ] && [ "$MINOR" -lt 10 ]); then
        echo "  ❌ 需要 Python 3.10+，当前版本太低"
        exit 1
    fi
else
    echo "  ❌ 未找到 python3，请先安装 Python 3.10+"
    echo "  Ubuntu/Debian: sudo apt install python3 python3-venv python3-pip"
    echo "  macOS: brew install python@3.12"
    exit 1
fi

# 检查 C 编译器 (bcrypt 需要)
if ! command -v gcc &>/dev/null && ! command -v cc &>/dev/null; then
    echo "  ⚠️  未找到 C 编译器，bcrypt 可能安装失败（将使用替代方案）"
fi

# 检测架构
ARCH=$(uname -m)
echo "  架构: $ARCH"
if [[ "$ARCH" == "arm64" || "$ARCH" == "aarch64" ]]; then
    echo "  ℹ️  ARM 平台，部分依赖可能需要从源码编译（较慢）"
fi

# 检查前端
if [ -f "$INSTALL_DIR/backend/frontend_dist/index.html" ]; then
    echo "  前端: 已预构建 ✅"
else
    echo "  前端: 未构建 (需要 Node.js 重新构建，或跳过)"
fi

# 创建虚拟环境
echo ""
echo "[2/5] 创建 Python 虚拟环境并安装依赖..."
cd "$INSTALL_DIR/backend"

# 优先使用 venv，失败则用 virtualenv（部分 ARM Docker 环境无 ensurepip）
if python3 -m venv venv 2>/dev/null; then
    echo "  ✅ 虚拟环境已创建 (venv)"
else
    echo "  ⚠️  venv 不可用，尝试 virtualenv..."
    pip3 install virtualenv -q 2>/dev/null || pip install virtualenv -q 2>/dev/null || true
    if command -v virtualenv &>/dev/null; then
        virtualenv venv
        echo "  ✅ 虚拟环境已创建 (virtualenv)"
    else
        python3 -m virtualenv venv && echo "  ✅ 虚拟环境已创建 (virtualenv)" || {
            echo "  ❌ 无法创建虚拟环境，请安装 python3-venv 或 virtualenv"
            exit 1
        }
    fi
fi

source venv/bin/activate
pip install --upgrade pip -q

# 尝试安装，bcrypt 失败则降级
if ! pip install -r requirements.txt -q 2>/dev/null; then
    echo "  ⚠️  部分依赖安装失败，尝试降级..."
    pip install $(grep -v '^bcrypt' requirements.txt | grep -v '^#' | grep -v '^$') -q 2>/dev/null
    pip install "passlib[bcrypt]>=1.7.4" -q 2>/dev/null || pip install passlib -q 2>/dev/null || true
    echo "  ✅ 降级安装完成"
else
    echo "  ✅ 依赖安装完成"
fi

# 安装核心安全引擎：Rust (预编译 wheel) → 纯 Python 兜底
# 注意：.pyx 源文件不在发布包中，核心安全逻辑由 Rust wheel 提供
echo ""
echo "  安装安全引擎..."
CORE_ENGINE="python"

# 方案1: 下载预编译的 Rust wheel (最安全，不需要编译器)
ARCH=$(python3 -c "import platform; print(platform.machine().lower())")
SYS=$(python3 -c "import platform; print(platform.system().lower())")
PYVER=$(python3 -c "import sys; print(f'cp{sys.version_info.major}{sys.version_info.minor}')")

# 确定 wheel 文件名模式
if [ "$SYS" = "linux" ]; then
    # Linux: manylinux 格式
    if [ "$ARCH" = "aarch64" ]; then
        WHEEL_PATTERN="clawmemory_core-2.1.0-${PYVER}-${PYVER}-manylinux*_aarch64.whl"
    else
        WHEEL_PATTERN="clawmemory_core-2.1.0-${PYVER}-${PYVER}-manylinux*_x86_64.whl"
    fi
elif [ "$SYS" = "darwin" ]; then
    WHEEL_PATTERN="clawmemory_core-2.1.0-${PYVER}-${PYVER}-macosx*_${ARCH}.whl"
elif [ "$SYS" = "windows" ]; then
    WHEEL_PATTERN="clawmemory_core-2.1.0-${PYVER}-${PYVER}-win_${ARCH}.whl"
else
    WHEEL_PATTERN="clawmemory_core-2.1.0-${PYVER}-${PYVER}-*.whl"
fi

WHEEL_URL="https://github.com/860016/Clawmemory/releases/latest/download/${WHEEL_PATTERN}"
echo "  尝试下载 Rust 安全引擎: ${WHEEL_PATTERN}"
if curl -sfL "$WHEEL_URL" -o /tmp/clawmemory_core.whl 2>/dev/null; then
    if pip install /tmp/clawmemory_core.whl -q 2>/dev/null; then
        echo "  ✅ Rust 安全引擎已安装 (预编译wheel)"
        CORE_ENGINE="rust"
    else
        echo "  ⚠️  wheel 下载成功但安装失败（架构/版本不匹配）"
    fi
    rm -f /tmp/clawmemory_core.whl
else
    echo "  ⚠️  未找到匹配的预编译 wheel"
fi

# 方案2: 本地编译 Rust (需要 rustc，仅开发者使用)
if [ "$CORE_ENGINE" = "python" ] && [ -d "$INSTALL_DIR/backend/clawmemory_core/src" ] && command -v rustc &>/dev/null; then
    echo "  尝试本地编译 Rust 引擎..."
    if command -v maturin &>/dev/null || pip install maturin -q 2>/dev/null; then
        cd "$INSTALL_DIR/backend/clawmemory_core"
        if maturin develop --release 2>&1; then
            echo "  ✅ Rust 安全引擎已编译安装"
            CORE_ENGINE="rust"
        else
            echo "  ⚠️  Rust 编译失败"
        fi
    fi
fi

# 方案3: Cython 编译 (.pyx → .so，需要 C 编译器)
if [ "$CORE_ENGINE" = "python" ] && [ -f "$INSTALL_DIR/backend/setup_cython.py" ] && command -v gcc &>/dev/null; then
    echo "  尝试 Cython 编译安全引擎..."
    if pip install cython -q 2>/dev/null; then
        cd "$INSTALL_DIR/backend"
        if python3 setup_cython.py build_ext --inplace 2>&1; then
            # 验证是否生成了 .so 文件
            if find app/core -name "*.so" -o -name "*.pyd" 2>/dev/null | head -1 | grep -q .; then
                echo "  ✅ Cython 安全引擎已编译 (中等安全)"
                CORE_ENGINE="cython"
            else
                echo "  ⚠️  Cython 编译完成但未生成 .so 文件"
            fi
        else
            echo "  ⚠️  Cython 编译失败"
        fi
    fi
fi

if [ "$CORE_ENGINE" = "python" ]; then
    echo "  ⚠️  使用纯 Python 模式（安全性较低）"
    echo "  ⚠️  建议安装 clawmemory-core Rust wheel 以获得 RSA 硬验证保护"
fi

# 配置环境变量
echo ""
echo "[3/5] 配置环境变量..."

# Auto-detect OpenClaw Gateway from ~/.openclaw/openclaw.json
GATEWAY_URL="http://localhost:18789"
GATEWAY_API_KEY=""
OPENCLAW_HOME="${OPENCLAW_STATE_DIR:-$HOME/.openclaw}"
if [ -d "$OPENCLAW_HOME" ] && [ -f "$OPENCLAW_HOME/openclaw.json" ]; then
    echo "  检测到 OpenClaw 配置目录: $OPENCLAW_HOME"
    # Parse openclaw.json (JSON5 format) to extract gateway port + auth token
    DETECTED=$(python3 -c "
import json, re, sys

config_path = '$OPENCLAW_HOME/openclaw.json'
try:
    with open(config_path, 'r', encoding='utf-8') as f:
        raw = f.read()
    # Strip JSON5 comments and trailing commas for stdlib json
    raw = re.sub(r'//.*?$', '', raw, flags=re.MULTILINE)
    raw = re.sub(r'/\*.*?\*/', '', raw, flags=re.DOTALL)
    raw = re.sub(r',\s*([}\]])', r'\1', raw)
    data = json.loads(raw)
    gw = data.get('gateway', {}) or {}
    port = gw.get('port', 18789)
    auth = gw.get('auth', {}) or {}
    token = auth.get('token', '') or auth.get('password', '')
    remote = gw.get('remote', {}) or {}
    url = remote.get('url', '')
    if url:
        url = url.replace('ws://', 'http://').replace('wss://', 'https://')
    else:
        url = f'http://localhost:{port}'
    print(f'{url}|{token}')
except Exception as e:
    print('|')
" 2>/dev/null)
    if [ -n "$DETECTED" ] && [ "$DETECTED" != "|" ]; then
        DETECTED_URL=$(echo "$DETECTED" | cut -d'|' -f1)
        DETECTED_KEY=$(echo "$DETECTED" | cut -d'|' -f2)
        if [ -n "$DETECTED_URL" ]; then
            GATEWAY_URL="$DETECTED_URL"
            echo "  ✅ 自动检测到 Gateway URL: $GATEWAY_URL"
        fi
        if [ -n "$DETECTED_KEY" ]; then
            GATEWAY_API_KEY="$DETECTED_KEY"
            echo "  ✅ 自动检测到 Gateway API Key"
        fi
    fi
fi

if [ ! -f .env ]; then
    SECRET_KEY=$(python3 -c "import secrets; print(secrets.token_urlsafe(32))")

    cat > .env << EOF
CLAWMEMORY_SECRET_KEY=${SECRET_KEY}
CLAWMEMORY_DATA_DIR=${INSTALL_DIR}/data
CLAWMEMORY_DEBUG=false
CLAWMEMORY_PORT=${BACKEND_PORT}
CLAWMEMORY_LICENSE_SERVER_URL=${LICENSE_SERVER}
CLAWMEMORY_RSA_PUBLIC_KEY_PATH=${INSTALL_DIR}/backend/keys/public.pem
CLAWMEMORY_CORS_ORIGINS=["*"]
EOF
    echo "  ✅ .env 已创建"
else
    echo "  ⏭️  .env 已存在，跳过"
fi

# 创建必要目录
mkdir -p "$INSTALL_DIR/data"
mkdir -p "$INSTALL_DIR/backend/data"
mkdir -p "$INSTALL_DIR/backend/keys"

# RSA 公钥
echo ""
echo "[4/5] RSA 公钥..."
if [ ! -f "$INSTALL_DIR/backend/keys/public.pem" ]; then
    echo "  正在从授权服务器获取公钥..."
    PUBKEY_FETCHED=false

    # 方案1: 直接从授权服务器获取
    if curl -sf "$LICENSE_SERVER/api/v1/public-key" -o "$INSTALL_DIR/backend/keys/public.pem" 2>/dev/null; then
        # 验证文件内容是否为有效 PEM
        if head -1 "$INSTALL_DIR/backend/keys/public.pem" 2>/dev/null | grep -q "BEGIN"; then
            echo "  ✅ 公钥已从授权服务器获取"
            PUBKEY_FETCHED=true
        else
            rm -f "$INSTALL_DIR/backend/keys/public.pem"
        fi
    fi

    # 方案2: 尝试备用路径 /api/v1/public-key/pem
    if [ "$PUBKEY_FETCHED" = false ]; then
        if curl -sf "$LICENSE_SERVER/api/v1/public-key/pem" -o "$INSTALL_DIR/backend/keys/public.pem" 2>/dev/null; then
            if head -1 "$INSTALL_DIR/backend/keys/public.pem" 2>/dev/null | grep -q "BEGIN"; then
                echo "  ✅ 公钥已从备用路径获取"
                PUBKEY_FETCHED=true
            else
                rm -f "$INSTALL_DIR/backend/keys/public.pem"
            fi
        fi
    fi

    # 方案3: 尝试从 GitHub releases 下载
    if [ "$PUBKEY_FETCHED" = false ]; then
        if curl -sfL "https://github.com/860016/Clawmemory/releases/latest/download/public.pem" -o "$INSTALL_DIR/backend/keys/public.pem" 2>/dev/null; then
            if head -1 "$INSTALL_DIR/backend/keys/public.pem" 2>/dev/null | grep -q "BEGIN"; then
                echo "  ✅ 公钥已从 GitHub Releases 获取"
                PUBKEY_FETCHED=true
            else
                rm -f "$INSTALL_DIR/backend/keys/public.pem"
            fi
        fi
    fi

    if [ "$PUBKEY_FETCHED" = false ]; then
        echo "  ⚠️  无法自动获取公钥（授权服务器可能不可达）"
        echo "  请手动将授权平台的公钥文件复制到:"
        echo "  $INSTALL_DIR/backend/keys/public.pem"
        echo "  或在启动后通过管理界面上传公钥"
    fi
else
    echo "  ✅ 公钥已存在"
fi

# 重新构建前端 (仅显式请求时)
if [ "$REBUILD_FRONTEND" = true ]; then
    echo ""
    echo "[5/5] 重新构建前端..."
    if command -v node &>/dev/null; then
        cd "$INSTALL_DIR/frontend"
        npm install --silent 2>/dev/null
        npm run build
        echo "  ✅ 前端构建完成"
    else
        echo "  ❌ 未找到 Node.js，无法构建前端"
        echo "  使用已预构建的前端"
    fi
else
    echo ""
    echo "[5/5] 前端: 使用预构建版本"
    if [ -f "$INSTALL_DIR/backend/frontend_dist/index.html" ]; then
        echo "  ✅ 前端已就绪"
    else
        echo "  ⚠️  前端未构建，如需 Web 界面请运行: bash install.sh --rebuild-frontend"
    fi
fi

# 初始化数据库
cd "$INSTALL_DIR/backend"
source venv/bin/activate
python3 -c "from app.database import init_db; init_db(); print('  ✅ 数据库初始化完成')"

# 创建启动脚本
cat > "$INSTALL_DIR/start.sh" << START_EOF
#!/bin/bash
cd "$INSTALL_DIR/backend"
source venv/bin/activate
exec uvicorn app.main:app --host 0.0.0.0 --port ${BACKEND_PORT}
START_EOF
chmod +x "$INSTALL_DIR/start.sh"

# 创建停止脚本
cat > "$INSTALL_DIR/stop.sh" << STOP_EOF
#!/bin/bash
pkill -f "uvicorn app.main:app.*--port ${BACKEND_PORT}" 2>/dev/null && echo "已停止" || echo "未在运行"
STOP_EOF
chmod +x "$INSTALL_DIR/stop.sh"

echo ""
echo "============================================"
echo "  ✅ 安装完成！"
echo "============================================"
echo ""
echo "启动服务:  bash $INSTALL_DIR/start.sh"
echo "停止服务:  bash $INSTALL_DIR/stop.sh"
echo "访问地址:  http://localhost:${BACKEND_PORT}"
echo ""
echo "检查状态:  curl http://localhost:${BACKEND_PORT}/api/v1/install-status"
echo "Docker:    bash install.sh --docker"
echo "重装前端:  bash install.sh --rebuild-frontend"
echo "开机自启:  sudo bash install_service.sh"
