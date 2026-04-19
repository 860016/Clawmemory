from contextlib import asynccontextmanager
from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from fastapi.staticfiles import StaticFiles
from fastapi.responses import FileResponse
from sqlalchemy import text
from starlette.types import ASGIApp, Receive, Scope, Send
from app.config import settings
from app.database import init_db
from app.services.setup_service import ensure_data_dirs
from app.routers import auth, memories, license, backups, knowledge, file_watcher, wiki, pro_features, openclaw_memories
import asyncio
import time
import logging


logger = logging.getLogger("clawmemory.api")


# ========== Background Scheduler ==========

_scheduler_task: asyncio.Task | None = None


async def _background_scheduler():
    """Background task for auto decay and auto backup"""
    from app.services.license_service import is_feature_enabled
    from app.config import settings

    while True:
        await asyncio.sleep(3600)  # Check every hour
        try:
            # Auto Decay
            if is_feature_enabled("auto_decay"):
                from app.database import SessionLocal
                from app.models.memory import Memory
                from app.services import license_service as core
                import time as _time

                db = SessionLocal()
                try:
                    memories = db.query(Memory).filter(Memory.user_id == 1).all()
                    memory_data = [
                        {"id": m.id, "importance": m.importance,
                         "last_accessed_at": m.last_accessed_at.timestamp() if m.last_accessed_at else 0}
                        for m in memories
                    ]
                    results = core.decay_batch(memory_data)
                    for r in results:
                        m = db.query(Memory).filter(Memory.id == r["memory_id"]).first()
                        if m:
                            if r["should_prune"]:
                                db.delete(m)
                            else:
                                m.importance = r["new_importance"]
                    db.commit()
                    logger.info("Auto decay applied: %d memories processed", len(results))
                except Exception as e:
                    logger.warning("Auto decay failed: %s", e)
                finally:
                    db.close()

            # Auto Backup
            if is_feature_enabled("auto_backup"):
                import json
                from pathlib import Path

                schedule_file = settings.data_dir / "backup_schedule.json"
                if schedule_file.exists():
                    schedule = json.loads(schedule_file.read_text())
                    if schedule.get("enabled"):
                        from app.database import SessionLocal
                        from app.services.backup_service import BackupService

                        db = SessionLocal()
                        try:
                            svc = BackupService(db)
                            svc.create_backup(1, notes="Auto backup")
                            logger.info("Auto backup created")
                        except Exception as e:
                            logger.warning("Auto backup failed: %s", e)
                        finally:
                            db.close()

        except Exception as e:
            logger.warning("Background scheduler error: %s", e)


# ========== ASGI-level 中间件 ==========

class ErrorHandlerASGI:
    """错误处理中间件"""
    def __init__(self, app: ASGIApp):
        self.app = app

    async def __call__(self, scope: Scope, receive: Receive, send: Send):
        if scope["type"] not in ("http", "websocket"):
            await self.app(scope, receive, send)
            return
        try:
            await self.app(scope, receive, send)
        except Exception as e:
            if scope["type"] == "http":
                from starlette.responses import JSONResponse
                logger.error("Unhandled exception: %s", e, exc_info=True)
                response = JSONResponse(
                    status_code=500,
                    content={"error": "INTERNAL_ERROR", "code": "SYS_001", "message": "Internal server error"},
                )
                await response(scope, receive, send)
            else:
                await send({"type": "websocket.close", "code": 1011, "reason": "Internal error"})


class RequestLoggingASGI:
    """请求日志中间件"""
    def __init__(self, app: ASGIApp):
        self.app = app

    async def __call__(self, scope: Scope, receive: Receive, send: Send):
        if scope["type"] != "http":
            await self.app(scope, receive, send)
            return
        start = time.perf_counter()
        path = scope.get("path", "")
        method = scope.get("method", "")

        async def send_with_log(message):
            if message["type"] == "http.response.start":
                duration_ms = (time.perf_counter() - start) * 1000
                status = message.get("status", 0)
                if path != "/api/v1/health":
                    logger.info("%s %s %d %.1fms", method, path, status, duration_ms)
            await send(message)

        await self.app(scope, receive, send_with_log)


