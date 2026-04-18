import httpx
import platform
import logging
from app.config import settings

logger = logging.getLogger(__name__)


class StatsService:
    async def send_heartbeat(self, db, license_key: str | None = None):
        """Send heartbeat to license server"""
        if not license_key:
            return
        try:
            async with httpx.AsyncClient(timeout=10) as client:
                await client.post(
                    f"{settings.license_server_url}/api/v1/heartbeat",
                    json={
                        "license_key": license_key,
<<<<<<< HEAD
                        "version": "1.2.0",
=======
                        "version": "2.0.0",
>>>>>>> fb055c7 (feat: v3.0 - Wiki知识库 + 科技感UI + i18n + Rust PyO3核心 + Pro功能)
                        "os": platform.system(),
                        "memory_count": self._count_memories(db),
                    }
                )
        except Exception as e:
            logger.warning(f"Heartbeat failed: {e}")

<<<<<<< HEAD
    async def send_ping(self, install_id: str):
        """Send anonymous install ping (open-source)"""
        try:
            async with httpx.AsyncClient(timeout=5) as client:
                await client.post(
                    f"{settings.license_server_url}/api/v1/ping",
                    json={"install_id": install_id, "version": "1.1.0", "os": platform.system()}
                )
        except Exception:
            pass

=======
>>>>>>> fb055c7 (feat: v3.0 - Wiki知识库 + 科技感UI + i18n + Rust PyO3核心 + Pro功能)
    def _count_memories(self, db) -> int:
        from app.models.memory import Memory
        return db.query(Memory).count()


stats_service = StatsService()
