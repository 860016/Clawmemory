import logging
from datetime import datetime
from apscheduler.schedulers.background import BackgroundScheduler
from app.database import SessionLocal
from app.pro.pro_loader import get_pro_class
import asyncio

logger = logging.getLogger(__name__)

scheduler = BackgroundScheduler()


def generate_daily_report_sync():
    """定时任务：生成日报（同步包装）"""
    logger.info("Starting daily report generation...")
    db = SessionLocal()
    try:
        cls = get_pro_class("daily_report_service", "DailyReportService")
        if cls is None:
            logger.warning("Pro module not available, skipping daily report")
            return

        service = cls(db)
        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)
        try:
            report = loop.run_until_complete(service.generate_report())
            if report:
                logger.info(f"Daily report generated: {report.report_date}")
            else:
                logger.info("No data for today, skipped report generation")
        finally:
            loop.close()
    except Exception as e:
        logger.error(f"Daily report generation failed: {e}")
    finally:
        db.close()


def init_scheduler():
    """初始化定时任务"""
    scheduler.add_job(
        generate_daily_report_sync,
        "cron",
        hour=22,
        minute=0,
        id="daily_report",
        replace_existing=True,
    )
    logger.info("Scheduler initialized: daily report at 22:00")


def start_scheduler():
    """启动调度器"""
    if not scheduler.running:
        scheduler.start()
        logger.info("Scheduler started")


def stop_scheduler():
    """停止调度器"""
    if scheduler.running:
        scheduler.shutdown()
        logger.info("Scheduler stopped")
