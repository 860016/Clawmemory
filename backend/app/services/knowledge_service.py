import json
from sqlalchemy.orm import Session
from app.models.knowledge import Entity, Relation


class KnowledgeService:
    def __init__(self, db: Session):
        self.db = db

    # Entity CRUD
    def list_entities(self, entity_type: str | None = None) -> list[Entity]:
        q = self.db.query(Entity).filter(Entity.user_id == 1)
        if entity_type:
            q = q.filter(Entity.entity_type == entity_type)
        return q.all()

    def create_entity(self, data: dict) -> Entity:
        if "properties" in data and isinstance(data["properties"], dict):
            data["properties"] = json.dumps(data["properties"])
        entity = Entity(user_id=1, **data)
        self.db.add(entity)
        self.db.commit()
        self.db.refresh(entity)
        return entity

    def delete_entity(self, entity_id: int) -> bool:
        entity = self.db.query(Entity).filter(Entity.id == entity_id, Entity.user_id == 1).first()
        if not entity:
            return False
        self.db.delete(entity)
        self.db.commit()
        return True

    def update_entity(self, entity_id: int, data: dict) -> Entity | None:
        entity = self.db.query(Entity).filter(Entity.id == entity_id, Entity.user_id == 1).first()
        if not entity:
            return None
        for k, v in data.items():
            if v is not None:
                if k == "properties" and isinstance(v, dict):
                    v = json.dumps(v)
                setattr(entity, k, v)
        self.db.commit()
        self.db.refresh(entity)
        return entity

    # Relation CRUD
    def list_relations(self) -> list[Relation]:
        return self.db.query(Relation).filter(Relation.user_id == 1).all()

    def create_relation(self, data: dict) -> Relation:
        relation = Relation(user_id=1, **data)
        self.db.add(relation)
        self.db.commit()
        self.db.refresh(relation)
        return relation

    def delete_relation(self, relation_id: int) -> bool:
        relation = self.db.query(Relation).filter(Relation.id == relation_id, Relation.user_id == 1).first()
        if not relation:
            return False
        self.db.delete(relation)
        self.db.commit()
        return True

    # Graph data for visualization
<<<<<<< HEAD
    def get_graph_data(self, user_id: int) -> dict:
        entities = self.list_entities(user_id)
        relations = self.list_relations(user_id)
        nodes = [{"id": e.id, "name": e.name, "type": e.entity_type} for e in entities]
        edges = [{"source_id": r.source_id, "target_id": r.target_id, "relation_type": r.relation_type} for r in relations]
=======
    def get_graph_data(self) -> dict:
        entities = self.list_entities()
        relations = self.list_relations()
        nodes = [{"id": e.id, "name": e.name, "type": e.entity_type, "description": e.description} for e in entities]
        edges = [{"source_id": r.source_id, "target_id": r.target_id, "relation_type": r.relation_type, "description": r.description} for r in relations]
>>>>>>> fb055c7 (feat: v3.0 - Wiki知识库 + 科技感UI + i18n + Rust PyO3核心 + Pro功能)
        return {"nodes": nodes, "edges": edges}
