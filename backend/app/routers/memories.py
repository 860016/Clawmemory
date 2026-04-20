from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session
from app.database import get_db
from app.middleware.auth import get_current_user
from app.schemas.memory import MemoryCreate, MemoryUpdate
from app.services.memory_service import MemoryService
from app.services.vector_service import vector_service
from app.services.fts_service import FtsService
from app.services.license_service import is_feature_enabled, decay_batch, get_decay_stats, get_stage_info
from app.config import settings
import json
from pathlib import Path

router = APIRouter(prefix="/api/v1/memories", tags=["memories"])


def _to_response(m) -> dict:
    return {
        "id": m.id, "layer": m.layer, "key": m.key, "value": m.value,
        "importance": m.importance, "access_count": m.access_count,
        "tags": json.loads(m.tags) if m.tags else [],
        "source": m.source, "is_encrypted": m.is_encrypted,
        "status": m.status or "active",
        "decay_stage": m.decay_stage or 0,
        "trashed_at": str(m.trashed_at) if m.trashed_at else None,
        "created_at": str(m.created_at), "updated_at": str(m.updated_at),
    }


@router.get("")
def list_memories(
    layer: str | None = None, 
    status: str | None = None,
    page: int = Query(1, ge=1), 
    size: int = Query(20, ge=1, le=100),
    _=Depends(get_current_user), db: Session = Depends(get_db),
):
    svc = MemoryService(db)
    items, total = svc.list_memories(layer, page, size, status)
    return {"items": [_to_response(m) for m in items], "total": total}


@router.post("", status_code=201)
def create_memory(req: MemoryCreate, _=Depends(get_current_user), db: Session = Depends(get_db)):
    svc = MemoryService(db)
    memory = svc.create(req.model_dump())
    vector_service.add_memory(1, memory.id, f"{memory.key}: {memory.value}", metadata={"layer": memory.layer})
    return _to_response(memory)


@router.get("/search/keyword")
def search_keyword(q: str, limit: int = Query(20, ge=1, le=50), _=Depends(get_current_user), db: Session = Depends(get_db)):
    fts = FtsService(db)
    return fts.search(1, q, limit)


@router.get("/search/semantic")
def search_semantic(q: str, limit: int = Query(10, ge=1, le=30), _=Depends(get_current_user), db: Session = Depends(get_db)):
    results = vector_service.search(1, q, limit)
    svc = MemoryService(db)
    enriched = []
    for r in results:
        memory = svc.get(r["memory_id"])
        if memory:
            enriched.append({
                "id": memory.id, "key": memory.key, "value": memory.value,
                "layer": memory.layer, "score": r["score"], "source": memory.source,
            })
    return enriched


@router.get("/{memory_id}")
def get_memory(memory_id: int, _=Depends(get_current_user), db: Session = Depends(get_db)):
    svc = MemoryService(db)
    memory = svc.get(memory_id)
    if not memory:
        raise HTTPException(status_code=404, detail="Memory not found")
    return _to_response(memory)


@router.put("/{memory_id}")
def update_memory(memory_id: int, req: MemoryUpdate, _=Depends(get_current_user), db: Session = Depends(get_db)):
    svc = MemoryService(db)
    memory = svc.update(memory_id, req.model_dump(exclude_none=True))
    if not memory:
        raise HTTPException(status_code=404, detail="Memory not found")
    vector_service.add_memory(1, memory.id, f"{memory.key}: {memory.value}", metadata={"layer": memory.layer})
    return _to_response(memory)


@router.delete("/{memory_id}")
def delete_memory(memory_id: int, _=Depends(get_current_user), db: Session = Depends(get_db)):
    svc = MemoryService(db)
    if not svc.delete(memory_id):
        raise HTTPException(status_code=404, detail="Memory not found")
    vector_service.delete_memory(1, memory_id)
    return {"message": "Memory deleted"}


# ========== 回收站 API ==========

@router.get("/trash")
def list_trash(
    page: int = Query(1, ge=1), 
    size: int = Query(20, ge=1, le=100),
    _=Depends(get_current_user), db: Session = Depends(get_db),
):
    """查看回收站中的记忆 - Pro 功能"""
    from app.services.license_service import is_feature_enabled
    
    # 检查是否是 Pro 用户
    if not is_feature_enabled("auto_decay"):
        raise HTTPException(status_code=403, detail="回收站是 Pro 功能，请先升级")
    
    svc = MemoryService(db)
    items, total = svc.list_memories(None, page, size, status="trashed")
    return {"items": [_to_response(m) for m in items], "total": total}


@router.post("/trash/{memory_id}/restore")
def restore_from_trash(memory_id: int, _=Depends(get_current_user), db: Session = Depends(get_db)):
    """从回收站恢复记忆 - Pro 功能"""
    from app.services.license_service import is_feature_enabled
    
    # 检查是否是 Pro 用户
    if not is_feature_enabled("auto_decay"):
        raise HTTPException(status_code=403, detail="回收站是 Pro 功能，请先升级")
    
    svc = MemoryService(db)
    memory = svc.restore(memory_id)
    if not memory:
        raise HTTPException(status_code=404, detail="Memory not found in trash")
    vector_service.add_memory(1, memory.id, f"{memory.key}: {memory.value}", metadata={"layer": memory.layer})
    return {"message": "Memory restored", "memory": _to_response(memory)}


