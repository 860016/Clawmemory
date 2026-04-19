import base64
import json
import logging
from datetime import datetime, timezone
from sqlalchemy.orm import Session
from app.models.license import License
from app.config import settings, APP_VERSION
import httpx

logger = logging.getLogger(__name__)

# ========== 加载核心模块：C/CPython (预编译 wheel) → 纯 Python 兜底 ==========
_CORE_ENGINE = "none"

# 1. 尝试 C/CPython — 最高安全
try:
    from clawmemory_core import (
        check_feature, set_license, get_tier, reset as _reset,
        verify_integrity as _verify_integrity,
        verify_license as _core_verify_license,
        get_build_info,
        # Pro: memory decay
        calculate_decay, should_prune, reinforce, decay_memory, decay_batch, get_decay_stats,
        # Pro: conflict resolver
        detect_conflict, resolve_conflict, scan_for_conflicts, get_conflict_summary,
        # Pro: token router
        estimate_complexity, route_model, get_routing_stats,
    )
    _build = get_build_info()
    if "c-cpython" in _build:
        _CORE_ENGINE = "c"
        USING_CORE = True
        if "cng" in _build:
            logger.info("Core engine: C/CPython + Windows CNG — high security, RSA verification")
        else:
            logger.info("Core engine: C/CPython + OpenSSL — high security, RSA verification")
    else:
        _CORE_ENGINE = "rust"
        USING_CORE = True
        logger.info("Core engine: Rust (PyO3) — maximum security")
