#!/bin/bash
# ClawMemory 全平台编译脚本
# 支持: Windows/Linux/macOS × x86/ARM

set -e

VERSION=${1:-"1.0.0"}
OUTPUT_DIR="releases/v${VERSION}"

echo "=========================================="
echo "ClawMemory 全平台编译"
echo "版本: ${VERSION}"
echo "=========================================="

# 创建输出目录
mkdir -p "${OUTPUT_DIR}"

# 编译函数
build_target() {
    local os=$1
    local arch=$2
    local output_name=$3
    local tags=$4
    
    echo ""
    echo "编译: ${os}/${arch} ..."
    
    env GOOS=${os} GOARCH=${arch} CGO_ENABLED=0 \
        go build -tags "${tags}" -ldflags "-s -w -X main.Version=${VERSION}" \
        -o "${OUTPUT_DIR}/${output_name}" ./cmd/server
    
    if [ $? -eq 0 ]; then
        echo "✅ 成功: ${output_name}"
        
        # 压缩
        if [ "${os}" = "windows" ]; then
            zip -j "${OUTPUT_DIR}/${output_name%.exe}.zip" "${OUTPUT_DIR}/${output_name}" >/dev/null 2>&1
        else
            tar -czf "${OUTPUT_DIR}/${output_name}.tar.gz" -C "${OUTPUT_DIR}" "${output_name}" >/dev/null 2>&1
        fi
    else
        echo "❌ 失败: ${output_name}"
    fi
}

# ========== 开源版 ==========
echo ""
echo "📦 编译开源版..."

# Windows
build_target "windows" "amd64" "clawmemory-windows-amd64.exe" ""
build_target "windows" "arm64" "clawmemory-windows-arm64.exe" ""

# Linux
build_target "linux" "amd64" "clawmemory-linux-amd64" ""
build_target "linux" "arm64" "clawmemory-linux-arm64" ""
build_target "linux" "386" "clawmemory-linux-386" ""

# macOS
build_target "darwin" "amd64" "clawmemory-darwin-amd64" ""
build_target "darwin" "arm64" "clawmemory-darwin-arm64" ""

# ========== 生成校验和 ==========
echo ""
echo "🔐 生成校验和..."
cd "${OUTPUT_DIR}"
sha256sum * > checksums.txt
cd ../..

# ========== 生成发布说明 ==========
echo ""
echo "📝 生成发布说明..."
cat > "${OUTPUT_DIR}/README.txt" << EOF
ClawMemory v${VERSION}
===================

## 系统要求
- Windows 10+ (x64/ARM64)
- Linux (x64/ARM64/x86)
- macOS 11+ (Intel/Apple Silicon)

## 文件说明
- clawmemory-*: 开源版（免费），Pro 功能通过授权激活使用

## 快速开始
1. 下载对应平台的文件
2. 直接运行（无需安装）
3. 打开浏览器访问 http://localhost:8765

## 校验
sha256sum -c checksums.txt
EOF

# ========== 完成 ==========
echo ""
echo "=========================================="
echo "编译完成!"
echo "输出目录: ${OUTPUT_DIR}"
echo "=========================================="
echo ""
ls -lh "${OUTPUT_DIR}"
