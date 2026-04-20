import json
from datetime import datetime, timezone
from sqlalchemy.orm import Session
from sqlalchemy import or_
from app.models.memory import Memory


class MemoryService:
    def __init__(self, db: Session):
        self.db = db

    def create(self, data: dict) -> Memory:
        memory = Memory(
            user_id=1,
            layer=data["layer"],
            key=data["key"],
            value=data["value"],
            importance=data.get("importance", 0.5),
            tags=json.dumps(data.get("tags", [])),
            source=data.get("source", "manual"),
            status="active",
            decay_stage=0,
        )
        self.db.add(memory)
        self.db.commit()
        self.db.refresh(memory)
        return memory

    def get(self, memory_id: int) -> Memory | None:
        return self.db.query(Memory).filter(Memory.id == memory_id).first()

    def list_memories(
        self, 
        layer: str | None = None, 
        page: int = 1, 
        size: int = 20,
        status: str | None = None,
    ) -> tuple[list[Memory], int]:
        query = self.db.query(Memory).filter(Memory.user_id == 1)
        
        if layer:
            query = query.filter(Memory.layer == layer)
        
        if status:
            query = query.filter(Memory.status == status)
        else:
            query = query.filter(Memory.status != "trashed")
        
        total = query.count()
        items = query.order_by(Memory.updated_at.desc()).offset((page - 1) * size).limit(size).all()
        return items, total

    def update(self, memory_id: int, data: dict) -> Memory | None:
        memory = self.get(memory_id)
        if not memory:
            return None
        
        if "value" in data and data["value"] is not None:
            memory.value = data["value"]
        if "importance" in data and data["importance"] is not None:
            memory.importance = data["importance"]
        if "tags" in data and data["tags"] is not None:
            memory.tags = json.dumps(data["tags"])
        if "layer" in data and data["layer"] is not None:
            memory.layer = data["layer"]
        
        memory.last_accessed_at = datetime.now(timezone.utc)
        memory.access_count += 1
        
        self.db.commit()
        self.db.refresh(memory)
        return memory

    def delete(self, memory_id: int) -> bool:
        memory = self.get(memory_id)
        if not memory:
            return False
        self.db.delete(memory)
        self.db.commit()
        return True

    def restore(self, memory_id: int) -> Memory | None:
        """恢复记忆（从 archived 或 trashed 状态恢复为 active）"""
        memory = self.get(memory_id)
        if not memory:
            return None
        
        memory.status = "active"
        memory.decay_stage = 0
        memory.trashed_at = None
        memory.last_accessed_at = datetime.now(timezone.utc)
        memory.access_count += 1
        
        self.db.commit()
        self.db.refresh(memory)
        return memory

    def empty_trash(self) -> int:
        """清空回收站，返回删除的数量"""
        count = self.db.query(Memory).filter(
            Memory.user_id == 1,
            Memory.status == "trashed",
        ).count()
        
        self.db.query(Memory).filter(
            Memory.user_id == 1,
            Memory.status == "trashed",
        ).delete()
        
        self.db.commit()
        return count

    def search_keywords(self, query: str, limit: int = 20) -> list[dict]:
        """FTS5 full-text search"""
        results = self.db.execute(
            "SELECT m.id, m.key, m.value, m.layer, m.source, m.importance, f.rank as score "
            "FROM memories_fts f JOIN memories m ON m.id = f.rowid "
            "WHERE f.memories_fts MATCH :query AND m.user_id = 1 AND m.status != 'trashed' "
            "ORDER BY f.rank LIMIT :limit",
            {"query": query, "limit": limit}
        ).fetchall()
        return [dict(r._mapping) for r in results]

    def bulk_create(self, memories: list[dict]) -> int:
        """Batch insert memories, returns count created"""
        count = 0
        for data in memories:
            memory = Memory(
                user_id=1,
                layer=data["layer"],
                key=data["key"],
                value=data["value"],
                importance=data.get("importance", 0.5),
                tags=json.dumps(data.get("tags", [])),
                source=data.get("source", "manual"),
                status="active",
                decay_stage=0,
            )
            self.db.add(memory)
            count += 1
        self.db.commit()
        return count