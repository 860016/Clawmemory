import logging
from app.core.pro_loader import get_pro_module

logger = logging.getLogger(__name__)


class _LLMServiceStub:
    """LLM 服务存根，Pro 模块未安装时提供空实现"""

    @property
    def available(self) -> bool:
        return False

    async def extract_entities(self, text: str) -> list[dict]:
        return []

    async def extract_relations(self, text: str, entity_names: list[str]) -> list[dict]:
        return []

    async def extract_knowledge_card(self, conversation: str, is_complete: bool = True) -> dict:
        return {}

    async def disambiguate_entity(self, name: str, context: str, candidates: dict[int, str]) -> dict:
        return {}

    async def classify_text(self, text: str, categories: list[str]) -> dict:
        return {}


def _get_llm_service():
    """获取 LLM 服务实例（优先从 Pro 模块加载）"""
    pro_module = get_pro_module("llm_service")
    if pro_module and hasattr(pro_module, "llm_service"):
        return pro_module.llm_service
    return _LLMServiceStub()


llm_service = _get_llm_service()
