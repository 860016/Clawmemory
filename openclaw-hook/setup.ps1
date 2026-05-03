$ErrorActionPreference = "Stop"

Write-Host ""
Write-Host "🧠 ClawMemory OpenClaw Hook - Setup"
Write-Host "====================================="

$HOOK_DIR = "$env:USERPROFILE\.openclaw\hooks\clawmemory"

if (Test-Path $HOOK_DIR) {
  Write-Host "⚠️  Hook already exists at $HOOK_DIR"
  $reply = Read-Host "Overwrite? (y/N)"
  if ($reply -ne "y" -and $reply -ne "Y") {
    Write-Host "Aborted."
    exit 0
  }
  Remove-Item -Recurse -Force $HOOK_DIR
}

$SCRIPT_DIR = Split-Path -Parent $MyInvocation.MyCommand.Path
New-Item -ItemType Directory -Path $HOOK_DIR -Force | Out-Null
Copy-Item "$SCRIPT_DIR\HOOK.md" "$HOOK_DIR\HOOK.md"
Copy-Item "$SCRIPT_DIR\handler.ts" "$HOOK_DIR\handler.ts"
Copy-Item "$SCRIPT_DIR\package.json" "$HOOK_DIR\package.json"

Write-Host ""
Write-Host "✅ Hook installed to $HOOK_DIR"
Write-Host ""

if (-not $env:CLAWMEMORY_URL) {
  Write-Host "⚠️  CLAWMEMORY_URL not set. Using default: http://localhost:8765"
  Write-Host "   To set: `$env:CLAWMEMORY_URL = 'http://localhost:8765'"
}

if (-not $env:CLAWMEMORY_API_KEY) {
  Write-Host "⚠️  CLAWMEMORY_API_KEY not set!"
  Write-Host "   1. Open ClawMemory at http://localhost:8765"
  Write-Host "   2. Go to Settings > API Keys"
  Write-Host "   3. Create a new key"
  Write-Host "   4. Set: `$env:CLAWMEMORY_API_KEY = 'cm_your_key_here'"
}

Write-Host ""
Write-Host "To enable the hook:"
Write-Host "  openclaw hooks enable clawmemory"
Write-Host ""
Write-Host "To restart OpenClaw:"
Write-Host "  openclaw gateway restart"
