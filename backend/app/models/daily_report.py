from sqlalchemy import Column, Integer, String, Text, Float, DateTime, JSON
from sqlalchemy.sql import func
from app.database import Base


class DailyReport(Base):
    __tablename__ = "daily_reports"

    id = Column(Integer, primary_key=True, autoincrement=True)
    user_id = Column(Integer, default=1, nullable=False, index=True)
    report_date = Column(String(10), nullable=False, index=True)  # YYYY-MM-DD
    summary = Column(Text, nullable=True)
    highlights = Column(JSON, nullable=True)  # Array of strings
    knowledge_gained = Column(JSON, nullable=True)  # Array of strings
    pending_tasks = Column(JSON, nullable=True)  # Array of strings
    tomorrow_suggestions = Column(JSON, nullable=True)  # Array of strings
    stats = Column(JSON, nullable=True)  # {new_memories, new_entities, updated_wiki, total_tokens}
    raw_data = Column(JSON, nullable=True)  # 原始数据，供重新生成使用
    ai_model = Column(String(50), nullable=True)  # 使用的 AI 模型
    generated_at = Column(DateTime, server_default=func.now())
    is_pushed = Column(Integer, default=0)  # 0=未推送, 1=已推送
    push_time = Column(DateTime, nullable=True)
