from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from app.database import get_db
from app.middleware.auth import get_current_user
from app.schemas.knowledge import EntityCreate, EntityUpdate, RelationCreate
from app.services.knowledge_service import KnowledgeService
from app.services.license_service import is_feature_enabled

router = APIRouter(prefix="/api/v1/knowledge", tags=["knowledge"])


# --- Entities ---
@router.get("/entities")
def list_entities(entity_type: str | None = None, _=Depends(get_current_user), db: Session = Depends(get_db)):
    svc = KnowledgeService(db)
    return svc.list_entities(entity_type)


@router.post("/entities")
def create_entity(data: EntityCreate, _=Depends(get_current_user), db: Session = Depends(get_db)):
    # 免费版限制：50 个实体
    svc = KnowledgeService(db)
    existing = svc.list_entities()
    if len(existing) >= 50 and not is_feature_enabled("unlimited_graph"):
        raise HTTPException(
            status_code=403,
            detail="免费版最多 50 个实体，升级 Pro 解锁无限实体"
        )
    return svc.create_entity(data.model_dump())


@router.delete("/entities/{entity_id}")
def delete_entity(entity_id: int, _=Depends(get_current_user), db: Session = Depends(get_db)):
    svc = KnowledgeService(db)
    if not svc.delete_entity(entity_id):
        raise HTTPException(status_code=404, detail="Entity not found")
    return {"message": "Entity deleted"}


@router.put("/entities/{entity_id}")
def update_entity(entity_id: int, data: EntityUpdate, _=Depends(get_current_user), db: Session = Depends(get_db)):
    svc = KnowledgeService(db)
    entity = svc.update_entity(entity_id, data.model_dump(exclude_unset=True))
    if not entity:
        raise HTTPException(status_code=404, detail="Entity not found")
    return entity


# --- Relations ---
@router.get("/relations")
def list_relations(_=Depends(get_current_user), db: Session = Depends(get_db)):
    svc = KnowledgeService(db)
    return svc.list_relations()


@router.post("/relations")
def create_relation(data: RelationCreate, _=Depends(get_current_user), db: Session = Depends(get_db)):
    # 免费版限制：100 个关系
    svc = KnowledgeService(db)
    existing = svc.list_relations()
    if len(existing) >= 100 and not is_feature_enabled("unlimited_graph"):
        raise HTTPException(
            status_code=403,
            detail="免费版最多 100 个关系，升级 Pro 解锁无限"
        )
    return svc.create_relation(data.model_dump())


@router.delete("/relations/{relation_id}")
def delete_relation(relation_id: int, _=Depends(get_current_user), db: Session = Depends(get_db)):
    svc = KnowledgeService(db)
    if not svc.delete_relation(relation_id):
        raise HTTPException(status_code=404, detail="Relation not found")
    return {"message": "Relation deleted"}


# --- Graph ---
@router.get("/graph")
def get_graph_data(_=Depends(get_current_user), db: Session = Depends(get_db)):
    svc = KnowledgeService(db)
    return svc.get_graph_data()
