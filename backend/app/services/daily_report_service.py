import logging
from app.core.pro_loader import get_pro_class

logger = logging.getLogger(__name__)


def get_daily_report_service_class():
    """获取 DailyReportService 类（从 Pro 模块加载）"""
    cls = get_pro_class("daily_report_service", "DailyReportService")
    if cls is None:
        logger.warning("Pro module not available, daily report feature disabled")
    return cls
