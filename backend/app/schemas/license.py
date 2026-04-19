from pydantic import BaseModel


class LicenseActivateRequest(BaseModel):
    license_key: str


class LicenseStatusResponse(BaseModel):
    active: bool
    tier: str
    type: str
    features: list[str]
    expires_at: str | None
    device_slot: str
    max_devices: int
    device_count: int
    license_key: str
    is_valid: bool


class LicenseActivateResponse(BaseModel):
    valid: bool
    tier: str = ""
    features: list[str] = []
    expires_at: str | None = None
    device_slot: str = ""
    message: str = ""
