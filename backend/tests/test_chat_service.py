from app.services.auth_service import AuthService
from app.services.chat_service import ChatService
from app.models.agent import Agent


def setup_user_and_agent(db):
    auth = AuthService(db)
    user = auth.create_user("testuser", "password123")
    agent = Agent(user_id=user.id, name="main", is_default=True)
    db.add(agent)
    db.commit()
    db.refresh(agent)
    return user, agent


def test_create_session(db):
    user, agent = setup_user_and_agent(db)
    svc = ChatService(db)
    session = svc.create_session(user.id, agent.id)
    assert session.id is not None
    assert session.title == "New Chat"


def test_save_and_get_messages(db):
    user, agent = setup_user_and_agent(db)
    svc = ChatService(db)
    session = svc.create_session(user.id, agent.id)
    svc.save_message(session.id, "user", "Hello")
    svc.save_message(session.id, "assistant", "Hi there!")
    messages = svc.get_messages(session.id, user.id)
    assert len(messages) == 2
    assert messages[0].role == "user"
    assert messages[1].role == "assistant"


def test_list_sessions(db):
    user, agent = setup_user_and_agent(db)
    svc = ChatService(db)
    svc.create_session(user.id, agent.id, "Chat 1")
    svc.create_session(user.id, agent.id, "Chat 2")
    sessions = svc.get_sessions(user.id)
    assert len(sessions) == 2


def test_delete_session(db):
    user, agent = setup_user_and_agent(db)
    svc = ChatService(db)
    session = svc.create_session(user.id, agent.id)
    assert svc.delete_session(session.id, user.id) is True
    assert svc.get_session(session.id, user.id) is None
