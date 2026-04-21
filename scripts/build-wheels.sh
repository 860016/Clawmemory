#!/bin/bash
# ============================================
# ClawMemory Core — 本地多平台 wheel 构建
#
# 使用 cibuildwheel + Docker 构建：
#   - Linux x86_64 / aarch64 (通过 Docker/QEMU)
#   - macOS x86_64 / arm64 (需要 macOS 主机)
#   - Windows x86_64 (需要 Windows 主机 + MSVC)
#
# 用法:
#   ./scripts/build-wheels.sh          # 构建当前平台
#   ./scripts/build-wheels.sh linux    # 构建 Linux (需要 Docker)
#   ./scripts/build-wheels.sh all      # 构建所有平台
# ============================================

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
CORE_DIR="$PROJECT_DIR/backend/clawmemory_core"
VERSION=$(python3 -c "import re; m=re.search(r'version=[\"']([^\"']+)[\"']', open('$CORE_DIR/setup.py').read()); print(m.group(1))" 2>/dev/null || echo "2.4.0")

echo "============================================"
echo "  ClawMemory Core Wheel Builder v${VERSION}"
echo "============================================"
echo "源码目录: $CORE_DIR"
echo ""

# 安装依赖
pip install --quiet cibuildwheel setuptools wheel 2>/dev/null || pip3 install --quiet cibuildwheel setuptools wheel 2>/dev/null

PLATFORM=${1:-"current"}

build_current() {
    echo "[1/1] 构建当前平台 wheel..."
    cd "$CORE_DIR"
    python setup.py bdist_wheel
    echo ""
    echo "✅ 当前平台 wheel:"
    ls -la dist/*.whl 2>/dev/null || echo "  (未生成 wheel)"
}

build_linux() {
    echo ""
    echo "========== Linux x86_64 =========="
    export CIBW_ARCHS=x86_64
    export CIBW_BUILD="cp310-* cp311-* cp312-* cp313-*"
    # CIBW_BEFORE_BUILD_LINUX 只跑在 manylinux, CIBW_BEFORE_BUILD 是所有平台的 fallback（包括 musllinux）
    export CIBW_BEFORE_BUILD_LINUX="pip install setuptools wheel && yum install -y openssl openssl-devel || true"
    export CIBW_BEFORE_BUILD="pip install setuptools wheel && (yum install -y openssl openssl-devel 2>/dev/null || apk add openssl-dev 2>/dev/null || true)"
    export CIBW_ENVIRONMENT_LINUX='CFLAGS="-O2"'
    cibuildwheel --platform linux --output-dir dist "$CORE_DIR"

    echo ""
    echo "========== Linux aarch64 =========="
    export CIBW_ARCHS=aarch64
    cibuildwheel --platform linux --output-dir dist "$CORE_DIR"
}

build_macos() {
    echo ""
    echo "========== macOS =========="
    export CIBW_ARCHS="x86_64 arm64"
    export CIBW_BUILD="cp310-* cp311-* cp312-* cp313-*"
    export CIBW_BEFORE_BUILD_MACOS="pip install setuptools wheel"
    export CIBW_ENVIRONMENT_MACOS='CFLAGS="-O2"'
    cibuildwheel --platform macos --output-dir dist "$CORE_DIR"
}

build_windows() {
    echo ""
    echo "========== Windows =========="
    export CIBW_ARCHS=AMD64
    export CIBW_BUILD="cp310-* cp311-* cp312-* cp313-*"
    export CIBW_BEFORE_BUILD_WINDOWS="pip install setuptools wheel"
    export CIBW_ENVIRONMENT_WINDOWS='CFLAGS="/O2"'
    cibuildwheel --platform windows --output-dir dist "$CORE_DIR"
}

case "$PLATFORM" in
    current)
        build_current
        ;;
    linux)
        build_linux
        ;;
    macos)
        build_macos
        ;;
    windows)
        build_windows
        ;;
    all)
        build_linux
        build_macos
        build_windows
        ;;
    *)
        echo "用法: $0 [current|linux|macos|windows|all]"
        exit 1
        ;;
esac

echo ""
echo "============================================"
echo "  ✅ 构建完成！"
echo "============================================"
echo ""
echo "Wheel 文件:"
ls -la "$CORE_DIR/dist/"*.whl 2>/dev/null || ls -la dist/*.whl 2>/dev/null || echo "  (未找到 wheel)"
echo ""
echo "发布步骤:"
echo "  1. 在 GitHub 创建 Release: https://github.com/860016/Clawmemory/releases/new"
echo "  2. 上传所有 .whl 文件"
echo "  3. install.sh 会自动下载匹配的 wheel"
