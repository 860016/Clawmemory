# ClawMemory 全平台编译脚本 (PowerShell)
# 支持: Windows/Linux/macOS × x86/ARM

param(
    [string]$Version = "1.0.0"
)

$OutputDir = "releases/v${Version}"

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "ClawMemory 全平台编译" -ForegroundColor Cyan
Write-Host "版本: ${Version}" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan

# 创建输出目录
New-Item -ItemType Directory -Force -Path $OutputDir | Out-Null

# 编译函数
function Build-Target {
    param(
        [string]$Os,
        [string]$Arch,
        [string]$OutputName,
        [string]$Tags = ""
    )
    
    Write-Host ""
    Write-Host "编译: ${Os}/${Arch} ..." -ForegroundColor Yellow
    
    $env:GOOS = $Os
    $env:GOARCH = $Arch
    $env:CGO_ENABLED = "0"
    
    $ldflags = "-s -w -X main.Version=${Version}"
    
    if ($Tags) {
        go build -tags $Tags -ldflags $ldflags -o "${OutputDir}/${OutputName}" ./cmd/server
    } else {
        go build -ldflags $ldflags -o "${OutputDir}/${OutputName}" ./cmd/server
    }
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ 成功: ${OutputName}" -ForegroundColor Green
        
        # 压缩
        $CompressPath = Join-Path $OutputDir $OutputName
        if ($Os -eq "windows") {
            Compress-Archive -Path $CompressPath -DestinationPath "${OutputDir}/${OutputName%.exe}.zip" -Force
        } else {
            # 使用 tar（Git Bash 或 WSL）
            $BaseName = Split-Path $OutputName -Leaf
            tar -czf "${OutputDir}/${OutputName}.tar.gz" -C $OutputDir $BaseName 2>$null
        }
    } else {
        Write-Host "❌ 失败: ${OutputName}" -ForegroundColor Red
    }
    
    # 清理环境变量
    Remove-Item Env:\GOOS -ErrorAction SilentlyContinue
    Remove-Item Env:\GOARCH -ErrorAction SilentlyContinue
    Remove-Item Env:\CGO_ENABLED -ErrorAction SilentlyContinue
}

# ========== 开源版 ==========
Write-Host ""
Write-Host "📦 编译开源版..." -ForegroundColor Magenta

# Windows
Build-Target -Os "windows" -Arch "amd64" -OutputName "clawmemory-windows-amd64.exe"
Build-Target -Os "windows" -Arch "arm64" -OutputName "clawmemory-windows-arm64.exe"

# Linux
Build-Target -Os "linux" -Arch "amd64" -OutputName "clawmemory-linux-amd64"
Build-Target -Os "linux" -Arch "arm64" -OutputName "clawmemory-linux-arm64"
Build-Target -Os "linux" -Arch "386" -OutputName "clawmemory-linux-386"

# macOS
Build-Target -Os "darwin" -Arch "amd64" -OutputName "clawmemory-darwin-amd64"
Build-Target -Os "darwin" -Arch "arm64" -OutputName "clawmemory-darwin-arm64"

# ========== 生成校验和 ==========
Write-Host ""
Write-Host "🔐 生成校验和..." -ForegroundColor Yellow

$ChecksumFile = Join-Path $OutputDir "checksums.txt"
Get-ChildItem $OutputDir -File | Where-Object { $_.Name -ne "checksums.txt" -and $_.Name -ne "README.txt" } | ForEach-Object {
    $hash = Get-FileHash $_.FullName -Algorithm SHA256
    "$($hash.Hash)  $($_.Name)" | Out-File -Append -FilePath $ChecksumFile
}

# ========== 生成发布说明 ==========
Write-Host ""
Write-Host "📝 生成发布说明..." -ForegroundColor Yellow

$ReadmeContent = @"
ClawMemory v${Version}
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
Get-FileHash -Algorithm SHA256 <文件名>
"@

$ReadmeContent | Out-File -FilePath (Join-Path $OutputDir "README.txt")

# ========== 完成 ==========
Write-Host ""
Write-Host "==========================================" -ForegroundColor Green
Write-Host "编译完成!" -ForegroundColor Green
Write-Host "输出目录: ${OutputDir}" -ForegroundColor Green
Write-Host "==========================================" -ForegroundColor Green
Write-Host ""

Get-ChildItem $OutputDir | Format-Table Name, @{Name="Size";Expression={"{0:N2} MB" -f ($_.Length/1MB)}}, LastWriteTime
