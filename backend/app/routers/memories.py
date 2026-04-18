from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session
from app.database import get_db
from app.middleware.auth import get_current_user
from app.schemas.memory import MemoryCreate, MemoryUpdate
from app.services.memory_service import MemoryService
from app.services.vector_service import vector_service
from app.services.fts_service import FtsService
from app.services.license_service import is_feature_enabled
import json

router = APIRouter(prefix="/api/v1/memories", tags=["memories"])


def _to_response(m) -> dict:
    return {
        "id": m.id, "layer": m.layer, "key": m.key, "value": m.value,
        "importance": m.importance, "access_count": m.access_count,
        "tags": json.loads(m.tags) if m.tags else [],
        "source": m.source, "is_encrypted": m.is_encrypted,
        "created_at": str(m.created_at), "updated_at": str(m.updated_at),
    }


@router.get("")
def list_memories(
    layer: str | None = None, page: int = Query(1, ge=1), size: int = Query(20, ge=1, le=100),
    _=Depends(get_current_user), db: Session = Depends(get_db),
):
    svc = MemoryService(db)
    items, total = svc.list_memories(layer, page, size)
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
