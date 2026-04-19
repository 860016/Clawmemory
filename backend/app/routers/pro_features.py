"""Pro Features API — Memory Decay, Conflict Resolution, Token Routing, AI Extract, Auto Graph, Backup Schedule"""
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from pydantic import BaseModel
from typing import Optional
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
    stats = core.get_decay_stats(memory_data)
    # Add per-memory decay details
    decay_details = []
    for m in memories:
        age_seconds = 0
        if m.last_accessed_at:
            import time
            age_seconds = max(time.time() - m.last_accessed_at.timestamp(), 0)
        elif m.updated_at:
            import time
            age_seconds = max(time.time() - m.updated_at.timestamp(), 0)
        new_importance = core.calculate_decay(m.importance, age_seconds)
        decay_details.append({
            "id": m.id, "key": m.key, "layer": m.layer,
            "original_importance": m.importance,
            "current_importance": round(new_importance, 4),
            "age_days": round(age_seconds / 86400, 1),
            "should_prune": core.should_prune(new_importance),
        })
    return {**stats, "memories": decay_details}


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
    auto_deleted = 0
    for r in results:
        m = db.query(Memory).filter(Memory.id == r["memory_id"]).first()
        if m:
            if r["should_prune"]:
                # Auto-delete very low importance memories
                db.delete(m)
                auto_deleted += 1
            else:
                m.importance = r["new_importance"]
                updated += 1
            pruned += 1 if r["should_prune"] else 0
    db.commit()
    return {"updated": updated, "pruned_candidates": pruned, "auto_deleted": auto_deleted}


@router.post("/reinforce/{memory_id}")
def reinforce_memory(memory_id: int, _=Depends(get_current_user), db: Session = Depends(get_db)):
    """Reinforce a memory's importance (requires reinforce feature)"""
    if not is_feature_enabled("reinforce"):
        raise HTTPException(status_code=403, detail="Pro feature: reinforce")
    from app.models.memory import Memory
    m = db.query(Memory).filter(Memory.id == memory_id, Memory.user_id == 1).first()
    if not m:
        raise HTTPException(status_code=404, detail="Memory not found")
    old_importance = m.importance
    m.importance = core.reinforce(m.importance)
    from datetime import datetime, timezone
    m.last_accessed_at = datetime.now(timezone.utc)
    db.commit()
    return {"id": m.id, "old_importance": old_importance, "new_importance": m.importance}


@router.get("/prune-suggest")
def prune_suggest(_=Depends(get_current_user), db: Session = Depends(get_db)):
    """Get pruning suggestions for low-importance memories (requires prune_suggest)"""
    if not is_feature_enabled("prune_suggest"):
        raise HTTPException(status_code=403, detail="Pro feature: prune suggest")
    import time
    from app.models.memory import Memory
    memories = db.query(Memory).filter(Memory.user_id == 1).all()
    now = time.time()
    suggestions = []
    for m in memories:
        age = now - (m.last_accessed_at.timestamp() if m.last_accessed_at else (m.updated_at.timestamp() if m.updated_at else 0))
        decayed = core.calculate_decay(m.importance, age)
        if core.should_prune(decayed):
            suggestions.append({
                "id": m.id, "key": m.key, "layer": m.layer,
                "current_importance": m.importance,
                "decayed_importance": round(decayed, 4),
                "age_days": round(age / 86400, 1),
                "reason": "below_threshold" if decayed < 0.05 else "low_importance",
            })
    suggestions.sort(key=lambda x: x["decayed_importance"])
    return {"suggestions": suggestions, "total": len(suggestions)}


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
    # If merge strategy, actually apply it
    if result.get("action") == "merge" and strategy == "merge":
        conflict = conflicts[conflict_index]
        # Keep both but update the older one
        memory_b = db.query(Memory).filter(Memory.id == conflict.get("memory_b_id")).first()
        if memory_b:
            merged_value = f"{conflict.get('value_a', '')} | {conflict.get('value_b', '')}"
            memory_b.value = merged_value
            db.commit()
    return result


# ========== Token Routing ==========

@router.post("/token/route")
def token_route(message: str, context_length: int = 0, _=Depends(get_current_user)):
    """Estimate complexity and route to appropriate model (requires smart_router)"""
    if not is_feature_enabled("smart_router"):
        raise HTTPException(status_code=403, detail="Pro feature: smart router")
    complexity = core.estimate_complexity(message, context_length)
    # Simple model routing based on complexity
    if complexity > 0.7:
        model = "gpt-4"
        reason = "high_complexity"
    elif complexity > 0.4:
        model = "gpt-4o-mini"
        reason = "medium_complexity"
    else:
        model = "gpt-3.5-turbo"
        reason = "low_complexity"
    return {
        "complexity": round(complexity, 4),
        "selected_model": model,
        "routing_reason": reason,
        "estimated_tokens": len(message.split()) * 2,  # rough estimate
    }


