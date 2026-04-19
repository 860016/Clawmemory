"""ClawMemory Password Reset Token Generator

Usage:
    cd /path/to/clawmemory/backend
    python -m app.utils.reset_password

This generates a one-time reset token valid for 15 minutes.
The token file is stored in data/reset_token.json.
After using the token, it is automatically deleted.
"""

import sys
import json
import secrets
from datetime import datetime, timezone, timedelta
from pathlib import Path


def generate_reset_token():
    # Detect data directory
    data_dir = Path("/app/data")
    if not data_dir.exists():
        # Local dev mode
        data_dir = Path(__file__).resolve().parent.parent.parent / "data"
    
    data_dir.mkdir(parents=True, exist_ok=True)
    token_path = data_dir / "reset_token.json"
    
    # Generate token
    token = secrets.token_urlsafe(24)
    expires_at = datetime.now(timezone.utc) + timedelta(minutes=15)
    
    token_data = {
        "token": token,
        "expires_at": expires_at.isoformat(),
        "created_at": datetime.now(timezone.utc).isoformat(),
    }
    
    token_path.write_text(json.dumps(token_data, indent=2))
    
    # Print instructions with actual installation path
    backend_dir = Path(__file__).resolve().parent.parent.parent
    
    print("=" * 60)
    print("ClawMemory Password Reset Token")
    print("=" * 60)
    print()
    print(f"  Token: {token}")
    print(f"  Expires: 15 minutes")
    print()
    print("  To reset your password, open in browser:")
    print(f"  http://localhost:8765/#/reset-password?token={token}")
    print()
    print("  Or use curl:")
    print(f'  curl -X POST http://localhost:8765/api/v1/auth/reset-password')
    print()
    print("  NOTE: This token can only be used once.")
    print("  If expired, run this command again to generate a new one.")
    print()
    print(f"  Installation path: {backend_dir}")
    print("=" * 60)


if __name__ == "__main__":
    generate_reset_token()
