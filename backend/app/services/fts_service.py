from sqlalchemy.orm import Session
from sqlalchemy import text


class FtsService:
    def __init__(self, db: Session):
        self.db = db

    def search(self, user_id: int, query: str, limit: int = 20) -> list[dict]:
        results = self.db.execute(text(
            "SELECT m.id, m.key, m.value, m.layer, m.source, f.rank "
            "FROM memories_fts f JOIN memories m ON m.id = f.rowid "
            "WHERE f.memories_fts MATCH :q AND m.user_id = :uid "
            "ORDER BY f.rank LIMIT :lim"
        ), {"q": query, "uid": user_id, "lim": limit}).fetchall()
        return [dict(r._mapping) for r in results]

    def rebuild_index(self):
        """Rebuild FTS5 index (run periodically for consistency)"""
        self.db.execute(text("INSERT INTO memories_fts(memories_fts) VALUES('rebuild')"))
        self.db.commit()
