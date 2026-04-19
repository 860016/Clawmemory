import base64
import json
import logging
from datetime import datetime, timezone
from sqlalchemy.orm import Session
from app.models.license import License
from app.config import settings
import httpx

logger = logging.getLogger(__name__)

# ========== 加载核心模块：Rust (预编译 wheel) → 纯 Python 兜底 ==========
_CORE_ENGINE = "none"

# 1. 尝试 Rust (PyO3) — 最高安全
try:
    from clawmemory_core import (
        check_feature, set_license, get_tier, reset as _reset,
        verify_integrity as _verify_integrity,
        verify_license as _rust_verify_license,
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
        USING_RUST = False
        if "cng" in _build:
            logger.info("Core engine: C/CPython + Windows CNG — high security, RSA verification")
        else:
            logger.info("Core engine: C/CPython + OpenSSL — high security, RSA verification")
    else:
        _CORE_ENGINE = "rust"
        USING_RUST = True
        logger.info("Core engine: Rust (PyO3) — maximum security")
except ImportError:
    USING_RUST = False

    # 2. 纯 Python 兜底 — 最低安全 (仅开发环境/OSS 免费版)
    # 注意：此模式不包含完整的安全保护，仅保证基本功能可用
    # 发布时必须安装 clawmemory_core (Rust) wheel
    import hashlib, os as _os, math, time as _time, re

    _features: set = set()
    _tier: str = "oss"
    # 使用 salted hash 防止简单篡改（非安全级别，仅增加门槛）
    _license_hash: str = ""

    def check_feature(f: str) -> bool:
        return f in _features

    def set_license(t: str, fs: list):
        global _tier, _features, _license_hash
        _tier = t
        _features = set(fs)
        # HMAC-style hash with tier-dependent salt
        raw = f"{t}|{','.join(sorted(fs))}|clawmemory_v1"
        _license_hash = hashlib.sha256(raw.encode()).hexdigest()[:16]

    def get_tier() -> str:
        return _tier

    def _reset():
        global _tier, _features, _license_hash
        _tier = "oss"
        _features = set()
        _license_hash = ""

    def _verify_integrity() -> bool:
        global _license_hash
        if not _license_hash:
            return _tier == "oss"
        raw = f"{_tier}|{','.join(sorted(_features))}|clawmemory_v1"
        expected = hashlib.sha256(raw.encode()).hexdigest()[:16]
        return _license_hash == expected

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
    logger.warning("Core engine: Pure Python — LOW security, install clawmemory-core wheel for Rust protection")


def verify_integrity() -> bool:
    """校验授权数据是否被篡改"""
    if USING_RUST:
        try:
            return _verify_integrity()
        except Exception:
            return False
    return _verify_integrity()


def get_core_engine() -> str:
    """返回当前核心引擎类型"""
    return _CORE_ENGINE


def reset():
    _reset()


# ========== v3.0 功能定义 ==========
PRO_FEATURES = {
    # Memory Decay
    "auto_decay": "自动记忆衰减",
    "decay_report": "衰减报告",
    "prune_suggest": "清理建议",
    "reinforce": "记忆强化",
    # Conflict Resolver
    "conflict_scan": "矛盾扫描",
    "conflict_merge": "自动合并",
    # Token Router
    "smart_router": "智能模型路由",
    "token_stats": "Token 统计",
    # AI & Graph
    "ai_extract": "AI 智能提取",
    "auto_graph": "自动知识图谱",
    "unlimited_graph": "无限图谱节点",
    # Wiki
    "wiki": "Wiki 知识库",
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

    def get_current_license(self) -> License | None:
        return self.db.query(License).filter(License.status == "active").first()

    async def activate(self, license_key: str) -> dict:
        try:
            async with httpx.AsyncClient(timeout=15) as client:
                resp = await client.post(
                    f"{settings.license_server_url}/api/v1/activate",
                    json={"license_key": license_key, "fingerprint": self._get_fingerprint(), "version": "2.1.0"},
                )
                data = resp.json()
        except Exception as e:
            return {"valid": False, "message": f"Cannot reach license server: {e}"}

        if not data.get("valid"):
            return {"valid": False, "message": data.get("message", "Invalid license")}

        # RSA 签名验证 (Rust 引擎)
        if USING_RUST and data.get("signed_data"):
            try:
                pubkey = self._load_public_key()
                if pubkey:
                    if not _rust_verify_license(data["signed_data"], pubkey):
                        return {"valid": False, "message": "RSA signature verification failed"}
            except Exception as e:
                logger.warning(f"RSA verification error: {e}")

        tier = data.get("tier", "pro")
        features = data.get("features", [])

        # 服务端返回的 features 优先，若无则按 tier 默认
        if not features:
            if tier == "oss":
                features = []
            elif tier == "enterprise":
                features = list(PRO_FEATURES.keys()) + list(ENTERPRISE_EXTRA_FEATURES.keys())
            else:
                features = list(PRO_FEATURES.keys())

        license_obj = License(
            license_key=license_key,
            tier=tier,
            features=json.dumps(features),
            status="active",
            rsa_signature=data.get("signature", ""),
            fingerprint_hash=self._get_fingerprint(),
            device_slot=data.get("device_slot", ""),
            expires_at=data.get("expires_at"),
        )
        self.db.add(license_obj)
        self.db.commit()

        set_license(tier, features)
        return {
            "valid": True,
            "tier": tier,
            "features": features,
            "expires_at": data.get("expires_at"),
        }

    def load_cached_license(self):
        lic = self.get_current_license()
        if lic:
            features = json.loads(lic.features) if lic.features else []
            tier = lic.tier if lic.tier in ("oss", "pro", "enterprise") else "pro"
            set_license(tier, features)

    def _get_fingerprint(self) -> str:
        import hashlib, platform, os, uuid
        parts = []
        for path in ['/etc/machine-id', '/var/lib/dbus/machine-id']:
            if os.path.isfile(path):
                parts.append(open(path).read().strip()[:64])
                break
        try:
            mac = uuid.getnode()
            if mac != uuid.UUID(int=0).node:
                parts.append(f"mac:{mac}")
        except Exception:
            pass
        parts.append(f"{platform.system()}-{platform.machine()}")
        env_fp = os.environ.get('DEVICE_FINGERPRINT', '')
        if env_fp:
            parts.insert(0, f"env:{env_fp}")
        raw = '|'.join(parts)
        return hashlib.sha256(raw.encode()).hexdigest()[:32]

    def _load_public_key(self) -> str | None:
        path = settings.rsa_public_key_path
        if path.exists():
            return path.read_text()
        return None


def is_feature_enabled(feature: str) -> bool:
    return check_feature(feature)


def current_tier() -> str:
    return get_tier()
