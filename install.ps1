# ClawMemory Windows 安装脚本 v3.1 (完整版)
# 用法: powershell -ExecutionPolicy Bypass -File install.ps1
# 特性: 自动检测依赖、构建前端、编译后端、创建统一目录、生成启动脚本
param(
    [string]$LicenseServer = "https://auth.bestu.top",
    [int]$Port = 8765,
    [switch]$RebuildFrontend,
    [switch]$Upgrade,
    [string]$InstallPath = "",
    [switch]$AutoStart
)

$ErrorActionPreference = "Stop"
$Host.UI.RawUI.WindowTitle = "ClawMemory 安装程序 v3.1"
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path

if ($InstallPath -eq "") {
    $InstallDir = $ScriptDir
} else {
    $InstallDir = $InstallPath
    if (-not (Test-Path $InstallDir)) {
        try {
            New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
            Write-Host "✓ 创建安装目录: $InstallDir" -ForegroundColor Green
        } catch {
            Write-Host "✗ 无法创建目录: $InstallDir" -ForegroundColor Red
            exit 1
        }
    }
}

$DataDir = "$InstallDir\data"
$SkillsDir = "$DataDir\skills"
$BackupsDir = "$DataDir\backups"
$EnvFile = "$InstallDir\go-backend\.env"
$ExeFile = "$InstallDir\go-backend\clawmemory.exe"

function Write-Step($step, $total, $message) {
    Write-Host "[$step/$total] $message" -ForegroundColor Yellow
}

function Write-Success($message) {
    Write-Host "  ✓ $message" -ForegroundColor Green
}

function Write-Error($message) {
    Write-Host "  ✗ $message" -ForegroundColor Red
}

function Write-Info($message) {
    Write-Host "  ℹ $message" -ForegroundColor Cyan
}

Write-Host ""
Write-Host "╔══════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║     ClawMemory 安装程序 v3.1 (Windows)   ║" -ForegroundColor Cyan
Write-Host "╚══════════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host ""
Write-Info "安装目录: $InstallDir"
Write-Info "数据目录: $DataDir"
Write-Info "技能目录: $SkillsDir"
Write-Info "备份目录: $BackupsDir"
Write-Info "服务端口: $Port"
Write-Info "授权服务器: $LicenseServer"
if ($Upgrade) { Write-Info "模式: 升级 (保留配置和数据)" }
if ($AutoStart) { Write-Info "自动启动: 是" }
Write-Host ""

$totalSteps = 6

Write-Step 1 $totalSteps "检查环境依赖..."
Write-Host ""

$hasGo = $false
$hasNode = $false
$hasGit = $false

try {
    $goVer = go version 2>&1
    $hasGo = $true
    Write-Success "Go 环境: $goVer"
} catch {
    Write-Error "Go 未安装"
    Write-Info "下载地址: https://go.dev/dl/"
    Write-Info "请安装 Go 1.21+ 后重试"
    exit 1
}

try {
    $nodeVer = node --version 2>&1
    $hasNode = $true
    Write-Success "Node.js: $nodeVer"
} catch {
    Write-Info "Node.js 未安装 (可选，用于构建前端)"
}

try {
    $gitVer = git --version 2>&1
    $hasGit = $true
    Write-Success "Git: $gitVer"
} catch {
    Write-Info "Git 未安装 (可选，用于技能安装)"
}

Write-Step 2 $totalSteps "构建前端..."
Write-Host ""

$frontendReady = Test-Path "$InstallDir\go-backend\frontend_dist\index.html"

if ($RebuildFrontend -or (-not $frontendReady)) {
    if ($hasNode) {
        Write-Info "开始构建前端..."
        Set-Location "$InstallDir\frontend"

        try {
            npm install --prefer-offline --no-audit --no-fund 2>$null
            Write-Success "npm install 完成"
            
            npm run build 2>&1 | ForEach-Object { 
                if ($_ -match "(error|Error|ERROR)") { Write-Error $_ }
                elseif ($_ -match "(built|warning|✓)") { Write-Success $_ }
            }
            
            if (Test-Path "$InstallDir\frontend\dist") {
                if (Test-Path "$InstallDir\go-backend\frontend_dist") {
                    Remove-Item "$InstallDir\go-backend\frontend_dist" -Recurse -Force
                }
                Copy-Item -Path "$InstallDir\frontend\dist" -Destination "$InstallDir\go-backend\frontend_dist" -Recurse -Force
                Write-Success "前端已复制到 go-backend/frontend_dist"
            }
        } catch {
            Write-Error "前端构建失败: $_"
            Write-Info "将使用预构建版本（如果存在）"
        }

        Set-Location "$InstallDir"
    } else {
        if ($frontendReady) {
            Write-Success "使用预构建的前端文件"
        } else {
            Write-Warning "未找到 Node.js 且无预构建前端"
            Write-Info "前端功能可能不可用"
        }
    }
} else {
    Write-Success "前端已就绪 (跳过构建)"
}

