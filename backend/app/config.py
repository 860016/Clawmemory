import os
import logging
from pathlib import Path
from pydantic_settings import BaseSettings

logger = logging.getLogger(__name__)


def _detect_data_dir() -> Path:
    """Docker: /app/data; local dev: backend/data"""
    docker_data = Path("/app/data")
    if docker_data.exists():
        return docker_data
    return Path(__file__).resolve().parent.parent / "data"


class Settings(BaseSettings):
    # Paths
    base_dir: Path = Path(__file__).resolve().parent.parent
    data_dir: Path = _detect_data_dir()
    db_path: Path = data_dir / "clawmemory.db"
    chroma_path: Path = data_dir / "chroma"

    # Auth — 单用户密码
    access_password: str = ""  # 空则不需要密码
    secret_key: str = "clawmemory-secret-change-me"
    algorithm: str = "HS256"
    access_token_expire_minutes: int = 1440  # 24h

    # Server
    host: str = "0.0.0.0"
    port: int = 8765
    debug: bool = False

    # License
    license_server_url: str = "https://auth.bestu.top"
    rsa_public_key_path: Path = base_dir / "keys" / "public.pem"

    # CORS (前端预构建后同源部署，"*" 可用；开发模式需指定 localhost)
    cors_origins: list[str] = ["*"]

    class Config:
        env_file = ".env"
        env_prefix = "CLAWMEMORY_"


settings = Settings()
