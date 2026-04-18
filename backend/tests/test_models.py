from app.models import User, Agent, Memory, ChatSession


def test_create_user(db):
    user = User(username="testuser", hashed_password="hashed", role="admin")
    db.add(user)
    db.commit()
    assert user.id is not None
    assert user.username == "testuser"
    assert user.role == "admin"


def test_user_agent_relationship(db):
    user = User(username="testuser", hashed_password="hashed")
    db.add(user)
    db.commit()
    agent = Agent(user_id=user.id, name="main", is_default=True)
    db.add(agent)
    db.commit()
    assert len(user.agents) == 1
    assert user.agents[0].name == "main"


def test_create_memory(db):
    user = User(username="testuser", hashed_password="hashed")
    db.add(user)
    db.commit()
    memory = Memory(user_id=user.id, layer="preference", key="language", value="Chinese")
    db.add(memory)
    db.commit()
    assert memory.id is not None
    assert memory.layer == "preference"
