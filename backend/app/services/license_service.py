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

# ========== v3.0 授权架构 ==========
#
# 安全设计原则：
#   1. 授权状态存储在本地数据库 + 内存中
#   2. clawmemory_core (C/Rust wheel) 是可选增强，提供 RSA 硬验证
#   3. 没有 clawmemory_core 时，授权平台验证 + 本地数据库即可激活
#   4. Pro 功能模块在激活后自动下载安装
#
# 引擎层级：
#   - clawmemory_core → RSA 硬验证 + 高性能计算
#   - .pyx 编译版 → 中等性能计算
#   - 纯 Python 兜底 → 基础功能

_LICENSE_ENGINE = "none"
_COMPUTE_ENGINE = "none"

# ========== Python 层授权状态（不依赖 clawmemory_core） ==========

_python_tier = "oss"
_python_features: list[str] = []


def _has_clawmemory_core() -> bool:
    return importlib.util.find_spec("clawmemory_core") is not None


def _get_core_functions():
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
    logger.info("clawmemory_core not installed, using Python license management")

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
        logger.info("Compute engine: .pyx compiled — compute functions available")
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


# ========== 授权功能：优先使用 clawmemory_core，回退到 Python ==========

def check_feature(feature: str) -> bool:
    funcs = _get_core_functions()
    if funcs:
        return funcs["check_feature"](feature)
    return feature in _python_features


def get_tier() -> str:
    funcs = _get_core_functions()
    if funcs:
        return funcs["get_tier"]()
    return _python_tier


def set_license(tier: str, features: list):
    global _python_tier, _python_features
    funcs = _get_core_functions()
    if funcs:
        return funcs["set_license"](tier, features)
    _python_tier = tier
    _python_features = list(features)
    logger.info("License set (Python): tier=%s, features=%d", tier, len(features))


def reset():
    global _python_tier, _python_features
    funcs = _get_core_functions()
    if funcs:
        return funcs["reset"]()
    _python_tier = "oss"
    _python_features = []


def verify_integrity() -> bool:
    funcs = _get_core_functions()
    if funcs:
        try:
            return funcs["verify_integrity"]()
        except Exception:
            return False
    return True


def _core_verify_license(license_data_b64: str, public_key_pem: str) -> dict:
    funcs = _get_core_functions()
    if funcs:
        return funcs["verify_license"](license_data_b64, public_key_pem)
    return {}


def _python_verify_license(license_data_b64: str, public_key_pem: str) -> dict:
    try:
        from cryptography.hazmat.primitives import hashes, serialization
        from cryptography.hazmat.primitives.asymmetric import padding

        public_key = serialization.load_pem_public_key(public_key_pem.encode())
        payload_b64, signature_b64 = license_data_b64.rsplit(".", 1)
        payload = base64.urlsafe_b64decode(payload_b64 + "==")
        signature = base64.urlsafe_b64decode(signature_b64 + "==")

        public_key.verify(signature, payload, padding.PKCS1v15(), hashes.SHA256())
        return json.loads(payload)
    except ImportError:
        logger.warning("cryptography library not available, skipping RSA verify")
        return {}
    except Exception as e:
        logger.warning("Python RSA verify failed: %s", e)
        return {}


# ========== 辅助函数 ==========

def get_license_engine() -> str:
    if _has_clawmemory_core():
        funcs = _get_core_functions()
        if funcs:
            _build = funcs["get_build_info"]()
            if "c-cpython" in _build:
                return "c"
            return "rust"
    return "python"


def get_compute_engine() -> str:
    return _COMPUTE_ENGINE


def get_engine_info() -> dict:
    has_core = _has_clawmemory_core()
    return {
        "license_engine": get_license_engine(),
        "compute_engine": _COMPUTE_ENGINE,
        "has_clawmemory_core": has_core,
        "can_activate": True,
    }


def is_feature_enabled(feature: str) -> bool:
    return check_feature(feature)


