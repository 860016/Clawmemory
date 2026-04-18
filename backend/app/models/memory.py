from sqlalchemy import Column, Integer, String, Text, Float, DateTime, Boolean
from sqlalchemy.sql import func
from app.database import Base


class Memory(Base):
    __tablename__ = "memories"

    id = Column(Integer, primary_key=True, autoincrement=True)
    user_id = Column(Integer, default=1, nullable=False, index=True)  # 单用户模式，固定为1
    layer = Column(String(20), nullable=False)  # preference/knowledge/short_term/private
    key = Column(String(200), nullable=False)
    value = Column(Text, nullable=False)
    importance = Column(Float, default=0.5)
    access_count = Column(Integer, default=0)
    last_accessed_at = Column(DateTime, nullable=True)
    is_encrypted = Column(Boolean, default=False)
    tags = Column(Text, nullable=True)  # JSON array stored as text
    source = Column(String(50), default="manual")  # manual/import/auto
    created_at = Column(DateTime, server_default=func.now())
    updated_at = Column(DateTime, server_default=func.now(), onupdate=func.now())
