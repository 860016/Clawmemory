from fastapi import Depends, HTTPException, status
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
from app.utils.security import decode_access_token

security = HTTPBearer(auto_error=False)


async def get_current_user(
    credentials: HTTPAuthorizationCredentials | None = Depends(security),
):
    """单用户模式：始终要求认证。如果设置了密码则验证密码，如果未设置密码则拒绝访问（引导用户设置密码）。"""
    # 未设置密码时也要求认证 — 引导用户先设置密码
    if credentials is None:
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail="Authentication required")

    payload = decode_access_token(credentials.credentials)
    if payload is None:
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail="Invalid token")

    return True


# 兼容旧代码 — set-password 端点使用可选认证
async def get_optional_user(
    credentials: HTTPAuthorizationCredentials | None = Depends(security),
):
    """可选认证：用于 set-password 等端点，未设置密码时允许访问。"""
    from app.config import settings

    # 没有设置密码，无需认证
    if not settings.access_password:
        return True

    # 设置了密码，需要验证
    if credentials is None:
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail="Authentication required")

    payload = decode_access_token(credentials.credentials)
    if payload is None:
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail="Invalid token")

    return True
