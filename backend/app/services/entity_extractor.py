import json
import re
import logging
from typing import Optional
from sqlalchemy.orm import Session
from app.models.memory import Memory
from app.models.knowledge import Entity, Relation
from app.services.llm_service import llm_service
from app.services.license_service import is_feature_enabled

logger = logging.getLogger(__name__)

VALID_ENTITY_TYPES = {"person", "organization", "location", "concept", "technology", "project", "event", "tool", "product"}

TYPE_KEYWORDS = {
    "person": ["person", "user", "name", "用户", "人", "先生", "女士", "老师", "博士"],
    "organization": ["company", "org", "team", "公司", "团队", "组织", "部门", "大学", "学校"],
    "location": ["city", "country", "location", "城市", "国家", "地方", "省", "区"],
    "technology": ["framework", "library", "language", "tool", "技术", "框架", "库", "语言"],
    "project": ["project", "app", "system", "项目", "系统", "应用", "产品"],
    "event": ["event", "incident", "conference", "事件", "会议", "活动"],
    "tool": ["tool", "utility", "cli", "工具", "插件", "脚本"],
}


class EntityExtractor:
    """实体提取引擎：LLM + 规则双引擎"""

    def __init__(self, db: Session):
        self.db = db

    async def extract_from_memory(self, memory_id: int) -> dict:
        """从单条记忆提取实体"""
        memory = self.db.query(Memory).filter(Memory.id == memory_id).first()
        if not memory:
            return {"entities": 0, "error": "Memory not found"}

        text = f"{memory.key}: {memory.value}"
        return await self._extract(text, memory_id)

    async def extract_batch(self, memory_ids: Optional[list[int]] = None, limit: int = 50) -> dict:
        """批量提取实体"""
        query = self.db.query(Memory).filter(Memory.user_id == 1, Memory.status != "trashed")
        if memory_ids:
            query = query.filter(Memory.id.in_(memory_ids))
        else:
            # 只处理未提取过的记忆
            extracted_ids = self.db.query(Entity.source_memory_id).filter(
                Entity.source_memory_id.isnot(None)
            ).distinct().all()
            extracted_id_set = {r[0] for r in extracted_ids if r[0]}
            query = query.filter(Memory.id.notin_(extracted_id_set))

        memories = query.limit(limit).all()
        if not memories:
            return {"entities_extracted": 0, "relations_extracted": 0, "message": "No memories to process"}

        total_entities = 0
        total_relations = 0

        for m in memories:
            text = f"{m.key}: {m.value}"
            result = await self._extract(text, m.id)
            total_entities += result.get("entities", 0)
            total_relations += result.get("relations", 0)

        return {
            "entities_extracted": total_entities,
            "relations_extracted": total_relations,
            "memories_processed": len(memories),
        }

    async def _extract(self, text: str, source_memory_id: int) -> dict:
        """核心提取逻辑"""
        if is_feature_enabled("ai_extract") and llm_service.available:
            return await self._llm_extract(text, source_memory_id)
        return self._rule_extract(text, source_memory_id)

    async def _llm_extract(self, text: str, source_memory_id: int) -> dict:
        """LLM 实体提取"""
        entities_data = await llm_service.extract_entities(text)
        created_entities = []
        existing_entities = {e.name.lower(): e for e in self.db.query(Entity).filter(Entity.user_id == 1).all()}

        for ed in entities_data:
            name = ed.get("name", "").strip()
            if not name or len(name) < 2 or len(name) > 100:
                continue

            entity_type = ed.get("type", "concept").lower()
            if entity_type not in VALID_ENTITY_TYPES:
                entity_type = "concept"

            if name.lower() in existing_entities:
                existing_entities[name.lower()].confidence = max(
                    existing_entities[name.lower()].confidence,
                    ed.get("confidence", 0.5)
                )
                continue

            entity = Entity(
                user_id=1,
                name=name,
                entity_type=entity_type,
                description=ed.get("description", ""),
                source_memory_id=source_memory_id,
                confidence=ed.get("confidence", 0.5),
                extract_method="llm",
                canonical_name=name,
            )
            self.db.add(entity)
            created_entities.append(entity)
            existing_entities[name.lower()] = entity

        self.db.flush()

        # 提取关系
        entity_names = [e.name for e in created_entities] if created_entities else list(existing_entities.keys())
        relations_data = await llm_service.extract_relations(text, entity_names)
        created_relations = self._create_relations(relations_data, existing_entities, source_memory_id, "llm")

        self.db.commit()
        return {"entities": len(created_entities), "relations": len(created_relations)}

    def _rule_extract(self, text: str, source_memory_id: int) -> dict:
        """规则引擎实体提取"""
        found_terms = set()
        entity_patterns = [
            r'"([^"]+)"',
            r'「([^」]+)」',
            r'【([^】]+)】',
            r'《([^》]+)》',
        ]

        for pattern in entity_patterns:
            for match in re.finditer(pattern, text):
                term = match.group(1).strip()
                if 2 <= len(term) <= 50:
                    found_terms.add(term)

        key_parts = re.split(r'[_\-\s]+', text.split(":")[0] if ":" in text else text)
        for part in key_parts:
            if 3 <= len(part) <= 50 and part.lower() not in ('the', 'and', 'for', 'with', 'this', 'that'):
                found_terms.add(part)

        existing_entities = {e.name.lower(): e for e in self.db.query(Entity).filter(Entity.user_id == 1).all()}
        created_entities = []

        for term in found_terms:
            if term.lower() in existing_entities:
                continue

            entity_type = self._infer_type(term, text)
            entity = Entity(
                user_id=1,
                name=term,
                entity_type=entity_type,
                description=f"Auto-extracted from memory",
                source_memory_id=source_memory_id,
                confidence=0.6,
                extract_method="rule",
                canonical_name=term,
            )
            self.db.add(entity)
            created_entities.append(entity)
            existing_entities[term.lower()] = entity

        self.db.flush()

        # 共现关系
        created_relations = self._co_occurrence_relations(text, existing_entities, source_memory_id)
        self.db.commit()
        return {"entities": len(created_entities), "relations": len(created_relations)}

    def _infer_type(self, term: str, context: str) -> str:
        """推断实体类型"""
        lower_term = term.lower()
        lower_context = context.lower()
        for etype, keywords in TYPE_KEYWORDS.items():
            if any(kw in lower_term or kw in lower_context for kw in keywords):
                return etype
        return "concept"

    def _create_relations(self, relations_data: list[dict], existing_entities: dict, source_memory_id: int, method: str) -> list:
        """创建关系"""
        created = []
        for rd in relations_data:
            source_name = rd.get("source", "").lower()
            target_name = rd.get("target", "").lower()
            source_entity = existing_entities.get(source_name)
            target_entity = existing_entities.get(target_name)

            if not source_entity or not target_entity:
                continue
            if source_entity.id == target_entity.id:
                continue

            exists = self.db.query(Relation).filter(
                Relation.source_id == source_entity.id,
                Relation.target_id == target_entity.id,
                Relation.relation_type == rd.get("relation_type", "related_to"),
            ).first()
            if exists:
                continue

            relation = Relation(
                user_id=1,
                source_id=source_entity.id,
                target_id=target_entity.id,
                relation_type=rd.get("relation_type", "related_to"),
                description=rd.get("description", ""),
                confidence=rd.get("confidence", 0.5),
                discover_method=method,
                source_memory_id=source_memory_id,
                weight=rd.get("confidence", 0.5),
            )
            self.db.add(relation)
            created.append(relation)
        return created

    def _co_occurrence_relations(self, text: str, existing_entities: dict, source_memory_id: int) -> list:
        """共现关系发现"""
        text_lower = text.lower()
        matched = [e for name, e in existing_entities.items() if name in text_lower]
        created = []

        for i in range(len(matched)):
            for j in range(i + 1, len(matched)):
                exists = self.db.query(Relation).filter(
                    Relation.source_id == matched[i].id,
                    Relation.target_id == matched[j].id,
                ).first()
                if exists:
                    continue

                relation = Relation(
                    user_id=1,
                    source_id=matched[i].id,
                    target_id=matched[j].id,
                    relation_type="co_occurs",
                    description=f"Co-mentioned in memory",
                    confidence=0.4,
                    discover_method="co_occurrence",
                    source_memory_id=source_memory_id,
                    weight=0.4,
                )
                self.db.add(relation)
                created.append(relation)
        return created
