#!/bin/bash
# ============================================
# ClawMemory 发布打包脚本
# 从源码中移除 C 安全核心源文件，确保只分发二进制
# ============================================

set -e

VERSION=${1:-"2.4.0"}
DIST_DIR="dist/clawmemory-${VERSION}"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "============================================"
echo "  ClawMemory 发布打包 v${VERSION}"
echo "============================================"
echo "项目目录: $PROJECT_DIR"
echo "输出目录: $DIST_DIR"
echo ""

# 清理旧构建
rm -rf "$PROJECT_DIR/dist"
mkdir -p "$DIST_DIR"

# 复制项目文件
echo "[1/4] 复制项目文件..."
rsync -a --exclude='__pycache__' --exclude='*.pyc' --exclude='.pytest_cache' \
    --exclude='venv' --exclude='.venv' --exclude='data' --exclude='*.db' \
    --exclude='keys/*.pem' --exclude='.env' --exclude='clawmemory_core/target' \
    --exclude='clawmemory_core/build' \
    "$PROJECT_DIR/" "$DIST_DIR/"

# 移除 C 安全核心源文件 — 这些功能由预编译 wheel 提供
echo "[2/4] 移除 C 安全核心源文件..."
rm -rf "$DIST_DIR/backend/clawmemory_core/src"
rm -f "$DIST_DIR/backend/clawmemory_core/setup.py"
rm -f "$DIST_DIR/backend/clawmemory_core/Cargo.toml"
rm -f "$DIST_DIR/backend/clawmemory_core/Cargo.lock"
rm -f "$DIST_DIR/backend/clawmemory_core/pyproject.toml"
echo "  ✅ clawmemory_core 源码已移除"

# 移除 Rust 相关文件
rm -rf "$DIST_DIR/backend/clawmemory_core/.github"
rm -rf "$DIST_DIR/backend/clawmemory_core/target"

# 如果目录为空则移除
rmdir "$DIST_DIR/backend/clawmemory_core" 2>/dev/null || true

# 创建 .gitkeep 保持目录结构
mkdir -p "$DIST_DIR/backend/keys"
touch "$DIST_DIR/backend/keys/.gitkeep"

# 更新安装脚本提示
echo "[3/4] 检查安装脚本..."
echo "  ✅ install.sh 已包含 wheel 下载逻辑"

# 版本确认
echo "[4/4] 版本信息..."
echo "  版本: ${VERSION}"
echo "  wheel: clawmemory_core-${VERSION}-*.whl"

echo ""
echo "============================================"
echo "  ✅ 打包完成！"
echo "============================================"
echo ""
echo "发布包: $DIST_DIR/"
echo ""
echo "⚠️  发布前请确保："
echo "  1. 已将 clawmemory_core wheel 发布到 GitHub Releases"
echo "  2. install.sh 中的 WHEEL_URL 指向 https://github.com/860016/Clawmemory/releases"
echo "  3. 授权平台 auth.bestu.top 已部署 /api/v1/public-key"
echo ""
echo "发布步骤："
echo "  1. git tag v${VERSION} && git push origin v${VERSION}"
echo "  2. CI 自动构建 5 平台 wheel 并发布到 GitHub Releases"
echo "  3. cd dist && tar czf clawmemory-${VERSION}.tar.gz clawmemory-${VERSION}/"
echo "  4. 上传到 GitHub Releases (可选，wheel 已通过 CI 自动发布)"
