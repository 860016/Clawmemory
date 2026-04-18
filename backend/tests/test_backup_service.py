import tempfile
import os
import zipfile
import json


def test_backup_service_crud():
    """Test BackupService create/list/restore/delete"""
    from app.database import SessionLocal, init_db
    from app.services.backup_service import BackupService
    from app.models.memory import Memory
    from app.models.user import User
    from app.utils.security import hash_password

    # Use temp directory
    tmpdir = tempfile.mkdtemp()
    os.environ["OPENCLAW_DATA_DIR"] = tmpdir

    init_db()
    db = SessionLocal()

    try:
        # Create test user
        user = User(username="testbackup", email="backup@test.com", hashed_password=hash_password("test123"))
        db.add(user)
        db.commit()
        db.refresh(user)

        # Add some memories
        for i in range(3):
            mem = Memory(user_id=user.id, key=f"test_key_{i}", value=f"test_value_{i}", layer="short_term")
            db.add(mem)
        db.commit()

        # Create backup
        svc = BackupService(db)
        backup = svc.create_backup(user.id, notes="test backup")
        assert backup is not None
        assert backup.memory_count == 3
        assert backup.notes == "test backup"

        # List backups
        backups = svc.list_backups(user.id)
        assert len(backups) >= 1

        # Verify zip file content
        with zipfile.ZipFile(backup.file_path) as zf:
            data = json.loads(zf.read("memories.json"))
            assert len(data) == 3
            assert data[0]["key"] == "test_key_0"

        # Delete backup
        assert svc.delete_backup(backup.id, user.id) is True
        assert len(svc.list_backups(user.id)) == 0
    finally:
        db.close()


def test_backup_service_not_found():
    """Test BackupService with non-existent backup"""
    from app.database import SessionLocal, init_db
    from app.services.backup_service import BackupService

    init_db()
    db = SessionLocal()
    try:
        svc = BackupService(db)
        assert svc.restore_backup(99999, 99999) == 0
        assert svc.delete_backup(99999, 99999) is False
    finally:
        db.close()
