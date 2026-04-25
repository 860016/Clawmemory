#!/bin/bash
# ClawMemory Linux/macOS 安装脚本 v3.1 (完整版)
# 用法: bash install.sh [选项]
# 特性: 自动检测依赖、构建前端、编译后端、创建统一目录、生成启动脚本

set -e

LICENSE_SERVER="https://auth.bestu.top"
PORT=8765
UPGRADE=false
INSTALL_PATH=""
AUTO_START=false
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

for arg in "$@"; do
  case $arg in
    --license-server=*) LICENSE_SERVER="${arg#*=}" ;;
    --port=*) PORT="${arg#*=}" ;;
    --upgrade) UPGRADE=true ;;
    --install-path=*) INSTALL_PATH="${arg#*=}" ;;
    --auto-start) AUTO_START=true ;;
    -h|--help)
      echo "╔══════════════════════════════════════════╗"
      echo "║   ClawMemory 安装程序 v3.1 (Linux/macOS) ║"
      echo "╚══════════════════════════════════════════╝"
      echo ""
      echo "用法: bash install.sh [选项]"
      echo ""
      echo "选项:"
      echo "  --license-server=URL   授权服务器地址"
      echo "                       默认: https://auth.bestu.top"
      echo "  --port=PORT           后端端口 (默认: 8765)"
      echo "  --upgrade             升级模式 (保留配置和数据)"
      echo "  --install-path=PATH   自定义安装路径"
      echo "  --auto-start          安装完成后自动启动服务"
      echo "  -h, --help            显示帮助信息"
      echo ""
      echo "示例:"
      echo "  bash install.sh                              # 默认安装到当前目录"
      echo "  bash install.sh --install-path=/opt/clawmemory  # 自定义路径"
      echo "  bash install.sh --upgrade                    # 升级模式"
      echo "  bash install.sh --auto-start                 # 安装后启动"
      exit 0
      ;;
    *) echo "未知参数: $arg"; exit 1 ;;
  esac
done

if [ "$INSTALL_PATH" = "" ]; then
    INSTALL_DIR="$SCRIPT_DIR"
else
    INSTALL_DIR="$INSTALL_PATH"
    mkdir -p "$INSTALL_DIR" 2>/dev/null || {
        echo "✗ 无法创建目录: $INSTALL_PATH"
        exit 1
    }
fi

DATA_DIR="$INSTALL_DIR/data"
SKILLS_DIR="$DATA_DIR/skills"
BACKUPS_DIR="$DATA_DIR/backups"
ENV_FILE="$INSTALL_DIR/go-backend/.env"
EXE_FILE="$INSTALL_DIR/go-backend/clawmemory"

TOTAL_STEPS=6

print_step() {
    echo ""
    echo "[$1/$TOTAL_STEPS] $2" | awk '{printf "\033[33m%s\033[0m\n", $0}'
}

print_success() {
    echo "  ✓ $1" | awk '{printf "\033[32m%s\033[0m\n", $0}'
}

print_error() {
    echo "  ✗ $1" | awk '{printf "\033[31m%s\033[0m\n", $0}'
}

print_info() {
    echo "  ℹ $1" | awk '{printf "\033[36m%s\033[0m\n", $0}'
}

echo ""
echo "╔══════════════════════════════════════════╗"
echo "║   ClawMemory 安装程序 v3.1 (Linux/macOS) ║"
echo "╚══════════════════════════════════════════╝"
echo ""
print_info "安装目录: $INSTALL_DIR"
print_info "数据目录: $DATA_DIR"
print_info "技能目录: $SKILLS_DIR"
print_info "备份目录: $BACKUPS_DIR"
print_info "服务端口: $PORT"
print_info "授权服务器: $LICENSE_SERVER"
[ "$UPGRADE" = true ] && print_info "模式: 升级 (保留配置和数据)"
[ "$AUTO_START" = true ] && print_info "自动启动: 是"

print_step 1 "检查环境依赖..."
echo ""

HAS_GO=false
HAS_NODE=false
HAS_GIT=false

if command -v go &>/dev/null; then
    HAS_GO=true
    print_success "Go 环境: $(go version)"
else
    print_error "Go 未安装"
    print_info "下载地址: https://go.dev/dl/"
    print_info "请安装 Go 1.21+ 后重试"
    exit 1
