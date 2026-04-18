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
    features = [f for f in [
        "ai_extract", "auto_graph", "unlimited_graph",
        "auto_decay", "decay_report", "prune_suggest", "reinforce",
        "conflict_scan", "conflict_merge",
        "smart_router", "token_stats",
        "wiki", "auto_backup",
    ] if is_feature_enabled(f)]

    lic = db.query(License).filter(License.status == "active").first()
    expires_at = None
    device_slot = ""
    max_devices = 1
    device_count = 0
    license_type = "none"
    if lic:
        expires_at = str(lic.expires_at) if lic.expires_at else None
        device_slot = lic.device_slot or ""
        license_type = lic.tier
        if device_slot and "/" in device_slot:
            parts = device_slot.split("/")
            device_count = int(parts[0])
            max_devices = int(parts[1])

    return {
        "active": active,
        "tier": tier,
        "type": license_type,
        "features": features,
        "expires_at": expires_at,
        "device_slot": device_slot,
        "max_devices": max_devices,
        "device_count": device_count,
        "is_valid": True,
    }


@router.get("/status")
def get_license_status(_=Depends(get_current_user), db: Session = Depends(get_db)):
    from app.services.license_service import USING_RUST
    if not USING_RUST:
        try:
            from app.services.license_service import verify_integrity
            if not verify_integrity():
                return {"active": False, "tier": "oss", "features": [], "is_valid": False}
        except Exception as e:
            import logging
            logging.getLogger(__name__).warning(f"License integrity check failed: {e}")
    return _build_license_info(db)


@router.get("/info")
def get_license_info(_=Depends(get_current_user), db: Session = Depends(get_db)):
    from app.services.license_service import USING_RUST
    if not USING_RUST:
        try:
            from app.services.license_service import verify_integrity
            if not verify_integrity():
                return {"active": False, "tier": "oss", "features": [], "is_valid": False}
        except Exception as e:
            import logging
            logging.getLogger(__name__).warning(f"License integrity check failed: {e}")
    return _build_license_info(db)


@router.post("/activate")
async def activate_license(req: LicenseActivateRequest, _=Depends(get_current_user), db: Session = Depends(get_db)):
    svc = LicenseService(db)
    result = await svc.activate(req.license_key)
    if result.get("valid"):
        return _build_license_info(db)
    return result


@router.post("/deactivate")
def deactivate_license(_=Depends(get_current_user), db: Session = Depends(get_db)):
    lic = db.query(License).filter(License.status == "active").first()
    if lic:
        lic.status = "revoked"
        db.commit()
    reset()
    return {"active": False, "tier": "oss", "message": "License deactivated"}
