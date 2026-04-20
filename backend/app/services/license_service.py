import base64
import json
import logging
import importlib.util
from datetime import datetime, timezone
from sqlalchemy.orm import Session
from app.models.license import License
from app.config import settings, APP_VERSION
import httpx

logger = logging.getLogger(__name__)

# ========== 授权引擎与计算引擎分离架构 ==========
#
# 安全设计原则：
#   1. 授权检查不依赖模块级变量，每次调用时重新验证
#   2. 使用 importlib.util.find_spec() 检查 clawmemory_core 是否真的存在
#   3. 即使修改了 Python 代码，也无法绕过真正的授权检查
#
# 授权引擎 (License Engine):
#   - 只有 clawmemory_core (C wheel) 才能激活 Pro/Enterprise
#   - .pyx 编译版和纯 Python 都无法设置授权，永远锁定为 OSS
#
# 计算引擎 (Compute Engine):
#   - clawmemory_core → 最高性能 + 安全
#   - .pyx 编译版 → 中等性能，可用于计算但无授权
#   - 纯 Python 兜底 → 基础功能可用

_LICENSE_ENGINE = "none"
_COMPUTE_ENGINE = "none"

# ========== 核心安全函数：检查 clawmemory_core 是否真的存在 ==========

def _has_clawmemory_core() -> bool:
    """
    检查 clawmemory_core 是否真的存在
    
    使用 importlib.util.find_spec() 而不是 try/except，
    这样即使修改了模块级变量，也无法绕过此检查。
    """
    return importlib.util.find_spec("clawmemory_core") is not None


def _get_core_functions():
    """
    从 clawmemory_core 获取授权函数
    
    每次调用时重新导入，确保检查的是真正的模块。
    """
    if not _has_clawmemory_core():
        return None
    
    try:
        import clawmemory_core
        return {
            "check_feature": clawmemory_core.check_feature,
            "set_license": clawmemory_core.set_license,
            "get_tier": clawmemory_core.get_tier,
            "reset": clawmemory_core.reset,
            "verify_integrity": clawmemory_core.verify_integrity,
            "verify_license": clawmemory_core.verify_license,
            "get_build_info": clawmemory_core.get_build_info,
        }
    except (ImportError, AttributeError):
        return None


# ========== Step 1: 初始化时尝试加载 clawmemory_core ==========

_core_funcs = _get_core_functions()
if _core_funcs:
    _build = _core_funcs["get_build_info"]()
    if "c-cpython" in _build:
        _LICENSE_ENGINE = "c"
        if "cng" in _build:
            logger.info("License engine: C/CPython + Windows CNG — high security RSA verification")
        else:
            logger.info("License engine: C/CPython + OpenSSL — high security RSA verification")
    else:
        _LICENSE_ENGINE = "rust"
        logger.info("License engine: Rust (PyO3) — maximum security")
    
    # 尝试加载计算功能
    try:
        import clawmemory_core as _core
        calculate_decay = _core.calculate_decay
        should_prune = _core.should_prune
        reinforce = _core.reinforce
        decay_memory = _core.decay_memory
        decay_batch = _core.decay_batch
        get_decay_stats = _core.get_decay_stats
        detect_conflict = _core.detect_conflict
        resolve_conflict = _core.resolve_conflict
        scan_for_conflicts = _core.scan_for_conflicts
        get_conflict_summary = _core.get_conflict_summary
        estimate_complexity = _core.estimate_complexity
        route_model = _core.route_model
        get_routing_stats = _core.get_routing_stats
        _COMPUTE_ENGINE = _LICENSE_ENGINE
        logger.info("Compute engine: %s (from clawmemory_core)", _COMPUTE_ENGINE)
    except (ImportError, AttributeError):
        logger.warning("clawmemory_core missing compute functions, will try .pyx fallback")
else:
    logger.info("clawmemory_core not installed, license activation will be blocked")

# ========== Step 2: 加载 .pyx 计算引擎（如果需要） ==========

