from app.websocket.manager import ConnectionManager
from app.websocket.session_map import SessionMap


def test_session_map_register():
    sm = SessionMap()
    sm.register(ws_id=1001, user_id=1, agent_id=1)
    info = sm.get(1001)
    assert info is not None
    assert info.user_id == 1
    assert info.agent_id == 1


def test_session_map_set_chat():
    sm = SessionMap()
    sm.register(ws_id=1001, user_id=1, agent_id=1)
    sm.set_chat_session(ws_id=1001, chat_session_id=42)
    info = sm.get(1001)
    assert info.chat_session_id == 42


def test_session_map_remove():
    sm = SessionMap()
    sm.register(ws_id=1001, user_id=1, agent_id=1)
    sm.remove(1001)
    assert sm.get(1001) is None


def test_session_map_get_by_user():
    sm = SessionMap()
    sm.register(ws_id=1001, user_id=1, agent_id=1)
    sm.register(ws_id=1002, user_id=1, agent_id=2)
    sm.register(ws_id=1003, user_id=2, agent_id=3)
    results = sm.get_by_user(user_id=1)
    assert len(results) == 2


def test_manager_online_status():
    m = ConnectionManager()
    assert m.is_online(user_id=1) is False
