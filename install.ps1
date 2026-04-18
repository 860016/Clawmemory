# ClawMemory Windows 安装脚本
# 用法: powershell -ExecutionPolicy Bypass -File install.ps1
param(
    [string]$LicenseServer = "https://auth.bestu.top",
    [int]$Port = 8765,
    [switch]$RebuildFrontend
)

$ErrorActionPreference = "Stop"
$InstallDir = Split-Path -Parent $MyInvocation.MyCommand.Path

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "  ClawMemory 安装程序 (Windows)" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host "安装目录: $InstallDir"
Write-Host "后端端口: $Port"
Write-Host "授权服务器: $LicenseServer"
Write-Host ""

# 检查 Python
Write-Host "[1/5] 检查环境..." -ForegroundColor Yellow
try {
    $pyVer = python --version 2>&1
    Write-Host "  Python: $pyVer"
} catch {
    Write-Host "  未找到 Python，请先安装 Python 3.10+" -ForegroundColor Red
    Write-Host "  下载: https://www.python.org/downloads/" -ForegroundColor Red
    exit 1
}

# 检查前端
if (Test-Path "$InstallDir\backend\frontend_dist\index.html") {
    Write-Host "  前端: 已预构建" -ForegroundColor Green
} else {
    Write-Host "  前端: 未构建" -ForegroundColor Yellow
}

# 创建虚拟环境
Write-Host ""
Write-Host "[2/5] 创建 Python 虚拟环境并安装依赖..." -ForegroundColor Yellow
Set-Location "$InstallDir\backend"

# 优先使用 venv，失败则用 virtualenv
$venvCreated = $false
try {
    python -m venv venv 2>$null
    if ($LASTEXITCODE -eq 0) {
        Write-Host "  虚拟环境已创建 (venv)" -ForegroundColor Green
        $venvCreated = $true
    }
} catch {}

if (-not $venvCreated) {
    Write-Host "  venv 不可用，尝试 virtualenv..." -ForegroundColor Yellow
    pip install virtualenv -q 2>$null
    try {
        python -m virtualenv venv 2>$null
        if ($LASTEXITCODE -eq 0) {
            Write-Host "  虚拟环境已创建 (virtualenv)" -ForegroundColor Green
            $venvCreated = $true
        }
    } catch {}
}

if (-not $venvCreated) {
    Write-Host "  无法创建虚拟环境，请安装 python3-venv 或 virtualenv" -ForegroundColor Red
    exit 1
}

& "$InstallDir\backend\venv\Scripts\pip.exe" install --upgrade pip -q

& "$InstallDir\backend\venv\Scripts\pip.exe" install -r requirements.txt -q
if ($LASTEXITCODE -ne 0) {
    Write-Host "  bcrypt 安装失败，尝试降级..." -ForegroundColor Yellow
    $content = Get-Content requirements.txt | Where-Object { $_ -notmatch '^bcrypt' }
    $content | Set-Content requirements.txt
    Add-Content requirements.txt "passlib[bcrypt]>=1.7.4"
    & "$InstallDir\backend\venv\Scripts\pip.exe" install -r requirements.txt -q
}
Write-Host "  依赖安装完成" -ForegroundColor Green

# 安装安全引擎：Rust wheel → 纯 Python 兜底
# 注意：.pyx 源文件不在发布包中，核心安全逻辑由 Rust wheel 提供
Write-Host ""
Write-Host "  安装安全引擎..." -ForegroundColor Yellow
$CoreEngine = "python"

# Detect architecture
$Arch = & "$InstallDir\backend\venv\Scripts\python.exe" -c "import platform; print(platform.machine().lower())"
$PyVer = & "$InstallDir\backend\venv\Scripts\python.exe" -c "import sys; print(f'cp{sys.version_info.major}{sys.version_info.minor}')"

