#!/usr/bin/env bash
set -euo pipefail

echo ""
echo "ClawMemory — OpenClaw AGENTS.md Setup"
echo "=========================================="

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

CLAWMEMORY_URL="${CLAWMEMORY_URL:-http://localhost:8765}"
CLAWMEMORY_API_KEY="${CLAWMEMORY_API_KEY:-}"

if [ -z "$CLAWMEMORY_API_KEY" ]; then
    echo ""
    echo "ERROR: CLAWMEMORY_API_KEY is required!"
    echo "  1. Open ClawMemory at $CLAWMEMORY_URL"
    echo "  2. Go to Settings > API Keys"
    echo "  3. Create a new key"
    echo "  4. Set: export CLAWMEMORY_API_KEY='cm_your_key_here'"
    echo ""
    exit 1
fi

AGENTS_MD=""

if [ -f "AGENTS.md" ]; then
    AGENTS_MD="$(pwd)/AGENTS.md"
elif [ -f "$HOME/.openclaw/AGENTS.md" ]; then
    AGENTS_MD="$HOME/.openclaw/AGENTS.md"
else
    AGENTS_MD="$(pwd)/AGENTS.md"
fi

TEMPLATE="$SCRIPT_DIR/AGENTS.md.template"
if [ ! -f "$TEMPLATE" ]; then
    echo "ERROR: AGENTS.md.template not found in $SCRIPT_DIR"
    exit 1
fi

CONTENT=$(sed \
    -e "s|{{CLAWMEMORY_URL}}|$CLAWMEMORY_URL|g" \
    -e "s|{{CLAWMEMORY_API_KEY}}|$CLAWMEMORY_API_KEY|g" \
    "$TEMPLATE")

MARKER="# ClawMemory — AI Memory Backend"
MARKER_END="<!-- END CLAWMEMORY -->"

if [ -f "$AGENTS_MD" ]; then
    if grep -q "$MARKER" "$AGENTS_MD"; then
        echo "Updating existing ClawMemory section in $AGENTS_MD ..."
        TEMP_FILE=$(mktemp)
        awk -v marker="$MARKER" -v marker_end="$MARKER_END" -v content="$CONTENT" '
            $0 == marker { print content; found=1; next }
            $0 == marker_end { found=0; next }
            !found { print }
        ' "$AGENTS_MD" > "$TEMP_FILE"
        echo "$MARKER_END" >> "$TEMP_FILE"
        mv "$TEMP_FILE" "$AGENTS_MD"
    else
        echo "Appending ClawMemory instructions to $AGENTS_MD ..."
        echo "" >> "$AGENTS_MD"
        echo "$CONTENT" >> "$AGENTS_MD"
        echo "$MARKER_END" >> "$AGENTS_MD"
    fi
else
    echo "Creating $AGENTS_MD ..."
    echo "$CONTENT" > "$AGENTS_MD"
    echo "$MARKER_END" >> "$AGENTS_MD"
fi

echo ""
echo "✅ Done! ClawMemory instructions written to: $AGENTS_MD"
echo ""
echo "OpenClaw will now:"
echo "  • Auto-save conversations to ClawMemory after each reply"
echo "  • Search ClawMemory when recalling past context"
echo "  • Maintain memories during idle time"
echo ""
echo "Restart OpenClaw to activate the new AGENTS.md instructions."
echo ""
