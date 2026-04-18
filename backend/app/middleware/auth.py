from fastapi import Depends, HTTPException, status
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
from app.utils.security import decode_access_token

security = HTTPBearer(auto_error=False)


async def get_optional_user(
    credentials: HTTPAuthorizationCredentials | None = Depends(security),
):
    """单用户模式：可选认证。如果设置了密码则验证，否则直接放行。"""
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


# 兼容旧代码的别名
get_current_user = get_optional_user
