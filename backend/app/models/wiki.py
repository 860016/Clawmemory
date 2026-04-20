from sqlalchemy import Column, Integer, String, Text, DateTime, Boolean, Float
from sqlalchemy.sql import func
from app.database import Base


class WikiPage(Base):
    __tablename__ = "wiki_pages"

    id = Column(Integer, primary_key=True, autoincrement=True)
    user_id = Column(Integer, default=1, nullable=False, index=True)
    title = Column(String(300), nullable=False)
    slug = Column(String(300), nullable=False, unique=True)
    content = Column(Text, nullable=False, default="")
    category = Column(String(100), nullable=True)
    tags = Column(Text, nullable=True)  # JSON array as text
    is_pinned = Column(Boolean, default=False)
    parent_id = Column(Integer, nullable=True)  # for nested pages
    status = Column(String(20), default="draft")  # draft/in_progress/completed/archived
    ai_generated = Column(Boolean, default=False)  # 是否 AI 生成
    ai_confidence = Column(Float, default=0.0)  # AI 生成置信度
    source_conversation = Column(Text, nullable=True)  # 来源对话记录
    key_decisions = Column(Text, nullable=True)  # JSON: 关键决策
    action_items = Column(Text, nullable=True)  # JSON: 待办事项
    summary = Column(Text, nullable=True)  # AI 生成的摘要
    created_at = Column(DateTime, server_default=func.now())
    updated_at = Column(DateTime, server_default=func.now(), onupdate=func.now())
