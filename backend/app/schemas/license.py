from pydantic import BaseModel


class LicenseActivateRequest(BaseModel):
    license_key: str


class LicenseStatusResponse(BaseModel):
    tier: str
    features: list[str]
    expires_at: str | None
    device_slot: str
    is_valid: bool


class LicenseActivateResponse(BaseModel):
    valid: bool
    tier: str = ""
    features: list[str] = []
    expires_at: str | None = None
    device_slot: str = ""
    message: str = ""
