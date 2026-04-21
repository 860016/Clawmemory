import logging
from datetime import datetime, timedelta
from typing import Optional
from sqlalchemy.orm import Session
from app.models.daily_report import DailyReport
from app.models.memory import Memory
from app.models.entity import Entity
from app.models.wiki_page import WikiPage

logger = logging.getLogger(__name__)


class DailyReportService:
    """日报服务（免费功能）"""

    def __init__(self, db: Session):
        self.db = db

    def list_reports(self, limit: int = 30) -> list:
        """获取日报列表"""
        return (
            self.db.query(DailyReport)
            .order_by(DailyReport.report_date.desc())
            .limit(limit)
            .all()
        )

    def get_report(self, date_str: str) -> Optional[DailyReport]:
        """获取指定日期的日报"""
        return (
            self.db.query(DailyReport)
            .filter(DailyReport.report_date == date_str)
            .first()
        )

    async def generate_report(self, target_date: Optional[datetime] = None) -> Optional[DailyReport]:
        """生成日报"""
        if target_date is None:
            target_date = datetime.now()

        date_str = target_date.strftime("%Y-%m-%d")
        start_dt = datetime(target_date.year, target_date.month, target_date.day)
        end_dt = start_dt + timedelta(days=1)

        # 统计当日数据
        new_memories = (
            self.db.query(Memory)
            .filter(Memory.created_at >= start_dt, Memory.created_at < end_dt)
            .count()
        )

        new_entities = (
            self.db.query(Entity)
            .filter(Entity.created_at >= start_dt, Entity.created_at < end_dt)
            .count()
        )

        updated_wiki = (
            self.db.query(WikiPage)
            .filter(WikiPage.updated_at >= start_dt, WikiPage.updated_at < end_dt)
            .count()
        )

        if new_memories == 0 and new_entities == 0 and updated_wiki == 0:
            logger.info(f"No data for {date_str}, skipping report")
            return None

        # 检查是否已存在
        existing = self.get_report(date_str)
        if existing:
            logger.info(f"Report already exists for {date_str}")
            return existing

        # 创建日报
        report = DailyReport(
            report_date=date_str,
            summary=f"今日新增 {new_memories} 条记忆，{new_entities} 个实体，更新 {updated_wiki} 个 Wiki 页面。",
            highlights=[],
            knowledge_gained=[],
            pending_tasks=[],
            tomorrow_suggestions=[],
            stats={
                "new_memories": new_memories,
                "new_entities": new_entities,
                "updated_wiki": updated_wiki,
            },
            ai_model="local",
        )

        self.db.add(report)
        self.db.commit()
        self.db.refresh(report)

        logger.info(f"Daily report created for {date_str}")
        return report

    def get_stats_summary(self, days: int = 7) -> dict:
        """获取统计摘要"""
        start_date = (datetime.now() - timedelta(days=days)).strftime("%Y-%m-%d")

        reports = (
            self.db.query(DailyReport)
            .filter(DailyReport.report_date >= start_date)
            .all()
        )

        total_memories = sum(r.stats.get("new_memories", 0) for r in reports if r.stats)
        total_entities = sum(r.stats.get("new_entities", 0) for r in reports if r.stats)
        total_wiki = sum(r.stats.get("updated_wiki", 0) for r in reports if r.stats)

        return {
            "period_days": days,
            "report_count": len(reports),
            "total_memories": total_memories,
            "total_entities": total_entities,
            "total_wiki_updated": total_wiki,
        }
