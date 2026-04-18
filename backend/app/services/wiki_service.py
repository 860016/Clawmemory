import json
import re
from sqlalchemy.orm import Session
from sqlalchemy import or_
from app.models.wiki import WikiPage


def _slugify(title: str) -> str:
    """Generate URL-friendly slug from title"""
    slug = re.sub(r'[^\w\s-]', '', title.lower())
    slug = re.sub(r'[\s_]+', '-', slug).strip('-')
    return slug or 'untitled'


class WikiService:
    def __init__(self, db: Session):
        self.db = db

    def list_pages(self, category: str | None = None, parent_id: int | None = None) -> list[WikiPage]:
        q = self.db.query(WikiPage).filter(WikiPage.user_id == 1)
        if category:
            q = q.filter(WikiPage.category == category)
        if parent_id is not None:
            q = q.filter(WikiPage.parent_id == parent_id)
        return q.order_by(WikiPage.is_pinned.desc(), WikiPage.updated_at.desc()).all()

    def get_page(self, page_id: int) -> WikiPage | None:
        return self.db.query(WikiPage).filter(WikiPage.id == page_id, WikiPage.user_id == 1).first()

    def get_by_slug(self, slug: str) -> WikiPage | None:
        return self.db.query(WikiPage).filter(WikiPage.slug == slug, WikiPage.user_id == 1).first()

    def create_page(self, data: dict) -> WikiPage:
        slug = data.get('slug') or _slugify(data['title'])
        # Ensure unique slug
        existing = self.get_by_slug(slug)
        if existing:
            slug = f"{slug}-{int(existing.id) + 1}"
        tags = data.get('tags')
        if isinstance(tags, list):
            tags = json.dumps(tags, ensure_ascii=False)
        page = WikiPage(
            user_id=1,
            title=data['title'],
            slug=slug,
            content=data.get('content', ''),
            category=data.get('category'),
            tags=tags,
            is_pinned=data.get('is_pinned', False),
            parent_id=data.get('parent_id'),
        )
        self.db.add(page)
        self.db.commit()
        self.db.refresh(page)
        return page

    def update_page(self, page_id: int, data: dict) -> WikiPage | None:
        page = self.get_page(page_id)
        if not page:
            return None
        for k, v in data.items():
            if k == 'tags' and isinstance(v, list):
                v = json.dumps(v, ensure_ascii=False)
            if k == 'title' and v != page.title:
                page.slug = _slugify(v)
            if v is not None:
                setattr(page, k, v)
        self.db.commit()
        self.db.refresh(page)
        return page

    def delete_page(self, page_id: int) -> bool:
        page = self.get_page(page_id)
        if not page:
            return False
        # Reassign children to parent
        children = self.db.query(WikiPage).filter(WikiPage.parent_id == page_id).all()
        for child in children:
            child.parent_id = page.parent_id
        self.db.delete(page)
        self.db.commit()
        return True

    def search_pages(self, query: str, limit: int = 20) -> list[WikiPage]:
        q = self.db.query(WikiPage).filter(WikiPage.user_id == 1)
        pattern = f"%{query}%"
        q = q.filter(or_(WikiPage.title.ilike(pattern), WikiPage.content.ilike(pattern)))
        return q.order_by(WikiPage.updated_at.desc()).limit(limit).all()

    def get_categories(self) -> list[str]:
        rows = self.db.query(WikiPage.category).filter(
            WikiPage.user_id == 1, WikiPage.category.isnot(None)
        ).distinct().all()
        return [r[0] for r in rows if r[0]]

    def get_tree(self) -> list[dict]:
        """Return pages as a tree structure"""
        pages = self.db.query(WikiPage).filter(WikiPage.user_id == 1).order_by(
            WikiPage.is_pinned.desc(), WikiPage.title
        ).all()
        page_map = {}
        roots = []
        for p in pages:
            node = {"id": p.id, "title": p.title, "slug": p.slug, "category": p.category,
                    "is_pinned": p.is_pinned, "parent_id": p.parent_id, "children": []}
            page_map[p.id] = node
        for p in pages:
            node = page_map[p.id]
            if p.parent_id and p.parent_id in page_map:
                page_map[p.parent_id]["children"].append(node)
            else:
                roots.append(node)
        return roots
