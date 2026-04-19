from sqlalchemy.orm import Session
from sqlalchemy import text


class FtsService:
    def __init__(self, db: Session):
        self.db = db

    def search(self, user_id: int, query: str, limit: int = 20) -> list[dict]:
        # Sanitize FTS5 query: wrap terms in quotes to avoid syntax errors
        safe_q = " ".join(f'"{t.strip()}"' for t in query.split() if t.strip())
        if not safe_q:
            return []
        results = self.db.execute(text(
            "SELECT m.id, m.key, m.value, m.layer, m.source, m.importance, "
            "m.access_count, m.tags, m.created_at, m.updated_at, f.rank "
            "FROM memories_fts f JOIN memories m ON m.id = f.rowid "
            "WHERE f.memories_fts MATCH :q AND m.user_id = :uid "
            "ORDER BY f.rank LIMIT :lim"
        ), {"q": safe_q, "uid": user_id, "lim": limit}).fetchall()
        import json
        return [{
            "id": r.id, "key": r.key, "value": r.value,
            "layer": r.layer, "source": r.source,
            "importance": r.importance or 0.5,
            "access_count": r.access_count or 0,
            "tags": json.loads(r.tags) if r.tags else [],
            "created_at": str(r.created_at) if r.created_at else None,
            "updated_at": str(r.updated_at) if r.updated_at else None,
        } for r in results]

    def rebuild_index(self):
        """Rebuild FTS5 index (run periodically for consistency)"""
        self.db.execute(text("INSERT INTO memories_fts(memories_fts) VALUES('rebuild')"))
        self.db.commit()
