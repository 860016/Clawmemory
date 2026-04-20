# ClawMemory Windows 安装脚本
# 用法: powershell -ExecutionPolicy Bypass -File install.ps1
param(
    [string]$LicenseServer = "https://auth.bestu.top",
    [int]$Port = 8765,
    [switch]$RebuildFrontend,
    [switch]$Upgrade
)

$ErrorActionPreference = "Stop"
$InstallDir = Split-Path -Parent $MyInvocation.MyCommand.Path

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "  ClawMemory 安装程序 v2.8.2 (Windows)" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host "安装目录: $InstallDir"
Write-Host "后端端口: $Port"
Write-Host "授权服务器: $LicenseServer"
if ($Upgrade) { Write-Host "模式: 升级 (保留配置和数据)" -ForegroundColor Yellow }
Write-Host ""

# 检查 Python
Write-Host "[1/6] 检查环境..." -ForegroundColor Yellow
try {
    $pyVer = python --version 2>&1
    Write-Host "  Python: $pyVer"
    $pyVerNum = ($pyVer -replace 'Python ', '' -split '\.')[0..1] -join '.'
    if ([double]$pyVerNum -lt 3.10) {
        Write-Host "  需要 Python 3.10+，当前版本太低" -ForegroundColor Red
        exit 1
    }
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
    if (Get-Command node -ErrorAction SilentlyContinue) {
        Write-Host "  检测到 Node.js，将自动构建前端" -ForegroundColor Cyan
        $RebuildFrontend = $true
    }
}

# 创建/复用虚拟环境
Write-Host ""
Write-Host "[2/6] 创建 Python 虚拟环境并安装依赖..." -ForegroundColor Yellow
Set-Location "$InstallDir\backend"

if ((Test-Path "$InstallDir\backend\venv\Scripts\python.exe") -and (Test-Path "$InstallDir\backend\venv\Scripts\pip.exe")) {
    Write-Host "  虚拟环境已存在，复用中" -ForegroundColor Green
} else {
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
}

& "$InstallDir\backend\venv\Scripts\pip.exe" install --upgrade pip -q

# 安装依赖（不修改 requirements.txt，降级使用临时文件方式）
$reqResult = & "$InstallDir\backend\venv\Scripts\pip.exe" install -r requirements.txt -q 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "  bcrypt 安装失败，尝试降级..." -ForegroundColor Yellow
    $tmpReq = "$env:TEMP\clawmemory_requirements.txt"
    Get-Content requirements.txt | Where-Object { $_ -notmatch '^bcrypt' } | Set-Content $tmpReq -Encoding UTF8
    & "$InstallDir\backend\venv\Scripts\pip.exe" install -r $tmpReq -q 2>$null
    & "$InstallDir\backend\venv\Scripts\pip.exe" install "passlib[bcrypt]>=1.7.4" -q 2>$null
    Remove-Item $tmpReq -ErrorAction SilentlyContinue
}
Write-Host "  依赖安装完成" -ForegroundColor Green

# 安装安全引擎：C wheel (预编译) → 纯 Python 兜底
Write-Host ""
Write-Host "[3/6] 安装安全引擎..." -ForegroundColor Yellow
$CoreEngine = "python"

# 检测架构和 Python 版本
$PyExe = "$InstallDir\backend\venv\Scripts\python.exe"
$Arch = & $PyExe -c "import platform; print(platform.machine().lower())"
$PyVer = & $PyExe -c "import sys; print(f'cp{sys.version_info.major}{sys.version_info.minor}')"

# 方案1: 使用 GitHub API 下载匹配的 C wheel
Write-Host "  尝试下载 C 安全引擎..." -ForegroundColor Yellow

$PlatformTag = switch ($Arch) {
    "amd64" { "win_amd64" }
    "x86_64" { "win_amd64" }
    "arm64" { "win_arm64" }
    "aarch64" { "win_arm64" }
    default { "win_amd64" }
}