class RateLimitASGI:
    """限流中间件"""
    def __init__(self, app: ASGIApp, requests_per_minute: int = 60, burst: int = 10):
        self.app = app
        self.requests_per_minute = requests_per_minute
        self.burst = burst
        self._requests: dict = {}

    def _cleanup(self, ip: str, now: float) -> None:
        cutoff = now - 60
        if ip in self._requests:
            self._requests[ip] = [t for t in self._requests[ip] if t > cutoff]

    async def __call__(self, scope: Scope, receive: Receive, send: Send):
        if scope["type"] != "http":
            await self.app(scope, receive, send)
            return

        path = scope.get("path", "")
        if path.startswith("/assets") or path == "/api/v1/health" or path == "/":
            await self.app(scope, receive, send)
            return

        ip = "unknown"
        for header_name, header_value in scope.get("headers", []):
            if header_name == b"x-forwarded-for":
                ip = header_value.decode().split(",")[0].strip()
                break
        if ip == "unknown":
            client = scope.get("client")
            if client:
                ip = client[0]

        now = time.monotonic()
        self._cleanup(ip, now)

        requests = self._requests.get(ip, [])
        recent = [t for t in requests if t > now - 1]
        if len(recent) >= self.burst:
            from starlette.responses import JSONResponse
            response = JSONResponse(
                status_code=429,
                content={"error": "RATE_LIMITED", "message": "Too many requests."},
            )
            await response(scope, receive, send)
            return

        if len(requests) >= self.requests_per_minute:
            from starlette.responses import JSONResponse
            response = JSONResponse(
                status_code=429,
                content={"error": "RATE_LIMITED", "message": "Rate limit exceeded."},
            )
            await response(scope, receive, send)
            return

        if ip not in self._requests:
            self._requests[ip] = []
        self._requests[ip].append(now)

        await self.app(scope, receive, send)


@asynccontextmanager
async def lifespan(app: FastAPI):
    ensure_data_dirs()
    init_db()
    # 加载缓存的授权信息
    from app.database import SessionLocal
    from app.services.license_service import LicenseService
    db = SessionLocal()
    try:
        LicenseService(db).load_cached_license()
    except Exception as e:
        logger.warning(f"Failed to load cached license: {e}")
    finally:
        db.close()
    # 启动后台调度器
    global _scheduler_task
    _scheduler_task = asyncio.create_task(_background_scheduler())
    yield
    # 停止后台调度器
    if _scheduler_task:
        _scheduler_task.cancel()


app = FastAPI(
    title="ClawMemory",
    version="2.2.0",
    docs_url="/api/docs",
    redoc_url="/api/redoc",
    lifespan=lifespan,
)

