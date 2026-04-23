# ClawMemory Windows 安装脚本 v2.0
# 用法: powershell -ExecutionPolicy Bypass -File install.ps1
param(
    [string]$LicenseServer = "https://auth.bestu.top",
    [int]$Port = 8765,
    [switch]$RebuildFrontend,
    [switch]$Upgrade,
    [switch]$UseGoBackend
)

$ErrorActionPreference = "Stop"
$InstallDir = Split-Path -Parent $MyInvocation.MyCommand.Path

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "  ClawMemory 安装程序 v2.0 (Windows)" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host "安装目录: $InstallDir"
Write-Host "后端端口: $Port"
Write-Host "授权服务器: $LicenseServer"
if ($Upgrade) { Write-Host "模式: 升级 (保留配置和数据)" -ForegroundColor Yellow }
if ($UseGoBackend) { Write-Host "后端: Go (高性能)" -ForegroundColor Cyan }
else { Write-Host "后端: Python" -ForegroundColor Cyan }
Write-Host ""

# ==================== Go 后端安装 ====================
if ($UseGoBackend) {
    Write-Host "[1/5] 检查 Go 环境..." -ForegroundColor Yellow
    try {
        $goVer = go version 2>&1
        Write-Host "  Go: $goVer" -ForegroundColor Green
    } catch {
        Write-Host "  未找到 Go，请先安装 Go 1.21+" -ForegroundColor Red
        Write-Host "  下载: https://go.dev/dl/" -ForegroundColor Red
        exit 1
    }

    # 检查前端
    if (Test-Path "$InstallDir\go-backend\frontend_dist\index.html") {
        Write-Host "  前端: 已预构建" -ForegroundColor Green
    } else {
        Write-Host "  前端: 未构建" -ForegroundColor Yellow
        if (Get-Command node -ErrorAction SilentlyContinue) {
            Write-Host "  检测到 Node.js，将自动构建前端" -ForegroundColor Cyan
            $RebuildFrontend = $true
        }
    }

    # 前端构建
    Write-Host ""
    Write-Host "[2/5] 前端构建..." -ForegroundColor Yellow
    if ($RebuildFrontend -or -not (Test-Path "$InstallDir\go-backend\frontend_dist\index.html")) {
        if (Get-Command node -ErrorAction SilentlyContinue) {
            Write-Host "  构建前端..." -ForegroundColor Cyan
            Set-Location "$InstallDir\frontend"
            npm install --silent 2>$null
            npm run build
            # 复制到 go-backend
            if (Test-Path "$InstallDir\frontend\dist") {
                Copy-Item -Path "$InstallDir\frontend\dist\*" -Destination "$InstallDir\go-backend\frontend_dist\" -Recurse -Force
                Write-Host "  前端已复制到 go-backend" -ForegroundColor Green
            }
            Set-Location "$InstallDir"
        } else {
            Write-Host "  未找到 Node.js，跳过前端构建" -ForegroundColor Yellow
        }
    } else {
        Write-Host "  前端已就绪" -ForegroundColor Green
    }

    # 编译 Go 后端
    Write-Host ""
    Write-Host "[3/5] 编译 Go 后端..." -ForegroundColor Yellow
    Set-Location "$InstallDir\go-backend"
    go build -o clawmemory.exe ./cmd/server
    if ($LASTEXITCODE -eq 0) {
        Write-Host "  Go 后端编译成功" -ForegroundColor Green
    } else {
        Write-Host "  Go 后端编译失败" -ForegroundColor Red
        exit 1
    }
    Set-Location "$InstallDir"

    # 配置环境变量
    Write-Host ""
    Write-Host "[4/5] 配置环境变量..." -ForegroundColor Yellow
    if (-not (Test-Path "$InstallDir\go-backend\.env")) {
        $secretKey = -join ((48..57) + (65..90) + (97..122) | Get-Random -Count 32 | ForEach-Object { [char]$_ })
        @"
SECRET_KEY=$secretKey
PORT=$Port
DATA_DIR=$InstallDir\data
LICENSE_SERVER_URL=$LicenseServer
"@ | Set-Content "$InstallDir\go-backend\.env" -Encoding UTF8
        Write-Host "  .env 已创建" -ForegroundColor Green
    } else {
        Write-Host "  .env 已存在，跳过" -ForegroundColor Yellow
    }

    # 创建目录
    New-Item -ItemType Directory -Force -Path "$InstallDir\data" | Out-Null

    # 创建启动脚本
    Write-Host ""
    Write-Host "[5/5] 创建启动脚本..." -ForegroundColor Yellow
    @"
@echo off
cd /d "$InstallDir\go-backend"
clawmemory.exe
"@ | Set-Content "$InstallDir\start.bat" -Encoding ASCII

    @"
@echo off
for /f "tokens=5" %%%%a in ('netstat -aon ^| findstr :$Port ^| findstr LISTENING') do (
    taskkill /pid %%%%a /f >nul 2>&1
    echo ClawMemory 已停止 (PID: %%%%a)
    exit /b 0
)
echo ClawMemory 未在运行 (端口 $Port)
"@ | Set-Content "$InstallDir\stop.bat" -Encoding ASCII

    Write-Host ""
    Write-Host "============================================" -ForegroundColor Green
    Write-Host "  安装完成!" -ForegroundColor Green
    Write-Host "============================================" -ForegroundColor Green
    Write-Host ""
    Write-Host "启动服务:  $InstallDir\start.bat" -ForegroundColor White
    Write-Host "停止服务:  $InstallDir\stop.bat" -ForegroundColor White
    Write-Host "访问地址:  http://localhost:$Port" -ForegroundColor White
    Write-Host "后端引擎:  Go (高性能)" -ForegroundColor Cyan
    exit 0
}

