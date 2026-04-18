from fastapi import APIRouter, Depends
from sqlalchemy.orm import Session
from app.database import get_db
from app.middleware.auth import get_current_user
<<<<<<< HEAD
from app.models.user import User
from app.models.license import License
from app.schemas.license import LicenseActivateRequest, LicenseStatusResponse
=======
from app.models.license import License
from app.schemas.license import LicenseActivateRequest
>>>>>>> fb055c7 (feat: v3.0 - Wiki知识库 + 科技感UI + i18n + Rust PyO3核心 + Pro功能)
from app.services.license_service import LicenseService, current_tier, is_feature_enabled, reset

router = APIRouter(prefix="/api/v1/license", tags=["license"])


<<<<<<< HEAD
def _build_license_info(current_user: User, db: Session) -> dict:
    """构建前端期望的授权信息格式"""
    tier = current_tier()
    active = tier != "oss"
    features = [f for f in ["graph", "backup", "decay", "routing", "conflict", "api", "sso", "audit", "timetravel", "offline"] if is_feature_enabled(f)]

    # 从 DB 查询详细信息
=======
def _build_license_info(db: Session) -> dict:
    """构建前端期望的授权信息格式"""
    tier = current_tier()
    active = tier != "oss"
    features = [f for f in [
        "ai_extract", "auto_graph", "unlimited_graph",
        "auto_decay", "decay_report", "prune_suggest", "reinforce",
        "conflict_scan", "conflict_merge",
        "smart_router", "token_stats",
    ] if is_feature_enabled(f)]

>>>>>>> fb055c7 (feat: v3.0 - Wiki知识库 + 科技感UI + i18n + Rust PyO3核心 + Pro功能)
    lic = db.query(License).filter(License.status == "active").first()
    expires_at = None
    device_slot = ""
    max_devices = 1
    device_count = 0
<<<<<<< HEAD
    if lic:
        expires_at = str(lic.expires_at) if lic.expires_at else None
        device_slot = lic.device_slot or ""
        # 解析 device_slot "2/3" 格式
=======
    license_type = "none"
    if lic:
        expires_at = str(lic.expires_at) if lic.expires_at else None
        device_slot = lic.device_slot or ""
        license_type = lic.tier  # pro_annual / pro_lifetime
>>>>>>> fb055c7 (feat: v3.0 - Wiki知识库 + 科技感UI + i18n + Rust PyO3核心 + Pro功能)
        if device_slot and "/" in device_slot:
            parts = device_slot.split("/")
            device_count = int(parts[0])
            max_devices = int(parts[1])

    return {
        "active": active,
        "tier": tier,
<<<<<<< HEAD
=======
        "type": license_type,
>>>>>>> fb055c7 (feat: v3.0 - Wiki知识库 + 科技感UI + i18n + Rust PyO3核心 + Pro功能)
        "features": features,
        "expires_at": expires_at,
        "device_slot": device_slot,
        "max_devices": max_devices,
        "device_count": device_count,
        "is_valid": True,
    }


@router.get("/status")
<<<<<<< HEAD
def get_license_status(current_user: User = Depends(get_current_user), db: Session = Depends(get_db)):
    # 完整性校验：纯 Python 模式下检测运行时篡改
    from app.services.license_service import USING_RUST
    if not USING_RUST:
=======
def get_license_status(_=Depends(get_current_user), db: Session = Depends(get_db)):
    from app.services.license_service import USING_CYTHON, USING_RUST
    if not USING_RUST and not USING_CYTHON:
>>>>>>> fb055c7 (feat: v3.0 - Wiki知识库 + 科技感UI + i18n + Rust PyO3核心 + Pro功能)
        try:
            from app.services.license_service import verify_integrity
            if not verify_integrity():
                return {"active": False, "tier": "oss", "features": [], "is_valid": False}
        except Exception:
            pass
    return _build_license_info(db)

<<<<<<< HEAD
    return _build_license_info(current_user, db)


@router.get("/info")
def get_license_info(current_user: User = Depends(get_current_user), db: Session = Depends(get_db)):
    """前端授权信息接口 — 返回 {active, tier, features, expires_at, ...}"""
    from app.services.license_service import USING_RUST
    if not USING_RUST:
=======

@router.get("/info")
def get_license_info(_=Depends(get_current_user), db: Session = Depends(get_db)):
    from app.services.license_service import USING_CYTHON, USING_RUST
    if not USING_RUST and not USING_CYTHON:
>>>>>>> fb055c7 (feat: v3.0 - Wiki知识库 + 科技感UI + i18n + Rust PyO3核心 + Pro功能)
        try:
            from app.services.license_service import verify_integrity
            if not verify_integrity():
                return {"active": False, "tier": "oss", "features": [], "is_valid": False}
        except Exception:
            pass
<<<<<<< HEAD

    return _build_license_info(current_user, db)
=======
    return _build_license_info(db)
>>>>>>> fb055c7 (feat: v3.0 - Wiki知识库 + 科技感UI + i18n + Rust PyO3核心 + Pro功能)


@router.post("/activate")
async def activate_license(req: LicenseActivateRequest, _=Depends(get_current_user), db: Session = Depends(get_db)):
    svc = LicenseService(db)
    result = await svc.activate(req.license_key)
<<<<<<< HEAD
    # 激活成功后，返回前端期望的格式
    if result.get("valid"):
        return _build_license_info(current_user, db)
=======
    if result.get("valid"):
        return _build_license_info(db)
>>>>>>> fb055c7 (feat: v3.0 - Wiki知识库 + 科技感UI + i18n + Rust PyO3核心 + Pro功能)
    return result


@router.post("/deactivate")
<<<<<<< HEAD
def deactivate_license(current_user: User = Depends(get_current_user), db: Session = Depends(get_db)):
    # 清除 DB 中的授权记录
=======
def deactivate_license(_=Depends(get_current_user), db: Session = Depends(get_db)):
>>>>>>> fb055c7 (feat: v3.0 - Wiki知识库 + 科技感UI + i18n + Rust PyO3核心 + Pro功能)
    lic = db.query(License).filter(License.status == "active").first()
    if lic:
        lic.status = "revoked"
        db.commit()
    reset()
    return {"active": False, "tier": "oss", "message": "License deactivated"}
