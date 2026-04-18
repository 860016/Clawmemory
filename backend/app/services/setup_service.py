from pathlib import Path
from app.config import settings


def ensure_data_dirs():
    dirs = [
        settings.data_dir,
        settings.chroma_path,
        settings.data_dir / "backups",
        settings.data_dir / "exports",
        settings.base_dir / "keys",
    ]
    for d in dirs:
        d.mkdir(parents=True, exist_ok=True)
