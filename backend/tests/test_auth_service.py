from app.services.auth_service import AuthService
from app.utils.security import hash_password, verify_password, create_access_token, decode_access_token


def test_password_hashing():
    hashed = hash_password("mypassword")
    assert verify_password("mypassword", hashed) is True
    assert verify_password("wrongpassword", hashed) is False


def test_create_and_decode_token():
    token = create_access_token({"sub": "testuser", "role": "admin", "user_id": 1})
    payload = decode_access_token(token)
    assert payload is not None
    assert payload["sub"] == "testuser"
    assert payload["role"] == "admin"


def test_authenticate_user(db):
    svc = AuthService(db)
    svc.create_user("testuser", "password123", role="admin")
    result = svc.authenticate("testuser", "password123")
    assert result is not None
    assert result["role"] == "admin"
    assert "access_token" in result


def test_authenticate_wrong_password(db):
    svc = AuthService(db)
    svc.create_user("testuser", "password123")
    result = svc.authenticate("testuser", "wrongpassword")
    assert result is None


def test_is_initialized(db):
    svc = AuthService(db)
    assert svc.is_initialized() is False
    svc.create_user("admin", "password", role="admin")
    assert svc.is_initialized() is True
