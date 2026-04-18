import tempfile
import os


def test_model_service_crud():
    """Test ModelService CRUD operations"""
    os.environ["OPENCLAW_DATA_DIR"] = tempfile.mkdtemp()
    from app.services.model_service import ModelService
    svc = ModelService()
    svc._models = []
    svc._save()

    # Create
    m = svc.add_model({"name": "GPT-4", "provider": "openai", "model_id": "gpt-4"})
    assert m["name"] == "GPT-4"
    assert len(svc.list_models()) == 1

    # Read
    fetched = svc.get_model(m["id"])
    assert fetched is not None
    assert fetched["name"] == "GPT-4"

    # Update
    updated = svc.update_model(m["id"], {"temperature": 0.5})
    assert updated is not None
    assert updated["temperature"] == 0.5

    # Delete
    assert svc.delete_model(m["id"]) is True
    assert len(svc.list_models()) == 0

    # Default
    m1 = svc.add_model({"name": "Default", "provider": "openai", "model_id": "gpt-4", "is_default": True})
    m2 = svc.add_model({"name": "Other", "provider": "anthropic", "model_id": "claude-3"})
    default = svc.get_default()
    assert default["id"] == m1["id"]


def test_model_service_not_found():
    """Test ModelService with non-existent model"""
    os.environ["OPENCLAW_DATA_DIR"] = tempfile.mkdtemp()
    from app.services.model_service import ModelService
    svc = ModelService()
    svc._models = []
    svc._save()

    assert svc.get_model("nonexistent") is None
    assert svc.update_model("nonexistent", {"name": "x"}) is None
    assert svc.delete_model("nonexistent") is False