if _COMPUTE_ENGINE == "none":
    try:
        from app.core.memory_decay import (
            calculate_decay, should_prune, reinforce, decay_memory, decay_batch, get_decay_stats,
            get_decay_stage, should_archive, should_trash, should_prune_from_trash, get_stage_info,
        )
        from app.core.conflict_resolver import (
            detect_conflict, resolve_conflict, scan_for_conflicts, get_conflict_summary,
        )
        from app.core.token_router import (
            estimate_complexity, route_model, get_routing_stats,
        )
        _COMPUTE_ENGINE = "pyx"
        logger.info("Compute engine: .pyx compiled — compute functions available, but NO license activation")
    except ImportError as e:
        logger.warning("Failed to load .pyx compute modules: %s", e)

# ========== Step 3: 计算功能兜底 ==========

if _COMPUTE_ENGINE == "none":
    logger.warning("Compute engine: using pure Python fallback — basic functionality only")

    def calculate_decay(importance: float, age_days: float) -> float:
        return importance

    def should_prune(importance: float) -> bool:
        return False

    def reinforce(importance: float, factor: float = 1.5) -> float:
        return importance

    def get_decay_stage(age_days: float) -> int:
        return 0

    def should_archive(age_days: float) -> bool:
        return False

    def should_trash(age_days: float) -> bool:
        return False

    def should_prune_from_trash(trashed_days: float) -> bool:
        return False

    def decay_memory(memory_id: int, current_importance: float, last_accessed_at: float,
                      current_status: str = "active", trashed_at: float = 0.0, now: float = 0.0) -> dict:
        return {
            "memory_id": memory_id,
            "new_importance": current_importance,
            "new_status": current_status,
            "should_prune": False,
            "decay_stage": 0,
            "age_days": 0,
        }

    def decay_batch(memories: list, now: float = 0.0) -> list:
        return []

    def get_decay_stats(memories: list, now: float = 0.0) -> dict:
        return {
            "total": 0, "active": 0, "archived": 0, "trashed": 0,
            "to_archive": 0, "to_trash": 0, "to_prune": 0, "avg_importance": 0.0,
        }

    def get_stage_info() -> dict:
        return {
            "stage_1_days": 15, "stage_2_days": 30, "stage_3_days": 60, "trash_expire_days": 30,
            "description": {"stage_0": "纯 Python 兜底模式，无衰减"},
        }

    def detect_conflict(memory_a: dict, memory_b: dict):
        return None

    def resolve_conflict(conflict: dict, strategy: str = None) -> dict:
        return {"conflict": conflict, "strategy": "none", "winner": None, "action": "none"}

    def scan_for_conflicts(memories: list) -> list:
        return []

    def get_conflict_summary(conflicts: list) -> dict:
        return {"total": 0, "by_severity": {"low": 0, "medium": 0, "high": 0}, "auto_resolvable": 0, "needs_review": 0}

    def estimate_complexity(message: str, context_length: int = 0, has_code: bool = False, has_reasoning: bool = False) -> float:
        return 0.0

    def route_model(message: str, available_models: list, context_length: int = 0, user_tier: str = "oss", budget_remaining: float = None) -> dict:
        return {"selected_model": None, "complexity": 0.0, "routing_reason": "no_engine", "estimated_cost": 0.0, "estimated_tokens": 0}

    def get_routing_stats(routing_history: list) -> dict:
        return {"total": 0, "distribution": {}, "avg_complexity": 0.0, "total_cost": 0.0}


# ========== 授权功能：每次调用时重新检查 clawmemory_core ==========

def check_feature(feature: str) -> bool:
    """
    检查功能是否启用
    
    安全设计：每次调用时重新检查 clawmemory_core 是否存在，
    不依赖模块级变量，防止通过修改代码绕过。
    """
    funcs = _get_core_functions()
    if funcs:
        return funcs["check_feature"](feature)
    return False


def get_tier() -> str:
    """
    获取当前授权等级
    
    安全设计：每次调用时重新检查 clawmemory_core 是否存在。
    """
    funcs = _get_core_functions()
    if funcs:
        return funcs["get_tier"]()
    return "oss"


