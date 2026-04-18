from sqlalchemy import Column, Integer, String, Text, DateTime
from sqlalchemy.sql import func
from app.database import Base


class Backup(Base):
    __tablename__ = "backups"

    id = Column(Integer, primary_key=True, autoincrement=True)
    user_id = Column(Integer, nullable=False, index=True)
    filename = Column(String(200), nullable=False)
    file_path = Column(String(500), nullable=False)
    file_size = Column(Integer, default=0)
    memory_count = Column(Integer, default=0)
    backup_type = Column(String(20), default="manual")  # manual/auto/scheduled
    status = Column(String(20), default="completed")  # completed/failed/in_progress
    notes = Column(Text, nullable=True)
    created_at = Column(DateTime, server_default=func.now())
