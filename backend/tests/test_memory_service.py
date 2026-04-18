from app.services.auth_service import AuthService
from app.services.memory_service import MemoryService


def setup_user(db):
    auth = AuthService(db)
    return auth.create_user("testuser", "password123")


def test_create_memory(db):
    user = setup_user(db)
    svc = MemoryService(db)
    memory = svc.create(user.id, {"layer": "preference", "key": "language", "value": "Chinese"})
    assert memory.id is not None
    assert memory.layer == "preference"


def test_list_memories(db):
    user = setup_user(db)
    svc = MemoryService(db)
    svc.create(user.id, {"layer": "preference", "key": "lang", "value": "CN"})
    svc.create(user.id, {"layer": "knowledge", "key": "python", "value": "3.10"})
    items, total = svc.list_memories(user.id)
    assert total == 2


def test_list_by_layer(db):
    user = setup_user(db)
    svc = MemoryService(db)
    svc.create(user.id, {"layer": "preference", "key": "lang", "value": "CN"})
    svc.create(user.id, {"layer": "knowledge", "key": "python", "value": "3.10"})
    items, total = svc.list_memories(user.id, layer="preference")
    assert total == 1
    assert items[0].layer == "preference"


def test_update_memory(db):
    user = setup_user(db)
    svc = MemoryService(db)
    memory = svc.create(user.id, {"layer": "preference", "key": "theme", "value": "dark"})
    updated = svc.update(memory.id, user.id, {"value": "light"})
    assert updated.value == "light"


def test_delete_memory(db):
    user = setup_user(db)
    svc = MemoryService(db)
    memory = svc.create(user.id, {"layer": "short_term", "key": "temp", "value": "x"})
    assert svc.delete(memory.id, user.id) is True
    assert svc.get(memory.id, user.id) is None