fi

if command -v node &>/dev/null; then
    HAS_NODE=true
    print_success "Node.js: $(node --version)"
else
    print_info "Node.js 未安装 (可选，用于构建前端)"
fi

if command -v git &>/dev/null; then
    HAS_GIT=true
    print_success "Git: $(git --version)"
else
    print_info "Git 未安装 (可选，用于技能安装)"
fi

print_step 2 "构建前端..."
echo ""

FRONTEND_READY=false
if [ -f "$INSTALL_DIR/go-backend/frontend_dist/index.html" ]; then
    FRONTEND_READY=true
fi

if [ ! "$FRONTEND_READY" = true ] || [ "$1" = "--rebuild" ]; then
    if [ "$HAS_NODE" = true ]; then
        print_info "开始构建前端..."
        cd "$INSTALL_DIR/frontend"
        
        npm install --prefer-offline --no-audit --no-fund 2>/dev/null || true
        print_success "npm install 完成"
        
        if npm run build 2>&1; then
            print_success "npm run build 完成"
        else
            print_error "前端构建失败，将使用预构建版本（如果存在）"
        fi
        
        if [ -d "$INSTALL_DIR/frontend/dist" ]; then
            rm -rf "$INSTALL_DIR/go-backend/frontend_dist" 2>/dev/null || true
            cp -r "$INSTALL_DIR/frontend/dist" "$INSTALL_DIR/go-backend/frontend_dist"
            print_success "前端已复制到 go-backend/frontend_dist"
        fi
        
        cd "$INSTALL_DIR"
    else
        if [ "$FRONTEND_READY" = true ]; then
            print_success "使用预构建的前端文件"
        else
            print_warning "未找到 Node.js 且无预构建前端"
            print_info "前端功能可能不可用"
        fi
    fi
else
    print_success "前端已就绪 (跳过构建)"
fi

print_step 3 "编译 Go 后端..."
echo ""

cd "$INSTALL_DIR/go-backend"

if go build -o clawmemory ./cmd/server 2>&1; then
    if [ -f "$EXE_FILE" ]; then
        EXE_SIZE=$(du -h "$EXE_FILE" | cut -f1)
        print_success "Go 后端编译成功 ($EXE_SIZE)"
    else
        print_success "Go 后端编译成功"
    fi
else
    print_error "Go 后端编译失败"
    exit 1
fi

cd "$INSTALL_DIR"

print_step 4 "配置环境..."
echo ""

if [ ! -f "$ENV_FILE" ]; then
    SECRET_KEY=$(openssl rand -hex 32 2>/dev/null || cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1)
    
    cat > "$ENV_FILE" << EOF
HOST=0.0.0.0
SECRET_KEY=$SECRET_KEY
PORT=$PORT
DATA_DIR=$DATA_DIR
SKILLS_DIR=$SKILLS_DIR
BACKUPS_DIR=$BACKUPS_DIR
LICENSE_SERVER_URL=$LICENSE_SERVER
EOF
    
    print_success ".env 配置文件已创建"
    print_info "SECRET_KEY 已自动生成"
elif [ "$UPGRADE" = true ]; then
    print_success "升级模式，保留现有 .env 配置"
else
    print_success ".env 配置文件已存在"
fi

echo ""
print_info "创建统一目录结构..."

create_dir() {
    mkdir -p "$1" 2>/dev/null
    print_success "$2: $1"
}

create_dir "$DATA_DIR" "数据目录"
create_dir "$SKILLS_DIR" "技能目录"
create_dir "$BACKUPS_DIR" "备份目录"
create_dir "$DATA_DIR/keys" "密钥目录"
create_dir "$DATA_DIR/uploads" "上传目录"

print_step 5 "生成启动脚本..."
echo ""

cat > "$INSTALL_DIR/start.sh" << EOF
#!/bin/bash
cd "\$(dirname "\$0")/go-backend"
echo "============================================"
echo "  Starting ClawMemory..."
echo "  Port: $PORT"
echo "  Data: $DATA_DIR"
echo "============================================"
./clawmemory
EOF
chmod +x "$INSTALL_DIR/start.sh"

