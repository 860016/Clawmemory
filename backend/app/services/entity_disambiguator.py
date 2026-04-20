import json
import logging
from sqlalchemy.orm import Session
from sqlalchemy import or_
from app.models.knowledge import Entity
from app.services.llm_service import llm_service
from app.services.license_service import is_feature_enabled

logger = logging.getLogger(__name__)


class EntityDisambiguator:
    """实体消歧引擎：别名 → 向量 → LLM 三级"""

    def __init__(self, db: Session):
        self.db = db

    def find_candidates(self, name: str, context: str = "") -> list[Entity]:
        """查找可能的匹配实体"""
        candidates = []
        name_lower = name.lower()

        # Step 1: 精确/别名匹配
        exact = self.db.query(Entity).filter(
            or_(
                Entity.name == name,
                Entity.canonical_name == name,
            )
        ).all()
        candidates.extend(exact)

        # Step 2: 模糊匹配（别名 JSON 中搜索）
        all_entities = self.db.query(Entity).filter(Entity.user_id == 1).all()
        for e in all_entities:
            if e in candidates:
                continue
            if e.aliases:
                try:
                    aliases = json.loads(e.aliases)
                    if name_lower in [a.lower() for a in aliases]:
                        candidates.append(e)
                except (json.JSONDecodeError, TypeError):
                    pass

        # Step 3: 名称相似度匹配（编辑距离）
        if not candidates:
            for e in all_entities:
                if self._similarity(name_lower, e.name.lower()) > 0.8:
                    candidates.append(e)

        return candidates

    async def disambiguate(self, name: str, context: str) -> Entity | None:
        """消歧：确定 name 在 context 中指向哪个已有实体"""
        candidates = self.find_candidates(name, context)
        if not candidates:
            return None
        if len(candidates) == 1:
            return candidates[0]

        # Step 3: LLM 消歧（Pro 功能）
        if is_feature_enabled("ai_extract") and llm_service.available:
            descriptions = {
                e.id: f"{e.name}({e.entity_type}): {e.description or ''}"
                for e in candidates
            }
            result = await llm_service.disambiguate_entity(name, context, descriptions)
            entity_id = result.get("entity_id")
            if entity_id and entity_id in descriptions:
                return self.db.query(Entity).filter(Entity.id == entity_id).first()

        # 无 LLM 时：按 entity_type 上下文匹配
        return self._type_context_match(candidates, context)

    def merge_entities(self, source_id: int, target_id: int) -> Entity | None:
        """合并两个实体（保留目标，迁移源的关系和别名）"""
        source = self.db.query(Entity).filter(Entity.id == source_id).first()
        target = self.db.query(Entity).filter(Entity.id == target_id).first()
        if not source or not target:
            return None
        if source.id == target.id:
            return target

        from app.models.knowledge import Relation

        # 迁移关系
        self.db.query(Relation).filter(
            Relation.source_id == source_id
        ).update({"source_id": target_id})
        self.db.query(Relation).filter(
            Relation.target_id == source_id
        ).update({"target_id": target_id})

        # 合并别名
        target_aliases = set()
        if target.aliases:
            try:
                target_aliases = set(json.loads(target.aliases))
            except (json.JSONDecodeError, TypeError):
                pass
        target_aliases.add(source.name)
        if source.aliases:
            try:
                target_aliases.update(json.loads(source.aliases))
            except (json.JSONDecodeError, TypeError):
                pass
        target.aliases = json.dumps(list(target_aliases))

        # 合并描述
        if source.description and not target.description:
            target.description = source.description

        # 删除源实体
        self.db.delete(source)
        self.db.commit()
        return target

    def add_alias(self, entity_id: int, alias: str) -> Entity | None:
        """为实体添加别名"""
        entity = self.db.query(Entity).filter(Entity.id == entity_id).first()
        if not entity:
            return None

        aliases = set()
        if entity.aliases:
            try:
                aliases = set(json.loads(entity.aliases))
            except (json.JSONDecodeError, TypeError):
                pass
        aliases.add(alias)
        entity.aliases = json.dumps(list(aliases))
        self.db.commit()
        return entity

    def _type_context_match(self, candidates: list[Entity], context: str) -> Entity | None:
        """基于上下文类型的匹配"""
        context_lower = context.lower()
        type_scores = {}
        for c in candidates:
            score = 0
            # 检查实体类型关键词是否在上下文中
            if c.entity_type in context_lower:
                score += 2
            # 检查描述中的关键词
            if c.description:
                desc_words = set(c.description.lower().split())
                context_words = set(context_lower.split())
                overlap = len(desc_words & context_words)
                score += overlap
            type_scores[c.id] = score

        if not type_scores:
            return candidates[0]

        best_id = max(type_scores, key=type_scores.get)
        return next(c for c in candidates if c.id == best_id)

    def _similarity(self, s1: str, s2: str) -> float:
        """计算字符串相似度（简单编辑距离）"""
        if not s1 or not s2:
            return 0.0
        len1, len2 = len(s1), len(s2)
        if len1 == 0 or len2 == 0:
            return 0.0

        # 简单字符重叠
        set1, set2 = set(s1), set(s2)
        intersection = len(set1 & set2)
        union = len(set1 | set2)
        return intersection / union if union > 0 else 0.0
