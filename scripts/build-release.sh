#!/bin/bash
# ============================================
# ClawMemory 发布打包脚本
# 从源码中移除 .pyx 安全核心源文件，确保只分发二进制
# ============================================

set -e

VERSION=${1:-"1.0.0"}
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
    "$PROJECT_DIR/" "$DIST_DIR/"

# 移除安全核心源文件 — 这些功能由 Rust wheel 提供
echo "[2/4] 移除 .pyx 源文件（安全核心）..."
rm -f "$DIST_DIR/backend/app/core/feature_gate.pyx"
rm -f "$DIST_DIR/backend/app/core/license_verifier.pyx"
echo "  ✅ feature_gate.pyx 已移除"
echo "  ✅ license_verifier.pyx 已移除"

# 保留非安全敏感的 .pyx 文件（业务功能模块）
echo "  ⏭️  token_router.pyx 保留（业务功能）"
echo "  ⏭️  memory_decay.pyx 保留（业务功能）"
echo "  ⏭️  conflict_resolver.pyx 保留（业务功能）"

# 移除 Cython 编译脚本（不再需要客户端编译）
rm -f "$DIST_DIR/backend/setup_cython.py"

# 移除 Rust 源码（仅通过预编译 wheel 分发）
echo "[3/4] 移除 Rust 源码（通过预编译 wheel 分发）..."
rm -rf "$DIST_DIR/backend/clawmemory_core/src"
rm -rf "$DIST_DIR/backend/clawmemory_core/.github"
rm -rf "$DIST_DIR/backend/clawmemory_core/target"
rm -f "$DIST_DIR/backend/clawmemory_core/Cargo.toml"
rm -f "$DIST_DIR/backend/clawmemory_core/Cargo.lock"
echo "  ✅ clawmemory_core 源码已移除"

# 只保留 pyproject.toml 用于 maturin（如果需要本地编译）
# 但发布时不应该需要，所以也移除
rm -f "$DIST_DIR/backend/clawmemory_core/pyproject.toml"

# 如果目录为空则移除
rmdir "$DIST_DIR/backend/clawmemory_core" 2>/dev/null || true

# 创建 .gitkeep 保持目录结构
mkdir -p "$DIST_DIR/backend/keys"
touch "$DIST_DIR/backend/keys/.gitkeep"

# 更新安装脚本：移除 Cython 编译步骤
echo "[4/4] 更新安装脚本..."
# 安装脚本已内置下载 Rust wheel 的逻辑，无需修改

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
echo "  4. 上传到 GitHub Releases"