@router.get("/token/stats")
def token_stats(_=Depends(get_current_user), db: Session = Depends(get_db)):
    """Get token usage statistics (requires token_stats)"""
    if not is_feature_enabled("token_stats"):
        raise HTTPException(status_code=403, detail="Pro feature: token stats")
    from app.models.memory import Memory
    memories = db.query(Memory).filter(Memory.user_id == 1).all()
    total_tokens = sum(len((m.key + " " + m.value).split()) for m in memories)
    layer_tokens = {}
    for m in memories:
        layer_tokens.setdefault(m.layer, 0)
        layer_tokens[m.layer] += len((m.key + " " + m.value).split())
    return {
        "total_estimated_tokens": total_tokens,
        "total_memories": len(memories),
        "by_layer": layer_tokens,
        "avg_tokens_per_memory": round(total_tokens / max(len(memories), 1), 1),
    }


# ========== AI Extract ==========

class AIExtractRequest(BaseModel):
    memory_ids: Optional[list[int]] = None

@router.post("/ai/extract")
def ai_extract(req: AIExtractRequest, _=Depends(get_current_user), db: Session = Depends(get_db)):
    """Extract entities and relations from memories using NLP heuristics (requires ai_extract)"""
    if not is_feature_enabled("ai_extract"):
        raise HTTPException(status_code=403, detail="Pro feature: AI extract")

    from app.models.memory import Memory
    from app.models.knowledge import Entity, Relation
    import re

    if req.memory_ids:
        memories = db.query(Memory).filter(Memory.id.in_(req.memory_ids), Memory.user_id == 1).all()
    else:
        memories = db.query(Memory).filter(Memory.user_id == 1).all()

    # Heuristic extraction: find capitalized terms, quoted terms, key patterns
    extracted_entities = []
    extracted_relations = []

    # Common patterns for entity extraction
    entity_patterns = [
        r'"([^"]+)"',           # Quoted terms
        r'「([^」]+)」',         # Chinese quoted terms
        r'【([^】]+)】',         # Chinese bracketed terms
    ]

    existing_entities = {e.name.lower(): e for e in db.query(Entity).filter(Entity.user_id == 1).all()}

    for m in memories:
        text = f"{m.key} {m.value}"
        found_terms = set()

        for pattern in entity_patterns:
            for match in re.finditer(pattern, text):
                term = match.group(1).strip()
                if len(term) >= 2 and len(term) <= 50:
                    found_terms.add(term)

        # Extract from key patterns like "user_name", "project_name"
        key_parts = re.split(r'[_\-\s]+', m.key)
        for part in key_parts:
            if len(part) >= 3 and not part.lower() in ('the', 'and', 'for', 'with'):
                found_terms.add(part)

        for term in found_terms:
            if term.lower() not in existing_entities:
                # Determine entity type heuristically
                entity_type = "concept"
                if any(kw in term.lower() for kw in ['project', 'proj', 'app', '系统', '项目']):
                    entity_type = "project"
                elif any(kw in term.lower() for kw in ['person', 'user', 'name', '用户', '人']):
                    entity_type = "person"
                elif any(kw in term.lower() for kw in ['tool', 'lib', 'framework', '工具', '库']):
                    entity_type = "tool"
                elif any(kw in term.lower() for kw in ['event', 'incident', '事件']):
                    entity_type = "event"

                entity = Entity(
                    user_id=1,
                    name=term,
                    entity_type=entity_type,
                    description=f"Auto-extracted from memory: {m.key}",
                )
                db.add(entity)
                extracted_entities.append({"name": term, "type": entity_type, "source_memory": m.key})
                existing_entities[term.lower()] = entity

    db.commit()

    # Now try to find relations between entities in the same memory
    for m in memories:
        text = f"{m.key} {m.value}".lower()
        matched = [e for name, e in existing_entities.items() if name in text]
        for i in range(len(matched)):
            for j in range(i + 1, len(matched)):
                # Check if relation already exists
                exists = db.query(Relation).filter(
                    Relation.source_id == matched[i].id,
                    Relation.target_id == matched[j].id,
                ).first()
                if not exists:
                    rel = Relation(
                        user_id=1,
                        source_id=matched[i].id,
                        target_id=matched[j].id,
                        relation_type="co_occurs",
                        description=f"Both mentioned in: {m.key}",
                    )
                    db.add(rel)
                    extracted_relations.append({
                        "source": matched[i].name,
                        "target": matched[j].name,
                        "type": "co_occurs",
                    })

    db.commit()
    return {
        "entities_extracted": len(extracted_entities),
        "relations_extracted": len(extracted_relations),
        "entities": extracted_entities,
        "relations": extracted_relations,
    }