# Map Python arch to wheel platform tag
$PlatformTag = switch ($Arch) {
    "amd64" { "win_amd64" }
    "x86_64" { "win_amd64" }
    "arm64" { "win_arm64" }
    "aarch64" { "win_arm64" }
    default { "win_amd64" }
}

$WheelUrl = "https://github.com/860016/Clawmemory/releases/latest/download/clawmemory_core-2.0.0-$PyVer-$PyVer-${PlatformTag}.whl"

try {
    Write-Host "  尝试下载 Rust 安全引擎 ($PlatformTag)..." -ForegroundColor Yellow
    $wheelPath = "$env:TEMP\clawmemory_core.whl"
    Invoke-WebRequest -Uri $WheelUrl -OutFile $wheelPath -TimeoutSec 30 -ErrorAction Stop
    & "$InstallDir\backend\venv\Scripts\pip.exe" install $wheelPath -q 2>$null
    if ($LASTEXITCODE -eq 0) {
        Write-Host "  Rust 安全引擎已安装 (预编译wheel: $PlatformTag)" -ForegroundColor Green
        $CoreEngine = "rust"
    } else {
        Write-Host "  wheel 下载成功但安装失败（架构/版本不匹配）" -ForegroundColor Yellow
    }
    Remove-Item $wheelPath -ErrorAction SilentlyContinue
} catch {
    Write-Host "  未找到匹配的预编译 wheel ($PlatformTag)" -ForegroundColor Yellow
    if ($Arch -in @("arm64", "aarch64")) {
        Write-Host "  ARM64 预编译 wheel 暂不可用，将使用纯 Python 模式" -ForegroundColor Yellow
    }
}

# 方案2: 本地编译 Rust (仅开发者使用，需要 rustc 和 clawmemory_core/src 目录)
if ($CoreEngine -eq "python" -and (Test-Path "$InstallDir\backend\clawmemory_core\src") -and (Get-Command rustc -ErrorAction SilentlyContinue)) {
    Write-Host "  尝试本地编译 Rust 引擎..." -ForegroundColor Yellow
    & "$InstallDir\backend\venv\Scripts\pip.exe" install maturin -q 2>$null
    Set-Location "$InstallDir\backend\clawmemory_core"
    maturin develop --release 2>&1 | Write-Host
    if ($LASTEXITCODE -eq 0) {
        Write-Host "  Rust 安全引擎已编译安装" -ForegroundColor Green
        $CoreEngine = "rust"
    }
}

# 方案3: Cython 编译 (.pyx → .pyd，需要 C 编译器或 Visual Studio Build Tools)
if ($CoreEngine -eq "python" -and (Test-Path "$InstallDir\backend\setup_cython.py")) {
    $hasCompiler = $false
    try { $null = Get-Command cl -ErrorAction Stop; $hasCompiler = $true } catch {}
    if (-not $hasCompiler) {
        try { $null = Get-Command gcc -ErrorAction Stop; $hasCompiler = $true } catch {}
    }
    if ($hasCompiler) {
        Write-Host "  尝试 Cython 编译安全引擎..." -ForegroundColor Yellow
        & "$InstallDir\backend\venv\Scripts\pip.exe" install cython -q 2>$null
        Set-Location "$InstallDir\backend"
        & "$InstallDir\backend\venv\Scripts\python.exe" setup_cython.py build_ext --inplace 2>&1 | Write-Host
        $pydFiles = Get-ChildItem "$InstallDir\backend\app\core" -Filter "*.pyd" -ErrorAction SilentlyContinue
        if ($pydFiles.Count -gt 0) {
            Write-Host "  Cython 安全引擎已编译 (中等安全)" -ForegroundColor Green
            $CoreEngine = "cython"
        } else {
            Write-Host "  Cython 编译完成但未生成 .pyd 文件" -ForegroundColor Yellow
        }
    }
}

