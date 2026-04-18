from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from pathlib import Path
from app.database import get_db
from app.middleware.auth import get_current_user
from app.services.memory_service import MemoryService
from app.services.vector_service import vector_service
from app.utils.file_watcher import file_watcher
from app.config import settings
from pydantic import BaseModel

router = APIRouter(prefix="/api/v1/file-watcher", tags=["file-watcher"])


class WatchRequest(BaseModel):
    watch_dir: str


@router.post("/start")
def start_watching(
    req: WatchRequest,
    _=Depends(get_current_user),
    db: Session = Depends(get_db),
):
    """Start watching a directory for memory file changes."""
    watch_path = Path(req.watch_dir)
    if not watch_path.exists():
        raise HTTPException(status_code=400, detail="Watch directory does not exist")

    memory_service = MemoryService(db)
    file_watcher.start_watching(
        user_id=1,
        agent_id=None,
        watch_dir=watch_path,
        memory_service=memory_service,
        vector_service=vector_service,
    )
    return {"message": f"Started watching: {req.watch_dir}"}


@router.post("/stop")
def stop_watching(_=Depends(get_current_user)):
    """Stop watching a directory."""
    file_watcher.stop_watching(1)
    return {"message": "Stopped watching"}


@router.get("/status")
def get_watcher_status(_=Depends(get_current_user)):
    """Get file watcher status."""
    return {
        "running": file_watcher.observer.is_alive(),
        "active_watches": len(file_watcher._watches),
    }
