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

# ========== 引擎状态 ==========
_LICENSE_ENGINE = "none"
_COMPUTE_ENGINE = "none"

# Python 层授权状态
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


# 初始化引擎
_core_funcs = _get_core_functions()
if _core_funcs:
    _build = _core_funcs["get_build_info"]()
    if "c-cpython" in _build:
        _LICENSE_ENGINE = "c"
        logger.info("License engine: C/CPython — high security")
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
        logger.warning("clawmemory_core missing compute functions")

# 加载 .pyx 计算引擎
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
        logger.info("Compute engine: pyx")
    except ImportError:
        logger.info("No pyx modules, using pure Python")

# 纯 Python 兜底
if _COMPUTE_ENGINE == "none":
    from app.services.memory_decay_py import (
        calculate_decay, should_prune, reinforce, decay_memory, decay_batch, get_decay_stats,
        get_decay_stage, should_archive, should_trash, should_prune_from_trash, get_stage_info,
    )
    from app.services.conflict_resolver_py import (
        detect_conflict, resolve_conflict, scan_for_conflicts, get_conflict_summary,
    )
    from app.services.token_router_py import (
        estimate_complexity, route_model, get_routing_stats,
    )
    _COMPUTE_ENGINE = "python"
    logger.info("Compute engine: pure Python")


# ========== 授权管理 ==========

def _set_python_license(license_data: dict) -> None:
    global _python_tier, _python_features
    _python_tier = license_data.get("tier", "oss")
    _python_features = license_data.get("features", [])
    logger.info("Python license set: tier=%s, features=%s", _python_tier, _python_features)


def _reset_python_license() -> None:
    global _python_tier, _python_features
    _python_tier = "oss"
    _python_features = []
    logger.info("Python license reset")


def get_tier() -> str:
    if _core_funcs:
        return _core_funcs["get_tier"]()
    return _python_tier


def check_feature(feature: str) -> bool:
    if _core_funcs:
        return _core_funcs["check_feature"](feature)
    return feature in _python_features


def reset() -> None:
    if _core_funcs:
        _core_funcs["reset"]()
    _reset_python_license()


def get_license_engine() -> str:
    return _LICENSE_ENGINE


def get_compute_engine() -> str:
    return _COMPUTE_ENGINE


def get_license_status() -> dict:
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

        # 1. 调用授权服务器
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

        # 2. RSA 签名验证（如果服务器返回了签名）
        signature_b64 = data.get("signature", "")
        if signature_b64:
            pubkey = self._load_public_key()
            if pubkey:
                verify_result = self._verify_signature(signature_b64, pubkey)
                if verify_result is False:
                    # 验证失败，尝试刷新公钥
                    logger.warning("RSA 验证失败，尝试刷新公钥...")
                    pubkey = self._load_public_key(force_refresh=True)
                    if pubkey:
                        verify_result = self._verify_signature(signature_b64, pubkey)
                
                if verify_result is False:
                    logger.error("RSA 签名验证失败")
                    return {"valid": False, "message": "RSA 签名验证失败，授权可能被篡改"}
                elif verify_result is True:
                    logger.info("RSA 签名验证通过")
                else:
                    logger.info("跳过 RSA 验证（缺少验证库）")
            else:
                logger.warning("无法获取公钥，跳过 RSA 验证")

        # 3. 保存授权信息到数据库
        try:
            # 停用旧的授权
            self.db.query(License).filter(License.status == "active").update({
                "status": "inactive"
            })

            # 创建新的授权记录
            license_data = {
                "license_key": license_key,
                "tier": data.get("tier", "pro"),
                "status": "active",
                "device_fingerprint": fingerprint,
                "device_name": device_name,
                "expires_at": data.get("expires_at"),
                "device_slot": data.get("device_slot", ""),
                "features": json.dumps(data.get("features", [])),
                "pro_download_url": data.get("pro_download_url", ""),
                "pro_fallback_urls": json.dumps(data.get("pro_fallback_urls", [])),
            }

            lic = License(**license_data)
            self.db.add(lic)
            self.db.commit()

            # 4. 设置内存中的授权状态
            if _core_funcs:
                _core_funcs["set_license"](license_data)
            _set_python_license(license_data)

            return {
                "valid": True,
                "message": "激活成功",
                "tier": data.get("tier", "pro"),
                "expires_at": data.get("expires_at"),
                "features": data.get("features", []),
                "pro_download_url": data.get("pro_download_url", ""),
                "pro_fallback_urls": data.get("pro_fallback_urls", []),
            }

        except Exception as e:
            self.db.rollback()
            logger.error(f"保存授权信息失败: {e}")
            return {"valid": False, "message": f"保存授权信息失败: {e}"}

    async def deactivate(self) -> dict:
        lic = self.get_active_license()
        if not lic:
            return {"success": False, "message": "没有激活的授权"}

        try:
            async with httpx.AsyncClient(timeout=15) as client:
                await client.post(
                    f"{settings.license_server_url}/api/v1/deactivate",
                    json={
                        "license_key": lic.license_key,
                        "fingerprint": self._get_fingerprint(),
                    },
                )
        except Exception as e:
            logger.warning(f"通知服务器停用失败: {e}")

        try:
            lic.status = "inactive"
            self.db.commit()
            reset()
            return {"success": True, "message": "已停用授权"}
        except Exception as e:
            self.db.rollback()
            return {"success": False, "message": f"停用失败: {e}"}

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
        """加载公钥，支持缓存和强制刷新"""
        path = settings.rsa_public_key_path
        
        # 先尝试从缓存加载
        if not force_refresh and path.exists():
            try:
                content = path.read_text().strip()
                if "BEGIN PUBLIC KEY" in content:
                    return content
            except Exception:
                pass
        
        # 从服务器获取最新公钥
        try:
            pubkey_url = f"{settings.license_server_url}/api/v1/public-key"
            resp = httpx.get(pubkey_url, timeout=10)
            if resp.status_code == 200:
                data = resp.json()
                pem = data.get("public_key", "").strip()
                if pem and "BEGIN PUBLIC KEY" in pem:
                    path.parent.mkdir(parents=True, exist_ok=True)
                    path.write_text(pem)
                    logger.info("已从服务器获取最新公钥")
                    return pem
        except Exception as e:
            logger.warning(f"从服务器获取公钥失败: {e}")

        return None

    def _verify_signature(self, signature_b64: str, public_key_pem: str) -> bool | None:
        """
        验证 RSA 签名
        
        返回:
            True: 验证成功
            False: 验证失败（签名无效）
            None: 无法验证（缺少库）
        """
        # 1. 尝试使用 clawmemory_core
        if _core_funcs:
            try:
                result = _core_funcs["verify_license"](signature_b64, public_key_pem)
                if result:
                    return True
                return False
            except Exception as e:
                logger.warning(f"clawmemory_core 验证失败: {e}")

        # 2. 尝试使用 Python cryptography
        try:
            from cryptography.hazmat.primitives import hashes, serialization
            from cryptography.hazmat.primitives.asymmetric import padding

            public_key = serialization.load_pem_public_key(public_key_pem.encode())
            payload_b64, sig_b64 = signature_b64.rsplit(".", 1)
            payload = base64.urlsafe_b64decode(payload_b64 + "==")
            signature = base64.urlsafe_b64decode(sig_b64 + "==")

            public_key.verify(signature, payload, padding.PKCS1v15(), hashes.SHA256())
            return True
        except ImportError:
            logger.warning("缺少 cryptography 库，跳过 RSA 验证")
            return None
        except Exception as e:
            logger.warning(f"Python RSA 验证失败: {e}")
            return False