if ($CoreEngine -eq "python") {
    Write-Host "  使用纯 Python 模式（安全性较低）" -ForegroundColor Yellow
    Write-Host "  建议安装 clawmemory-core Rust wheel 以获得 RSA 硬验证保护" -ForegroundColor Yellow
    Write-Host "  安装 Rust (https://rustup.rs) 后可本地编译获得更强保护" -ForegroundColor Yellow
}

# 配置环境变量
Write-Host ""
Write-Host "[3/5] 配置环境变量..." -ForegroundColor Yellow

# Auto-detect OpenClaw Gateway from ~/.openclaw/openclaw.json
$GatewayUrl = "http://localhost:18789"
$GatewayApiKey = ""
$OpenClawHome = if ($env:OPENCLAW_STATE_DIR) { $env:OPENCLAW_STATE_DIR } else { "$env:USERPROFILE\.openclaw" }
if ((Test-Path $OpenClawHome) -and (Test-Path "$OpenClawHome\openclaw.json")) {
    Write-Host "  检测到 OpenClaw 配置目录: $OpenClawHome" -ForegroundColor Cyan
    # Parse openclaw.json (JSON5 format) to extract gateway port + auth token
    $detected = & python -c "
import json, re, sys

config_path = r'$OpenClawHome\openclaw.json'
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
except Exception:
    print('|')
" 2>$null
    if ($detected -and $detected.Contains("|")) {
        $parts = $detected -split "\|", 2
        if ($parts[0]) {
            $GatewayUrl = $parts[0]
            Write-Host "  自动检测到 Gateway URL: $GatewayUrl" -ForegroundColor Green
        }
        if ($parts[1]) {
            $GatewayApiKey = $parts[1]
            Write-Host "  自动检测到 Gateway API Key" -ForegroundColor Green
        }
    }
}

if (-not (Test-Path "$InstallDir\backend\.env")) {
    $secretKey = -join ((48..57) + (65..90) + (97..122) | Get-Random -Count 32 | ForEach-Object { [char]$_ })
    
    @"
OPENCLAW_SECRET_KEY=$secretKey
OPENCLAW_DATA_DIR=$InstallDir\data
OPENCLAW_DEBUG=false
OPENCLAW_PORT=$Port
OPENCLAW_LICENSE_SERVER_URL=$LicenseServer
OPENCLAW_RSA_PUBLIC_KEY_PATH=./keys/public.pem
OPENCLAW_CORS_ORIGINS=["*"]
OPENCLAW_OPENCLAW_GATEWAY_URL=$GatewayUrl
OPENCLAW_OPENCLAW_API_KEY=$GatewayApiKey
"@ | Set-Content "$InstallDir\backend\.env" -Encoding UTF8
    Write-Host "  .env 已创建" -ForegroundColor Green
} else {
    Write-Host "  .env 已存在，跳过"
}

# 创建目录
New-Item -ItemType Directory -Force -Path "$InstallDir\data" | Out-Null
New-Item -ItemType Directory -Force -Path "$InstallDir\backend\data" | Out-Null
New-Item -ItemType Directory -Force -Path "$InstallDir\backend\keys" | Out-Null

