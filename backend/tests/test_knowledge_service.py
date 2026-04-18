import uuid
from app.database import SessionLocal, init_db
from app.services.knowledge_service import KnowledgeService
from app.models.user import User
from app.utils.security import hash_password


def test_knowledge_service_entity_crud():
    """Test KnowledgeService entity CRUD"""
    init_db()
    db = SessionLocal()
    unique = uuid.uuid4().hex[:8]
    try:
        user = User(username=f"test_kn_{unique}", email=f"kn_{unique}@test.com", hashed_password=hash_password("test123"))
        db.add(user)
        db.commit()
        db.refresh(user)

        svc = KnowledgeService(db)

        # Create entity
        entity = svc.create_entity(user.id, {
            "name": "Python",
            "entity_type": "language",
            "description": "A programming language",
            "properties": {"typed": True, "year": 1991},
        })
        assert entity.name == "Python"
        assert entity.entity_type == "language"

        # Update entity
        updated = svc.update_entity(entity.id, user.id, {"name": "Python3", "description": "Updated"})
        assert updated is not None
        assert updated.name == "Python3"
        assert updated.description == "Updated"

        # List entities
        entities = svc.list_entities(user.id)
        assert len(entities) >= 1

        # List by type
        typed_entities = svc.list_entities(user.id, entity_type="language")
        assert len(typed_entities) >= 1

        # Delete entity
        assert svc.delete_entity(entity.id, user.id) is True
        assert len(svc.list_entities(user.id)) == 0
    finally:
        db.close()


def test_knowledge_service_relation_crud():
    """Test KnowledgeService relation CRUD + graph"""
    init_db()
    db = SessionLocal()
    unique = uuid.uuid4().hex[:8]
    try:
        user = User(username=f"test_rel_{unique}", email=f"rel_{unique}@test.com", hashed_password=hash_password("test123"))
        db.add(user)
        db.commit()
        db.refresh(user)

        svc = KnowledgeService(db)

        # Create two entities
        e1 = svc.create_entity(user.id, {"name": "Python", "entity_type": "language"})
        e2 = svc.create_entity(user.id, {"name": "FastAPI", "entity_type": "framework"})

        # Create relation
        rel = svc.create_relation(user.id, {
            "source_id": e1.id,
            "target_id": e2.id,
            "relation_type": "powers",
        })
        assert rel.relation_type == "powers"

        # List relations
        rels = svc.list_relations(user.id)
        assert len(rels) >= 1

        # Graph data
        graph = svc.get_graph_data(user.id)
        assert len(graph["nodes"]) >= 2
        assert len(graph["edges"]) >= 1

        # Delete relation
        assert svc.delete_relation(rel.id, user.id) is True
    finally:
        db.close()
