from fastapi import APIRouter, Depends, HTTPException, UploadFile, File
from fastapi.responses import FileResponse
from sqlalchemy.orm import Session
from app.database import get_db
from app.middleware.auth import get_current_user
from app.schemas.backup import BackupCreate
from app.services.backup_service import BackupService
from pathlib import Path
import tempfile

router = APIRouter(prefix="/api/v1/backups", tags=["backups"])


@router.get("")
def list_backups(_=Depends(get_current_user), db: Session = Depends(get_db)):
    svc = BackupService(db)
    return svc.list_backups(1)


@router.post("")
def create_backup(data: BackupCreate, _=Depends(get_current_user), db: Session = Depends(get_db)):
    svc = BackupService(db)
    return svc.create_backup(1, data.notes)


@router.get("/{backup_id}/download")
def download_backup(backup_id: int, _=Depends(get_current_user), db: Session = Depends(get_db)):
    svc = BackupService(db)
    path = svc.download_backup(backup_id, 1)
    if not path:
        raise HTTPException(status_code=404, detail="Backup not found")
    return FileResponse(path=str(path), filename=path.name, media_type="application/zip")


@router.post("/upload")
async def upload_backup(file: UploadFile = File(...), _=Depends(get_current_user), db: Session = Depends(get_db)):
    if not file.filename or not file.filename.endswith(".zip"):
        raise HTTPException(status_code=400, detail="Only .zip files are accepted")
    svc = BackupService(db)
    with tempfile.NamedTemporaryFile(suffix=".zip", delete=False) as tmp:
        content = await file.read()
        tmp.write(content)
        tmp_path = Path(tmp.name)
    try:
        backup = svc.upload_backup(1, tmp_path, file.filename)
        return backup
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))
    finally:
        tmp_path.unlink(missing_ok=True)


@router.post("/{backup_id}/restore")
def restore_backup(backup_id: int, _=Depends(get_current_user), db: Session = Depends(get_db)):
    svc = BackupService(db)
    count = svc.restore_backup(backup_id, 1)
    if count == 0:
        raise HTTPException(status_code=404, detail="Backup not found")
    return {"message": f"Restored {count} memories"}


@router.delete("/{backup_id}")
def delete_backup(backup_id: int, _=Depends(get_current_user), db: Session = Depends(get_db)):
    svc = BackupService(db)
    if not svc.delete_backup(backup_id, 1):
        raise HTTPException(status_code=404, detail="Backup not found")
    return {"message": "Backup deleted"}
