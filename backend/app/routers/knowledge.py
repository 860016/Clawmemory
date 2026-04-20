from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session
from app.database import get_db
from app.middleware.auth import get_current_user
from app.schemas.knowledge import EntityCreate, EntityUpdate, RelationCreate
from app.services.knowledge_service import KnowledgeService
from app.services.license_service import is_feature_enabled

router = APIRouter(prefix="/api/v1/knowledge", tags=["knowledge"])


# --- Entities ---
@router.get("/entities")
def list_entities(entity_type: str | None = None, search: str | None = Query(None), _=Depends(get_current_user), db: Session = Depends(get_db)):
    svc = KnowledgeService(db)
    entities = svc.list_entities(entity_type)
    if search:
        q = search.lower()
        entities = [e for e in entities if q in (e.name or "").lower() or q in (e.entity_type or "").lower() or q in (e.description or "").lower()]
    return entities


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


# --- Graph Analysis (Pro) ---
@router.get("/analysis/centrality")
def get_centrality(_=Depends(get_current_user), db: Session = Depends(get_db)):
    """中心度分析（Pro 功能）"""
    if not is_feature_enabled("ai_extract"):
        raise HTTPException(status_code=403, detail="Pro feature: graph analysis")
    from app.services.graph_analyzer import GraphAnalyzer
    analyzer = GraphAnalyzer(db)
    return analyzer.centrality_analysis()


@router.get("/analysis/communities")
def get_communities(_=Depends(get_current_user), db: Session = Depends(get_db)):
    """社区发现（Pro 功能）"""
    if not is_feature_enabled("ai_extract"):
        raise HTTPException(status_code=403, detail="Pro feature: graph analysis")
    from app.services.graph_analyzer import GraphAnalyzer
    analyzer = GraphAnalyzer(db)
    return analyzer.community_detection()


@router.get("/analysis/stats")
def get_graph_stats(_=Depends(get_current_user), db: Session = Depends(get_db)):
    """图谱统计"""
    from app.services.graph_analyzer import GraphAnalyzer
    analyzer = GraphAnalyzer(db)
    return analyzer.get_graph_stats()


@router.get("/analysis/path/{source_id}/{target_id}")
def get_shortest_path(source_id: int, target_id: int, _=Depends(get_current_user), db: Session = Depends(get_db)):
    """两实体间最短路径"""
    from app.services.graph_analyzer import GraphAnalyzer
    analyzer = GraphAnalyzer(db)
    return analyzer.shortest_path(source_id, target_id)


@router.get("/analysis/neighbors/{entity_id}")
def get_entity_neighbors(entity_id: int, depth: int = Query(2, ge=1, le=3), _=Depends(get_current_user), db: Session = Depends(get_db)):
    """实体邻居探索"""
    from app.services.graph_analyzer import GraphAnalyzer
    analyzer = GraphAnalyzer(db)
    return analyzer.entity_neighbors(entity_id, depth)


# --- Entity Semantic Search ---
@router.get("/entities/search")
def search_entities_semantic(q: str, limit: int = Query(10, ge=1, le=30), _=Depends(get_current_user), db: Session = Depends(get_db)):
    """语义搜索实体"""
    from app.services.vector_service import vector_service
    results = vector_service.search_entities(1, q, limit)
    return results
