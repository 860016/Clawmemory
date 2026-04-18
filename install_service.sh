#!/bin/bash
# 将 ClawMemory 注册为 systemd 服务 (开机自启)
set -e

INSTALL_DIR="$(cd "$(dirname "$0")" && pwd)"
PORT=$(grep OPENCLAW_PORT "$INSTALL_DIR/backend/.env" 2>/dev/null | cut -d= -f2 || echo "8765")

cat > /etc/systemd/system/clawmemory.service << EOF
[Unit]
Description=ClawMemory - OpenClaw Memory Management
After=network.target

[Service]
Type=simple
User=$(whoami)
WorkingDirectory=$INSTALL_DIR/backend
ExecStart=$INSTALL_DIR/backend/venv/bin/uvicorn app.main:app --host 0.0.0.0 --port $PORT
Restart=always
RestartSec=5
Environment=PATH=$INSTALL_DIR/backend/venv/bin:/usr/local/bin:/usr/bin

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable clawmemory
systemctl start clawmemory

echo "✅ 服务已注册并启动"
echo "  状态: systemctl status clawmemory"
echo "  日志: journalctl -u clawmemory -f"
