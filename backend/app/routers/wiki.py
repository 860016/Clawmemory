from fastapi import APIRouter, Depends, HTTPException
from pydantic import BaseModel
from sqlalchemy.orm import Session
from typing import Optional
from app.database import get_db
from app.middleware.auth import get_current_user
from app.services.wiki_service import WikiService

router = APIRouter(prefix="/api/v1/wiki", tags=["wiki"])


class WikiPageCreate(BaseModel):
    title: str
    content: str = ""
    category: Optional[str] = None
    tags: Optional[list[str]] = None
    is_pinned: bool = False
    parent_id: Optional[int] = None


class WikiPageUpdate(BaseModel):
    title: Optional[str] = None
    content: Optional[str] = None
    category: Optional[str] = None
    tags: Optional[list[str]] = None
    is_pinned: Optional[bool] = None
    parent_id: Optional[int] = None


@router.get("/pages")
def list_pages(
    category: str | None = None,
    parent_id: int | None = None,
    _=Depends(get_current_user),
    db: Session = Depends(get_db),
):
    svc = WikiService(db)
    pages = svc.list_pages(category, parent_id)
    return [
        {
            "id": p.id, "title": p.title, "slug": p.slug,
            "category": p.category, "tags": _parse_tags(p.tags),
            "is_pinned": p.is_pinned, "parent_id": p.parent_id,
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
        "parent_id": page.parent_id,
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
         "updated_at": p.updated_at.isoformat() if p.updated_at else None}
        for p in pages
    ]


@router.get("/categories")
def list_categories(_=Depends(get_current_user), db: Session = Depends(get_db)):
    svc = WikiService(db)
    return svc.get_categories()


def _parse_tags(tags: str | None) -> list[str]:
    if not tags:
        return []
    try:
        import json
        return json.loads(tags)
    except Exception:
        return []
