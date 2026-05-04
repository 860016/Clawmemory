$ErrorActionPreference = "Stop"

Write-Host ""
Write-Host "ClawMemory — OpenClaw AGENTS.md Setup"
Write-Host "=========================================="

$SCRIPT_DIR = Split-Path -Parent $MyInvocation.MyCommand.Path

$CLAWMEMORY_URL = if ($env:CLAWMEMORY_URL) { $env:CLAWMEMORY_URL } else { "http://localhost:8765" }
$CLAWMEMORY_API_KEY = if ($env:CLAWMEMORY_API_KEY) { $env:CLAWMEMORY_API_KEY } else { "" }

if (-not $CLAWMEMORY_API_KEY) {
    Write-Host ""
    Write-Host "ERROR: CLAWMEMORY_API_KEY is required!" -ForegroundColor Red
    Write-Host "  1. Open ClawMemory at $CLAWMEMORY_URL"
    Write-Host "  2. Go to Settings > API Keys"
    Write-Host "  3. Create a new key"
    Write-Host "  4. Set: `$env:CLAWMEMORY_API_KEY = 'cm_your_key_here'"
    Write-Host ""
    exit 1
}

$AGENTS_MD = ""

if (Test-Path ".\AGENTS.md") {
    $AGENTS_MD = (Resolve-Path ".\AGENTS.md").Path
} elseif (Test-Path "$env:USERPROFILE\.openclaw\AGENTS.md") {
    $AGENTS_MD = "$env:USERPROFILE\.openclaw\AGENTS.md"
} else {
    $AGENTS_MD = (Resolve-Path ".").Path + "\AGENTS.md"
}

$TEMPLATE = Join-Path $SCRIPT_DIR "AGENTS.md.template"
if (-not (Test-Path $TEMPLATE)) {
    Write-Host "ERROR: AGENTS.md.template not found in $SCRIPT_DIR" -ForegroundColor Red
    exit 1
}

$CONTENT = Get-Content $TEMPLATE -Raw
$CONTENT = $CONTENT -replace '\{\{CLAWMEMORY_URL\}\}', $CLAWMEMORY_URL
$CONTENT = $CONTENT -replace '\{\{CLAWMEMORY_API_KEY\}\}', $CLAWMEMORY_API_KEY

$MARKER_START = "## 🧠 ClawMemory Auto-Record"
$MARKER_END = "<!-- END CLAWMEMORY -->"

if (Test-Path $AGENTS_MD) {
    $existing = Get-Content $AGENTS_MD -Raw
    if ($existing -match [regex]::Escape($MARKER_START)) {
        Write-Host "Updating existing ClawMemory section in $AGENTS_MD ..."
        $pattern = [regex]::Escape($MARKER_START) + "[\s\S]*?" + [regex]::Escape($MARKER_END) + "\s*"
        $updated = $existing -replace $pattern, ""
        $updated = $updated.TrimEnd() + "`n`n" + $CONTENT
        Set-Content -Path $AGENTS_MD -Value $updated -NoNewline
    } else {
        Write-Host "Appending ClawMemory instructions to $AGENTS_MD ..."
        $existing = $existing.TrimEnd() + "`n`n" + $CONTENT
        Set-Content -Path $AGENTS_MD -Value $existing -NoNewline
    }
} else {
    Write-Host "Creating $AGENTS_MD ..."
    Set-Content -Path $AGENTS_MD -Value $CONTENT -NoNewline
}

Write-Host ""
Write-Host "Done! ClawMemory instructions written to: $AGENTS_MD" -ForegroundColor Green
Write-Host ""
Write-Host "OpenClaw will now:"
Write-Host "  * Auto-save conversations to ClawMemory after each reply"
Write-Host "  * Search ClawMemory when recalling past context"
Write-Host "  * Maintain memories during idle time"
Write-Host ""
Write-Host "Restart OpenClaw to activate the new AGENTS.md instructions."
Write-Host ""
