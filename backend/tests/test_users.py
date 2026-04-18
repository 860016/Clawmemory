def setup_admin(client):
    resp = client.post("/api/v1/auth/init", json={"username": "admin", "password": "admin123"})
    return resp.json()["access_token"]


def test_list_users(client):
    token = setup_admin(client)
    resp = client.get("/api/v1/users", headers={"Authorization": f"Bearer {token}"})
    assert resp.status_code == 200
    assert resp.json()["total"] == 1


def test_create_user(client):
    token = setup_admin(client)
    resp = client.post("/api/v1/users", json={
        "username": "newuser", "password": "pass123", "role": "user"
    }, headers={"Authorization": f"Bearer {token}"})
    assert resp.status_code == 200
    assert resp.json()["username"] == "newuser"


def test_create_duplicate_user(client):
    token = setup_admin(client)
    client.post("/api/v1/users", json={"username": "newuser", "password": "pass123"},
                 headers={"Authorization": f"Bearer {token}"})
    resp = client.post("/api/v1/users", json={"username": "newuser", "password": "pass123"},
                        headers={"Authorization": f"Bearer {token}"})
    assert resp.status_code == 400


def test_delete_user(client):
    token = setup_admin(client)
    create_resp = client.post("/api/v1/users", json={"username": "todelete", "password": "pass123"},
                               headers={"Authorization": f"Bearer {token}"})
    user_id = create_resp.json()["id"]
    resp = client.delete(f"/api/v1/users/{user_id}", headers={"Authorization": f"Bearer {token}"})
    assert resp.status_code == 200
