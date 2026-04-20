from sqlalchemy import Column, Integer, String, Text, DateTime, Float, ForeignKey
from sqlalchemy.sql import func
from app.database import Base


class Entity(Base):
    __tablename__ = "entities"

    id = Column(Integer, primary_key=True, autoincrement=True)
    user_id = Column(Integer, default=1, nullable=False, index=True)
    name = Column(String(200), nullable=False)
    entity_type = Column(String(50), nullable=False)
    description = Column(Text, nullable=True)
    properties = Column(Text, nullable=True)  # JSON
    source_memory_id = Column(Integer, nullable=True)  # 来源记忆 ID
    confidence = Column(Float, default=1.0)  # 提取置信度
    extract_method = Column(String(20), default="manual")  # manual/llm/rule
    canonical_name = Column(String(200), nullable=True)  # 标准化名称（消歧用）
    aliases = Column(Text, nullable=True)  # JSON: 别名列表
    embedding_id = Column(String(100), nullable=True)  # 向量索引 ID
    created_at = Column(DateTime, server_default=func.now())
    updated_at = Column(DateTime, server_default=func.now(), onupdate=func.now())


class Relation(Base):
    __tablename__ = "relations"

    id = Column(Integer, primary_key=True, autoincrement=True)
    user_id = Column(Integer, default=1, nullable=False, index=True)
    source_id = Column(Integer, ForeignKey("entities.id", ondelete="CASCADE"), nullable=False)
    target_id = Column(Integer, ForeignKey("entities.id", ondelete="CASCADE"), nullable=False)
    relation_type = Column(String(100), nullable=False)
    description = Column(Text, nullable=True)
    confidence = Column(Float, default=1.0)  # 关系置信度
    discover_method = Column(String(20), default="manual")  # manual/llm/co_occurrence/inferred
    source_memory_id = Column(Integer, nullable=True)  # 来源记忆 ID
    weight = Column(Float, default=1.0)  # 关系权重（用于图算法）
    created_at = Column(DateTime, server_default=func.now())
