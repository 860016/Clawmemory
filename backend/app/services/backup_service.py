import json
import zipfile
from datetime import datetime
from pathlib import Path
from sqlalchemy.orm import Session
from app.models.backup import Backup
from app.models.memory import Memory
from app.config import settings


class BackupService:
    def __init__(self, db: Session):
        self.db = db

    def create_backup(self, user_id: int, notes: str | None = None) -> Backup:
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        filename = f"backup_{user_id}_{timestamp}.zip"
        backup_dir = settings.data_dir / "backups"
        backup_dir.mkdir(parents=True, exist_ok=True)
        filepath = backup_dir / filename

        # Export all memories
        memories = self.db.query(Memory).filter(Memory.user_id == user_id).all()
        export_data = []
        for m in memories:
            export_data.append({
                "layer": m.layer, "key": m.key, "value": m.value,
                "importance": m.importance, "tags": m.tags, "source": m.source,
            })

        # Create zip with JSON
        with zipfile.ZipFile(filepath, 'w', zipfile.ZIP_DEFLATED) as zf:
            zf.writestr("memories.json", json.dumps(export_data, indent=2, ensure_ascii=False))

        backup = Backup(
            user_id=user_id,
            filename=filename,
            file_path=str(filepath),
            file_size=filepath.stat().st_size,
            memory_count=len(memories),
            backup_type="manual",
            notes=notes,
        )
        self.db.add(backup)
        self.db.commit()
        self.db.refresh(backup)
        return backup

    def list_backups(self, user_id: int) -> list[Backup]:
        return self.db.query(Backup).filter(Backup.user_id == user_id).order_by(Backup.created_at.desc()).all()

    def restore_backup(self, backup_id: int, user_id: int) -> int:
        backup = self.db.query(Backup).filter(Backup.id == backup_id, Backup.user_id == user_id).first()
        if not backup:
            return 0
        with zipfile.ZipFile(backup.file_path) as zf:
            data = json.loads(zf.read("memories.json"))
        count = 0
        for item in data:
            memory = Memory(user_id=user_id, **item)
            self.db.add(memory)
            count += 1
        self.db.commit()
        return count

    def download_backup(self, backup_id: int, user_id: int) -> Path | None:
        """Return the file path for download. Caller handles streaming."""
        backup = self.db.query(Backup).filter(Backup.id == backup_id, Backup.user_id == user_id).first()
        if not backup:
            return None
        path = Path(backup.file_path)
        if not path.exists():
            return None
        return path

    def upload_backup(self, user_id: int, file_path: Path, original_filename: str) -> Backup:
        """Import an uploaded backup file and register it."""
        backup_dir = settings.data_dir / "backups"
        backup_dir.mkdir(parents=True, exist_ok=True)

        # Copy uploaded file to backup dir
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        filename = f"uploaded_{user_id}_{timestamp}.zip"
        dest = backup_dir / filename

        # Validate it's a valid zip with memories.json
        with zipfile.ZipFile(file_path, 'r') as zf:
            if "memories.json" not in zf.namelist():
                raise ValueError("Invalid backup file: missing memories.json")
            data = json.loads(zf.read("memories.json"))
            memory_count = len(data) if isinstance(data, list) else 0

        # Copy file to backup directory
        import shutil
        shutil.copy2(file_path, dest)

        backup = Backup(
            user_id=user_id,
            filename=filename,
            file_path=str(dest),
            file_size=dest.stat().st_size,
            memory_count=memory_count,
            backup_type="uploaded",
            notes=f"Uploaded from {original_filename}",
        )
        self.db.add(backup)
        self.db.commit()
        self.db.refresh(backup)
        return backup

    def delete_backup(self, backup_id: int, user_id: int) -> bool:
        backup = self.db.query(Backup).filter(Backup.id == backup_id, Backup.user_id == user_id).first()
        if not backup:
            return False
        Path(backup.file_path).unlink(missing_ok=True)
        self.db.delete(backup)
        self.db.commit()
        return True
