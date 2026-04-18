"""Integration tests for general API routes."""


def test_health_check(client):
    resp = client.get("/api/v1/health")
    assert resp.status_code == 200
    assert resp.json()["status"] == "ok"


def test_license_status(auth_client):
    resp = auth_client.get("/api/v1/license/status")
    assert resp.status_code == 200
    data = resp.json()
    assert data["tier"] == "oss"
    assert data["is_valid"] is True


def test_license_deactivate(auth_client):
    resp = auth_client.post("/api/v1/license/deactivate")
    assert resp.status_code == 200


def test_models_list(auth_client):
    resp = auth_client.get("/api/v1/models")
    assert resp.status_code == 200


def test_models_crud_requires_admin(auth_client):
    """Non-admin user should be forbidden from creating models."""
    resp = auth_client.post("/api/v1/models", json={
        "name": "test-model", "provider": "openai", "model_id": "gpt-4o-mini",
    })
    assert resp.status_code in (403, 401)


def test_backups_list(auth_client):
    resp = auth_client.get("/api/v1/backups")
    assert resp.status_code == 200


def test_knowledge_entities(auth_client):
    resp = auth_client.get("/api/v1/knowledge/entities")
    assert resp.status_code == 200


def test_knowledge_graph(auth_client):
    resp = auth_client.get("/api/v1/knowledge/graph")
    assert resp.status_code == 200


def test_skills_list(auth_client):
    resp = auth_client.get("/api/v1/skills")
    assert resp.status_code == 200


def test_nodes_list(auth_client):
    resp = auth_client.get("/api/v1/nodes")
    assert resp.status_code == 200


def test_unauthorized_access(client):
    """Unauthenticated requests should return 401."""
    assert client.get("/api/v1/memories").status_code == 401
    assert client.get("/api/v1/chat/sessions").status_code == 401
    assert client.get("/api/v1/models").status_code == 401
    assert client.get("/api/v1/license/status").status_code == 401