except ImportError:
    USING_CORE = False

    # 2. 纯 Python 兜底
    import hashlib, os as _os, math, time as _time, re

    _features: set = set()
    _tier: str = "oss"

    def check_feature(f: str) -> bool:
        return f in _features

    def set_license(t: str, fs: list):
        global _tier, _features
        _tier = t
        _features = set(fs)

    def get_tier() -> str:
        return _tier

    def _reset():
        global _tier, _features
        _tier = "oss"
        _features = set()

    def _verify_integrity() -> bool:
        """纯 Python 模式不做完整性校验（无安全意义），始终返回 True"""
        return True

    def _core_verify_license(license_data_b64: str, public_key_pem: str) -> bool:
        """纯 Python fallback: 用 cryptography 库做 RSA-SHA256 签名验证。"""
        try:
            from cryptography.hazmat.primitives import hashes, serialization
            from cryptography.hazmat.primitives.asymmetric import padding
            from cryptography.exceptions import InvalidSignature

            raw = base64.b64decode(license_data_b64)
            payload = json.loads(raw)
            data_str = payload.get("data", "")
            signature_b64 = payload.get("signature", "")
            if not data_str or not signature_b64:
                logger.warning("RSA verify: missing data or signature field")
                return False

            signature = base64.b64decode(signature_b64)
            public_key = serialization.load_pem_public_key(public_key_pem.encode())
            public_key.verify(
                signature,
                data_str.encode(),
                padding.PKCS1v15(),
                hashes.SHA256(),
            )
            return True
        except InvalidSignature:
            logger.warning("RSA verify: signature verification FAILED")
            return False
        except Exception as e:
            logger.error(f"RSA verify error: {e}")
            return False

    # --- Memory Decay (Python fallback) ---
    _HALF_LIFE_DEFAULT = 30 * 24 * 3600
    _MIN_IMPORTANCE = 0.05

    def calculate_decay(importance: float, age_seconds: float, half_life: float = _HALF_LIFE_DEFAULT) -> float:
        if half_life <= 0:
            return importance
        return max(importance * (2.0 ** (-age_seconds / half_life)), 0.0)

    def should_prune(importance: float) -> bool:
        return importance < _MIN_IMPORTANCE

    def reinforce(importance: float, factor: float = 1.5) -> float:
        reinforced = importance + (1.0 - importance) * (1.0 - math.exp(-0.5 * factor * importance))
        return min(reinforced, 1.0)

    def decay_memory(memory_id: int, current_importance: float, last_accessed_at: float, now: float = 0.0, half_life: float = _HALF_LIFE_DEFAULT) -> dict:
        if now == 0.0:
            now = _time.time()
        age_seconds = max(now - last_accessed_at, 0.0)
        new_importance = calculate_decay(current_importance, age_seconds, half_life)
        return {"memory_id": memory_id, "new_importance": round(new_importance, 6), "should_prune": new_importance < _MIN_IMPORTANCE, "age_seconds": age_seconds}

    def decay_batch(memories: list, now: float = 0.0, half_life: float = _HALF_LIFE_DEFAULT) -> list:
        return [decay_memory(m["id"], m["importance"], m["last_accessed_at"], now, half_life) for m in memories]

    def get_decay_stats(memories: list, now: float = 0.0) -> dict:
        if not memories:
            return {"total": 0, "prune_candidates": 0, "avg_importance": 0.0, "decayed_count": 0}
        if now == 0.0:
            now = _time.time()
        total = len(memories)
        prune_candidates = sum(1 for m in memories if calculate_decay(m.get("importance", 0.5), max(now - m.get("last_accessed_at", 0), 0)) < _MIN_IMPORTANCE)
        decayed_count = sum(1 for m in memories if calculate_decay(m.get("importance", 0.5), max(now - m.get("last_accessed_at", 0), 0)) < m.get("importance", 0.5) - 0.01)
        importance_sum = sum(calculate_decay(m.get("importance", 0.5), max(now - m.get("last_accessed_at", 0), 0)) for m in memories)
        return {"total": total, "prune_candidates": prune_candidates, "avg_importance": round(importance_sum / total, 4), "decayed_count": decayed_count}

    # --- Conflict Resolver (Python fallback) ---
    _SAFETY_KEYS = ["password", "secret", "api_key", "token", "permission", "access"]

    def detect_conflict(memory_a: dict, memory_b: dict):
        key_a, key_b = memory_a.get("key", ""), memory_b.get("key", "")
        if key_a != key_b and not (key_a.lower().replace('_', ' ').replace('-', ' ') in key_b.lower().replace('_', ' ').replace('-', ' ') or key_b.lower().replace('_', ' ').replace('-', ' ') in key_a.lower().replace('_', ' ').replace('-', ' ')):
            return None
        val_a, val_b = str(memory_a.get("value", "")), str(memory_b.get("value", ""))
        if val_a.strip().lower() == val_b.strip().lower():
            return None
        severity = "high" if any(sk in key_a.lower() for sk in _SAFETY_KEYS) else "medium" if re.search(r'\d{4}-\d{2}-\d{2}', val_a) and re.search(r'\d{4}-\d{2}-\d{2}', val_b) else "low"
        return {"memory_a_id": memory_a.get("id"), "memory_b_id": memory_b.get("id"), "key": key_a, "value_a": val_a, "value_b": val_b, "severity": severity, "layer_a": memory_a.get("layer", ""), "layer_b": memory_b.get("layer", ""), "source_a": memory_a.get("source", ""), "source_b": memory_b.get("source", "")}

    def resolve_conflict(conflict: dict, strategy: str = None) -> dict:
        if strategy is None:
            severity = conflict.get("severity", "low")
            strategy = "flag_for_review" if severity == "high" else "merge" if severity == "low" else "flag_for_review"
        return {"conflict": conflict, "strategy": strategy, "winner": "both" if strategy == "merge" else None, "action": "merge" if strategy == "merge" else strategy}

    def scan_for_conflicts(memories: list) -> list:
        by_key = {}
        for m in memories:
            k = m.get("key", "")
            if k:
                by_key.setdefault(k, []).append(m)
        conflicts = []
        for group in by_key.values():
            for i in range(len(group)):
                for j in range(i + 1, len(group)):
                    c = detect_conflict(group[i], group[j])
                    if c:
                        conflicts.append(c)
        return conflicts

    def get_conflict_summary(conflicts: list) -> dict:
        by_sev = {"low": 0, "medium": 0, "high": 0}
        for c in conflicts:
            by_sev[c.get("severity", "low")] += 1
        return {"total": len(conflicts), "by_severity": by_sev, "auto_resolvable": sum(1 for c in conflicts if c.get("severity") == "low"), "needs_review": sum(1 for c in conflicts if c.get("severity") != "low")}

    # --- Token Router (Python fallback) ---
    def estimate_complexity(message: str, context_length: int = 0, has_code: bool = False, has_reasoning: bool = False) -> float:
        score = min(math.log1p(len(message)) / math.log1p(4000), 1.0) * 0.25
        score += min(context_length / 50.0, 1.0) * 0.2
        if has_code or any(ind in message for ind in ["```", "def ", "class ", "import "]):
            score += 0.25
        if has_reasoning or any(kw in message.lower() for kw in ["分析", "解释", "why", "how", "compare", "evaluate"]):
            score += 0.3
        return min(score, 1.0)

    def route_model(message: str, available_models: list, context_length: int = 0, user_tier: str = "oss", budget_remaining: float = None) -> dict:
        complexity = estimate_complexity(message, context_length)
        return {"selected_model": available_models[0] if available_models else None, "complexity": round(complexity, 4), "routing_reason": "complexity_based", "estimated_cost": 0.0, "estimated_tokens": 0}

    def get_routing_stats(routing_history: list) -> dict:
        return {"total": len(routing_history), "distribution": {}, "avg_complexity": 0.0, "total_cost": 0.0}

    _CORE_ENGINE = "python"
    logger.warning("Core engine: Pure Python — LOW security, install clawmemory-core wheel for C/CPython protection")


def verify_integrity() -> bool:
    """校验授权数据是否被篡改"""
    if USING_CORE:
        try:
            return _verify_integrity()
        except Exception:
            return False
    # 纯 Python 模式不做完整性校验
    return True


