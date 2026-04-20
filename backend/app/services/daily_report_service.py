import json
import logging
from datetime import datetime, timedelta
from sqlalchemy.orm import Session
from app.models.memory import Memory
from app.models.knowledge import Entity, Relation
from app.models.wiki import WikiPage
from app.models.daily_report import DailyReport
from app.services.llm_service import llm_service
from app.services.license_service import is_feature_enabled

logger = logging.getLogger(__name__)


class DailyReportService:
    """日报生成服务"""

    def __init__(self, db: Session):
        self.db = db

    def _get_today_data(self, target_date: datetime | None = None) -> dict:
        """获取指定日期的数据"""
        if target_date is None:
            target_date = datetime.now()
        
        start = target_date.replace(hour=0, minute=0, second=0, microsecond=0)
        end = start + timedelta(days=1)

        # 当日新增记忆
        memories = self.db.query(Memory).filter(
            Memory.user_id == 1,
            Memory.created_at >= start,
            Memory.created_at < end,
            Memory.status == "active"
        ).all()

        # 当日新增实体
        entities = self.db.query(Entity).filter(
            Entity.user_id == 1,
            Entity.created_at >= start,
            Entity.created_at < end,
        ).all()

        # 当日新增关系
        relations = self.db.query(Relation).filter(
            Relation.user_id == 1,
            Relation.created_at >= start,
            Relation.created_at < end,
        ).all()

        # 当日更新的知识页面
        wiki_pages = self.db.query(WikiPage).filter(
            WikiPage.user_id == 1,
            WikiPage.updated_at >= start,
            WikiPage.updated_at < end,
        ).all()

        return {
            "memories": memories,
            "entities": entities,
            "relations": relations,
            "wiki_pages": wiki_pages,
        }

    def _build_context(self, data: dict) -> str:
        """构建 LLM 上下文"""
        parts = []

        # 记忆摘要
        if data["memories"]:
            parts.append("今日新增记忆：")
            for m in data["memories"][:50]:  # 限制数量
                parts.append(f"- [{m.layer}] {m.key}: {m.value[:100]}")

        # 实体摘要
        if data["entities"]:
            parts.append("\n今日新增实体：")
            for e in data["entities"]:
                parts.append(f"- {e.name} ({e.entity_type}): {e.description or '无描述'}")

        # 关系摘要
        if data["relations"]:
            parts.append("\n今日新增关系：")
            for r in data["relations"]:
                parts.append(f"- {r.relation_type} 连接实体 #{r.source_id} 和 #{r.target_id}")

        # 知识页面更新
        if data["wiki_pages"]:
            parts.append("\n今日更新的知识页面：")
            for w in data["wiki_pages"]:
                parts.append(f"- {w.title} (分类: {w.category or '未分类'})")

        return "\n".join(parts)

    async def generate_report(self, target_date: datetime | None = None) -> DailyReport | None:
        """生成日报"""
        if target_date is None:
            target_date = datetime.now()

        date_str = target_date.strftime("%Y-%m-%d")

        # 检查是否已存在
        existing = self.db.query(DailyReport).filter(
            DailyReport.user_id == 1,
            DailyReport.report_date == date_str,
        ).first()
        if existing:
            logger.info(f"Report already exists for {date_str}")
            return existing

        # 获取数据
        data = self._get_today_data(target_date)

        # 如果没有任何数据，跳过
        if not any([data["memories"], data["entities"], data["relations"], data["wiki_pages"]]):
            logger.info(f"No data for {date_str}, skipping report")
            return None

        # 构建上下文
        context = self._build_context(data)

        # 调用 LLM 生成日报
        report_data = await self._llm_generate(context, data)

        # 保存到数据库
        report = DailyReport(
            user_id=1,
            report_date=date_str,
            summary=report_data.get("summary", ""),
            highlights=report_data.get("highlights", []),
            knowledge_gained=report_data.get("knowledge_gained", []),
            pending_tasks=report_data.get("pending_tasks", []),
            tomorrow_suggestions=report_data.get("tomorrow_suggestions", []),
            stats={
                "new_memories": len(data["memories"]),
                "new_entities": len(data["entities"]),
                "new_relations": len(data["relations"]),
                "updated_wiki": len(data["wiki_pages"]),
            },
            raw_data={
                "memory_keys": [m.key for m in data["memories"]],
                "entity_names": [e.name for e in data["entities"]],
            },
            ai_model=llm_service.model if llm_service.available else "fallback",
        )

        self.db.add(report)
        self.db.commit()
        self.db.refresh(report)

        logger.info(f"Daily report generated for {date_str}")
        return report

    async def _llm_generate(self, context: str, data: dict) -> dict:
        """调用 LLM 生成日报内容"""
        if not llm_service.available:
            return self._fallback_report(context, data)

        messages = [
            {"role": "system", "content": (
                "You are an AI daily report generator. "
                "Analyze the given daily activity data and create a concise, insightful daily report. "
                "Output ONLY a JSON object, nothing else."
            )},
            {"role": "user", "content": (
                f"Generate a daily report based on today's activity:\n\n{context}\n\n"
                f"Output a JSON object with:\n"
                f"- summary: a one-sentence overview of today's activity (in Chinese)\n"
                f"- highlights: array of 3-5 key highlights or achievements (in Chinese)\n"
                f"- knowledge_gained: array of new knowledge or insights learned today (in Chinese)\n"
                f"- pending_tasks: array of unfinished tasks or ongoing discussions (in Chinese)\n"
                f"- tomorrow_suggestions: array of 2-3 suggestions for tomorrow based on pending items (in Chinese)\n\n"
                f"Output JSON only."
            )},
        ]

        try:
            result = await llm_service._call_json(messages, temperature=0.3)
            if isinstance(result, dict):
                return result
            return self._fallback_report(context, data)
        except Exception as e:
            logger.warning(f"LLM report generation failed: {e}")
            return self._fallback_report(context, data)

    def _fallback_report(self, context: str, data: dict) -> dict:
        """LLM 不可用时的降级方案"""
        highlights = []
        if data["memories"]:
            highlights.append(f"记录了 {len(data['memories'])} 条新记忆")
        if data["entities"]:
            highlights.append(f"新增了 {len(data['entities'])} 个知识实体")
        if data["relations"]:
            highlights.append(f"建立了 {len(data['relations'])} 个实体关系")
        if data["wiki_pages"]:
            highlights.append(f"更新了 {len(data['wiki_pages'])} 个知识页面")

        return {
            "summary": f"今日共记录 {len(data['memories'])} 条记忆，{len(data['entities'])} 个实体，{len(data['wiki_pages'])} 个知识页面更新",
            "highlights": highlights[:5],
            "knowledge_gained": [],
            "pending_tasks": [],
            "tomorrow_suggestions": [],
        }

    def get_report(self, date_str: str) -> DailyReport | None:
        """获取指定日期的日报"""
        return self.db.query(DailyReport).filter(
            DailyReport.user_id == 1,
            DailyReport.report_date == date_str,
        ).first()

    def list_reports(self, limit: int = 30) -> list[DailyReport]:
        """获取最近 N 天的日报"""
        return self.db.query(DailyReport).filter(
            DailyReport.user_id == 1,
        ).order_by(DailyReport.report_date.desc()).limit(limit).all()

    def get_stats_summary(self, days: int = 7) -> dict:
        """获取最近 N 天的统计摘要"""
        cutoff = (datetime.now() - timedelta(days=days)).strftime("%Y-%m-%d")
        reports = self.db.query(DailyReport).filter(
            DailyReport.user_id == 1,
            DailyReport.report_date >= cutoff,
        ).all()

        total_memories = 0
        total_entities = 0
        total_wiki = 0

        for r in reports:
            if r.stats:
                total_memories += r.stats.get("new_memories", 0)
                total_entities += r.stats.get("new_entities", 0)
                total_wiki += r.stats.get("updated_wiki", 0)

        return {
            "period_days": days,
            "report_count": len(reports),
            "total_memories": total_memories,
            "total_entities": total_entities,
            "total_wiki_updated": total_wiki,
        }


daily_report_service = DailyReportService