def set_license(tier: str, features: list):
    """
    设置授权状态
    
    安全设计：只有 clawmemory_core 存在时才能设置，
    否则抛出异常。
    """
    funcs = _get_core_functions()
    if funcs:
        return funcs["set_license"](tier, features)
    logger.warning("set_license() BLOCKED: clawmemory_core not installed")
    raise RuntimeError("License activation requires clawmemory_core (C/Rust wheel)")


def reset():
    """
    重置授权状态
    
    安全设计：只有 clawmemory_core 存在时才能重置。
    """
    funcs = _get_core_functions()
    if funcs:
        return funcs["reset"]()


def verify_integrity() -> bool:
    """
    校验授权数据是否被篡改
    """
    funcs = _get_core_functions()
    if funcs:
        try:
            return funcs["verify_integrity"]()
        except Exception:
            return False
    return True


def _core_verify_license(license_data_b64: str, public_key_pem: str) -> dict:
    """
    RSA 签名验证
    
    安全设计：只有 clawmemory_core 存在时才能验证。
    """
    funcs = _get_core_functions()
    if funcs:
        return funcs["verify_license"](license_data_b64, public_key_pem)
    logger.warning("RSA verify BLOCKED: clawmemory_core not installed")
    return {}


# ========== 辅助函数 ==========

def get_license_engine() -> str:
    """获取授权引擎类型: c/rust/none"""
    if _has_clawmemory_core():
        funcs = _get_core_functions()
        if funcs:
            _build = funcs["get_build_info"]()
            if "c-cpython" in _build:
                return "c"
            return "rust"
    return "none"


def get_compute_engine() -> str:
    """获取计算引擎类型: c/rust/pyx/none"""
    return _COMPUTE_ENGINE


def get_engine_info() -> dict:
    """获取引擎完整信息"""
    has_core = _has_clawmemory_core()
    return {
        "license_engine": get_license_engine(),
        "compute_engine": _COMPUTE_ENGINE,
        "has_clawmemory_core": has_core,
        "can_activate": has_core,
    }


def is_feature_enabled(feature: str) -> bool:
    """检查功能是否启用（别名）"""
    return check_feature(feature)


def current_tier() -> str:
    """获取当前授权等级（别名）"""
    return get_tier()


# ========== v3.0 功能定义 ==========

PRO_FEATURES = {
    "auto_decay": "自动记忆衰减",
    "decay_report": "衰减报告",
    "prune_suggest": "清理建议",
    "reinforce": "记忆强化",
    "conflict_scan": "矛盾扫描",
    "conflict_merge": "自动合并",
    "smart_router": "智能模型路由",
    "token_stats": "Token 统计",
    "ai_extract": "AI 智能提取",
    "auto_graph": "自动知识图谱",
    "unlimited_graph": "无限图谱节点",
    "wiki": "Wiki 知识库",
    "auto_backup": "自动备份",
}

ENTERPRISE_EXTRA_FEATURES = {
    "api_access": "API 完整访问",
    "sso": "SSO 单点登录",
    "audit_log": "审计日志",
    "time_travel": "时间回溯",
    "offline_mode": "离线模式",
}


# ========== LicenseService 类 ==========