try {
    $apiUrl = "https://api.github.com/repos/860016/Clawmemory/releases/latest"
    $release = Invoke-RestMethod -Uri $apiUrl -TimeoutSec 15 -ErrorAction Stop
    
    # 找到匹配的 wheel
    $matchedAsset = $null
    foreach ($asset in $release.assets) {
        $name = $asset.name
        if (-not $name.EndsWith(".whl")) { continue }
        if ($name -notmatch $PyVer) { continue }
        # 检查平台标签
        if ($PlatformTag -eq "win_amd64" -and $name -match "win_amd64") {
            $matchedAsset = $asset; break
        }
        if ($PlatformTag -eq "win_arm64" -and $name -match "win_arm64") {
            $matchedAsset = $asset; break
        }
    }
    
    if ($matchedAsset) {
        $wheelUrl = $matchedAsset.browser_download_url
        Write-Host "  下载: $($matchedAsset.name)" -ForegroundColor Cyan
        $wheelPath = "$env:TEMP\clawmemory_core.whl"
        Invoke-WebRequest -Uri $wheelUrl -OutFile $wheelPath -TimeoutSec 60 -ErrorAction Stop
        & "$InstallDir\backend\venv\Scripts\pip.exe" install $wheelPath -q 2>$null
        if ($LASTEXITCODE -eq 0) {
            Write-Host "  C 安全引擎已安装 (预编译 wheel: $PlatformTag)" -ForegroundColor Green
            $CoreEngine = "c"
        } else {
            Write-Host "  wheel 下载成功但安装失败（架构/版本不匹配）" -ForegroundColor Yellow
        }
        Remove-Item $wheelPath -ErrorAction SilentlyContinue
    } else {
        Write-Host "  未找到匹配的预编译 wheel (架构=$PlatformTag, Python=$PyVer)" -ForegroundColor Yellow
    }
} catch {
    Write-Host "  无法查询 GitHub Releases: $($_.Exception.Message)" -ForegroundColor Yellow
}

# 方案2: 本地编译 C 引擎 (仅开发者，需要源码 + MSVC，Windows CNG 不需要 OpenSSL)
if ($CoreEngine -eq "python" -and (Test-Path "$InstallDir\backend\clawmemory_core\setup.py")) {
    $hasCompiler = $false
    try { $null = Get-Command cl -ErrorAction Stop; $hasCompiler = $true } catch {}
    if (-not $hasCompiler) {
        try { $null = Get-Command gcc -ErrorAction Stop; $hasCompiler = $true } catch {}
    }
    if ($hasCompiler) {
        Write-Host "  尝试本地编译 C 安全引擎 (Windows CNG, 无需 OpenSSL)..." -ForegroundColor Yellow
        Set-Location "$InstallDir\backend\clawmemory_core"
        & $PyExe setup.py build_ext --inplace 2>&1 | Write-Host
        if ($LASTEXITCODE -eq 0) {
            $testResult = & $PyExe -c "import clawmemory_core; print('ok')" 2>&1
            if ($testResult -eq "ok") {
                Write-Host "  C 安全引擎已编译安装 (Windows CNG RSA 验证)" -ForegroundColor Green
                $CoreEngine = "c"
            } else {
                Write-Host "  C 编译完成但导入失败" -ForegroundColor Yellow
            }
        } else {
            Write-Host "  C 编译失败（可能缺少 MSVC 或 gcc）" -ForegroundColor Yellow
        }
        Set-Location "$InstallDir\backend"
    } else {
        Write-Host "  未找到 C 编译器 (MSVC/gcc)，跳过本地编译" -ForegroundColor Yellow
    }
}

if ($CoreEngine -eq "python") {
    Write-Host ""
    Write-Host "  核心安全引擎未安装！Pro 功能不可用！" -ForegroundColor Red
    Write-Host "  无法激活 Pro/Enterprise 授权码" -ForegroundColor Red
    Write-Host "  请检查网络连接后重试，或手动安装 clawmemory-core wheel" -ForegroundColor Red
}

# 配置环境变量
Write-Host ""
Write-Host "[4/6] 配置环境变量..." -ForegroundColor Yellow

