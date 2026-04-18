"""Integration tests for chat API routes."""
from fastapi.testclient import TestClient


def test_create_session(auth_client):
    resp = auth_client.post("/api/v1/chat/sessions", json={
        "agent_id": 1,
        "title": "Test Session",
    })
    assert resp.status_code == 200
    data = resp.json()
    assert data["title"] == "Test Session"


def test_list_sessions(auth_client):
    auth_client.post("/api/v1/chat/sessions", json={"agent_id": 1, "title": "Session 1"})
    resp = auth_client.get("/api/v1/chat/sessions")
    assert resp.status_code == 200
    data = resp.json()
    assert len(data) >= 1


def test_get_messages(auth_client):
    create_resp = auth_client.post("/api/v1/chat/sessions", json={"agent_id": 1, "title": "Msg Session"})
    session_id = create_resp.json()["id"]
    resp = auth_client.get(f"/api/v1/chat/sessions/{session_id}/messages")
    assert resp.status_code == 200


def test_delete_session(auth_client):
    create_resp = auth_client.post("/api/v1/chat/sessions", json={"agent_id": 1, "title": "Del Session"})
    session_id = create_resp.json()["id"]
    resp = auth_client.delete(f"/api/v1/chat/sessions/{session_id}")
    assert resp.status_code == 200


def test_unauthorized_chat(client):
    resp = client.get("/api/v1/chat/sessions")
    assert resp.status_code == 401