class LicenseService:
    def __init__(self, db: Session):
        self.db = db

    def get_active_license(self) -> License | None:
        """获取当前活跃的授权记录"""
        return self.db.query(License).filter(License.status == "active").first()

    async def activate(self, license_key: str) -> dict:
        """
        激活授权码 — 完整流程：
        1. 检查 clawmemory_core 是否存在
        2. 向授权平台发送激活请求
        3. RSA 签名验证
        4. 存入本地数据库
        5. 写入内存
        """
        # Step 0: 安全引擎检查 — 使用 _has_clawmemory_core() 而不是模块变量
        if not _has_clawmemory_core():
            logger.error("Cannot activate: clawmemory_core not installed")
            return {"valid": False, "message": "未安装核心安全引擎，请先安装 clawmemory-core 后再激活"}

        fingerprint = self._get_fingerprint()
        device_name = self._get_device_name()

        # Step 1: 向授权平台请求激活
        try:
            async with httpx.AsyncClient(timeout=15) as client:
                resp = await client.post(
                    f"{settings.license_server_url}/api/v1/activate",
                    json={
                        "license_key": license_key,
                        "fingerprint": fingerprint,
                        "device_name": device_name,
                        "version": APP_VERSION,
                    },
                )
                data = resp.json()
        except httpx.ConnectError:
            return {"valid": False, "message": "无法连接授权服务器，请检查网络连接"}
        except httpx.TimeoutException:
            return {"valid": False, "message": "连接授权服务器超时，请稍后重试"}
        except Exception as e:
            return {"valid": False, "message": f"激活请求失败: {e}"}

        # 授权平台返回错误
        if not data.get("valid"):
            return {"valid": False, "message": data.get("message", "授权码无效")}

        # Step 2: RSA 签名验证（核心安全环节）
        signature_b64 = data.get("signature", "")
        if signature_b64:
            pubkey = self._load_public_key()
            if not pubkey:
                logger.error("RSA public key not available, cannot verify license signature")
                return {"valid": False, "message": "无法获取 RSA 公钥，请检查网络连接后重试"}

            verified_data = _core_verify_license(signature_b64, pubkey)
            if not verified_data:
                # 验签失败 — 尝试刷新公钥后重试
                logger.warning("RSA verify failed with cached key, refreshing from server...")
                fresh_pubkey = self._load_public_key(force_refresh=True)
                if fresh_pubkey and fresh_pubkey != pubkey:
                    verified_data = _core_verify_license(signature_b64, fresh_pubkey)
                    if verified_data:
                        logger.info("RSA signature verification passed after key refresh")
                        pubkey = fresh_pubkey
                    else:
                        logger.warning("RSA signature verification FAILED even with fresh key")
                        return {"valid": False, "message": "RSA 签名验证失败，授权可能被篡改"}
                else:
                    logger.warning("RSA signature verification FAILED")
                    return {"valid": False, "message": "RSA 签名验证失败，授权可能被篡改"}
            else:
                logger.info("RSA signature verification passed")
        else:
            if _has_clawmemory_core():
                logger.error("No signature in activation response but core engine is active")
                return {"valid": False, "message": "授权服务器未返回签名数据"}
            logger.warning("No signature in response (no core engine, skipping RSA verify)")

        # Step 3: 提取授权数据
        tier = data.get("tier", "pro")
        features = data.get("features", [])
        expires_at = data.get("expires_at")
        device_slot = data.get("device_slot", "")

        # 如果 features 为空，按 tier 填充默认值
        if not features:
            if tier == "oss":
                features = []
            elif tier == "enterprise":
                features = list(PRO_FEATURES.keys()) + list(ENTERPRISE_EXTRA_FEATURES.keys())
            else:
                features = list(PRO_FEATURES.keys())

        # Step 4: 存入本地数据库
        self._deactivate_all()
        license_obj = License(
            license_key=license_key,
            tier=tier,
            features=json.dumps(features),
            status="active",
            rsa_signature=signature_b64,
            fingerprint_hash=fingerprint,
            device_name=device_name,
            device_slot=device_slot,
            expires_at=expires_at,
            last_verified_at=datetime.now(timezone.utc),
        )
        self.db.add(license_obj)
        self.db.commit()

        # Step 5: 写入内存（通过 clawmemory_core）
        set_license(tier, features)

        return {
            "valid": True,
            "tier": tier,
            "features": features,
            "expires_at": expires_at,
            "device_slot": device_slot,
            "pro_download_url": settings.pro_download_url,
            "pro_fallback_urls": settings.pro_fallback_urls,
        }

    async def deactivate(self) -> dict:
        """取消激活 — 本地 + 通知授权平台"""
        lic = self.get_active_license()
        if lic:
            await self._notify_deactivate(lic)
            lic.status = "revoked"
            self.db.commit()

        reset()
        return {"active": False, "tier": "oss", "message": "License deactivated"}

    async def _notify_deactivate(self, lic: License):
        """通知授权平台解绑设备"""
        try:
            async with httpx.AsyncClient(timeout=10) as client:
                await client.post(
                    f"{settings.license_server_url}/api/v1/deactivate",
                    json={
                        "license_key": lic.license_key,
                        "fingerprint": lic.fingerprint_hash or self._get_fingerprint(),
                    },
                )
                logger.info(f"Notified license server: device deactivated for {lic.license_key[:8]}***")
        except Exception as e:
            logger.warning(f"Failed to notify license server about deactivation: {e}")

    def _deactivate_all(self):
        """将所有 active 的授权标记为 revoked"""
        self.db.query(License).filter(License.status == "active").update({"status": "revoked"})
        self.db.flush()

    def load_cached_license(self):
        """
        启动时从本地数据库恢复授权状态
        
        安全设计：使用 _has_clawmemory_core() 检查，而不是模块变量。
        """
        if not _has_clawmemory_core():
            logger.warning("Cached license ignored: clawmemory_core not installed, running as OSS")
            return

        lic = self.get_active_license()
        if lic:
            features = json.loads(lic.features) if lic.features else []
            tier = lic.tier if lic.tier in ("oss", "pro", "enterprise") else "pro"
            set_license(tier, features)
            logger.info(f"Cached license loaded: tier={tier}, features={len(features)}")
        else:
            logger.info("No cached license found, running as OSS")

    def _get_fingerprint(self) -> str:
        """生成设备指纹"""
        import hashlib, platform, os, uuid
        parts = []

        env_fp = os.environ.get('DEVICE_FINGERPRINT', '')
        if env_fp:
            parts.append(f"env:{env_fp}")

        for path in ['/etc/machine-id', '/var/lib/dbus/machine-id']:
            if os.path.isfile(path):
                try:
                    parts.append(open(path).read().strip()[:64])
                except Exception:
                    pass
                break

        try:
            mac = uuid.getnode()
            if mac != uuid.UUID(int=0).node:
                parts.append(f"mac:{mac}")
        except Exception:
            pass

        parts.append(f"{platform.system()}-{platform.machine()}")

        raw = '|'.join(parts)
        return hashlib.sha256(raw.encode()).hexdigest()[:32]

    def _get_device_name(self) -> str:
        """生成人类可读的设备名称"""
        import platform, os, socket

        name = os.environ.get('DEVICE_NAME', '')
        if name:
            return name

        hostname = socket.gethostname()
        if hostname and hostname not in ('localhost', '0.0.0.0'):
            if not hostname.startswith('DESKTOP'):
                return f"{hostname} ({platform.system()})"

        return f"{platform.system()} {platform.machine()}"

    def _load_public_key(self, force_refresh: bool = False) -> str | None:
        """加载 RSA 公钥"""
        path = settings.rsa_public_key_path
        if not force_refresh and path.exists():
            content = path.read_text().strip()
            if content.startswith("-----BEGIN PUBLIC KEY-----"):
                return content

        try:
            import httpx as _httpx
            resp = _httpx.get(f"{settings.license_server_url}/api/v1/public-key", timeout=10)
            if resp.status_code == 200:
                pubkey = resp.text.strip()
                if pubkey.startswith("-----BEGIN PUBLIC KEY-----"):
                    self._cache_public_key(path, pubkey)
                    logger.info("RSA public key fetched from license server and cached locally")
                    return pubkey
                try:
                    data = resp.json()
                    pk = data.get("public_key", "")
                    if pk and pk.startswith("-----BEGIN PUBLIC KEY-----"):
                        self._cache_public_key(path, pk)
                        logger.info("RSA public key fetched from license server (JSON) and cached locally")
                        return pk
                except Exception:
                    pass
        except Exception as e:
            logger.warning(f"Failed to fetch public key from server: {e}")

        return None

    def _cache_public_key(self, path, content: str):
        """缓存公钥到本地文件"""
        try:
            path.parent.mkdir(parents=True, exist_ok=True)
            path.write_text(content)
        except Exception as e:
            logger.warning(f"Failed to cache public key: {e}")