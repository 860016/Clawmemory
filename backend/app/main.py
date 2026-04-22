from contextlib import asynccontextmanager
from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from fastapi.staticfiles import StaticFiles
from fastapi.responses import FileResponse
from sqlalchemy import text
from starlette.types import ASGIApp, Receive, Scope, Send
from app.config import settings, APP_VERSION
from app.database import init_db
from app.services.setup_service import ensure_data_dirs
from app.routers import auth, memories, license, backups, knowledge, file_watcher, wiki, openclaw_memories, openclaw_skills, daily_reports, chromadb
from app.pro.routers import pro_features
import asyncio
import time
import logging
import json


logger = logging.getLogger("clawmemory.api")


# ========== Background Scheduler ==========

_scheduler_task: asyncio.Task | None = None


async def _background_scheduler():
    """Background task for auto decay and auto backup
    
    新的衰减策略（逐步降级 + 回收站机制）：
    - 15天未访问：轻度衰减 10%
    - 30天未访问：中度衰减 30%，标记为 archived
    - 60天未访问：进入回收站（trashed）
    - 回收站保留 30天后自动清空
    
    用户可以通过设置开启/关闭自动衰减，Pro 用户默认开启。
    """
    from app.services.license_service import is_feature_enabled, current_tier
    from app.config import settings
    from datetime import datetime, timezone

    while True:
        await asyncio.sleep(3600)  # Check every hour
        try:
            # Auto Decay - Pro 功能，只有 Pro/Enterprise 用户才能使用
            from app.services.license_service import is_feature_enabled
            decay_enabled = is_feature_enabled("auto_decay")
            
            if decay_enabled:
                from app.database import SessionLocal
                from app.models.memory import Memory
                from app.services import license_service as core

                db = SessionLocal()
                try:
                    now = time.time()
                    
                    # 处理所有记忆（包括 active, archived, trashed）
                    memories = db.query(Memory).filter(Memory.user_id == 1).all()
                    memory_data = [
                        {
                            "id": m.id,
                            "importance": m.importance,
                            "last_accessed_at": m.last_accessed_at.timestamp() if m.last_accessed_at else 0,
                            "status": m.status or "active",
                            "trashed_at": m.trashed_at.timestamp() if m.trashed_at else 0,
                        }
                        for m in memories
                    ]
                    results = core.decay_batch(memory_data, now)
                    
                    archived_count = 0
                    trashed_count = 0
                    pruned_count = 0
                    
                    for r in results:
                        m = db.query(Memory).filter(Memory.id == r["memory_id"]).first()
                        if m:
                            # 更新重要性
                            m.importance = r["new_importance"]
                            m.decay_stage = r["decay_stage"]
                            
                            # 更新状态
                            new_status = r.get("new_status", m.status)
                            if new_status != m.status:
                                m.status = new_status
                                if new_status == "archived":
                                    archived_count += 1
                                elif new_status == "trashed":
                                    m.trashed_at = datetime.now(timezone.utc)
                                    trashed_count += 1
                            
                            # 回收站中的记忆超过 30天，永久删除
                            if r.get("should_prune"):
                                db.delete(m)
                                pruned_count += 1
                    
                    db.commit()
                    logger.info(
                        "Auto decay applied: %d processed, %d archived, %d trashed, %d pruned",
                        len(results), archived_count, trashed_count, pruned_count
                    )
                except Exception as e:
                    logger.warning("Auto decay failed: %s", e)
                finally:
                    db.close()

            # Auto Backup
            if is_feature_enabled("auto_backup"):
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
    # 启动 APScheduler（日报等定时任务）
    try:
        from app.scheduler import init_scheduler, start_scheduler
        init_scheduler()
        start_scheduler()
        logger.info("APScheduler started")
    except Exception as e:
        logger.warning(f"Failed to start APScheduler: {e}")
    yield
    # 停止后台调度器
    if _scheduler_task:
        _scheduler_task.cancel()
    # 停止 APScheduler
    try:
        from app.scheduler import stop_scheduler
        stop_scheduler()
        logger.info("APScheduler stopped")
    except Exception:
        pass


app = FastAPI(
    title="ClawMemory",
    version=APP_VERSION,
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
app.include_router(openclaw_skills.router)
app.include_router(daily_reports.router)
app.include_router(chromadb.router)


@app.get("/api/v1/health")
async def health_check():
    return {"status": "ok", "version": APP_VERSION}


@app.get("/api/v1/install-status")
async def install_status():
    """ClawMemory 安装状态报告"""
    from app.services.license_service import current_tier, is_feature_enabled, get_compute_engine, get_engine_info

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

    engine_info = get_engine_info()
    compute_engine = get_compute_engine()

    if compute_engine in ("rust", "c"):
        security_level = "high"
    elif compute_engine == "pyx":
        security_level = "medium"
    else:
        security_level = "basic"

    return {
        "plugin": "ClawMemory",
        "version": APP_VERSION,
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
                "type": compute_engine,
                "level": security_level,
                "can_activate": engine_info["can_activate"],
                "description": {
                    "rust": "Rust (PyO3) — 最高安全，RSA 硬验证",
                    "c": "C/CPython — 高安全，RSA 硬验证",
                    "pyx": ".pyx 编译版 — 中等安全，计算可用，授权需 Pro 模块",
                    "none": "纯 Python 兜底 — 基础功能可用",
                }.get(compute_engine, "unknown"),
            },
        },
        "next_steps": _get_next_steps(is_licensed, frontend_ready, pubkey_exists, engine_info),
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

        # 关系数
        relation_count = 0
        try:
            from app.models.knowledge import Relation
            relation_count = db.query(Relation).count()
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
            "relationCount": relation_count,
            "wikiCount": wiki_count,
            "layerStats": layer_stats,
            "recentMemories": recent_list,
            "license": license_info,
            "passwordSet": bool(settings.access_password),
            "version": APP_VERSION,
        }
    finally:
        db.close()


