#!/bin/bash
# ============================================
# ClawMemory 一键安装脚本 (Linux/macOS)
# 用法: bash install.sh [选项]
#   --license-server=URL   授权服务器地址
#   --port=PORT            后端端口 (默认 8765)
#   --docker               使用 Docker 安装
#   --rebuild-frontend     重新构建前端 (需要 Node.js)
#   --upgrade              升级模式 (保留 .env 和数据)
# ============================================

set -e

# 默认配置
LICENSE_SERVER="https://auth.bestu.top"
BACKEND_PORT=8765
USE_DOCKER=false
REBUILD_FRONTEND=false
UPGRADE_MODE=false
INSTALL_DIR="$(cd "$(dirname "$0")" && pwd)"

# 解析参数
for arg in "$@"; do
  case $arg in
    --license-server=*) LICENSE_SERVER="${arg#*=}" ;;
    --port=*) BACKEND_PORT="${arg#*=}" ;;
    --docker) USE_DOCKER=true ;;
    --rebuild-frontend) REBUILD_FRONTEND=true ;;
    --upgrade) UPGRADE_MODE=true ;;
    -h|--help) echo "用法: bash install.sh [--license-server=URL] [--port=PORT] [--docker] [--rebuild-frontend] [--upgrade]"; exit 0 ;;
    *) echo "未知参数: $arg (用 -h 查看帮助)"; exit 1 ;;
  esac
done

echo "============================================"
echo "  ClawMemory 安装程序 v2.8.0"
echo "============================================"
echo "安装目录: $INSTALL_DIR"
echo "后端端口: $BACKEND_PORT"
echo "授权服务器: $LICENSE_SERVER"
echo "安装方式: $([ "$USE_DOCKER" = true ] && echo 'Docker' || echo '本地')"
[ "$UPGRADE_MODE" = true ] && echo "模式: 升级 (保留配置和数据)"
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
    echo "  echo 'DEVICE_FINGERPRINT=my-device-123' > .env"
    exit 0
fi

# ==================== 本地安装 ====================

# 检查 Python
echo ""
echo "[1/6] 检查环境..."
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

# 检查 C 编译器 (bcrypt 可能需要)
if ! command -v gcc &>/dev/null && ! command -v cc &>/dev/null; then
    echo "  ⚠️  未找到 C 编译器，bcrypt 可能安装失败（将使用替代方案）"
fi

# 检测架构
ARCH=$(uname -m)
echo "  架构: $ARCH"

# 检查前端
if [ -f "$INSTALL_DIR/backend/frontend_dist/index.html" ]; then
    echo "  前端: 已预构建 ✅"
else
    echo "  前端: 未构建"
    if command -v node &>/dev/null; then
        echo "  ℹ️  检测到 Node.js，将自动构建前端"
        REBUILD_FRONTEND=true
    fi
fi

# 创建/复用虚拟环境
echo ""
echo "[2/6] 创建 Python 虚拟环境并安装依赖..."
cd "$INSTALL_DIR/backend"

if [ -d "venv" ] && [ -f "venv/bin/activate" ]; then
    echo "  ✅ 虚拟环境已存在，复用中"
    source venv/bin/activate
else
    # 优先使用 venv，失败则用 virtualenv
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
fi

pip install --upgrade pip -q

# 安装依赖（不修改 requirements.txt，降级使用临时方式）
if ! pip install -r requirements.txt -q 2>/dev/null; then
    echo "  ⚠️  部分依赖安装失败，尝试降级 bcrypt..."
    # 创建临时文件去掉 bcrypt
    grep -v '^bcrypt' requirements.txt > /tmp/requirements_no_bcrypt.txt 2>/dev/null
    pip install -r /tmp/requirements_no_bcrypt.txt -q 2>/dev/null || true
    pip install "passlib[bcrypt]>=1.7.4" -q 2>/dev/null || pip install passlib -q 2>/dev/null || true
    rm -f /tmp/requirements_no_bcrypt.txt
    echo "  ✅ 降级安装完成"
else
    echo "  ✅ 依赖安装完成"
fi

# 安装核心安全引擎：C wheel (预编译) → 纯 Python 兜底
echo ""
echo "[3/6] 安装安全引擎..."
CORE_ENGINE="python"

# 检测 Python 版本和架构
PY_CPVER=$(python3 -c "import sys; print(f'cp{sys.version_info.major}{sys.version_info.minor}')")
PY_ARCH=$(python3 -c "import platform; print(platform.machine().lower())")
PY_SYS=$(python3 -c "import platform; print(platform.system().lower())")

