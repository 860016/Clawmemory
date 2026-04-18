from sqlalchemy import Column, Integer, String, Text, DateTime, Boolean
from sqlalchemy.sql import func
from app.database import Base


class License(Base):
    __tablename__ = "licenses"

    id = Column(Integer, primary_key=True, autoincrement=True)
    license_key = Column(String(64), unique=True, nullable=False)
    tier = Column(String(20), nullable=False)  # oss/pro/enterprise
    features = Column(Text, nullable=True)  # JSON array
    status = Column(String(20), default="active")  # active/revoked/expired
    rsa_signature = Column(Text, nullable=True)
    fingerprint_hash = Column(String(64), nullable=True)
    device_name = Column(String(128), nullable=True)
    device_slot = Column(String(10), nullable=True)  # e.g. "2/3"
    last_verified_at = Column(DateTime, nullable=True)
    expires_at = Column(DateTime, nullable=True)
    created_at = Column(DateTime, server_default=func.now())