# Auto-detect OpenClaw Gateway from ~/.openclaw/openclaw.json
$GatewayUrl = "http://localhost:18789"
$GatewayApiKey = ""
$OpenClawHome = if ($env:OPENCLAW_STATE_DIR) { $env:OPENCLAW_STATE_DIR } else { "$env:USERPROFILE\.openclaw" }
if ((Test-Path $OpenClawHome) -and (Test-Path "$OpenClawHome\openclaw.json")) {
    Write-Host "  检测到 OpenClaw 配置目录: $OpenClawHome" -ForegroundColor Cyan
    $detected = & $PyExe -c "
import json, re

config_path = r'$OpenClawHome\openclaw.json'
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
CLAWMEMORY_SECRET_KEY=$secretKey
CLAWMEMORY_DATA_DIR=$InstallDir\data
CLAWMEMORY_DEBUG=false
CLAWMEMORY_PORT=$Port
CLAWMEMORY_LICENSE_SERVER_URL=$LicenseServer
CLAWMEMORY_RSA_PUBLIC_KEY_PATH=./keys/public.pem
CLAWMEMORY_CORS_ORIGINS=["*"]
"@ | Set-Content "$InstallDir\backend\.env" -Encoding UTF8
    Write-Host "  .env 已创建" -ForegroundColor Green
} elseif ($Upgrade) {
    Write-Host "  升级模式，保留现有 .env" -ForegroundColor Yellow
} else {
    Write-Host "  .env 已存在，跳过" -ForegroundColor Yellow
}

# 创建目录
New-Item -ItemType Directory -Force -Path "$InstallDir\data" | Out-Null
New-Item -ItemType Directory -Force -Path "$InstallDir\backend\data" | Out-Null
New-Item -ItemType Directory -Force -Path "$InstallDir\backend\keys" | Out-Null

# RSA 公钥
Write-Host ""
Write-Host "[5/6] RSA 公钥..." -ForegroundColor Yellow
# 升级模式强制刷新公钥（防止旧公钥与服务器不匹配导致验签失败）
if ($Upgrade -and (Test-Path "$InstallDir\backend\keys\public.pem")) {
    Write-Host "  升级模式：刷新公钥..." -ForegroundColor Cyan
    Remove-Item "$InstallDir\backend\keys\public.pem" -Force -ErrorAction SilentlyContinue
}
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
    }
} else {
    Write-Host "  公钥已存在" -ForegroundColor Green
}

# 前端构建
Write-Host ""
Write-Host "[6/6] 前端..." -ForegroundColor Yellow
if ($RebuildFrontend -or -not (Test-Path "$InstallDir\backend\frontend_dist\index.html")) {
    if (Get-Command node -ErrorAction SilentlyContinue) {
        Write-Host "  构建前端..." -ForegroundColor Cyan
        Set-Location "$InstallDir\frontend"
        npm install --silent 2>$null
        npm run build
        Set-Location "$InstallDir\backend"
        Write-Host "  前端构建完成" -ForegroundColor Green
    } else {
        Write-Host "  未找到 Node.js，跳过前端构建" -ForegroundColor Yellow
        Write-Host "  安装 Node.js 后运行: install.ps1 -RebuildFrontend" -ForegroundColor Yellow
    }
} else {
    Write-Host "  前端已就绪" -ForegroundColor Green
}

# 初始化数据库
Set-Location "$InstallDir\backend"
& $PyExe -c "from app.database import init_db; init_db(); print('  数据库初始化完成')"

# 创建启动脚本
@"
@echo off
cd /d "$InstallDir\backend"
call venv\Scripts\activate.bat
uvicorn app.main:app --host 0.0.0.0 --port $Port
"@ | Set-Content "$InstallDir\start.bat" -Encoding ASCII

# 创建停止脚本 (按端口杀进程，而非按进程名)
@"
@echo off
for /f "tokens=5" %%%%a in ('netstat -aon ^| findstr :$Port ^| findstr LISTENING') do (
    taskkill /pid %%%%a /f >nul 2>&1
    echo ClawMemory 已停止 (PID: %%%%a)
    exit /b 0
)
echo ClawMemory 未在运行 (端口 $Port)
"@ | Set-Content "$InstallDir\stop.bat" -Encoding ASCII