# RSA 公钥
Write-Host ""
Write-Host "[4/5] RSA 公钥..." -ForegroundColor Yellow
if (-not (Test-Path "$InstallDir\backend\keys\public.pem")) {
    $pubKeyFetched = $false

    # 方案1: 直接从授权服务器获取
    try {
        Write-Host "  正在从授权服务器获取公钥..." -ForegroundColor Cyan
        $pubKey = Invoke-RestMethod -Uri "$LicenseServer/api/v1/public-key" -TimeoutSec 10 -ErrorAction Stop
        if ($pubKey -and $pubKey.ToString().Contains("BEGIN")) {
            $pubKey | Set-Content "$InstallDir\backend\keys\public.pem" -Encoding UTF8
            Write-Host "  公钥已从授权服务器获取" -ForegroundColor Green
            $pubKeyFetched = $true
        }
    } catch {}

    # 方案2: 尝试备用路径
    if (-not $pubKeyFetched) {
        try {
            $pubKey = Invoke-RestMethod -Uri "$LicenseServer/api/v1/public-key/pem" -TimeoutSec 10 -ErrorAction Stop
            if ($pubKey -and $pubKey.ToString().Contains("BEGIN")) {
                $pubKey | Set-Content "$InstallDir\backend\keys\public.pem" -Encoding UTF8
                Write-Host "  公钥已从备用路径获取" -ForegroundColor Green
                $pubKeyFetched = $true
            }
        } catch {}
    }

    # 方案3: 尝试从 GitHub Releases 下载
    if (-not $pubKeyFetched) {
        try {
            $pubKeyPath = "$env:TEMP\clawmemory_public.pem"
            Invoke-WebRequest -Uri "https://github.com/860016/Clawmemory/releases/latest/download/public.pem" -OutFile $pubKeyPath -TimeoutSec 15 -ErrorAction Stop
            $content = Get-Content $pubKeyPath -Raw
            if ($content -and $content.Contains("BEGIN")) {
                Copy-Item $pubKeyPath "$InstallDir\backend\keys\public.pem" -Force
                Write-Host "  公钥已从 GitHub Releases 获取" -ForegroundColor Green
                $pubKeyFetched = $true
            }
            Remove-Item $pubKeyPath -ErrorAction SilentlyContinue
        } catch {}
    }

    if (-not $pubKeyFetched) {
        Write-Host "  无法自动获取公钥（授权服务器可能不可达）" -ForegroundColor Yellow
        Write-Host "  请手动将公钥文件放到: $InstallDir\backend\keys\public.pem" -ForegroundColor Yellow
        Write-Host "  或在启动后通过管理界面上传公钥" -ForegroundColor Yellow
    }
} else {
    Write-Host "  公钥已存在" -ForegroundColor Green
}

# 前端
Write-Host ""
if ($RebuildFrontend) {
    Write-Host "[5/5] 重新构建前端..." -ForegroundColor Yellow
    Set-Location "$InstallDir\frontend"
    npm install --silent 2>$null
    npm run build
    Write-Host "  前端构建完成" -ForegroundColor Green
} else {
    Write-Host "[5/5] 前端: 使用预构建版本" -ForegroundColor Yellow
    if (Test-Path "$InstallDir\backend\frontend_dist\index.html") {
        Write-Host "  前端已就绪" -ForegroundColor Green
    } else {
        Write-Host "  前端未构建，如需 Web 界面请运行: install.ps1 -RebuildFrontend" -ForegroundColor Yellow
    }
}

# 初始化数据库
Set-Location "$InstallDir\backend"
& "$InstallDir\backend\venv\Scripts\python.exe" -c "from app.database import init_db; init_db(); print('  数据库初始化完成')"

# 创建启动脚本
@"
@echo off
cd /d "$InstallDir\backend"
call venv\Scripts\activate.bat
uvicorn app.main:app --host 0.0.0.0 --port $Port
"@ | Set-Content "$InstallDir\start.bat" -Encoding ASCII

@"
@echo off
taskkill /f /im uvicorn.exe 2>nul
echo ClawMemory 已停止
"@ | Set-Content "$InstallDir\stop.bat" -Encoding ASCII

Write-Host ""
Write-Host "============================================" -ForegroundColor Green
Write-Host "  安装完成!" -ForegroundColor Green
Write-Host "============================================" -ForegroundColor Green
Write-Host ""
Write-Host "启动服务:  $InstallDir\start.bat" -ForegroundColor White
Write-Host "停止服务:  $InstallDir\stop.bat" -ForegroundColor White
Write-Host "访问地址:  http://localhost:$Port" -ForegroundColor White
Write-Host "检查状态:  curl http://localhost:$Port/api/v1/install-status" -ForegroundColor White