cat > "$INSTALL_DIR/stop.sh" << EOF
#!/bin/bash
echo "Stopping ClawMemory on port \$PORT..."
if pkill -f "clawmemory" 2>/dev/null; then
    echo "✓ ClawMemory stopped successfully"
else
    echo "ClawMemory is not running"
fi
EOF
chmod +x "$INSTALL_DIR/stop.sh"

print_success "start.sh - 启动脚本已生成"
print_success "stop.sh  - 停止脚本已生成"

print_step 6 "验证安装..."
echo ""

check_file() {
    if [ -f "$1" ]; then
        print_success "$2: ✓"
        return 0
    else
        if [ "$3" = "required" ]; then
            print_error "$2: ✗ 缺失!"
            return 1
        else
            print_info "$2: ⚠ 不存在 (可选)"
            return 0
        fi
    fi
}

check_dir() {
    if [ -d "$1" ]; then
        print_success "$2: ✓"
        return 0
    else
        if [ "$3" = "required" ]; then
            print_error "$2: ✗ 缺失!"
            return 1
        else
            print_info "$2: ⚠ 不存在 (可选)"
            return 0
        fi
    fi
}

ALL_PASSED=true

check_file "$EXE_FILE" "可执行文件" "required" || ALL_PASSED=false
check_file "$INSTALL_DIR/go-backend/frontend_dist/index.html" "前端文件" "required" || ALL_PASSED=false
check_file "$ENV_FILE" "配置文件" "required" || ALL_PASSED=false
check_file "$INSTALL_DIR/start.sh" "启动脚本" "required" || ALL_PASSED=false
check_file "$INSTALL_DIR/stop.sh" "停止脚本" "required" || ALL_PASSED=false
check_dir "$DATA_DIR" "数据目录" "required" || ALL_PASSED=false
check_dir "$SKILLS_DIR" "技能目录" "optional"
check_dir "$BACKUPS_DIR" "备份目录" "optional"

echo ""

if [ "$ALL_PASSED" = true ]; then
    echo "╔══════════════════════════════════════════╗" | awk '{printf "\033[32m%s\033[0m\n", $0}'
    echo "║           ✅ 安装完成!                    ║" | awk '{printf "\033[32m%s\033[0m\n", $0}'
    echo "╚══════════════════════════════════════════╝" | awk '{printf "\033[32m%s\033[0m\n", $0}'
else
    echo "⚠️  安装完成但有警告，请检查上述缺失项" | awk '{printf "\033[33m%s\033[0m\n", $0}'
fi

echo ""
echo "📁 目录结构:" | awk '{printf "\033[36m%s\033[0m\n", $0}'
echo "   $INSTALL_DIR/"
echo "   ├── go-backend/          # 后端程序"
echo "   │   ├── clawmemory       # 可执行文件"
echo "   │   └── .env             # 配置文件"
echo "   ├── data/                # 数据目录 (统一)"
echo "   │   ├── skills/          # 技能插件"
echo "   │   ├── backups/         # 备份文件"
echo "   │   └── keys/            # 密钥文件"
echo "   ├── start.sh             # 启动脚本"
echo "   └── stop.sh              # 停止脚本"
echo ""
echo "🚀 快速开始:" | awk '{printf "\033[36m%s\033[0m\n", $0}'
echo "   1. 运行 ./start.sh 启动服务"
echo "   2. 打开浏览器访问 http://localhost:$PORT"
echo "   3. 首次访问需设置管理员密码"
echo ""
echo "💡 提示:" | awk '{printf "\033[33m%s\033[0m\n", $0}'
echo "   • 所有数据统一存储在 data/ 目录下"
echo "   • 技能安装自动保存到 data/skills/"
echo "   • 使用 ./stop.sh 可安全停止服务"
echo "   • 升级时使用 --upgrade 参数保留配置"

if [ "$INSTALL_PATH" != "" ]; then
    echo ""
    print_info "自定义安装路径: $INSTALL_PATH"
fi

if [ "$AUTO_START" = true ] && [ "$ALL_PASSED" = true ]; then
    echo ""
    print_info "正在启动服务..."
    "$INSTALL_DIR/start.sh" &
    sleep 2
    print_success "服务已启动!"
    print_info "请在浏览器中打开: http://localhost:$PORT"
fi

echo ""