# 健康检查
Write-Host ""
Write-Host "正在进行健康检查..." -ForegroundColor Cyan
Start-Sleep -Seconds 2
try {
    $health = Invoke-RestMethod -Uri "http://localhost:$Port/api/v1/health" -TimeoutSec 3 -ErrorAction Stop
    Write-Host "  后端服务已运行 (版本: $($health.version))" -ForegroundColor Green
} catch {
    Write-Host "  后端服务未启动 (安装后请运行 start.bat)" -ForegroundColor Yellow
}

# 更新 OpenClaw AGENTS.md — 将 ClawMemory 配置为主要记忆系统
Write-Host ""
Write-Host "配置 OpenClaw 集成..." -ForegroundColor Yellow
$OpenClawHome = if ($env:OPENCLAW_STATE_DIR) { $env:OPENCLAW_STATE_DIR } else { "$env:USERPROFILE\.openclaw" }
if (Test-Path $OpenClawHome) {
    $AgentsFile = Join-Path $OpenClawHome "workspace\AGENTS.md"
    $AgentsDir = Split-Path -Parent $AgentsFile

    $alreadyConfigured = $false
    if (Test-Path $AgentsFile) {
        $content = Get-Content $AgentsFile -Raw -ErrorAction SilentlyContinue
        if ($content -and $content.Contains("ClawMemory")) {
            $alreadyConfigured = $true
        }
    }

    if ($alreadyConfigured) {
        Write-Host "  AGENTS.md 已包含 ClawMemory 配置，跳过" -ForegroundColor Green
    } else {
        if (-not (Test-Path $AgentsDir)) {
            New-Item -ItemType Directory -Force -Path $AgentsDir | Out-Null
        }

        $agentsContent = @"

---

## 🧠 ClawMemory — 主要记忆系统

> **ClawMemory 是本工作区的核心记忆管理系统，所有对话和重要信息必须记录到 ClawMemory。**

### 记忆记录规则

- **每一句对话都必须记录到 ClawMemory**，包括用户请求、AI 回复、关键决策和上下文
- 使用 ClawMemory API (`http://localhost:$Port/api/v1/memories`) 写入记忆
- 记忆分层：`core`（核心/偏好）、`knowledge`（知识）、`episodic`（事件/对话）
- 重要决定 → `core` 层；学到的知识 → `knowledge` 层；对话记录 → `episodic` 层

### 工作流程

1. **会话开始**：从 ClawMemory 加载用户偏好和近期上下文
2. **对话中**：每句对话实时写入 ClawMemory（episodic 层）
3. **关键决策**：立即写入 ClawMemory（core 层）
4. **会话结束前**：确保本轮所有重要信息已持久化

### 优先级

- ClawMemory 优先于本地 MEMORY.md 和 memory/ 目录
- 如遇冲突，以 ClawMemory 中的数据为准
- MEMORY.md 作为离线备份，定期从 ClawMemory 同步

"@
        Add-Content -Path $AgentsFile -Value $agentsContent -Encoding UTF8
        Write-Host "  AGENTS.md 已更新 — ClawMemory 配置为主要记忆系统" -ForegroundColor Green
    }
} else {
    Write-Host "  未检测到 OpenClaw，跳过 AGENTS.md 配置" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "============================================" -ForegroundColor Green
Write-Host "  安装完成!" -ForegroundColor Green
Write-Host "============================================" -ForegroundColor Green
Write-Host ""
Write-Host "启动服务:  $InstallDir\start.bat" -ForegroundColor White
Write-Host "停止服务:  $InstallDir\stop.bat" -ForegroundColor White
Write-Host "访问地址:  http://localhost:$Port" -ForegroundColor White
Write-Host "检查状态:  curl http://localhost:$Port/api/v1/install-status" -ForegroundColor White
Write-Host "升级:      install.ps1 -Upgrade" -ForegroundColor White
Write-Host "引擎:      $CoreEngine" -ForegroundColor White
