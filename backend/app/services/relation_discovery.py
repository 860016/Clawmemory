import json
import logging
from typing import Optional
from sqlalchemy.orm import Session
from app.models.knowledge import Entity, Relation
from app.services.llm_service import llm_service
from app.services.license_service import is_feature_enabled

logger = logging.getLogger(__name__)

VALID_RELATION_TYPES = {
    "works_for", "uses", "depends_on", "located_in", "member_of",
    "created_by", "related_to", "part_of", "causes", "enables",
    "co_occurs", "prefers_alongside",
}


class RelationDiscovery:
    """关系发现引擎：LLM + 共现 + 推理三层"""

    def __init__(self, db: Session):
        self.db = db

    async def discover_from_text(self, text: str, entity_ids: list[int]) -> list[dict]:
        """从文本中发现实体间关系"""
        entities = self.db.query(Entity).filter(Entity.id.in_(entity_ids)).all()
        if len(entities) < 2:
            return []

        entity_names = [e.name for e in entities]

        if is_feature_enabled("ai_extract") and llm_service.available:
            return await self._llm_discover(text, entities, entity_names)
        return self._co_occurrence_discover(text, entities)

    async def _llm_discover(self, text: str, entities: list[Entity], entity_names: list[str]) -> list[dict]:
        """LLM 关系发现"""
        relations_data = await llm_service.extract_relations(text, entity_names)
        created = []

        existing = {(r.source_id, r.target_id, r.relation_type): r for r in self.db.query(Relation).all()}

        for rd in relations_data:
            source_name = rd.get("source", "").lower()
            target_name = rd.get("target", "").lower()
            source_entity = next((e for e in entities if e.name.lower() == source_name), None)
            target_entity = next((e for e in entities if e.name.lower() == target_name), None)

            if not source_entity or not target_entity:
                continue
            if source_entity.id == target_entity.id:
                continue

            rel_type = rd.get("relation_type", "related_to")
            if rel_type not in VALID_RELATION_TYPES:
                rel_type = "related_to"

            key = (source_entity.id, target_entity.id, rel_type)
            if key in existing:
                continue

            relation = Relation(
                user_id=1,
                source_id=source_entity.id,
                target_id=target_entity.id,
                relation_type=rel_type,
                description=rd.get("description", ""),
                confidence=rd.get("confidence", 0.5),
                discover_method="llm",
                weight=rd.get("confidence", 0.5),
            )
            self.db.add(relation)
            created.append(relation)

        self.db.commit()
        return [{"source": r.source_id, "target": r.target_id, "type": r.relation_type} for r in created]

    def _co_occurrence_discover(self, text: str, entities: list[Entity]) -> list[dict]:
        """改进的共现关系发现（句子级）"""
        sentences = re_split_sentences(text)
        created = []

        existing = {(r.source_id, r.target_id): r for r in self.db.query(Relation).all()}

        for sentence in sentences:
            sentence_lower = sentence.lower()
            matched = [e for e in entities if e.name.lower() in sentence_lower]

            for i in range(len(matched)):
                for j in range(i + 1, len(matched)):
                    key = (matched[i].id, matched[j].id)
                    if key in existing:
                        continue

                    relation = Relation(
                        user_id=1,
                        source_id=matched[i].id,
                        target_id=matched[j].id,
                        relation_type="co_occurs",
                        description=f"Co-mentioned in same sentence",
                        confidence=0.3,
                        discover_method="co_occurrence",
                        weight=0.3,
                    )
                    self.db.add(relation)
                    created.append(relation)
                    existing[key] = relation

        if created:
            self.db.commit()
        return [{"source": r.source_id, "target": r.target_id, "type": "co_occurs"} for r in created]

    async def infer_relations(self) -> list[dict]:
        """基于已有关系推理新关系（传递性推理）"""
        all_relations = self.db.query(Relation).all()
        all_entities = {e.id: e for e in self.db.query(Entity).all()}

        # 构建邻接表
        graph = {}
        for r in all_relations:
            graph.setdefault(r.source_id, []).append((r.target_id, r.relation_type))

        inferred = []
        existing = {(r.source_id, r.target_id, r.relation_type) for r in all_relations}

        # 传递性推理：A works_for B, B located_in C → A located_in C
        transitive_rules = {
            ("works_for", "located_in"): "located_in",
            ("part_of", "part_of"): "part_of",
            ("member_of", "part_of"): "member_of",
        }

        for src_id, edges in graph.items():
            for target_id, rel_type in edges:
                if target_id in graph:
                    for next_id, next_rel in graph[target_id]:
                        rule_key = (rel_type, next_rel)
                        if rule_key in transitive_rules:
                            inferred_type = transitive_rules[rule_key]
                            key = (src_id, next_id, inferred_type)
                            if key not in existing:
                                relation = Relation(
                                    user_id=1,
                                    source_id=src_id,
                                    target_id=next_id,
                                    relation_type=inferred_type,
                                    description=f"Inferred: {rel_type} + {next_rel}",
                                    confidence=0.5,
                                    discover_method="inferred",
                                    weight=0.5,
                                )
                                self.db.add(relation)
                                inferred.append(relation)
                                existing.add(key)

        if inferred:
            self.db.commit()
        return [{"source": r.source_id, "target": r.target_id, "type": r.relation_type} for r in inferred]


def re_split_sentences(text: str) -> list[str]:
    """按句子分割文本"""
    import re
    # 中英文句子分割
    sentences = re.split(r'[.!?。！？;；\n]+', text)
    return [s.strip() for s in sentences if len(s.strip()) > 5]