Write-Step 3 $totalSteps "编译 Go 后端..."
Write-Host ""

Set-Location "$InstallDir\go-backend"

try {
    $buildOutput = go build -o clawmemory.exe ./cmd/server 2>&1
    if ($LASTEXITCODE -eq 0) {
        if (Test-Path $ExeFile) {
            $exeSize = [math]::Round((Get-Item $ExeFile).Length / 1MB, 2)
            Write-Success "Go 后端编译成功 (${exeSize} MB)"
        } else {
            Write-Success "Go 后端编译成功"
        }
    } else {
        Write-Error "Go 后端编译失败"
        Write-Info "错误信息: $buildOutput"
        exit 1
    }
} catch {
    Write-Error "编译异常: $_"
    exit 1
}

Set-Location "$InstallDir"

Write-Step 4 $totalSteps "配置环境..."
Write-Host ""

if (-not (Test-Path $EnvFile)) {
    $secretKey = -join ((48..57) + (65..90) + (97..122) | Get-Random -Count 32 | ForEach-Object { [char]$_ })
    
    $envContent = @"
HOST=0.0.0.0
SECRET_KEY=$secretKey
PORT=$Port
DATA_DIR=$DataDir
SKILLS_DIR=$SkillsDir
BACKUPS_DIR=$BackupsDir
LICENSE_SERVER_URL=$LicenseServer
"@
    
    Set-Content -Path $EnvFile -Value $envContent -Encoding UTF8
    Write-Success ".env 配置文件已创建"
    Write-Info "SECRET_KEY 已自动生成"
} elseif ($Upgrade) {
    Write-Success "升级模式，保留现有 .env 配置"
} else {
    Write-Success ".env 配置文件已存在"
}

Write-Host ""
Write-Info "创建统一目录结构..."

$dirsToCreate = @(
    @{ Path = $DataDir; Name = "数据目录" },
    @{ Path = $SkillsDir; Name = "技能目录" },
    @{ Path = $BackupsDir; Name = "备份目录" },
    @{ Path = "$DataDir\keys"; Name = "密钥目录" },
    @{ Path = "$DataDir\uploads"; Name = "上传目录" }
)

foreach ($dir in $dirsToCreate) {
    New-Item -ItemType Directory -Force -Path $dir.Path | Out-Null
    Write-Success "$($dir.Name): $($dir.Path)"
}

Write-Step 5 $totalSteps "生成启动脚本..."
Write-Host ""

$startBatContent = @"
@echo off
title ClawMemory Server
cd /d "%~dp0go-backend"
echo ============================================
echo   Starting ClawMemory...
echo   Port: $Port
echo   Data: $DataDir
echo ============================================
clawmemory.exe
pause
"@

$stopBatContent = @"
@echo off
title Stop ClawMemory
echo Stopping ClawMemory on port $Port...
for /f "tokens=5" %%%%a in ('netstat -aon ^| findstr :$Port ^| findstr LISTENING') do (
    echo Found process PID: %%%%a
    taskkill /pid %%%%a /f >nul 2>&1
    if !errorlevel! equ 0 (
        echo ✓ ClawMemory stopped successfully (PID: %%%%a)
    ) else (
        echo ✗ Failed to stop process
    )
    goto :done
)
echo ClawMemory is not running on port $Port
:done
pause
"@

Set-Content -Path "$InstallDir\start.bat" -Value $startBatContent -Encoding ASCII
Set-Content -Path "$InstallDir\stop.bat" -Value $stopBatContent -Encoding ASCII

Write-Success "start.bat - 启动脚本已生成"
Write-Success "stop.bat  - 停止脚本已生成"

