from fastapi import APIRouter, Depends, HTTPException
from app.config import settings
from app.utils.security import verify_password, hash_password, create_access_token
from app.middleware.auth import get_optional_user
from pydantic import BaseModel

router = APIRouter(prefix="/api/v1/auth", tags=["auth"])


def _write_password_to_env(hashed: str):
    """Write hashed password to .env file"""
    env_path = settings.base_dir / ".env"
    lines = []
    if env_path.exists():
        lines = env_path.read_text().splitlines()

    found = False
    new_lines = []
    for line in lines:
        if line.startswith("CLAWMEMORY_ACCESS_PASSWORD="):
            new_lines.append(f"CLAWMEMORY_ACCESS_PASSWORD={hashed}")
            found = True
        else:
            new_lines.append(line)

    if not found:
        new_lines.append(f"CLAWMEMORY_ACCESS_PASSWORD={hashed}")

    env_path.write_text("\n".join(new_lines) + "\n")


class LoginRequest(BaseModel):
    password: str


class SetPasswordRequest(BaseModel):
    password: str


class ChangePasswordRequest(BaseModel):
    old_password: str
    new_password: str


@router.get("/init-status")
def init_status():
    """检查是否设置了访问密码"""
    return {"password_set": bool(settings.access_password), "password_required": True}


@router.post("/login")
def login(req: LoginRequest):
    """单用户登录 — 仅验证密码"""
    if not settings.access_password:
        raise HTTPException(status_code=400, detail="No password set, access is open")

    if not verify_password(req.password, settings.access_password):
        raise HTTPException(status_code=401, detail="Invalid password")

    token = create_access_token({"sub": "admin", "role": "admin"})
    return {"access_token": token, "token_type": "bearer"}


@router.post("/set-password")
def set_password(req: SetPasswordRequest, _=Depends(get_optional_user)):
    """首次设置或修改访问密码"""
    if len(req.password) < 4:
        raise HTTPException(status_code=400, detail="Password must be at least 4 characters")

    hashed = hash_password(req.password)
    _write_password_to_env(hashed)
    settings.access_password = hashed

    token = create_access_token({"sub": "admin", "role": "admin"})
    return {"access_token": token, "token_type": "bearer", "message": "Password set successfully"}


@router.post("/change-password")
def change_password(req: ChangePasswordRequest, _=Depends(get_optional_user)):
    """修改密码 — 必须验证旧密码"""
    if not settings.access_password:
        raise HTTPException(status_code=400, detail="No password set yet")

    if not verify_password(req.old_password, settings.access_password):
        raise HTTPException(status_code=401, detail="Current password is incorrect")

    if len(req.new_password) < 4:
        raise HTTPException(status_code=400, detail="Password must be at least 4 characters")

    hashed = hash_password(req.new_password)
    _write_password_to_env(hashed)
    settings.access_password = hashed
    return {"message": "Password changed successfully"}


@router.post("/reset-password")
def reset_password():
    """Reset password by removing it from .env — for forgot password flow.
    Only works when accessed from the server itself (no auth required).
    After reset, user must set a new password on next login.
    
    NOTE: This endpoint requires a valid reset token generated via CLI:
      python -m app.utils.reset_password
    """
    env_path = settings.base_dir / ".env"
    # Check for reset token file
    from pathlib import Path
    token_path = Path(settings.data_dir) / "reset_token.json"
    if not token_path.exists():
        raise HTTPException(status_code=403, detail="No reset token found. Run 'python -m app.utils.reset_password' on the server to generate one.")
    
    import json
    from datetime import datetime, timezone
    try:
        token_data = json.loads(token_path.read_text())
    except Exception:
        raise HTTPException(status_code=403, detail="Invalid reset token file")
    
    # Check if token is expired
    expires_at = datetime.fromisoformat(token_data.get("expires_at", ""))
    if datetime.now(timezone.utc) > expires_at:
        token_path.unlink(missing_ok=True)
        raise HTTPException(status_code=403, detail="Reset token has expired. Generate a new one.")
    
    # Token is valid — reset password
    if env_path.exists():
        lines = env_path.read_text().splitlines()
        new_lines = [line for line in lines if not line.startswith("CLAWMEMORY_ACCESS_PASSWORD=")]
        env_path.write_text("\n".join(new_lines) + "\n")
    settings.access_password = ""
    # Remove used token
    token_path.unlink(missing_ok=True)
    return {"message": "Password reset successful. Please set a new password."}


@router.get("/me")
def get_me(_=Depends(get_optional_user)):
    return {"username": "admin", "role": "admin"}
