import tempfile
import os
from app.services.vector_service import VectorService


def test_vector_service_unavailable():
    """Test that vector service degrades gracefully when chromadb is not installed"""
    svc = VectorService()
    # If chromadb is not installed, search should return empty list
    if not svc.available:
        results = svc.search(user_id=999, query="test")
        assert results == []
        assert svc.count(user_id=999) == 0


def test_vector_service_no_error_without_chromadb():
    """Ensure no errors when chromadb is not available"""
    svc = VectorService()
    # These should not raise exceptions even if chromadb is not installed
    svc.add_memory(user_id=999, memory_id=1, content="test")
    svc.delete_memory(user_id=999, memory_id=1)
    svc.delete_user_collection(user_id=999)
