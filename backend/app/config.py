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


def _detect_openclaw_dir() -> Path | None:
    """Find ~/.openclaw/ directory.
    
    Priority: OPENCLAW_STATE_DIR > OPENCLAW_HOME/.openclaw > ~/.openclaw
    """
    home = Path.home()
    candidates = [
        Path(os.environ.get("OPENCLAW_STATE_DIR", "")),
        # OPENCLAW_HOME overrides the home dir entirely
        Path(os.environ.get("OPENCLAW_HOME", home)) / ".openclaw" if os.environ.get("OPENCLAW_HOME") else None,
        home / ".openclaw",
    ]
    for p in candidates:
        if p and p.exists():
            return p
    return None


def _read_openclaw_gateway_config() -> tuple[str, str]:
    """Auto-detect Gateway URL and API Key from ~/.openclaw/openclaw.json.
    
    OpenClaw config format (JSON5):
      gateway.port: 18789
      gateway.auth.token: "your-token"
      gateway.auth.mode: "token" | "none" | "password"
    
    Returns (gateway_url, api_key). Falls back to defaults if not found.
    """
    openclaw_dir = _detect_openclaw_dir()
    if not openclaw_dir:
        return "http://localhost:18789", ""

    config_path = openclaw_dir / "openclaw.json"
    if not config_path.exists():
        return "http://localhost:18789", ""

    gateway_url = ""
    api_key = ""
    port = 18789

    try:
        # Try json5 first (supports comments + trailing commas)
        try:
            import json5
            with open(config_path, "r", encoding="utf-8") as f:
                data = json5.load(f)
        except ImportError:
            # Fallback: strip comments and trailing commas, then use stdlib json
            import re
            with open(config_path, "r", encoding="utf-8") as f:
                raw = f.read()
            # Remove single-line comments // ...
            raw = re.sub(r'//.*?$', '', raw, flags=re.MULTILINE)
            # Remove multi-line comments /* ... */
            raw = re.sub(r'/\*.*?\*/', '', raw, flags=re.DOTALL)
            # Remove trailing commas before } or ]
            raw = re.sub(r',\s*([}\]])', r'\1', raw)
            data = __import__("json").loads(raw)

        if isinstance(data, dict):
            gw = data.get("gateway", {})
            if isinstance(gw, dict):
                # Port
                if "port" in gw:
                    port = int(gw["port"])
                # Auth token
                auth = gw.get("auth", {})
                if isinstance(auth, dict):
                    api_key = auth.get("token", "") or auth.get("password", "")
                # Remote URL (for remote mode)
                remote = gw.get("remote", {})
                if isinstance(remote, dict) and remote.get("url"):
                    gateway_url = remote["url"].replace("ws://", "http://").replace("wss://", "https://")
    except Exception as e:
        logger.debug(f"Failed to parse openclaw.json: {e}")

    # Build URL from port if not set by remote
    if not gateway_url:
        gateway_url = f"http://localhost:{port}"
    if not api_key:
        api_key = ""

    logger.info(f"Auto-detected OpenClaw Gateway: {gateway_url} (from {config_path})")
    return gateway_url, api_key


# Auto-detect Gateway config at module load time
_detected_gateway_url, _detected_api_key = _read_openclaw_gateway_config()


class Settings(BaseSettings):
    # Paths
    base_dir: Path = Path(__file__).resolve().parent.parent
<<<<<<< HEAD
    data_dir: Path = _detect_data_dir()
    db_path: Path = data_dir / "openclaw.db"
=======
    data_dir: Path = base_dir / "data"
    db_path: Path = data_dir / "clawmemory.db"
>>>>>>> fb055c7 (feat: v3.0 - Wiki知识库 + 科技感UI + i18n + Rust PyO3核心 + Pro功能)
    chroma_path: Path = data_dir / "chroma"

    # Auth — 单用户密码
    access_password: str = ""  # 空则不需要密码
    secret_key: str = "clawmemory-secret-change-me"
    algorithm: str = "HS256"
    access_token_expire_minutes: int = 1440  # 24h

<<<<<<< HEAD
    # OpenClaw Gateway (auto-detected from ~/.openclaw/openclaw.json, overridable via .env)
    openclaw_gateway_url: str = _detected_gateway_url
    openclaw_api_key: str = _detected_api_key

=======
>>>>>>> fb055c7 (feat: v3.0 - Wiki知识库 + 科技感UI + i18n + Rust PyO3核心 + Pro功能)
    # Server
    host: str = "0.0.0.0"
    port: int = 8765
    debug: bool = False

    # License
    license_server_url: str = "https://license.yoursite.com"
    rsa_public_key_path: Path = base_dir / "keys" / "public.pem"

    # CORS (前端预构建后同源部署，"*" 可用；开发模式需指定 localhost)
    cors_origins: list[str] = ["*"]

    class Config:
        env_file = ".env"
        env_prefix = "CLAWMEMORY_"


settings = Settings()
