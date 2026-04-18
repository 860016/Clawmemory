import logging
from app.config import settings

logger = logging.getLogger(__name__)


class VectorService:
    """ChromaDB vector search service.
    
    Gracefully degrades when chromadb is not installed.
    """

    def __init__(self):
        self._client = None
        self._available = False
        try:
            import chromadb
            self._available = True
        except ImportError:
            logger.warning("chromadb not installed, vector search disabled")

    @property
    def available(self) -> bool:
        return self._available

    def _get_client(self):
        if not self._available:
            raise RuntimeError("chromadb not available")
        import chromadb
        from chromadb.config import Settings as ChromaSettings
        if self._client is None:
            self._client = chromadb.PersistentClient(
                path=str(settings.chroma_path),
                settings=ChromaSettings(anonymized_telemetry=False),
            )
        return self._client

    def get_collection(self, user_id: int):
        client = self._get_client()
        name = f"user_{user_id}_memories"
        return client.get_or_create_collection(
            name=name,
            metadata={"hnsw:space": "cosine"},
        )

    def add_memory(self, user_id: int, memory_id: int, content: str, metadata: dict | None = None):
        if not self._available:
            return
        collection = self.get_collection(user_id)
        collection.upsert(
            ids=[str(memory_id)],
            documents=[content],
            metadatas=[metadata or {}],
        )

    def search(self, user_id: int, query: str, n_results: int = 10) -> list[dict]:
        if not self._available:
            return []
        collection = self.get_collection(user_id)
        if collection.count() == 0:
            return []
        results = collection.query(
            query_texts=[query],
            n_results=min(n_results, collection.count()),
        )
        items = []
        if results["ids"] and results["ids"][0]:
            for i, doc_id in enumerate(results["ids"][0]):
                items.append({
                    "memory_id": int(doc_id),
                    "content": results["documents"][0][i] if results["documents"] else "",
                    "score": 1 - results["distances"][0][i] if results["distances"] else 0,
                    "metadata": results["metadatas"][0][i] if results["metadatas"] else {},
                })
        return items

    def delete_memory(self, user_id: int, memory_id: int):
        if not self._available:
            return
        collection = self.get_collection(user_id)
        collection.delete(ids=[str(memory_id)])

    def delete_user_collection(self, user_id: int):
        if not self._available:
            return
        try:
            self._get_client().delete_collection(f"user_{user_id}_memories")
        except Exception:
            pass

    def count(self, user_id: int) -> int:
        if not self._available:
            return 0
        collection = self.get_collection(user_id)
        return collection.count()


vector_service = VectorService()