def get_core_engine() -> str:
    return _CORE_ENGINE


def reset():
    _reset()


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


class LicenseService:
    def __init__(self, db: Session):
        self.db = db

    def get_active_license(self) -> License | None:
        """获取当前活跃的授权记录"""
        return self.db.query(License).filter(License.status == "active").first()

    async def activate(self, license_key: str) -> dict:
        """
        激活授权码 — 完整流程：
        1. 向授权平台发送激活请求
        2. RSA 签名验证
        3. 存入本地数据库
        4. 写入内存
        """
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
                # 公钥获取失败 — 尝试从签名数据中恢复
                logger.error("RSA public key not available, cannot verify license signature")
                return {"valid": False, "message": "无法获取 RSA 公钥，请检查网络连接后重试"}

            if not _core_verify_license(signature_b64, pubkey):
                logger.warning("RSA signature verification FAILED — license may be forged")
                return {"valid": False, "message": "RSA 签名验证失败，授权可能被篡改"}
            logger.info("RSA signature verification passed")
        else:
            # 无签名数据 — 仅在纯 Python 模式下允许（开发/测试）
            if USING_CORE:
                logger.error("No signature in activation response but core engine is active")
                return {"valid": False, "message": "授权服务器未返回签名数据"}
            logger.warning("No signature in response (Python fallback mode, skipping RSA verify)")

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

        # Step 4: 存入本地数据库 — 先清理旧授权，再写入新授权
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

        # Step 5: 写入内存
        set_license(tier, features)

        return {
            "valid": True,
            "tier": tier,
            "features": features,
            "expires_at": expires_at,
            "device_slot": device_slot,
        }

    async def deactivate(self) -> dict:
        """
        取消激活 — 本地 + 通知授权平台
        """
        lic = self.get_active_license()
        if lic:
            # 通知授权平台解绑设备
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
        """启动时从本地数据库恢复授权状态"""
        lic = self.get_active_license()
        if lic:
            features = json.loads(lic.features) if lic.features else []
            tier = lic.tier if lic.tier in ("oss", "pro", "enterprise") else "pro"
            set_license(tier, features)
            logger.info(f"Cached license loaded: tier={tier}, features={len(features)}")
        else:
            logger.info("No cached license found, running as OSS")

    def _get_fingerprint(self) -> str:
        """生成设备指纹 — 稳定的设备唯一标识"""
        import hashlib, platform, os, uuid
        parts = []

        # 1. 环境变量覆盖（Docker 部署推荐设置）
        env_fp = os.environ.get('DEVICE_FINGERPRINT', '')
        if env_fp:
            parts.append(f"env:{env_fp}")

        # 2. Linux machine-id
        for path in ['/etc/machine-id', '/var/lib/dbus/machine-id']:
            if os.path.isfile(path):
                try:
                    parts.append(open(path).read().strip()[:64])
                except Exception:
                    pass
                break

        # 3. MAC 地址
        try:
            mac = uuid.getnode()
            if mac != uuid.UUID(int=0).node:
                parts.append(f"mac:{mac}")
        except Exception:
            pass

        # 4. 系统+架构
        parts.append(f"{platform.system()}-{platform.machine()}")

        raw = '|'.join(parts)
        return hashlib.sha256(raw.encode()).hexdigest()[:32]

    def _get_device_name(self) -> str:
        """生成人类可读的设备名称"""
        import platform, os, socket

        # 1. 环境变量覆盖
        name = os.environ.get('DEVICE_NAME', '')
        if name:
            return name

        # 2. hostname（Docker 容器名、服务器名）
        hostname = socket.gethostname()
        if hostname and hostname not in ('localhost', '0.0.0.0'):
            # Docker 容器名通常是随机 hex，如 7275d4494039
            if not hostname.startswith('DESKTOP'):
                return f"{hostname} ({platform.system()})"

        # 3. 回退: OS + 架构
        return f"{platform.system()} {platform.machine()}"

    def _load_public_key(self) -> str | None:
        """
        加载 RSA 公钥 — 多策略降级：
        1. 本地文件
        2. 从授权服务器获取并缓存
        """
        path = settings.rsa_public_key_path
        if path.exists():
            content = path.read_text().strip()
            if content.startswith("-----BEGIN PUBLIC KEY-----"):
                return content

        # 从授权服务器获取
        try:
            import httpx as _httpx
            resp = _httpx.get(f"{settings.license_server_url}/api/v1/public-key", timeout=10)
            if resp.status_code == 200:
                pubkey = resp.text.strip()
                if pubkey.startswith("-----BEGIN PUBLIC KEY-----"):
                    self._cache_public_key(path, pubkey)
                    logger.info("RSA public key fetched from license server and cached locally")
                    return pubkey
                # 尝试 JSON 格式
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


def is_feature_enabled(feature: str) -> bool:
    return check_feature(feature)


def current_tier() -> str:
    return get_tier()
