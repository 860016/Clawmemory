#!/bin/bash
# ============================================
# ClawMemory Core — 私有仓库初始化
#
# 用法:
#   ./scripts/setup-private-repo.sh
#
# 前提:
#   1. 已安装 gh (GitHub CLI) 或有 GitHub API token
#   2. SSH key 已配置到 GitHub
#
# 步骤:
#   1. 在 GitHub 创建私有仓库 Clawmemory-Core
#   2. 推送 C 源码 + CI workflow
#   3. 配置 PAT token 用于 CI 推送到公开仓库
# ============================================

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
CORE_DIR="$PROJECT_DIR/backend/clawmemory_core"
PRIVATE_REPO="git@github.com:860016/Clawmemory-Core.git"
TEMP_DIR="/tmp/clawmemory-core-repo"

echo "============================================"
echo "  ClawMemory Core 私有仓库初始化"
echo "============================================"
echo ""

# 1. 创建私有仓库
echo "[1/4] 创建私有仓库 Clawmemory-Core..."
if command -v gh &>/dev/null; then
    gh repo create 860016/Clawmemory-Core --private --description "ClawMemory Core Engine (C/CPython) — 商业版核心" 2>/dev/null || echo "  仓库可能已存在"
else
    echo "  ⚠️  gh CLI 未安装，请手动创建:"
    echo "  https://github.com/new?name=Clawmemory-Core&visibility=private"
    echo ""
    read -p "  创建完成后按 Enter 继续..." _
fi

# 2. 准备仓库内容
echo "[2/4] 准备仓库内容..."
rm -rf "$TEMP_DIR"
mkdir -p "$TEMP_DIR"

# 复制 C 源码和构建文件
cp -r "$CORE_DIR/src" "$TEMP_DIR/src"
cp "$CORE_DIR/setup.py" "$TEMP_DIR/"
cp "$CORE_DIR/.github/workflows/build.yml" "$TEMP_DIR/.github/workflows/build.yml" 2>/dev/null || mkdir -p "$TEMP_DIR/.github/workflows" && cp "$CORE_DIR/.github/workflows/build.yml" "$TEMP_DIR/.github/workflows/build.yml"

# 创建 README
cat > "$TEMP_DIR/README.md" << 'EOF'
# ClawMemory Core Engine

商业版核心引擎 (C/CPython)，私有仓库。

## 构建

```bash
# 本地构建
python setup.py bdist_wheel

# 多平台构建 (需要 cibuildwheel + Docker)
pip install cibuildwheel
cibuildwheel --platform linux --output-dir dist .
cibuildwheel --platform windows --output-dir dist .
cibuildwheel --platform macos --output-dir dist .
```

## CI

- Push tag `v*` 触发自动构建 5 平台 wheel
- Wheel 自动发布到公开仓库 [Clawmemory](https://github.com/860016/Clawmemory) 的 Releases

## 安全

- Windows: CNG (CryptoAPI) RSA 验证
- Linux/macOS: OpenSSL PKCS1v15 + SHA256
EOF

# 创建 .gitignore
cat > "$TEMP_DIR/.gitignore" << 'EOF'
__pycache__/
*.pyc
*.so
*.pyd
*.egg-info/
dist/
build/
target/
.env
*.log
EOF

# 3. 初始化 Git 并推送
echo "[3/4] 推送到私有仓库..."
cd "$TEMP_DIR"
git init
git add -A
git commit -m "init: ClawMemory Core v2.4.0 (C/CPython engine)

- C/CPython extension with 20 exported functions
- Windows: CNG (CryptoAPI) RSA verification
- Linux/macOS: OpenSSL PKCS1v15 + SHA256
- cibuildwheel CI for 5 platforms
- Auto-release to public repo"

git branch -M main
git remote add origin "$PRIVATE_REPO"
git push -u origin main

echo ""
echo "[4/4] 配置 CI..."
echo ""
echo "⚠️  需要手动配置 GitHub Secret:"
echo ""
echo "  1. 打开 https://github.com/860016/Clawmemory-Core/settings/secrets/actions"
echo "  2. 创建 Secret: PUBLIC_REPO_PAT"
echo "  3. 值为你的 GitHub Personal Access Token (需要 repo 权限)"
echo "     创建 token: https://github.com/settings/tokens/new?scopes=repo"
echo ""
echo "  这样 CI 才能把 wheel 上传到公开仓库的 Releases"
echo ""
echo "============================================"
echo "  ✅ 私有仓库初始化完成！"
echo "============================================"
echo ""
echo "仓库地址: https://github.com/860016/Clawmemory-Core"
echo ""
echo "发布新版本:"
echo "  cd $TEMP_DIR"
echo "  # 修改版本号..."
echo "  git tag v2.4.0"
echo "  git push origin v2.4.0"
echo "  # CI 自动构建 5 平台 wheel 并发布"
