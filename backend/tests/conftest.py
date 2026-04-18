import pytest
from sqlalchemy import create_engine, event, text
from sqlalchemy.orm import sessionmaker
from fastapi.testclient import TestClient
from app.database import Base, get_db
from app.main import app

SQLALCHEMY_DATABASE_URL = "sqlite:///:memory:"
engine = create_engine(SQLALCHEMY_DATABASE_URL, connect_args={"check_same_thread": False})

# Enable foreign keys for in-memory SQLite
@event.listens_for(engine, "connect")
def set_sqlite_pragma(dbapi_conn, connection_record):
    cursor = dbapi_conn.cursor()
    cursor.execute("PRAGMA foreign_keys=ON")
    cursor.close()

TestingSessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)


def override_get_db():
    db = TestingSessionLocal()
    try:
        yield db
    finally:
        db.close()


app.dependency_overrides[get_db] = override_get_db


@pytest.fixture(autouse=True)
def setup_db():
    from app.models import (  # noqa: F401
        user, agent, chat, memory, knowledge, license as lic, skill, backup, notification
    )
    Base.metadata.drop_all(bind=engine)
    Base.metadata.create_all(bind=engine)
    # FTS5 with content=memories requires the table to exist first
    try:
        with engine.connect() as conn:
            conn.execute(text("""
                CREATE VIRTUAL TABLE IF NOT EXISTS memories_fts
                USING fts5(key, value, content=memories, content_rowid=id)
            """))
            conn.execute(text("""
                CREATE TRIGGER IF NOT EXISTS memories_ai AFTER INSERT ON memories BEGIN
                    INSERT INTO memories_fts(rowid, key, value) VALUES (new.id, new.key, new.value);
                END
            """))
            conn.execute(text("""
                CREATE TRIGGER IF NOT EXISTS memories_ad AFTER DELETE ON memories BEGIN
                    INSERT INTO memories_fts(memories_fts, rowid, key, value) VALUES('delete', old.id, old.key, old.value);
                END
            """))
            conn.execute(text("""
                CREATE TRIGGER IF NOT EXISTS memories_au AFTER UPDATE ON memories BEGIN
                    INSERT INTO memories_fts(memories_fts, rowid, key, value) VALUES('delete', old.id, old.key, old.value);
                    INSERT INTO memories_fts(rowid, key, value) VALUES (new.id, new.key, new.value);
                END
            """))
            conn.commit()
    except Exception:
        pass  # FTS5 may fail in some test environments
    yield
    Base.metadata.drop_all(bind=engine)


@pytest.fixture
def client():
    return TestClient(app)


@pytest.fixture
def db():
    db = TestingSessionLocal()
    try:
        yield db
    finally:
        db.close()


@pytest.fixture
def auth_client(client, db):
    """TestClient with authenticated user"""
    from app.models.user import User
    from app.utils.security import hash_password, create_access_token

    user = User(
        username="testuser",
        email="test@test.com",
        hashed_password=hash_password("test123"),
        role="user",
    )
    db.add(user)
    db.commit()
    db.refresh(user)

    token = create_access_token({"sub": user.username})
    client.headers["Authorization"] = f"Bearer {token}"
    client._test_user = user
    return client


@pytest.fixture
def admin_client(client, db):
    """TestClient with admin user"""
    from app.models.user import User
    from app.utils.security import hash_password, create_access_token

    user = User(
        username="admin",
        email="admin@test.com",
        hashed_password=hash_password("admin123"),
        role="admin",
    )
    db.add(user)
    db.commit()
    db.refresh(user)

    token = create_access_token({"sub": user.username})
    client.headers["Authorization"] = f"Bearer {token}"
    client._test_user = user
    return client
