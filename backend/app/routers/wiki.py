from fastapi import APIRouter, Depends, HTTPException
from pydantic import BaseModel
from sqlalchemy.orm import Session
from typing import Optional
from app.database import get_db
from app.middleware.auth import get_current_user
from app.services.wiki_service import WikiService
from app.services.license_service import is_feature_enabled

router = APIRouter(prefix="/api/v1/wiki", tags=["wiki"])


class WikiPageCreate(BaseModel):
    title: str
    content: str = ""
    category: Optional[str] = None
    tags: Optional[list[str]] = None
    is_pinned: bool = False
    parent_id: Optional[int] = None
    status: str = "draft"


class WikiPageUpdate(BaseModel):
    title: Optional[str] = None
    content: Optional[str] = None
    category: Optional[str] = None
    tags: Optional[list[str]] = None
    is_pinned: Optional[bool] = None
    parent_id: Optional[int] = None
    status: Optional[str] = None


class ConversationExtractRequest(BaseModel):
    conversation: str
    is_complete: bool = True


class RefinePageRequest(BaseModel):
    additional_context: str = ""


@router.get("/pages")
def list_pages(
    category: str | None = None,
    parent_id: int | None = None,
    status: str | None = None,
    _=Depends(get_current_user),
    db: Session = Depends(get_db),
):
    svc = WikiService(db)
    pages = svc.list_pages(category, parent_id)
    if status:
        pages = [p for p in pages if p.status == status]
    return [
        {
            "id": p.id, "title": p.title, "slug": p.slug,
            "category": p.category, "tags": _parse_tags(p.tags),
            "is_pinned": p.is_pinned, "parent_id": p.parent_id,
            "status": p.status, "ai_generated": p.ai_generated,
            "ai_confidence": p.ai_confidence, "summary": p.summary,
            "created_at": p.created_at.isoformat() if p.created_at else None,
            "updated_at": p.updated_at.isoformat() if p.updated_at else None,
        }
        for p in pages
    ]


@router.get("/pages/tree")
def get_page_tree(_=Depends(get_current_user), db: Session = Depends(get_db)):
    svc = WikiService(db)
    return svc.get_tree()


@router.get("/pages/{page_id}")
def get_page(page_id: int, _=Depends(get_current_user), db: Session = Depends(get_db)):
    svc = WikiService(db)
    page = svc.get_page(page_id)
    if not page:
        raise HTTPException(status_code=404, detail="Page not found")
    return {
        "id": page.id, "title": page.title, "slug": page.slug,
        "content": page.content, "category": page.category,
        "tags": _parse_tags(page.tags), "is_pinned": page.is_pinned,
        "parent_id": page.parent_id, "status": page.status,
        "ai_generated": page.ai_generated, "ai_confidence": page.ai_confidence,
        "summary": page.summary,
        "key_decisions": _parse_json(page.key_decisions),
        "action_items": _parse_json(page.action_items),
        "created_at": page.created_at.isoformat() if page.created_at else None,
        "updated_at": page.updated_at.isoformat() if page.updated_at else None,
    }


@router.post("/pages")
def create_page(data: WikiPageCreate, _=Depends(get_current_user), db: Session = Depends(get_db)):
    svc = WikiService(db)
    return svc.create_page(data.model_dump())


@router.put("/pages/{page_id}")
def update_page(page_id: int, data: WikiPageUpdate, _=Depends(get_current_user), db: Session = Depends(get_db)):
    svc = WikiService(db)
    page = svc.update_page(page_id, data.model_dump(exclude_unset=True))
    if not page:
        raise HTTPException(status_code=404, detail="Page not found")
    return page


@router.delete("/pages/{page_id}")
def delete_page(page_id: int, _=Depends(get_current_user), db: Session = Depends(get_db)):
    svc = WikiService(db)
    if not svc.delete_page(page_id):
        raise HTTPException(status_code=404, detail="Page not found")
    return {"message": "Page deleted"}


@router.get("/search")
def search_pages(q: str, limit: int = 20, _=Depends(get_current_user), db: Session = Depends(get_db)):
    svc = WikiService(db)
    pages = svc.search_pages(q, limit)
    return [
        {"id": p.id, "title": p.title, "slug": p.slug, "category": p.category,
         "status": p.status, "ai_generated": p.ai_generated,
         "updated_at": p.updated_at.isoformat() if p.updated_at else None}
        for p in pages
    ]


@router.get("/categories")
def list_categories(_=Depends(get_current_user), db: Session = Depends(get_db)):
    svc = WikiService(db)
    return svc.get_categories()


@router.get("/config")
def get_config(_=Depends(get_current_user)):
    """获取 LLM 配置状态"""
    from app.config import settings
    return {
        "llm_available": bool(settings.openai_api_key and settings.openai_base_url),
    }


@router.get("/stats")
def get_stats(_=Depends(get_current_user), db: Session = Depends(get_db)):
    """获取知识库统计"""
    from app.services.wiki_ai_service import WikiAIService
    svc = WikiAIService(db)
    return svc.get_stats()


@router.post("/ai/extract")
async def extract_from_conversation(
    req: ConversationExtractRequest,
    _=Depends(get_current_user),
    db: Session = Depends(get_db),
):
    """从对话中提取知识卡片（Pro 功能）"""
    if not is_feature_enabled("ai_extract"):
        raise HTTPException(status_code=403, detail="Pro feature: AI extract")

    from app.services.wiki_ai_service import WikiAIService
    svc = WikiAIService(db)
    try:
        page = await svc.create_card_from_conversation(req.conversation, req.is_complete)
        return {
            "id": page.id, "title": page.title, "status": page.status,
            "ai_generated": page.ai_generated, "ai_confidence": page.ai_confidence,
        }
    except RuntimeError as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.post("/pages/{page_id}/refine")
async def refine_page(
    page_id: int,
    req: RefinePageRequest,
    _=Depends(get_current_user),
    db: Session = Depends(get_db),
):
    """优化知识卡片（Pro 功能）"""
    if not is_feature_enabled("ai_extract"):
        raise HTTPException(status_code=403, detail="Pro feature: AI extract")

    from app.services.wiki_ai_service import WikiAIService
    svc = WikiAIService(db)
    page = await svc.refine_page(page_id, req.additional_context)
    if not page:
        raise HTTPException(status_code=404, detail="Page not found")
    return {"message": "Page refined", "ai_confidence": page.ai_confidence}


@router.post("/pages/{page_id}/mark-complete")
def mark_complete(page_id: int, _=Depends(get_current_user), db: Session = Depends(get_db)):
    """标记知识卡片为完成"""
    from app.services.wiki_ai_service import WikiAIService
    svc = WikiAIService(db)
    page = svc.mark_complete_sync(page_id)
    if not page:
        raise HTTPException(status_code=404, detail="Page not found")
    return {"message": "Page marked as complete", "status": page.status}


@router.post("/pages/{page_id}/mark-in-progress")
def mark_in_progress(page_id: int, _=Depends(get_current_user), db: Session = Depends(get_db)):
    """标记知识卡片为进行中"""
    from app.services.wiki_ai_service import WikiAIService
    svc = WikiAIService(db)
    page = svc.mark_in_progress_sync(page_id)
    if not page:
        raise HTTPException(status_code=404, detail="Page not found")
    return {"message": "Page marked as in progress", "status": page.status}


def _parse_tags(tags: str | None) -> list[str]:
    if not tags:
        return []
    try:
        import json
        return json.loads(tags)
    except Exception:
        return []


def _parse_json(data: str | None):
    if not data:
        return []
    try:
        import json
        return json.loads(data)
    except Exception:
        return []
