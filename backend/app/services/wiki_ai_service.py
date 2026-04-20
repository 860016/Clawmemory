import json
import logging
from sqlalchemy.orm import Session
from app.models.wiki import WikiPage
from app.services.llm_service import llm_service
from app.services.license_service import is_feature_enabled

logger = logging.getLogger(__name__)


class WikiAIService:
    """AI 知识卡片提取服务"""

    def __init__(self, db: Session):
        self.db = db

    async def extract_knowledge_card(self, conversation: str, is_complete: bool = True) -> dict:
        """从对话中提取知识卡片"""
        if not is_feature_enabled("ai_extract"):
            raise RuntimeError("Pro feature required")
        if not llm_service.available:
            raise RuntimeError("LLM not configured")

        return await llm_service.extract_knowledge_card(conversation, is_complete)

    async def create_card_from_conversation(self, conversation: str, is_complete: bool = True) -> WikiPage:
        """从对话创建知识卡片"""
        card_data = await self.extract_knowledge_card(conversation, is_complete)

        if not card_data:
            raise RuntimeError("Failed to extract knowledge card")

        title = card_data.get("title", "Untitled Knowledge Card")
        content = card_data.get("content", "")
        summary = card_data.get("summary", "")
        category = card_data.get("category", "discussion")
        tags = card_data.get("tags", [])
        status = card_data.get("status", "in_progress" if not is_complete else "completed")
        key_decisions = card_data.get("key_decisions", [])
        action_items = card_data.get("action_items", [])

        page = WikiPage(
            user_id=1,
            title=title,
            content=content,
            category=category,
            tags=json.dumps(tags, ensure_ascii=False) if tags else None,
            status=status,
            ai_generated=True,
            ai_confidence=0.8,
            source_conversation=conversation[:5000],  # 保存前5000字符
            summary=summary,
            key_decisions=json.dumps(key_decisions, ensure_ascii=False) if key_decisions else None,
            action_items=json.dumps(action_items, ensure_ascii=False) if action_items else None,
        )
        self.db.add(page)
        self.db.commit()
        self.db.refresh(page)
        return page

    async def refine_page(self, page_id: int, additional_context: str = "") -> WikiPage | None:
        """优化已有知识卡片"""
        page = self.db.query(WikiPage).filter(WikiPage.id == page_id).first()
        if not page:
            return None

        if not llm_service.available:
            raise RuntimeError("LLM not configured")

        messages = [
            {"role": "system", "content": (
                "You are an expert knowledge curator. "
                "Improve and refine the given knowledge card with additional context. "
                "Output ONLY a JSON object, nothing else."
            )},
            {"role": "user", "content": (
                f"Improve this knowledge card:\n\n"
                f"Current title: {page.title}\n"
                f"Current content: {page.content[:2000]}\n"
                f"Current summary: {page.summary or ''}\n\n"
                f"Additional context: {additional_context}\n\n"
                f"Output JSON with:\n"
                f"- title: improved title\n"
                f"- content: improved markdown content\n"
                f"- summary: improved summary\n"
                f"- tags: updated tags\n"
                f"- key_decisions: updated key decisions\n"
                f"- action_items: updated action items\n\n"
                f"Output JSON only."
            )},
        ]

        try:
            result = await llm_service._call_json(messages)
            if result.get("title"):
                page.title = result["title"]
            if result.get("content"):
                page.content = result["content"]
            if result.get("summary"):
                page.summary = result["summary"]
            if result.get("tags"):
                page.tags = json.dumps(result["tags"], ensure_ascii=False)
            if result.get("key_decisions"):
                page.key_decisions = json.dumps(result["key_decisions"], ensure_ascii=False)
            if result.get("action_items"):
                page.action_items = json.dumps(result["action_items"], ensure_ascii=False)

            page.ai_confidence = min(page.ai_confidence + 0.1, 1.0)
            self.db.commit()
            self.db.refresh(page)
        except Exception as e:
            logger.warning(f"Failed to refine page: {e}")

        return page

    async def mark_complete(self, page_id: int) -> WikiPage | None:
        """标记知识卡片为完成"""
        page = self.db.query(WikiPage).filter(WikiPage.id == page_id).first()
        if not page:
            return None

        page.status = "completed"
        page.action_items = json.dumps([], ensure_ascii=False)
        self.db.commit()
        self.db.refresh(page)
        return page

    async def mark_in_progress(self, page_id: int) -> WikiPage | None:
        """标记知识卡片为进行中"""
        page = self.db.query(WikiPage).filter(WikiPage.id == page_id).first()
        if not page:
            return None

        page.status = "in_progress"
        self.db.commit()
        self.db.refresh(page)
        return page

    def mark_complete_sync(self, page_id: int) -> WikiPage | None:
        """标记知识卡片为完成（同步版本）"""
        page = self.db.query(WikiPage).filter(WikiPage.id == page_id).first()
        if not page:
            return None

        page.status = "completed"
        page.action_items = json.dumps([], ensure_ascii=False)
        self.db.commit()
        self.db.refresh(page)
        return page

    def mark_in_progress_sync(self, page_id: int) -> WikiPage | None:
        """标记知识卡片为进行中（同步版本）"""
        page = self.db.query(WikiPage).filter(WikiPage.id == page_id).first()
        if not page:
            return None

        page.status = "in_progress"
        self.db.commit()
        self.db.refresh(page)
        return page

    def get_pages_by_status(self, status: str) -> list[WikiPage]:
        """按状态获取知识卡片"""
        return self.db.query(WikiPage).filter(
            WikiPage.user_id == 1,
            WikiPage.status == status,
        ).order_by(WikiPage.updated_at.desc()).all()

    def get_stats(self) -> dict:
        """获取知识库统计"""
        total = self.db.query(WikiPage).filter(WikiPage.user_id == 1).count()
        completed = self.db.query(WikiPage).filter(WikiPage.user_id == 1, WikiPage.status == "completed").count()
        in_progress = self.db.query(WikiPage).filter(WikiPage.user_id == 1, WikiPage.status == "in_progress").count()
        draft = self.db.query(WikiPage).filter(WikiPage.user_id == 1, WikiPage.status == "draft").count()
        ai_generated = self.db.query(WikiPage).filter(WikiPage.user_id == 1, WikiPage.ai_generated == True).count()

        return {
            "total": total,
            "completed": completed,
            "in_progress": in_progress,
            "draft": draft,
            "ai_generated": ai_generated,
        }
