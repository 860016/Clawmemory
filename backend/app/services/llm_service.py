import httpx
import json
import logging
from typing import Optional
from app.config import settings

logger = logging.getLogger(__name__)


class LLMService:
    """统一的 LLM 调用服务，支持 OpenAI 兼容 API"""

    def __init__(self):
        self.api_key = settings.openai_api_key
        self.base_url = settings.openai_base_url.rstrip("/")
        self.model = settings.llm_model
        self._available = bool(self.api_key)

    @property
    def available(self) -> bool:
        return self._available

    async def _call(self, messages: list[dict], temperature: float = 0.1, max_tokens: int = 4000) -> str:
        """调用 LLM API，返回文本响应"""
        if not self.available:
            raise RuntimeError("LLM not configured: set OPENAI_API_KEY")

        async with httpx.AsyncClient(timeout=60.0) as client:
            resp = await client.post(
                f"{self.base_url}/chat/completions",
                headers={
                    "Authorization": f"Bearer {self.api_key}",
                    "Content-Type": "application/json",
                },
                json={
                    "model": self.model,
                    "messages": messages,
                    "temperature": temperature,
                    "max_tokens": max_tokens,
                },
            )
            resp.raise_for_status()
            data = resp.json()
            return data["choices"][0]["message"]["content"]

    async def _call_json(self, messages: list[dict], temperature: float = 0.1) -> dict:
        """调用 LLM 并解析 JSON 响应"""
        text = await self._call(messages, temperature=temperature)
        # 尝试提取 JSON（可能包含 markdown 代码块）
        if "```" in text:
            # 提取第一个代码块
            start = text.find("```") + 3
            # 跳过语言标识
            nl = text.find("\n", start)
            if nl != -1 and nl < start + 10:
                start = nl + 1
            end = text.find("```", start)
            if end != -1:
                text = text[start:end].strip()
            else:
                text = text[start:].strip()
        return json.loads(text)

    async def extract_entities(self, text: str) -> list[dict]:
        """从文本中提取实体"""
        messages = [
            {"role": "system", "content": (
                "You are an expert at named entity recognition. "
                "Extract all meaningful entities from the given text. "
                "Output ONLY a JSON array, nothing else."
            )},
            {"role": "user", "content": (
                f"Extract entities from this text. For each entity, provide:\n"
                f"- name: the entity name (as it appears in text)\n"
                f"- type: one of [person, organization, location, concept, technology, project, event, tool, product]\n"
                f"- description: a brief description (1 sentence)\n"
                f"- confidence: a number 0-1 indicating confidence\n\n"
                f"Text: {text}\n\n"
                f"Output JSON array only."
            )},
        ]
        try:
            result = await self._call_json(messages)
            if isinstance(result, list):
                return result
            return []
        except Exception as e:
            logger.warning(f"Entity extraction failed: {e}")
            return []

    async def extract_relations(self, text: str, entity_names: list[str]) -> list[dict]:
        """从文本中提取实体间关系"""
        messages = [
            {"role": "system", "content": (
                "You are an expert at relation extraction. "
                "Find semantic relations between entities in the given text. "
                "Output ONLY a JSON array, nothing else."
            )},
            {"role": "user", "content": (
                f"Given the text and entity list, find relations between entities.\n\n"
                f"Text: {text}\n\n"
                f"Entities: {', '.join(entity_names)}\n\n"
                f"For each relation, provide:\n"
                f"- source: source entity name (must be from the entity list)\n"
                f"- target: target entity name (must be from the entity list)\n"
                f"- relation_type: one of [works_for, uses, depends_on, located_in, member_of, created_by, related_to, part_of, causes, enables]\n"
                f"- description: brief description of the relation\n"
                f"- confidence: 0-1\n\n"
                f"Output JSON array only."
            )},
        ]
        try:
            result = await self._call_json(messages)
            if isinstance(result, list):
                return result
            return []
        except Exception as e:
            logger.warning(f"Relation extraction failed: {e}")
            return []

    async def extract_knowledge_card(self, conversation: str, is_complete: bool = True) -> dict:
        """从对话中提取知识卡片"""
        status = "completed" if is_complete else "in_progress"
        messages = [
            {"role": "system", "content": (
                "You are an expert knowledge curator. "
                "Extract and synthesize knowledge from conversations. "
                "Create a well-structured knowledge card. "
                "Output ONLY a JSON object, nothing else."
            )},
            {"role": "user", "content": (
                f"Extract a knowledge card from this conversation.\n\n"
                f"Conversation: {conversation}\n\n"
                f"Discussion status: {status}\n\n"
                f"Output a JSON object with:\n"
                f"- title: concise, descriptive title\n"
                f"- summary: 2-3 sentence summary\n"
                f"- content: well-structured markdown content with key points, decisions, code snippets if any\n"
                f"- category: one of [project, technology, concept, tutorial, troubleshooting, discussion, reference]\n"
                f"- tags: array of relevant tags (3-5)\n"
                f"- status: '{status}'\n"
                f"- key_decisions: array of key decisions or conclusions\n"
                f"- action_items: array of pending actions (empty if completed)\n\n"
                f"Output JSON only."
            )},
        ]
        try:
            result = await self._call_json(messages)
            if isinstance(result, dict):
                return result
            return {}
        except Exception as e:
            logger.warning(f"Knowledge card extraction failed: {e}")
            return {}

    async def disambiguate_entity(self, name: str, context: str, candidates: dict[int, str]) -> dict:
        """消歧：确定 name 在 context 中指向哪个候选实体"""
        candidates_str = "\n".join([f"ID {k}: {v}" for k, v in candidates.items()])
        messages = [
            {"role": "system", "content": (
                "You are an expert at entity disambiguation. "
                "Determine which candidate entity the mention refers to. "
                "Output ONLY a JSON object, nothing else."
            )},
            {"role": "user", "content": (
                f"Which entity does '{name}' refer to in this context?\n\n"
                f"Context: {context}\n\n"
                f"Candidates:\n{candidates_str}\n\n"
                f"Output JSON with:\n"
                f"- entity_id: the ID of the best match (or null if none match)\n"
                f"- confidence: 0-1\n"
                f"- reason: brief explanation\n\n"
                f"Output JSON only."
            )},
        ]
        try:
            result = await self._call_json(messages)
            if isinstance(result, dict):
                return result
            return {}
        except Exception as e:
            logger.warning(f"Entity disambiguation failed: {e}")
            return {}

    async def classify_text(self, text: str, categories: list[str]) -> dict:
        """文本分类"""
        messages = [
            {"role": "system", "content": (
                "Classify the given text into one of the provided categories. "
                "Output ONLY a JSON object, nothing else."
            )},
            {"role": "user", "content": (
                f"Classify this text:\n\n{text}\n\n"
                f"Categories: {', '.join(categories)}\n\n"
                f"Output JSON with:\n"
                f"- category: the best matching category\n"
                f"- confidence: 0-1\n"
                f"- reason: brief explanation\n\n"
                f"Output JSON only."
            )},
        ]
        try:
            result = await self._call_json(messages)
            if isinstance(result, dict):
                return result
            return {}
        except Exception as e:
            logger.warning(f"Text classification failed: {e}")
            return {}


llm_service = LLMService()
