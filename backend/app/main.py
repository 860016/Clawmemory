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
<<<<<<< HEAD
from app.routers import auth, users, chat, memories, license, models, skills, nodes, backups, knowledge, file_watcher, agents, openclaw_memories
=======
from app.routers import auth, memories, license, backups, knowledge, file_watcher, wiki, pro_features
>>>>>>> fb055c7 (feat: v3.0 - Wiki知识库 + 科技感UI + i18n + Rust PyO3核心 + Pro功能)
import time
import logging

logger = logging.getLogger("clawmemory.api")


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
    except Exception:
        pass
    finally:
        db.close()
    yield


app = FastAPI(
<<<<<<< HEAD
    title="OpenClaw Memory Web",
    version="1.1.0",
=======
    title="ClawMemory",
    version="2.0.0",
>>>>>>> fb055c7 (feat: v3.0 - Wiki知识库 + 科技感UI + i18n + Rust PyO3核心 + Pro功能)
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

# Routers — 只保留核心功能
app.include_router(auth.router)
app.include_router(memories.router)
app.include_router(license.router)
app.include_router(backups.router)
app.include_router(knowledge.router)
app.include_router(file_watcher.router)
<<<<<<< HEAD
app.include_router(agents.router)
app.include_router(openclaw_memories.router)
=======
app.include_router(wiki.router)
app.include_router(pro_features.router)
>>>>>>> fb055c7 (feat: v3.0 - Wiki知识库 + 科技感UI + i18n + Rust PyO3核心 + Pro功能)


@app.get("/api/v1/health")
async def health_check():
<<<<<<< HEAD
    return {"status": "ok", "version": "1.1.0"}
=======
    return {"status": "ok", "version": "2.0.0"}
>>>>>>> fb055c7 (feat: v3.0 - Wiki知识库 + 科技感UI + i18n + Rust PyO3核心 + Pro功能)


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
    security_level = "high" if core_engine == "rust" else "low"

    return {
        "plugin": "ClawMemory",
<<<<<<< HEAD
        "version": "1.2.0",
=======
        "version": "2.0.0",
>>>>>>> fb055c7 (feat: v3.0 - Wiki知识库 + 科技感UI + i18n + Rust PyO3核心 + Pro功能)
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
                    "python": "纯 Python — 低安全，仅 OSS 免费版",
                }.get(core_engine, "unknown"),
            },
        },
        "next_steps": _get_next_steps(is_licensed, frontend_ready, pubkey_exists, core_engine),
    }


def _get_next_steps(is_licensed: bool, frontend_ready: bool, pubkey_ok: bool, core_engine: str) -> list:
    steps = []
    if core_engine == "python":
        steps.append("⚠️ 安全引擎为纯 Python 模式，授权保护较弱。建议安装 Rust 或 Cython 编译环境")
    if not pubkey_ok:
        steps.append("⚠️ RSA 公钥缺失，无法验证授权签名。请确保授权服务器可访问或手动放置 backend/keys/public.pem")
    if not is_licensed:
<<<<<<< HEAD
        steps.append("当前为 OSS 免费版，在「设置 → 授权管理」中输入授权码激活 Pro/Enterprise 功能")
=======
        steps.append("在「设置 → 授权管理」中输入授权码，激活 Pro 功能")
>>>>>>> fb055c7 (feat: v3.0 - Wiki知识库 + 科技感UI + i18n + Rust PyO3核心 + Pro功能)
    if not frontend_ready:
        steps.append("前端界面未安装，请运行 cd frontend && npm install && npm run build")
    if is_licensed and frontend_ready and pubkey_ok and core_engine != "python":
        steps.append("✅ 一切就绪！访问 http://localhost:8765 开始使用")
    return steps


<<<<<<< HEAD
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
=======
# Serve frontend static files
frontend_dist = settings.base_dir / "frontend" / "dist"
if frontend_dist.exists():
    @app.get("/{full_path:path}")
    async def serve_spa(full_path: str):
        if full_path.startswith("api/"):
            raise HTTPException(status_code=404)
        if full_path.startswith("assets/"):
            raise HTTPException(status_code=404)
        return FileResponse(str(frontend_dist / "index.html"))

    app.mount("/", StaticFiles(directory=str(frontend_dist), html=True), name="frontend")
>>>>>>> fb055c7 (feat: v3.0 - Wiki知识库 + 科技感UI + i18n + Rust PyO3核心 + Pro功能)