@router.delete("/trash")
def empty_trash(_=Depends(get_current_user), db: Session = Depends(get_db)):
    """清空回收站 - Pro 功能"""
    from app.services.license_service import is_feature_enabled
    
    # 检查是否是 Pro 用户
    if not is_feature_enabled("auto_decay"):
        raise HTTPException(status_code=403, detail="回收站是 Pro 功能，请先升级")
    
    svc = MemoryService(db)
    count = svc.empty_trash()
    return {"message": f"Trash emptied, {count} memories deleted"}


# ========== 不重要记忆 API ==========

@router.get("/archived")
def list_archived(
    page: int = Query(1, ge=1), 
    size: int = Query(20, ge=1, le=100),
    _=Depends(get_current_user), db: Session = Depends(get_db),
):
    """查看不重要记忆（archived）- Pro 功能"""
    from app.services.license_service import is_feature_enabled
    
    # 检查是否是 Pro 用户
    if not is_feature_enabled("auto_decay"):
        raise HTTPException(status_code=403, detail="不重要记忆是 Pro 功能，请先升级")
    
    svc = MemoryService(db)
    items, total = svc.list_memories(None, page, size, status="archived")
    return {"items": [_to_response(m) for m in items], "total": total}


@router.post("/archived/{memory_id}/restore")
def restore_from_archived(memory_id: int, _=Depends(get_current_user), db: Session = Depends(get_db)):
    """从不重要记忆恢复为正常记忆 - Pro 功能"""
    from app.services.license_service import is_feature_enabled
    
    # 检查是否是 Pro 用户
    if not is_feature_enabled("auto_decay"):
        raise HTTPException(status_code=403, detail="不重要记忆是 Pro 功能，请先升级")
    
    svc = MemoryService(db)
    memory = svc.restore(memory_id)
    if not memory:
        raise HTTPException(status_code=404, detail="Memory not found in archived")
    return {"message": "Memory restored to active", "memory": _to_response(memory)}


# ========== 衰减统计 API ==========

@router.get("/decay/stats")
def get_memory_decay_stats(_=Depends(get_current_user), db: Session = Depends(get_db)):
    """获取记忆衰减统计信息 - Pro 功能"""
    from app.services.license_service import is_feature_enabled
    
    # 检查是否是 Pro 用户
    if not is_feature_enabled("auto_decay"):
        raise HTTPException(status_code=403, detail="记忆衰减是 Pro 功能，请先升级")
    
    from app.models.memory import Memory
    import time
    
    memories = db.query(Memory).filter(Memory.user_id == 1).all()
    memory_data = [
        {
            "id": m.id,
            "importance": m.importance,
            "last_accessed_at": m.last_accessed_at.timestamp() if m.last_accessed_at else 0,
            "status": m.status or "active",
            "trashed_at": m.trashed_at.timestamp() if m.trashed_at else 0,
        }
        for m in memories
    ]
    
    stats = get_decay_stats(memory_data, time.time())
    stage_info = get_stage_info()
    
    return {
        "stats": stats,
        "stage_info": stage_info,
    }


@router.get("/decay/info")
def get_decay_stage_info(_=Depends(get_current_user)):
    """获取衰减阶段配置信息 - Pro 功能"""
    from app.services.license_service import is_feature_enabled
    
    # 检查是否是 Pro 用户
    if not is_feature_enabled("auto_decay"):
        raise HTTPException(status_code=403, detail="记忆衰减是 Pro 功能，请先升级")
    
    return get_stage_info()


# ========== 衰减设置 API ==========

@router.get("/decay/settings")
def get_decay_settings(_=Depends(get_current_user)):
    """获取衰减设置 - Pro 功能"""
    from app.services.license_service import is_feature_enabled
    
    # 检查是否是 Pro 用户
    if not is_feature_enabled("auto_decay"):
        raise HTTPException(status_code=403, detail="记忆衰减是 Pro 功能，请先升级")
    
    decay_settings_file = settings.data_dir / "decay_settings.json"
    if decay_settings_file.exists():
        return json.loads(decay_settings_file.read_text())
    
    return {
        "enabled": True,
        "description": "Pro 功能：自动记忆衰减",
    }


@router.post("/decay/settings")
def update_decay_settings(
    enabled: bool = Query(...),
    _=Depends(get_current_user),
):
    """更新衰减设置 - Pro 功能"""
    from app.services.license_service import is_feature_enabled
    
    # 检查是否是 Pro 用户
    if not is_feature_enabled("auto_decay"):
        raise HTTPException(status_code=403, detail="记忆衰减是 Pro 功能，请先升级")
    
    decay_settings_file = settings.data_dir / "decay_settings.json"
    settings.data_dir.mkdir(parents=True, exist_ok=True)
    
    decay_settings = {
        "enabled": enabled,
        "updated_at": str(Path(__file__).stat().st_mtime),
    }
    
    decay_settings_file.write_text(json.dumps(decay_settings, indent=2))
    
    return {
        "message": f"Auto decay {'enabled' if enabled else 'disabled'}",
        "settings": decay_settings,
    }