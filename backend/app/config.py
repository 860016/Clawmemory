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


def _read_version() -> str:
    """从项目根目录 VERSION 文件读取统一版本号"""
    version_file = Path(__file__).resolve().parent.parent.parent / "VERSION"
    if version_file.exists():
        return version_file.read_text().strip()
    return "0.0.0"


APP_VERSION = _read_version()


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

    # LLM / AI
    openai_api_key: str = ""
    openai_base_url: str = "https://api.openai.com/v1"
    llm_model: str = "gpt-4o-mini"

    # License
    license_server_url: str = "https://auth.bestu.top"
    rsa_public_key_path: Path = base_dir / "keys" / "public.pem"

    # Pro module download URLs (fixed, configured by admin)
    pro_download_url: str = ""
    pro_fallback_urls: list[str] = []

    # CORS (前端预构建后同源部署，"*" 可用；开发模式需指定 localhost)
    cors_origins: list[str] = ["*"]

    class Config:
        env_file = ".env"
        env_prefix = "CLAWMEMORY_"


settings = Settings()
