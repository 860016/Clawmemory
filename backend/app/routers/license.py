from fastapi import APIRouter, Depends, Query
from sqlalchemy.orm import Session
from app.database import get_db
from app.middleware.auth import get_current_user
from app.models.license import License
from app.schemas.license import LicenseActivateRequest
from app.services.license_service import LicenseService, current_tier, is_feature_enabled, reset
from app.core.pro_loader import is_pro_installed, get_pro_info, download_pro_package, uninstall_pro

router = APIRouter(prefix="/api/v1/license", tags=["license"])


def _build_license_info(db: Session) -> dict:
    """构建前端期望的授权信息格式"""
    tier = current_tier()
    active = tier != "oss"

    all_pro_features = [
        "ai_extract", "auto_graph", "unlimited_graph",
        "auto_decay", "decay_report", "prune_suggest", "reinforce",
        "conflict_scan", "conflict_merge",
        "smart_router", "token_stats",
        "wiki", "auto_backup",
    ]
    features = [f for f in all_pro_features if is_feature_enabled(f)]

    lic = db.query(License).filter(License.status == "active").first()
    expires_at = None
    device_slot = ""
    max_devices = 1
    device_count = 0
    license_type = "none"
    license_key_display = ""
    pro_download_url = ""
    pro_fallback_urls = []

    if lic:
        expires_at = str(lic.expires_at) if lic.expires_at else None
        device_slot = lic.device_slot or ""
        license_type = lic.tier
        if lic.license_key and len(lic.license_key) > 8:
            license_key_display = lic.license_key[:4] + "****" + lic.license_key[-4:]
        else:
            license_key_display = "****"
        if device_slot and "/" in device_slot:
            try:
                parts = device_slot.split("/")
                device_count = int(parts[0])
                max_devices = int(parts[1])
            except (ValueError, IndexError):
                pass
        # 从数据库读取 Pro 下载地址
        if lic.pro_download_url:
            pro_download_url = lic.pro_download_url
        if lic.pro_fallback_urls:
            try:
                import json
                pro_fallback_urls = json.loads(lic.pro_fallback_urls)
            except Exception:
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
        "pro_download_url": pro_download_url,
        "pro_fallback_urls": pro_fallback_urls,
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
        if result.get("pro_download_url"):
            info["pro_download_url"] = result["pro_download_url"]
        if result.get("pro_fallback_urls"):
            info["pro_fallback_urls"] = result["pro_fallback_urls"]
        return info
    return result


@router.post("/deactivate")
async def deactivate_license(_=Depends(get_current_user), db: Session = Depends(get_db)):
    svc = LicenseService(db)
    result = await svc.deactivate()
    return result


@router.get("/pro/status")
def get_pro_status(_=Depends(get_current_user)):
    """获取 Pro 模块安装状态"""
    return get_pro_info()


@router.post("/pro/install")
async def install_pro(
    url: str = Query(..., description="主下载链接"),
    fallback_urls: str | None = Query(None, description="备用下载链接（逗号分隔）"),
    _=Depends(get_current_user),
):
    """安装 Pro 模块"""
    fallback_list = [u.strip() for u in fallback_urls.split(",") if u.strip()] if fallback_urls else []
    success = await download_pro_package(url, fallback_list)
    if success:
        return {"success": True, "message": "Pro module installed", "info": get_pro_info()}
    return {"success": False, "message": "Failed to download Pro module from all sources"}


@router.post("/pro/uninstall")
def uninstall_pro_module(_=Depends(get_current_user)):
    """卸载 Pro 模块"""
    success = uninstall_pro()
    return {"success": success, "message": "Pro module uninstalled" if success else "Failed to uninstall"}
