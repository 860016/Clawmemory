# ClawMemory Go 跨平台编译脚本
# 用法: .\scripts\build.ps1 [版本号]

param(
    [string]$Version = "3.0.0"
)

$ProjectName = "clawmemory"
$BuildDir = "..\build"

# 创建构建目录
New-Item -ItemType Directory -Force -Path $BuildDir | Out-Null

# 版本信息
$LdFlags = "-s -w -X main.Version=$Version"

Write-Host "=== ClawMemory 构建系统 ===" -ForegroundColor Cyan
Write-Host "版本: $Version" -ForegroundColor Yellow

# 构建函数
function Build-Target {
    param(
        [string]$OS,
        [string]$Arch,
        [string]$OutputName,
        [string[]]$Tags = @()
    )

    Write-Host ""
    Write-Host "构建 $OS/$Arch..." -ForegroundColor Green

    $env:GOOS = $OS
    $env:GOARCH = $Arch
    $env:CGO_ENABLED = 1

    # Windows 需要特殊处理 CGO
    if ($OS -eq "windows") {
        $env:CGO_ENABLED = 1
    } else {
        $env:CGO_ENABLED = 0
    }

    $tagsFlag = ""
    if ($Tags.Count -gt 0) {
        $tagsFlag = "-tags `""$($Tags -join ',')`"""
    }

    $output = Join-Path $BuildDir $OutputName
    $cmd = "go build $tagsFlag -ldflags `""$LdFlags`"" -o `""$output`"" ..\cmd\server"

    Write-Host "执行: $cmd" -ForegroundColor Gray
    Invoke-Expression $cmd

    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ 构建成功: $OutputName" -ForegroundColor Green
    } else {
        Write-Host "✗ 构建失败: $OutputName" -ForegroundColor Red
    }
}

# 1. Windows AMD64 (带 Pro 功能)
Build-Target -OS "windows" -Arch "amd64" -OutputName "clawmemory-windows-amd64-pro.exe" -Tags @("pro")

# 2. Windows AMD64 (开源版)
Build-Target -OS "windows" -Arch "amd64" -OutputName "clawmemory-windows-amd64.exe"

# 3. Linux AMD64 (带 Pro 功能)
Build-Target -OS "linux" -Arch "amd64" -OutputName "clawmemory-linux-amd64-pro" -Tags @("pro")

# 4. Linux AMD64 (开源版)
Build-Target -OS "linux" -Arch "amd64" -OutputName "clawmemory-linux-amd64"

# 5. macOS AMD64 (带 Pro 功能)
Build-Target -OS "darwin" -Arch "amd64" -OutputName "clawmemory-darwin-amd64-pro" -Tags @("pro")

# 6. macOS AMD64 (开源版)
Build-Target -OS "darwin" -Arch "amd64" -OutputName "clawmemory-darwin-amd64"

# 7. macOS ARM64 (M1/M2, 带 Pro 功能)
Build-Target -OS "darwin" -Arch "arm64" -OutputName "clawmemory-darwin-arm64-pro" -Tags @("pro")

# 8. macOS ARM64 (M1/M2, 开源版)
Build-Target -OS "darwin" -Arch "arm64" -OutputName "clawmemory-darwin-arm64"

Write-Host ""
Write-Host "=== 构建完成 ===" -ForegroundColor Cyan
Write-Host "输出目录: $(Resolve-Path $BuildDir)" -ForegroundColor Yellow

# 列出构建结果
Get-ChildItem $BuildDir | ForEach-Object {
    $size = [math]::Round($_.Length / 1MB, 2)
    Write-Host "  $($_.Name) (${size} MB)" -ForegroundColor White
}