# ASGI-level 中间件
app.add_middleware(RateLimitASGI, requests_per_minute=60, burst=10)
app.add_middleware(RequestLoggingASGI)
app.add_middleware(ErrorHandlerASGI)
app.add_middleware(
    CORSMiddleware,
    allow_origins=settings.cors_origins,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Routers
app.include_router(auth.router)
app.include_router(memories.router)
app.include_router(license.router)
app.include_router(backups.router)
app.include_router(knowledge.router)
app.include_router(file_watcher.router)
app.include_router(wiki.router)
app.include_router(pro_features.router)
app.include_router(openclaw_memories.router)


@app.get("/api/v1/health")
async def health_check():
    return {"status": "ok", "version": "2.2.0"}


@app.get("/api/v1/install-status")
async def install_status():
    """ClawMemory 安装状态报告"""
    from app.services.license_service import current_tier, is_feature_enabled, get_core_engine

    tier = current_tier()
    is_licensed = tier != "oss"
    frontend_ready = frontend_dist.exists()
    pubkey_exists = settings.rsa_public_key_path.exists()

    try:
        from app.database import engine
        with engine.connect() as conn:
            conn.execute(text("SELECT 1"))
        db_ok = True
    except Exception:
        db_ok = False

    core_engine = get_core_engine()

    # 安全等级评估
    if core_engine == "rust":
        security_level = "high"
    elif core_engine == "c":
        security_level = "high"
    else:
        security_level = "low"

    return {
        "plugin": "ClawMemory",
        "version": "2.2.0",
        "status": "running" if db_ok else "degraded",
        "service_url": f"http://localhost:{settings.port}",
        "security_level": security_level,
        "checks": {
            "database": "ok" if db_ok else "error",
            "license": {
                "tier": tier,
                "activated": is_licensed,
                "public_key": "ok" if pubkey_exists else "missing",
            },
            "frontend": "installed" if frontend_ready else "not_built",
            "security_engine": {
                "type": core_engine,
                "level": security_level,
                "description": {
                    "rust": "Rust (PyO3) — 最高安全，RSA 硬验证",
                    "c": "C/CPython — 高安全，RSA 硬验证",
                    "python": "纯 Python — 低安全，仅 OSS 免费版",
                }.get(core_engine, "unknown"),
            },
        },
        "next_steps": _get_next_steps(is_licensed, frontend_ready, pubkey_exists, core_engine),
    }


@app.get("/api/v1/stats")
async def get_dashboard_stats():
    """Dashboard 统计数据 — 一次返回所有统计"""
    from app.database import SessionLocal
    from app.models.memory import Memory
    from app.services.license_service import current_tier, is_feature_enabled

    db = SessionLocal()
    try:
        # 记忆总数 + 按层级统计
        total = db.query(Memory).count()
        layer_stats = {}
        for layer in ['preference', 'knowledge', 'short_term', 'private']:
            count = db.query(Memory).filter(Memory.layer == layer).count()
            if count > 0:
                layer_stats[layer] = count

        # 实体数
        entity_count = 0
        try:
            from app.models.knowledge import Entity
            entity_count = db.query(Entity).count()
        except Exception:
            pass

        # Wiki 页面数
        wiki_count = 0
        try:
            from app.models.wiki import WikiPage
            wiki_count = db.query(WikiPage).count()
        except Exception:
            pass

        # 最近记忆
        recent = db.query(Memory).order_by(Memory.updated_at.desc()).limit(5).all()
        recent_list = [
            {
                "id": m.id, "key": m.key, "value": m.value,
                "layer": m.layer, "updated_at": str(m.updated_at) if m.updated_at else None,
            }
            for m in recent
        ]

        # 授权信息
        tier = current_tier()
        license_info = {"tier": tier, "active": tier != "oss"}
        if tier != "oss":
            from app.models.license import License
            lic = db.query(License).filter(License.status == "active").first()
            if lic:
                license_info.update({
                    "type": lic.tier,
                    "expires_at": str(lic.expires_at) if lic.expires_at else None,
                    "device_slot": lic.device_slot or "",
                })

        return {
            "memoryCount": total,
            "entityCount": entity_count,
            "wikiCount": wiki_count,
            "layerStats": layer_stats,
            "recentMemories": recent_list,
            "license": license_info,
            "passwordSet": bool(settings.access_password),
            "version": "2.2.0",
        }
    finally:
        db.close()


def _get_next_steps(is_licensed: bool, frontend_ready: bool, pubkey_ok: bool, core_engine: str) -> list:
    steps = []
    if core_engine == "python":
        steps.append("⚠️ 安全引擎为纯 Python 模式，授权保护较弱。建议安装 clawmemory-core (C/Rust 编译版)")
    if not pubkey_ok:
        steps.append("⚠️ RSA 公钥缺失，无法验证授权签名。请确保授权服务器可访问或手动放置 backend/keys/public.pem")
    if not is_licensed:
        steps.append("在「设置 → 授权管理」中输入授权码，激活 Pro 功能")
    if not frontend_ready:
        steps.append("前端界面未安装，请运行 cd frontend && npm install && npm run build")
    if is_licensed and frontend_ready and pubkey_ok and core_engine != "python":
        steps.append("✅ 一切就绪！访问 http://localhost:8765 开始使用")
    return steps


# Serve frontend static files (built Vue3 app)
# SPA fallback: 非 API 的 GET 请求返回 index.html，由 Vue Router 处理前端路由
frontend_dist = settings.base_dir / "frontend_dist"
if frontend_dist.exists():
    # 挂载 /assets 目录的静态资源（JS/CSS/图片）
    if (frontend_dist / "assets").exists():
        app.mount("/assets", StaticFiles(directory=str(frontend_dist / "assets")), name="frontend-assets")

    # SPA catch-all: 所有非 API 的 GET 请求返回 index.html
    # FastAPI 路由按注册顺序匹配，API 路由已先注册，不会被此路由截获
    @app.get("/{full_path:path}", include_in_schema=False)
    async def spa_fallback(full_path: str):
        from starlette.responses import FileResponse
        index_html = frontend_dist / "index.html"
        if index_html.exists():
            return FileResponse(str(index_html), media_type="text/html")
        return {"error": "Frontend not built"}
