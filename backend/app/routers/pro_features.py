"""Pro Features API — Memory Decay, Conflict Resolution, Token Routing"""
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from app.database import get_db
from app.middleware.auth import get_current_user
from app.services.license_service import is_feature_enabled
from app.services import license_service as core

router = APIRouter(prefix="/api/v1/pro", tags=["pro-features"])


# ========== Memory Decay ==========

@router.get("/decay/stats")
def get_decay_stats(_=Depends(get_current_user), db: Session = Depends(get_db)):
    """Get memory decay statistics (requires auto_decay or decay_report feature)"""
    if not is_feature_enabled("auto_decay") and not is_feature_enabled("decay_report"):
        raise HTTPException(status_code=403, detail="Pro feature: memory decay analysis")
    from app.models.memory import Memory
    memories = db.query(Memory).filter(Memory.user_id == 1).all()
    memory_data = [
        {"id": m.id, "importance": m.importance, "last_accessed_at": m.last_accessed_at.timestamp() if m.last_accessed_at else 0}
        for m in memories
    ]
    return core.get_decay_stats(memory_data)


@router.post("/decay/apply")
def apply_decay(_=Depends(get_current_user), db: Session = Depends(get_db)):
    """Apply decay to all memories and update importance (requires auto_decay)"""
    if not is_feature_enabled("auto_decay"):
        raise HTTPException(status_code=403, detail="Pro feature: auto decay")
    import time
    from app.models.memory import Memory
    memories = db.query(Memory).filter(Memory.user_id == 1).all()
    memory_data = [
        {"id": m.id, "importance": m.importance, "last_accessed_at": m.last_accessed_at.timestamp() if m.last_accessed_at else 0}
        for m in memories
    ]
    results = core.decay_batch(memory_data)
    updated = 0
    pruned = 0
    for r in results:
        m = db.query(Memory).filter(Memory.id == r["memory_id"]).first()
        if m:
            m.importance = r["new_importance"]
            updated += 1
            if r["should_prune"]:
                pruned += 1
    db.commit()
    return {"updated": updated, "pruned_candidates": pruned}


# ========== Conflict Resolution ==========

@router.get("/conflicts/scan")
def scan_conflicts(_=Depends(get_current_user), db: Session = Depends(get_db)):
    """Scan memories for conflicts (requires conflict_scan)"""
    if not is_feature_enabled("conflict_scan"):
        raise HTTPException(status_code=403, detail="Pro feature: conflict scan")
    from app.models.memory import Memory
    memories = db.query(Memory).filter(Memory.user_id == 1).all()
    memory_dicts = [
        {"id": m.id, "key": m.key, "value": m.value, "layer": m.layer, "source": m.source, "importance": m.importance}
        for m in memories
    ]
    conflicts = core.scan_for_conflicts(memory_dicts)
    summary = core.get_conflict_summary(conflicts)
    return {"conflicts": conflicts, "summary": summary}


@router.post("/conflicts/resolve/{conflict_index}")
def resolve_conflict_api(conflict_index: int, strategy: str = None, _=Depends(get_current_user), db: Session = Depends(get_db)):
    """Resolve a specific conflict (requires conflict_merge)"""
    if not is_feature_enabled("conflict_merge"):
        raise HTTPException(status_code=403, detail="Pro feature: conflict merge")
    from app.models.memory import Memory
    memories = db.query(Memory).filter(Memory.user_id == 1).all()
    memory_dicts = [
        {"id": m.id, "key": m.key, "value": m.value, "layer": m.layer, "source": m.source, "importance": m.importance}
        for m in memories
    ]
    conflicts = core.scan_for_conflicts(memory_dicts)
    if conflict_index >= len(conflicts):
        raise HTTPException(status_code=404, detail="Conflict not found")
    result = core.resolve_conflict(conflicts[conflict_index], strategy)
    return result


# ========== Token Routing ==========

@router.post("/token/route")
def token_route(message: str, context_length: int = 0, _=Depends(get_current_user)):
    """Estimate complexity and route to appropriate model (requires smart_router)"""
    if not is_feature_enabled("smart_router"):
        raise HTTPException(status_code=403, detail="Pro feature: smart router")
    result = core.estimate_complexity(message, context_length)
    return {"complexity": result}