Write-Step 6 $totalSteps "验证安装..."
Write-Host ""

$checks = @(
    @{ File = $ExeFile; Name = "可执行文件"; Required = $true },
    @{ File = "$InstallDir\go-backend\frontend_dist\index.html"; Name = "前端文件"; Required = $true },
    @{ File = $EnvFile; Name = "配置文件"; Required = $true },
    @{ File = "$InstallDir\start.bat"; Name = "启动脚本"; Required = $true },
    @{ File = "$InstallDir\stop.bat"; Name = "停止脚本"; Required = $true },
    @{ Dir = $DataDir; Name = "数据目录"; Required = $true },
    @{ Dir = $SkillsDir; Name = "技能目录"; Required = $false },
    @{ Dir = $BackupsDir; Name = "备份目录"; Required = $false }
)

$allPassed = $true

foreach ($check in $checks) {
    if ($check.File) {
        if (Test-Path $check.File) {
            Write-Success "$($check.Name): ✓"
        } else {
            if ($check.Required) {
                Write-Error "$($check.Name): ✗ 缺失!"
                $allPassed = $false
            } else {
                Write-Info "$($check.Name): ⚠ 不存在 (可选)"
            }
        }
    } elseif ($check.Dir) {
        if (Test-Path $check.Dir) {
            Write-Success "$($check.Name): ✓"
        } else {
            if ($check.Required) {
                Write-Error "$($check.Name): ✗ 缺失!"
                $allPassed = $false
            } else {
                Write-Info "$($check.Name): ⚠ 不存在 (可选)"
            }
        }
    }
}

Write-Host ""
if ($allPassed) {
    Write-Host "╔══════════════════════════════════════════╗" -ForegroundColor Green
    Write-Host "║           ✅ 安装完成!                    ║" -ForegroundColor Green
    Write-Host "╚══════════════════════════════════════════╝" -ForegroundColor Green
} else {
    Write-Host "⚠️ 安装完成但有警告，请检查上述缺失项" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "📁 目录结构:" -ForegroundColor Cyan
Write-Host "   $InstallDir\" -ForegroundColor White
Write-Host "   ├── go-backend/          # 后端程序" -ForegroundColor Gray
Write-Host "   │   ├── clawmemory.exe  # 可执行文件" -ForegroundColor Gray
Write-Host "   │   └── .env            # 配置文件" -ForegroundColor Gray
Write-Host "   ├── data/               # 数据目录 (统一)" -ForegroundColor Gray
Write-Host "   │   ├── skills/         # 技能插件" -ForegroundColor Gray
Write-Host "   │   ├── backups/        # 备份文件" -ForegroundColor Gray
Write-Host "   │   └── keys/           # 密钥文件" -ForegroundColor Gray
Write-Host "   ├── start.bat           # 启动脚本" -ForegroundColor Gray
Write-Host "   └── stop.bat            # 停止脚本" -ForegroundColor Gray
Write-Host ""
Write-Host "🚀 快速开始:" -ForegroundColor Cyan
Write-Host "   1. 双击 start.bat 启动服务" -ForegroundColor White
Write-Host "   2. 打开浏览器访问 http://localhost:$Port" -ForegroundColor White
Write-Host "   3. 首次访问需设置管理员密码" -ForegroundColor White
Write-Host ""
Write-Host "💡 提示:" -ForegroundColor Yellow
Write-Host "   • 所有数据统一存储在 data/ 目录下" -ForegroundColor Yellow
Write-Host "   • 技能安装自动保存到 data/skills/" -ForegroundColor Yellow
Write-Host "   • 使用 stop.bat 可安全停止服务" -ForegroundColor Yellow
Write-Host "   • 升级时使用 --Upgrade 参数保留配置" -ForegroundColor Yellow

if ($InstallPath -ne "") {
    Write-Host ""
    Write-Info "自定义安装路径: $InstallPath"
}

if ($AutoStart -and $allPassed) {
    Write-Host ""
    Write-Info "正在启动服务..."
    Start-Process -FilePath "$InstallDir\start.bat" -WorkingDirectory "$InstallDir"
    Start-Sleep -Seconds 2
    Write-Success "服务已启动!"
    Write-Info "请在浏览器中打开: http://localhost:$Port"
}

Write-Host ""
