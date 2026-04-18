from fastapi import HTTPException
from app.services.license_service import is_feature_enabled


def require_feature(feature: str):
    def decorator(func):
        async def wrapper(*args, **kwargs):
            if not is_feature_enabled(feature):
                raise HTTPException(status_code=403, detail=f"Feature '{feature}' requires Pro/Enterprise license")
            return await func(*args, **kwargs)
        return wrapper
    return decorator
