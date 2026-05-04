$ErrorActionPreference = "Stop"

Write-Host ""
Write-Host "ClawMemory OpenClaw Hook - Setup"
Write-Host "====================================="

$SCRIPT_DIR = Split-Path -Parent $MyInvocation.MyCommand.Path
$HOOK_PACK_DIR = $SCRIPT_DIR

if (-not (Test-Path "$HOOK_PACK_DIR\clawmemory\HOOK.md")) {
    Write-Host "ERROR: Invalid hook pack structure. Missing clawmemory\HOOK.md"
    exit 1
}

if (-not (Test-Path "$HOOK_PACK_DIR\clawmemory\handler.ts")) {
    Write-Host "ERROR: Invalid hook pack structure. Missing clawmemory\handler.ts"
    exit 1
}

if (-not (Test-Path "$HOOK_PACK_DIR\package.json")) {
    Write-Host "ERROR: Missing package.json in hook pack root"
    exit 1
}

$openclawCmd = Get-Command "openclaw" -ErrorAction SilentlyContinue
if (-not $openclawCmd) {
    Write-Host ""
    Write-Host "WARNING: openclaw CLI not found on PATH."
    Write-Host "  Please install OpenClaw first: https://openclaw.ai"
    Write-Host ""
    Write-Host "Falling back to manual installation..."

    $HOOK_DIR = "$env:USERPROFILE\.openclaw\hooks\clawmemory"
    if (Test-Path $HOOK_DIR) {
        Write-Host "  Hook already exists at $HOOK_DIR"
        $reply = Read-Host "  Overwrite? (y/N)"
        if ($reply -ne "y" -and $reply -ne "Y") {
            Write-Host "  Aborted."
            exit 0
        }
        Remove-Item -Recurse -Force $HOOK_DIR
    }

    New-Item -ItemType Directory -Path $HOOK_DIR -Force | Out-Null
    Copy-Item "$HOOK_PACK_DIR\clawmemory\HOOK.md" "$HOOK_DIR\HOOK.md"
    Copy-Item "$HOOK_PACK_DIR\clawmemory\handler.ts" "$HOOK_DIR\handler.ts"

    Write-Host ""
    Write-Host "  Hook files copied to $HOOK_DIR"
    Write-Host ""
    Write-Host "  IMPORTANT: You must manually enable the hook:"
    Write-Host "    openclaw hooks enable clawmemory"
    Write-Host "    openclaw gateway restart"
} else {
    Write-Host ""
    Write-Host "Installing hook pack via openclaw plugins install..."

    & openclaw plugins install $HOOK_PACK_DIR

    if ($LASTEXITCODE -ne 0) {
        Write-Host ""
        Write-Host "ERROR: openclaw plugins install failed (exit code $LASTEXITCODE)"
        Write-Host "Falling back to manual installation..."

        $HOOK_DIR = "$env:USERPROFILE\.openclaw\hooks\clawmemory"
        if (Test-Path $HOOK_DIR) {
            Write-Host "  Hook already exists at $HOOK_DIR"
            $reply = Read-Host "  Overwrite? (y/N)"
            if ($reply -ne "y" -and $reply -ne "Y") {
                Write-Host "  Skipping file copy. Continuing with enable..."
            } else {
                Remove-Item -Recurse -Force $HOOK_DIR
                New-Item -ItemType Directory -Path $HOOK_DIR -Force | Out-Null
                Copy-Item "$HOOK_PACK_DIR\clawmemory\HOOK.md" "$HOOK_DIR\HOOK.md"
                Copy-Item "$HOOK_PACK_DIR\clawmemory\handler.ts" "$HOOK_DIR\handler.ts"
                Write-Host "  Hook files copied to $HOOK_DIR"
            }
        } else {
            New-Item -ItemType Directory -Path $HOOK_DIR -Force | Out-Null
            Copy-Item "$HOOK_PACK_DIR\clawmemory\HOOK.md" "$HOOK_DIR\HOOK.md"
            Copy-Item "$HOOK_PACK_DIR\clawmemory\handler.ts" "$HOOK_DIR\handler.ts"
            Write-Host "  Hook files copied to $HOOK_DIR"
        }
    }

    Write-Host ""
    Write-Host "Enabling hook..."
    & openclaw hooks enable clawmemory

    Write-Host ""
    Write-Host "Checking hook status..."
    & openclaw hooks check
}

Write-Host ""
if (-not $env:CLAWMEMORY_URL) {
    Write-Host "WARNING: CLAWMEMORY_URL not set. Using default: http://localhost:8765"
    Write-Host "  To set: `$env:CLAWMEMORY_URL = 'http://localhost:8765'"
    Write-Host "  Or add to your shell profile for persistence."
}

if (-not $env:CLAWMEMORY_API_KEY) {
    Write-Host ""
    Write-Host "WARNING: CLAWMEMORY_API_KEY not set!"
    Write-Host "  1. Open ClawMemory at http://localhost:8765"
    Write-Host "  2. Go to Settings > API Keys"
    Write-Host "  3. Create a new key"
    Write-Host "  4. Set: `$env:CLAWMEMORY_API_KEY = 'cm_your_key_here'"
    Write-Host "  Or add to your shell profile for persistence."
}

Write-Host ""
Write-Host "Setup complete! Restart OpenClaw gateway to activate:"
Write-Host "  openclaw gateway restart"
Write-Host ""