# ==================== Python 后端安装 ====================
Write-Host "[1/6] 检查 Python 环境..." -ForegroundColor Yellow
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

# 安装依赖
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

# 安装安全引擎
Write-Host ""
Write-Host "[3/6] 安装安全引擎..." -ForegroundColor Yellow
$CoreEngine = "python"

$PyExe = "$InstallDir\backend\venv\Scripts\python.exe"
$Arch = & $PyExe -c "import platform; print(platform.machine().lower())"
$PyVer = & $PyExe -c "import sys; print(f'cp{sys.version_info.major}{sys.version_info.minor}')"

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
    
    $matchedAsset = $null
    foreach ($asset in $release.assets) {
        $name = $asset.name
        if (-not $name.EndsWith(".whl")) { continue }
        if ($name -notmatch $PyVer) { continue }
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
            Write-Host "  C 安全引擎已安装" -ForegroundColor Green
            $CoreEngine = "c"
        }
        Remove-Item $wheelPath -ErrorAction SilentlyContinue
    } else {
        Write-Host "  未找到匹配的预编译 wheel" -ForegroundColor Yellow
    }
} catch {
    Write-Host "  无法查询 GitHub Releases: $($_.Exception.Message)" -ForegroundColor Yellow
}

if ($CoreEngine -eq "python") {
    Write-Host "  使用纯 Python 安全引擎 (Pro 功能受限)" -ForegroundColor Yellow
}

# 配置环境变量
Write-Host ""
Write-Host "[4/6] 配置环境变量..." -ForegroundColor Yellow
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
if ($Upgrade -and (Test-Path "$InstallDir\backend\keys\public.pem")) {
    Write-Host "  升级模式：刷新公钥..." -ForegroundColor Cyan
    Remove-Item "$InstallDir\backend\keys\public.pem" -Force -ErrorAction SilentlyContinue
}
if (-not (Test-Path "$InstallDir\backend\keys\public.pem")) {
    $pubKeyFetched = $false

    try {
        Write-Host "  正在从授权服务器获取公钥..." -ForegroundColor Cyan
        $pubKey = Invoke-RestMethod -Uri "$LicenseServer/api/v1/public-key" -TimeoutSec 10 -ErrorAction Stop
        if ($pubKey -and $pubKey.ToString().Contains("BEGIN")) {
            $pubKey | Set-Content "$InstallDir\backend\keys\public.pem" -Encoding UTF8
            Write-Host "  公钥已从授权服务器获取" -ForegroundColor Green
            $pubKeyFetched = $true
        }
    } catch {}

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

    if (-not $pubKeyFetched) {
        Write-Host "  无法自动获取公钥" -ForegroundColor Yellow
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

Write-Host ""
Write-Host "============================================" -ForegroundColor Green
Write-Host "  安装完成!" -ForegroundColor Green
Write-Host "============================================" -ForegroundColor Green
Write-Host ""
Write-Host "启动服务:  $InstallDir\start.bat" -ForegroundColor White
Write-Host "停止服务:  $InstallDir\stop.bat" -ForegroundColor White
Write-Host "访问地址:  http://localhost:$Port" -ForegroundColor White
Write-Host "引擎:      $CoreEngine" -ForegroundColor White
