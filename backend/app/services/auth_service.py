from app.utils.security import hash_password, verify_password, create_access_token


class AuthService:
    """单用户模式的认证服务 — 不再依赖 User 模型"""

    @staticmethod
    def authenticate(password: str, hashed_password: str) -> dict | None:
        if not verify_password(password, hashed_password):
            return None
        token = create_access_token({"sub": "admin", "role": "admin"})
        return {"access_token": token, "token_type": "bearer", "role": "admin"}

    @staticmethod
    def hash_password(password: str) -> str:
        return hash_password(password)
