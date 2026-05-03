#!/bin/bash
set -e

echo "🧠 ClawMemory OpenClaw Hook - Setup"
echo "====================================="

HOOK_DIR="$HOME/.openclaw/hooks/clawmemory"

if [ -d "$HOOK_DIR" ]; then
  echo "⚠️  Hook already exists at $HOOK_DIR"
  read -p "Overwrite? (y/N) " -n 1 -r
  echo
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Aborted."
    exit 0
  fi
  rm -rf "$HOOK_DIR"
fi

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
mkdir -p "$HOOK_DIR"
cp "$SCRIPT_DIR/HOOK.md" "$HOOK_DIR/HOOK.md"
cp "$SCRIPT_DIR/handler.ts" "$HOOK_DIR/handler.ts"
cp "$SCRIPT_DIR/package.json" "$HOOK_DIR/package.json"

echo ""
echo "✅ Hook installed to $HOOK_DIR"
echo ""

if [ -z "$CLAWMEMORY_URL" ]; then
  echo "⚠️  CLAWMEMORY_URL not set. Using default: http://localhost:8765"
  echo "   To set: export CLAWMEMORY_URL=http://localhost:8765"
fi

if [ -z "$CLAWMEMORY_API_KEY" ]; then
  echo "⚠️  CLAWMEMORY_API_KEY not set!"
  echo "   1. Open ClawMemory at http://localhost:8765"
  echo "   2. Go to Settings → API Keys"
  echo "   3. Create a new key"
  echo "   4. Set: export CLAWMEMORY_API_KEY=cm_your_key_here"
fi

echo ""
echo "To enable the hook:"
echo "  openclaw hooks enable clawmemory"
echo ""
echo "To restart OpenClaw:"
echo "  openclaw gateway restart"
