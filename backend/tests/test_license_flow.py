from app.services.license_service import LicenseService, current_tier, is_feature_enabled, reset, set_license
from app.services.auth_service import AuthService


def test_license_service_init(db):
    svc = LicenseService(db)
    assert svc.get_current_license() is None


def test_feature_gate_default():
    reset()
    assert current_tier() == "oss"
    assert is_feature_enabled("graph") is False


def test_feature_gate_pro():
    set_license("pro", ["graph", "backup", "decay", "routing", "conflict"])
    assert current_tier() == "pro"
    assert is_feature_enabled("graph") is True
    assert is_feature_enabled("sso") is False
    reset()
