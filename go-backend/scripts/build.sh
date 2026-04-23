#!/bin/bash
# ClawMemory Go 跨平台编译脚本 (Linux/macOS)
# 用法: ./scripts/build.sh [版本号]

VERSION=${1:-"3.0.0"}
PROJECT_NAME="clawmemory"
BUILD_DIR="../build"

# 创建构建目录
mkdir -p "$BUILD_DIR"

# 版本信息
LDFLAGS="-s -w -X main.Version=$VERSION"

echo "=== ClawMemory 构建系统 ==="
echo "版本: $VERSION"

# 构建函数
build_target() {
    local os=$1
    local arch=$2
    local output=$3
    local tags=$4

    echo ""
    echo "构建 $os/$arch..."

    export GOOS=$os
    export GOARCH=$arch

    # Windows 和 Linux 需要 CGO
    if [ "$os" = "windows" ] || [ "$os" = "linux" ]; then
        export CGO_ENABLED=1
    else
        export CGO_ENABLED=0
    fi

    local tags_flag=""
    if [ -n "$tags" ]; then
        tags_flag="-tags \"$tags\""
    fi

    local output_path="$BUILD_DIR/$output"

    if go build $tags_flag -ldflags "$LDFLAGS" -o "$output_path" ../cmd/server; then
        echo "✓ 构建成功: $output"
    else
        echo "✗ 构建失败: $output"
    fi
}

# 1. Windows AMD64 (带 Pro 功能)
build_target "windows" "amd64" "clawmemory-windows-amd64-pro.exe" "pro"

# 2. Windows AMD64 (开源版)
build_target "windows" "amd64" "clawmemory-windows-amd64.exe" ""

# 3. Linux AMD64 (带 Pro 功能)
build_target "linux" "amd64" "clawmemory-linux-amd64-pro" "pro"

# 4. Linux AMD64 (开源版)
build_target "linux" "amd64" "clawmemory-linux-amd64" ""

# 5. Linux ARM64 (带 Pro 功能)
build_target "linux" "arm64" "clawmemory-linux-arm64-pro" "pro"

# 6. Linux ARM64 (开源版)
build_target "linux" "arm64" "clawmemory-linux-arm64" ""

# 7. macOS AMD64 (带 Pro 功能)
build_target "darwin" "amd64" "clawmemory-darwin-amd64-pro" "pro"

# 8. macOS AMD64 (开源版)
build_target "darwin" "amd64" "clawmemory-darwin-amd64" ""

# 9. macOS ARM64 (M1/M2, 带 Pro 功能)
build_target "darwin" "arm64" "clawmemory-darwin-arm64-pro" "pro"

# 10. macOS ARM64 (M1/M2, 开源版)
build_target "darwin" "arm64" "clawmemory-darwin-arm64" ""

echo ""
echo "=== 构建完成 ==="
echo "输出目录: $(cd "$BUILD_DIR" && pwd)"

# 列出构建结果
ls -lh "$BUILD_DIR"