@app.get("/api/v1/stats/usage")
async def get_usage_stats(days: int = 30):
    """详细使用统计 — 趋势图、来源分布、重要度分布、Token 估算、访问排行"""
    from app.database import SessionLocal
    from app.models.memory import Memory
    from sqlalchemy import func as sa_func
    from datetime import datetime, timedelta

    db = SessionLocal()
    try:
        now = datetime.utcnow()
        start_date = now - timedelta(days=days)

        # 1. 每日新增趋势
        daily_trend = []
        for i in range(days):
            day = start_date + timedelta(days=i)
            day_end = day + timedelta(days=1)
            count = db.query(Memory).filter(
                Memory.created_at >= day,
                Memory.created_at < day_end,
            ).count()
            daily_trend.append({"date": day.strftime("%Y-%m-%d"), "count": count})

        # 2. 来源分布
        source_dist = {}
        for m in db.query(Memory.source, sa_func.count(Memory.id)).group_by(Memory.source).all():
            source_dist[m[0] or "manual"] = m[1]

        # 3. 重要度分布
        importance_buckets = {"high": 0, "medium": 0, "low": 0}
        for m in db.query(Memory.importance).all():
            imp = m[0] or 0.5
            if imp >= 0.7:
                importance_buckets["high"] += 1
            elif imp >= 0.3:
                importance_buckets["medium"] += 1
            else:
                importance_buckets["low"] += 1

        # 4. Token 估算（按层级）
        memories = db.query(Memory).all()
        layer_tokens = {}
        total_tokens = 0
        for m in memories:
            tokens = len((m.key + " " + m.value).split())
            layer_tokens.setdefault(m.layer, 0)
            layer_tokens[m.layer] += tokens
            total_tokens += tokens

        # 5. 访问排行 Top 10
        top_accessed = db.query(Memory).filter(
            Memory.access_count > 0,
        ).order_by(Memory.access_count.desc()).limit(10).all()
        top_accessed_list = [
            {"id": m.id, "key": m.key, "access_count": m.access_count, "layer": m.layer}
            for m in top_accessed
        ]

        # 6. 记忆总 Token 按日累积趋势
        daily_token_trend = []
        cumulative = 0
        for day_data in daily_trend:
            day = datetime.strptime(day_data["date"], "%Y-%m-%d")
            day_end = day + timedelta(days=1)
            day_memories = db.query(Memory).filter(
                Memory.created_at >= day,
                Memory.created_at < day_end,
            ).all()
            day_tokens = sum(len((m.key + " " + m.value).split()) for m in day_memories)
            cumulative += day_tokens
            daily_token_trend.append({"date": day_data["date"], "tokens": day_tokens, "cumulative": cumulative})

        # 7. 操作计数（按来源和类型统计次数）
        operation_counts = {}
        operation_counts["manual"] = db.query(Memory).filter(Memory.source == "manual").count()
        operation_counts["import"] = db.query(Memory).filter(Memory.source == "import").count()
        operation_counts["auto"] = db.query(Memory).filter(Memory.source == "auto").count()

        # 8. 实体类型分布
        entity_type_dist = {}
        try:
            from app.models.knowledge import Entity
            for row in db.query(Entity.entity_type, sa_func.count(Entity.id)).group_by(Entity.entity_type).all():
                entity_type_dist[row[0] or "other"] = row[1]
        except Exception:
            pass

        return {
            "dailyTrend": daily_trend,
            "dailyTokenTrend": daily_token_trend,
            "sourceDistribution": source_dist,
            "importanceDistribution": importance_buckets,
            "tokenByLayer": layer_tokens,
            "totalEstimatedTokens": total_tokens,
            "topAccessed": top_accessed_list,
            "operationCounts": operation_counts,
            "entityTypeDistribution": entity_type_dist,
            "totalMemories": len(memories),
            "days": days,
        }
    finally:
        db.close()


def _get_next_steps(is_licensed: bool, frontend_ready: bool, pubkey_ok: bool, engine_info: dict) -> list:
    steps = []
    can_activate = engine_info.get("can_activate", False)
    compute_engine = engine_info.get("compute_engine", "none")
    if not can_activate:
        steps.append("ℹ️ 核心安全引擎未安装，Pro 功能将在激活时自动下载安装")
    if not pubkey_ok:
        steps.append("⚠️ RSA 公钥缺失，无法验证授权签名。请确保授权服务器可访问或手动放置 backend/keys/public.pem")
    if not is_licensed:
        steps.append("在「设置 → 授权管理」中输入授权码，激活 Pro 功能")
    if not frontend_ready:
        steps.append("前端界面未安装，请运行 cd frontend && npm install && npm run build")
    if is_licensed and frontend_ready and pubkey_ok:
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
