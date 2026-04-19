from fastapi import APIRouter, Depends
from sqlalchemy.orm import Session
from app.database import get_db
from app.middleware.auth import get_current_user
from app.models.license import License
from app.schemas.license import LicenseActivateRequest
from app.services.license_service import LicenseService, current_tier, is_feature_enabled, reset

router = APIRouter(prefix="/api/v1/license", tags=["license"])


def _build_license_info(db: Session) -> dict:
    """构建前端期望的授权信息格式"""
    tier = current_tier()
    active = tier != "oss"

    # 当前启用的功能列表
    all_pro_features = [
        "ai_extract", "auto_graph", "unlimited_graph",
        "auto_decay", "decay_report", "prune_suggest", "reinforce",
        "conflict_scan", "conflict_merge",
        "smart_router", "token_stats",
        "wiki", "auto_backup",
    ]
    features = [f for f in all_pro_features if is_feature_enabled(f)]

    # 从数据库读取额外信息
    lic = db.query(License).filter(License.status == "active").first()
    expires_at = None
    device_slot = ""
    max_devices = 1
    device_count = 0
    license_type = "none"
    license_key_display = ""

    if lic:
        expires_at = str(lic.expires_at) if lic.expires_at else None
        device_slot = lic.device_slot or ""
        license_type = lic.tier
        # 脱敏显示授权码
        if lic.license_key and len(lic.license_key) > 8:
            license_key_display = lic.license_key[:4] + "****" + lic.license_key[-4:]
        else:
            license_key_display = "****"
        # 解析 device_slot: "2/3" → device_count=2, max_devices=3
        if device_slot and "/" in device_slot:
            try:
                parts = device_slot.split("/")
                device_count = int(parts[0])
                max_devices = int(parts[1])
            except (ValueError, IndexError):
                pass

    return {
        "active": active,
        "tier": tier,
        "type": license_type,
        "features": features,
        "expires_at": expires_at,
        "device_slot": device_slot,
        "max_devices": max_devices,
        "device_count": device_count,
        "license_key": license_key_display,
        "is_valid": True,
    }


@router.get("/status")
def get_license_status(_=Depends(get_current_user), db: Session = Depends(get_db)):
    return _build_license_info(db)


@router.get("/info")
def get_license_info(_=Depends(get_current_user), db: Session = Depends(get_db)):
    return _build_license_info(db)


@router.post("/activate")
async def activate_license(req: LicenseActivateRequest, _=Depends(get_current_user), db: Session = Depends(get_db)):
    svc = LicenseService(db)
    result = await svc.activate(req.license_key)
    if result.get("valid"):
        info = _build_license_info(db)
        info["valid"] = True
        return info
    return result


@router.post("/deactivate")
async def deactivate_license(_=Depends(get_current_user), db: Session = Depends(get_db)):
    svc = LicenseService(db)
    result = await svc.deactivate()
    return result