def current_tier() -> str:
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
        return self.db.query(License).filter(License.status == "active").first()

    async def activate(self, license_key: str) -> dict:
        fingerprint = self._get_fingerprint()
        device_name = self._get_device_name()

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

        if not data.get("valid"):
            return {"valid": False, "message": data.get("message", "授权码无效")}

        # RSA 签名验证（可选增强）
        signature_b64 = data.get("signature", "")
        if signature_b64:
            pubkey = self._load_public_key()
            if pubkey:
                verified_data = _core_verify_license(signature_b64, pubkey)
                if not verified_data:
                    verified_data = _python_verify_license(signature_b64, pubkey)
                if verified_data:
                    logger.info("RSA signature verification passed")
                else:
                    # 验证失败，尝试从服务器获取最新公钥
                    fresh_pubkey = self._load_public_key(force_refresh=True)
                    if fresh_pubkey:
                        verified_data = _core_verify_license(signature_b64, fresh_pubkey)
                        if not verified_data:
                            verified_data = _python_verify_license(signature_b64, fresh_pubkey)
                        if verified_data:
                            logger.info("RSA signature verification passed with fresh key")
                    if not verified_data:
                        logger.warning("RSA signature verification FAILED")
                        return {"valid": False, "message": "RSA 签名验证失败，授权可能被篡改"}
            else:
                logger.warning("No public key available, skipping RSA verify")
        else:
            logger.info("No signature in response, server-side verification only")

        tier = data.get("tier", "pro")
        features = data.get("features", [])
        expires_at = data.get("expires_at")
        device_slot = data.get("device_slot", "")
        pro_download_url = data.get("pro_download_url", "")
        pro_fallback_urls = data.get("pro_fallback_urls", [])

        if not features:
            if tier == "oss":
                features = []
            elif tier == "enterprise":
                features = list(PRO_FEATURES.keys()) + list(ENTERPRISE_EXTRA_FEATURES.keys())
            else:
                features = list(PRO_FEATURES.keys())

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
            pro_download_url=pro_download_url,
            pro_fallback_urls=json.dumps(pro_fallback_urls),
        )
        self.db.add(license_obj)
        self.db.commit()

        set_license(tier, features)

        return {
            "valid": True,
            "tier": tier,
            "features": features,
            "expires_at": expires_at,
            "device_slot": device_slot,
            "pro_download_url": data.get("pro_download_url", ""),
            "pro_fallback_urls": data.get("pro_fallback_urls", []),
        }

    async def deactivate(self) -> dict:
        lic = self.get_active_license()
        if lic:
            await self._notify_deactivate(lic)
            lic.status = "revoked"
            self.db.commit()

        reset()
        return {"active": False, "tier": "oss", "message": "License deactivated"}

    async def _notify_deactivate(self, lic: License):
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
        self.db.query(License).filter(License.status == "active").update({"status": "revoked"})
        self.db.flush()

    def load_cached_license(self):
        lic = self.get_active_license()
        if lic:
            features = json.loads(lic.features) if lic.features else []
            tier = lic.tier if lic.tier in ("oss", "pro", "enterprise") else "pro"
            set_license(tier, features)
            logger.info(f"Cached license loaded: tier={tier}, features={len(features)}")
        else:
            logger.info("No cached license found, running as OSS")

    def _get_fingerprint(self) -> str:
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
        path = settings.rsa_public_key_path
        if not force_refresh and path.exists():
            try:
                return path.read_text()
            except Exception:
                pass

        try:
            pubkey_url = f"{settings.license_server_url}/api/v1/public-key"
            resp = httpx.get(pubkey_url, timeout=10)
            if resp.status_code == 200:
                data = resp.json()
                pem = data.get("public_key", "")
                if pem and "BEGIN PUBLIC KEY" in pem:
                    path.parent.mkdir(parents=True, exist_ok=True)
                    path.write_text(pem)
                    return pem
        except Exception as e:
            logger.warning(f"Failed to fetch public key from server: {e}")

        return None
