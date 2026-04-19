from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from pathlib import Path
from app.database import get_db
from app.middleware.auth import get_current_user
from app.services.openclaw_memory_scanner import (
    scan_openclaw_memories,
    scan_agent_memories,
    scan_openclaw_sqlite,
    _detect_openclaw_dir,
)
from app.services.memory_service import MemoryService
from app.services.vector_service import vector_service
from pydantic import BaseModel

router = APIRouter(prefix="/api/v1/openclaw-memories", tags=["openclaw-memories"])


class ImportRequest(BaseModel):
    agent_name: str
    memory_dir: str | None = None  # Override auto-detect
    target_agent_id: int | None = None  # Our system agent ID
    layer: str = "knowledge"  # Default layer for imported memories
    skip_existing: bool = True  # Skip memories with same key


@router.get("/scan")
def scan_memories(_=Depends(get_current_user)):
    """Scan OpenClaw memory directory and return discovered agents + file counts."""
    openclaw_dir = _detect_openclaw_dir()
    if not openclaw_dir:
        return {"found": False, "openclaw_dir": None, "agents": []}

    results = scan_openclaw_memories(openclaw_dir)
    # Don't return full memory content in scan — just counts
    summary = []
    for r in results:
        summary.append({
            "agent_name": r["agent_name"],
            "layout": r["layout"],
            "memory_dir": r["memory_dir"],
            "files": r["files"],
        })

    return {
        "found": True,
        "openclaw_dir": str(openclaw_dir),
        "agents": summary,
    }


@router.get("/scan/{agent_name}")
def scan_agent(
    agent_name: str,
    _=Depends(get_current_user),
):
    """Scan a specific agent's memories and return preview (first 50)."""
    openclaw_dir = _detect_openclaw_dir()
    if not openclaw_dir:
        raise HTTPException(status_code=404, detail="OpenClaw directory not found")

    results = scan_openclaw_memories(openclaw_dir)
    agent_data = None
    for r in results:
        if r["agent_name"] == agent_name:
            agent_data = r
            break

    if not agent_data:
        raise HTTPException(status_code=404, detail=f"Agent '{agent_name}' not found")

    # Preview: first 50 memories
    preview = agent_data["memories"][:50]
    return {
        "agent_name": agent_data["agent_name"],
        "memory_dir": agent_data["memory_dir"],
        "total": agent_data["files"],
        "preview": preview,
    }


@router.post("/import")
def import_memories(
    req: ImportRequest,
    _=Depends(get_current_user),
    db: Session = Depends(get_db),
):
    """Import OpenClaw memories into our system."""
    openclaw_dir = _detect_openclaw_dir()
    if not openclaw_dir:
        raise HTTPException(status_code=404, detail="OpenClaw directory not found")

    # Find agent's memories
    results = scan_openclaw_memories(openclaw_dir)
    agent_data = None
    for r in results:
        if r["agent_name"] == req.agent_name:
            agent_data = r
            break

    if not agent_data:
        raise HTTPException(status_code=404, detail=f"Agent '{req.agent_name}' not found in OpenClaw")

    memories = agent_data["memories"]
    if not memories:
        return {"imported": 0, "skipped": 0, "errors": 0}

    svc = MemoryService(db)
    imported = 0
    skipped = 0
    errors = 0

    # Get existing keys for skip logic
    existing_keys = set()
    if req.skip_existing:
        items, _ = svc.list_memories(page=1, size=10000)
        existing_keys = {m.key for m in items}

    for mem in memories:
        try:
            key = mem["key"]
            if req.skip_existing and key in existing_keys:
                skipped += 1
                continue

            data = {
                "key": key,
                "value": mem["value"],
                "layer": req.layer,
                "importance": mem.get("importance", 0.5),
                "tags": mem.get("tags", []),
                "source": "openclaw_import",
            }
            memory = svc.create(data)

            # Sync to ChromaDB
            try:
                vector_service.add_memory(
                    1,
                    memory.id,
                    f"{data['key']}: {data['value']}",
                    metadata={"layer": data["layer"], "source": "openclaw_import"},
                )
            except Exception:
                pass  # ChromaDB not available, skip

            imported += 1
        except Exception as e:
            errors += 1
            import logging
            logging.getLogger(__name__).warning(f"Import error for key={mem.get('key')}: {e}")

    return {
        "imported": imported,
        "skipped": skipped,
        "errors": errors,
        "total_source": len(memories),
    }


@router.get("/scan-sqlite")
def scan_sqlite_memories(_=Depends(get_current_user)):
    """Scan official OpenClaw memory SQLite database."""
    result = scan_openclaw_sqlite()
    return result
