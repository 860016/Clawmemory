#!/usr/bin/env bash
set -euo pipefail

echo ""
echo "ClawMemory OpenClaw Hook - Setup"
echo "====================================="

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
HOOK_PACK_DIR="$SCRIPT_DIR"

if [ ! -f "$HOOK_PACK_DIR/clawmemory/HOOK.md" ]; then
    echo "ERROR: Invalid hook pack structure. Missing clawmemory/HOOK.md"
    exit 1
fi

if [ ! -f "$HOOK_PACK_DIR/clawmemory/handler.ts" ]; then
    echo "ERROR: Invalid hook pack structure. Missing clawmemory/handler.ts"
    exit 1
fi

if [ ! -f "$HOOK_PACK_DIR/package.json" ]; then
    echo "ERROR: Missing package.json in hook pack root"
    exit 1
fi

if command -v openclaw &>/dev/null; then
    echo ""
    echo "Installing hook pack via openclaw plugins install..."

    if openclaw plugins install "$HOOK_PACK_DIR"; then
        echo ""
        echo "Enabling hook..."
        openclaw hooks enable clawmemory

        echo ""
        echo "Checking hook status..."
        openclaw hooks check
    else
        echo ""
        echo "WARNING: openclaw plugins install failed."
        echo "Falling back to manual installation..."

        HOOK_DIR="$HOME/.openclaw/hooks/clawmemory"
        rm -rf "$HOOK_DIR"
        mkdir -p "$HOOK_DIR"
        cp "$HOOK_PACK_DIR/clawmemory/HOOK.md" "$HOOK_DIR/HOOK.md"
        cp "$HOOK_PACK_DIR/clawmemory/handler.ts" "$HOOK_DIR/handler.ts"

        echo "  Hook files copied to $HOOK_DIR"
        echo ""
        echo "  IMPORTANT: You must manually enable the hook:"
        echo "    openclaw hooks enable clawmemory"
        echo "    openclaw gateway restart"
    fi
else
    echo ""
    echo "WARNING: openclaw CLI not found on PATH."
    echo "  Please install OpenClaw first: https://openclaw.ai"
    echo ""
    echo "Falling back to manual installation..."

    HOOK_DIR="$HOME/.openclaw/hooks/clawmemory"
    if [ -d "$HOOK_DIR" ]; then
        read -p "  Hook already exists at $HOOK_DIR. Overwrite? (y/N) " reply
        if [ "$reply" != "y" ] && [ "$reply" != "Y" ]; then
            echo "  Aborted."
            exit 0
        fi
        rm -rf "$HOOK_DIR"
    fi

    mkdir -p "$HOOK_DIR"
    cp "$HOOK_PACK_DIR/clawmemory/HOOK.md" "$HOOK_DIR/HOOK.md"
    cp "$HOOK_PACK_DIR/clawmemory/handler.ts" "$HOOK_DIR/handler.ts"

    echo ""
    echo "  Hook files copied to $HOOK_DIR"
    echo ""
    echo "  IMPORTANT: You must manually enable the hook:"
    echo "    openclaw hooks enable clawmemory"
    echo "    openclaw gateway restart"
fi

echo ""
if [ -z "${CLAWMEMORY_URL:-}" ]; then
    echo "WARNING: CLAWMEMORY_URL not set. Using default: http://localhost:8765"
    echo "  To set: export CLAWMEMORY_URL='http://localhost:8765'"
    echo "  Or add to your shell profile for persistence."
fi

if [ -z "${CLAWMEMORY_API_KEY:-}" ]; then
    echo ""
    echo "WARNING: CLAWMEMORY_API_KEY not set!"
    echo "  1. Open ClawMemory at http://localhost:8765"
    echo "  2. Go to Settings > API Keys"
    echo "  3. Create a new key"
    echo "  4. Set: export CLAWMEMORY_API_KEY='cm_your_key_here'"
    echo "  Or add to your shell profile for persistence."
fi

echo ""
echo "Setup complete! Restart OpenClaw gateway to activate:"
echo "  openclaw gateway restart"
echo ""