# 方案1: 使用 GitHub API 下载匹配的预编译 C wheel (最可靠)
echo "  尝试下载 C 安全引擎..."
API_URL="https://api.github.com/repos/860016/Clawmemory/releases/latest"

WHEEL_URL=$(curl -sf "$API_URL" 2>/dev/null | python3 -c "
import json, sys
try:
    data = json.load(sys.stdin)
    arch_key = '${PY_ARCH}'.lower()
    # 归一化架构名
    if arch_key in ('x86_64', 'amd64'):
        if '${PY_SYS}' == 'linux':
            arch_key = 'x86_64'
        elif '${PY_SYS}' == 'darwin':
            arch_key = 'x86_64' if arch_key in ('x86_64', 'amd64') else 'arm64'
        else:
            arch_key = 'amd64'
    elif arch_key == 'aarch64':
        arch_key = 'aarch64'
    for a in data.get('assets', []):
        name = a['name']
        if not name.endswith('.whl'):
            continue
        if '${PY_CPVER}' not in name:
            continue
        # 匹配架构
        name_lower = name.lower()
        if 'arm64' in name_lower and arch_key != 'arm64':
            continue
        if 'aarch64' in name_lower and arch_key != 'aarch64':
            continue
        if 'win_amd64' in name_lower and not (arch_key == 'amd64' and '${PY_SYS}' == 'windows'):
            continue
        if 'macosx' in name_lower and '${PY_SYS}' != 'darwin':
            continue
        if 'manylinux' in name_lower and '${PY_SYS}' != 'linux':
            continue
        # 架构匹配
        if '${PY_SYS}' == 'linux' and arch_key == 'x86_64' and 'x86_64' not in name_lower:
            continue
        if '${PY_SYS}' == 'linux' and arch_key == 'aarch64' and 'aarch64' not in name_lower:
            continue
        print(a['browser_download_url'])
        break
except Exception:
    pass
" 2>/dev/null)

if [ -n "$WHEEL_URL" ]; then
    echo "  下载: $WHEEL_URL"
    if curl -sfL --progress-bar "$WHEEL_URL" -o /tmp/clawmemory_core.whl 2>/dev/null; then
        if pip install /tmp/clawmemory_core.whl -q 2>/dev/null; then
            echo "  ✅ C 安全引擎已安装 (预编译 wheel)"
            CORE_ENGINE="c"
        else
            echo "  ⚠️  wheel 下载成功但安装失败（架构/版本不匹配）"
        fi
        rm -f /tmp/clawmemory_core.whl
    else
        echo "  ⚠️  wheel 下载失败"
    fi
else
    echo "  ⚠️  未找到匹配的预编译 wheel (架构=$PY_ARCH, Python=$PY_CPVER, 系统=$PY_SYS)"
fi

# 方案2: 本地编译 C 扩展 (仅开发者，需要源码 + gcc + libssl-dev)
if [ "$CORE_ENGINE" = "python" ] && [ -f "$INSTALL_DIR/backend/clawmemory_core/setup.py" ]; then
    echo "  尝试本地编译 C 安全引擎..."
    # 安装 OpenSSL 开发库
    if [ "$PY_SYS" = "linux" ]; then
        if command -v apt-get &>/dev/null; then
            sudo apt-get install -y libssl-dev 2>/dev/null || true
        elif command -v yum &>/dev/null; then
            sudo yum install -y openssl-devel 2>/dev/null || true
        elif command -v apk &>/dev/null; then
            sudo apk add openssl-dev 2>/dev/null || true
        fi
    elif [ "$PY_SYS" = "darwin" ]; then
        if command -v brew &>/dev/null; then
            brew install openssl 2>/dev/null || true
        fi
    fi
    cd "$INSTALL_DIR/backend/clawmemory_core"
    if python3 setup.py build_ext --inplace 2>&1; then
        if python3 -c "import clawmemory_core" 2>/dev/null; then
            echo "  ✅ C 安全引擎已编译安装 (RSA 硬验证)"
            CORE_ENGINE="c"
        else
            echo "  ⚠️  C 编译完成但导入失败"
        fi
    else
        echo "  ⚠️  C 编译失败（可能缺少编译器或 OpenSSL 开发库）"
    fi
    cd "$INSTALL_DIR/backend"
fi

if [ "$CORE_ENGINE" = "python" ]; then
    echo "  ⚠️  使用纯 Python 模式（安全性较低）"
    echo "  ⚠️  建议安装 clawmemory-core C 编译版以获得 RSA 硬验证保护"
fi

# 配置环境变量
echo ""
echo "[4/6] 配置环境变量..."

# Auto-detect OpenClaw Gateway from ~/.openclaw/openclaw.json
GATEWAY_URL="http://localhost:18789"
GATEWAY_API_KEY=""
OPENCLAW_HOME="${OPENCLAW_STATE_DIR:-$HOME/.openclaw}"
if [ -d "$OPENCLAW_HOME" ] && [ -f "$OPENCLAW_HOME/openclaw.json" ]; then
    echo "  检测到 OpenClaw 配置目录: $OPENCLAW_HOME"
    DETECTED=$(python3 -c "
import json, re

config_path = '$OPENCLAW_HOME/openclaw.json'
try:
    with open(config_path, 'r', encoding='utf-8') as f:
        raw = f.read()
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
except Exception:
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
elif [ "$UPGRADE_MODE" = true ]; then
    echo "  ⏭️  升级模式，保留现有 .env"
else
    echo "  ⏭️  .env 已存在，跳过"
fi

# 创建必要目录
mkdir -p "$INSTALL_DIR/data"
mkdir -p "$INSTALL_DIR/backend/data"
mkdir -p "$INSTALL_DIR/backend/keys"

# RSA 公钥
echo ""
echo "[5/6] RSA 公钥..."
if [ ! -f "$INSTALL_DIR/backend/keys/public.pem" ]; then
    echo "  正在从授权服务器获取公钥..."
    PUBKEY_FETCHED=false

    # 方案1: 直接从授权服务器获取
    if curl -sf "$LICENSE_SERVER/api/v1/public-key" -o "$INSTALL_DIR/backend/keys/public.pem" 2>/dev/null; then
        if head -1 "$INSTALL_DIR/backend/keys/public.pem" 2>/dev/null | grep -q "BEGIN"; then
            echo "  ✅ 公钥已从授权服务器获取"
            PUBKEY_FETCHED=true
        else
            rm -f "$INSTALL_DIR/backend/keys/public.pem"
        fi
    fi

    # 方案2: 尝试备用路径
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
    fi
else
    echo "  ✅ 公钥已存在"
fi

# 前端构建
echo ""
echo "[6/6] 前端..."
if [ "$REBUILD_FRONTEND" = true ] || [ ! -f "$INSTALL_DIR/backend/frontend_dist/index.html" ]; then
    if command -v node &>/dev/null; then
        echo "  构建前端..."
        cd "$INSTALL_DIR/frontend"
        npm install --silent 2>/dev/null
        npm run build
        cd "$INSTALL_DIR/backend"
        echo "  ✅ 前端构建完成"
    else
        echo "  ⚠️  未找到 Node.js，跳过前端构建"
        echo "  安装 Node.js 后运行: bash install.sh --rebuild-frontend"
    fi
else
    echo "  ✅ 前端已就绪"
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

# 创建停止脚本 (按端口杀进程，而非进程名)
cat > "$INSTALL_DIR/stop.sh" << STOP_EOF
#!/bin/bash
PID=\$(lsof -ti:\${1:-${BACKEND_PORT}} 2>/dev/null)
if [ -n "\$PID" ]; then
    kill \$PID 2>/dev/null && echo "已停止 ClawMemory (PID: \$PID)" || echo "停止失败"
else
    echo "ClawMemory 未在运行 (端口 ${BACKEND_PORT})"
fi
STOP_EOF
chmod +x "$INSTALL_DIR/stop.sh"

# 健康检查
echo ""
echo "正在进行健康检查..."
sleep 2
if curl -sf "http://localhost:${BACKEND_PORT}/api/v1/health" -o /dev/null 2>/dev/null; then
    echo "  ✅ 后端服务已运行"
else
    echo "  ℹ️  后端服务未启动 (安装后请运行 start.sh)"
fi

# 版本一致性检查
APP_VERSION=$(python3 -c "import app.main; print(getattr(app.main.app, 'version', 'unknown'))" 2>/dev/null || echo "unknown")
echo "  ℹ️  应用版本: $APP_VERSION | 引擎: $CORE_ENGINE"

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
echo "升级:      bash install.sh --upgrade"
echo "Docker:    bash install.sh --docker"
echo "重装前端:  bash install.sh --rebuild-frontend"
echo "开机自启:  sudo bash install_service.sh"
