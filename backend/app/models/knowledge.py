from sqlalchemy import Column, Integer, String, Text, DateTime, ForeignKey
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
    created_at = Column(DateTime, server_default=func.now())
