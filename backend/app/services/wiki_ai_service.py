import logging
from app.pro.pro_loader import get_pro_class

logger = logging.getLogger(__name__)


def get_wiki_ai_service_class():
    """获取 WikiAIService 类（从 Pro 模块加载）"""
    cls = get_pro_class("wiki_ai_service", "WikiAIService")
    if cls is None:
        logger.warning("Pro module not available, wiki AI feature disabled")
    return cls
