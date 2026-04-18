"""Unit tests for auth change-password endpoint."""
import pytest
from fastapi.testclient import TestClient
from app.models.user import User
from app.utils.security import hash_password, verify_password


def test_change_password_success(auth_client, db):
    """User can change password with correct old password."""
    resp = auth_client.put("/api/v1/auth/change-password", json={
        "old_password": "test123",
        "new_password": "newpass456",
    })
    assert resp.status_code == 200
    assert resp.json()["message"] == "Password changed successfully"

    # Verify password actually changed in DB
    user = db.query(User).filter(User.username == "testuser").first()
    assert verify_password("newpass456", user.hashed_password)


def test_change_password_wrong_old(auth_client, db):
    """Changing password with wrong old password should fail."""
    resp = auth_client.put("/api/v1/auth/change-password", json={
        "old_password": "wrongpass",
        "new_password": "newpass456",
    })
    assert resp.status_code == 400
    assert "incorrect" in resp.json()["detail"].lower()


def test_change_password_too_short(auth_client, db):
    """New password less than 6 chars should fail."""
    resp = auth_client.put("/api/v1/auth/change-password", json={
        "old_password": "test123",
        "new_password": "abc",
    })
    assert resp.status_code == 400
    assert "6 characters" in resp.json()["detail"]


def test_change_password_unauthorized(client):
    """Unauthenticated request should return 401."""
    resp = client.put("/api/v1/auth/change-password", json={
        "old_password": "test123",
        "new_password": "newpass456",
    })
    assert resp.status_code == 401


def test_auth_init_status(client):
    """Check init-status endpoint returns boolean."""
    resp = client.get("/api/v1/auth/init-status")
    assert resp.status_code == 200
    assert "initialized" in resp.json()
