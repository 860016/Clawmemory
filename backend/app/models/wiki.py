from sqlalchemy import Column, Integer, String, Text, DateTime, Boolean
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
    created_at = Column(DateTime, server_default=func.now())
    updated_at = Column(DateTime, server_default=func.now(), onupdate=func.now())