# ========== Auto Graph ==========

class AutoGraphRequest(BaseModel):
    overwrite: bool = False

@router.post("/auto-graph")
def auto_graph(req: AutoGraphRequest, _=Depends(get_current_user), db: Session = Depends(get_db)):
    """Auto-generate knowledge graph from memories (requires auto_graph)"""
    if not is_feature_enabled("auto_graph"):
        raise HTTPException(status_code=403, detail="Pro feature: auto graph")

    from app.models.memory import Memory
    from app.models.knowledge import Entity, Relation
    import re

    # If overwrite, clear existing graph
    if req.overwrite:
        db.query(Relation).filter(Relation.user_id == 1).delete()
        db.query(Entity).filter(Entity.user_id == 1).delete()
        db.commit()

    memories = db.query(Memory).filter(Memory.user_id == 1).all()
    existing_entities = {e.name.lower(): e for e in db.query(Entity).filter(Entity.user_id == 1).all()}

    # Group memories by layer for relationship inference
    layer_entities = {}
    entities_created = 0
    relations_created = 0

    for m in memories:
        # Each memory's key becomes an entity
        entity_name = m.key.replace('_', ' ').replace('-', ' ').title()
        layer = m.layer

        if entity_name.lower() not in existing_entities:
            entity_type_map = {
                "preference": "concept",
                "knowledge": "concept",
                "short_term": "event",
                "private": "concept",
            }
            entity = Entity(
                user_id=1,
                name=entity_name,
                entity_type=entity_type_map.get(layer, "concept"),
                description=m.value[:200] if m.value else "",
            )
            db.add(entity)
            db.flush()
            existing_entities[entity_name.lower()] = entity
            entities_created += 1
            layer_entities.setdefault(layer, []).append(entity)

    db.commit()

    # Create relations between entities in the same layer
    for layer, ents in layer_entities.items():
        for i in range(len(ents)):
            for j in range(i + 1, min(i + 5, len(ents))):  # Limit to nearby entities
                exists = db.query(Relation).filter(
                    Relation.source_id == ents[i].id,
                    Relation.target_id == ents[j].id,
                ).first()
                if not exists:
                    rel_type = "related_to"
                    if layer == "preference":
                        rel_type = "prefers_alongside"
                    elif layer == "knowledge":
                        rel_type = "related_to"
                    elif layer == "short_term":
                        rel_type = "co_occurs"
                    rel = Relation(
                        user_id=1,
                        source_id=ents[i].id,
                        target_id=ents[j].id,
                        relation_type=rel_type,
                        description=f"Same layer: {layer}",
                    )
                    db.add(rel)
                    relations_created += 1

    db.commit()
    return {
        "entities_created": entities_created,
        "relations_created": relations_created,
        "total_entities": db.query(Entity).filter(Entity.user_id == 1).count(),
        "total_relations": db.query(Relation).filter(Relation.user_id == 1).count(),
    }


# ========== Auto Backup Schedule ==========

@router.get("/backup/schedule")
def get_backup_schedule(_=Depends(get_current_user)):
    """Get auto backup schedule settings (requires auto_backup)"""
    if not is_feature_enabled("auto_backup"):
        raise HTTPException(status_code=403, detail="Pro feature: auto backup")
    from app.config import settings
    import json
    from pathlib import Path

    schedule_file = settings.data_dir / "backup_schedule.json"
    if schedule_file.exists():
        schedule = json.loads(schedule_file.read_text())
    else:
        schedule = {"enabled": False, "interval_hours": 24}
    return schedule


class BackupScheduleRequest(BaseModel):
    enabled: bool
    interval_hours: int = 24

@router.post("/backup/schedule")
def set_backup_schedule(req: BackupScheduleRequest, _=Depends(get_current_user)):
    """Set auto backup schedule (requires auto_backup)"""
    if not is_feature_enabled("auto_backup"):
        raise HTTPException(status_code=403, detail="Pro feature: auto backup")
    from app.config import settings
    import json

    schedule = {"enabled": req.enabled, "interval_hours": req.interval_hours}
    schedule_file = settings.data_dir / "backup_schedule.json"
    schedule_file.write_text(json.dumps(schedule, indent=2))
    return schedule
