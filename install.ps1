# ClawMemory Windows 安装脚本 v2.0
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
Write-Host "  ClawMemory 安装程序 v2.0 (Windows)" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host "安装目录: $InstallDir"
Write-Host "后端端口: $Port"
Write-Host "授权服务器: $LicenseServer"
if ($Upgrade) { Write-Host "模式: 升级 (保留配置和数据)" -ForegroundColor Yellow }
Write-Host ""

# 检查 Go 环境
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
            if (Test-Path "$InstallDir\go-backend\frontend_dist") {
                Remove-Item "$InstallDir\go-backend\frontend_dist" -Recurse -Force
            }
            Copy-Item -Path "$InstallDir\frontend\dist" -Destination "$InstallDir\go-backend\frontend_dist" -Recurse -Force
            Write-Host "  前端已复制到 go-backend" -ForegroundColor Green
        }
        Set-Location "$InstallDir"
    } else {
        Write-Host "  未找到 Node.js，跳过前端构建" -ForegroundColor Yellow
        Write-Host "  请确保 go-backend/frontend_dist 目录存在" -ForegroundColor Yellow
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
} elseif ($Upgrade) {
    Write-Host "  升级模式，保留现有 .env" -ForegroundColor Yellow
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