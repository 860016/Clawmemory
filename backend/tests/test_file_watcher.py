from app.utils.file_watcher import MemoryFileHandler


def test_handler_init():
    handler = MemoryFileHandler(user_id=1, agent_id=1, memory_service=None, vector_service=None)
    assert handler.user_id == 1
    assert handler.agent_id == 1
