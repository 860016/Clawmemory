from pydantic import BaseModel


class MemoryCreate(BaseModel):
    layer: str  # preference/knowledge/short_term/private
    key: str
    value: str
    importance: float = 0.5
    tags: list[str] | None = None
    source: str = "manual"


class MemoryUpdate(BaseModel):
    value: str | None = None
    importance: float | None = None
    tags: list[str] | None = None


class MemoryResponse(BaseModel):
    id: int
    layer: str
    key: str
    value: str
    importance: float
    access_count: int
    tags: list[str] | None
    source: str
    is_encrypted: bool
    created_at: str
    updated_at: str


class MemorySearchResult(BaseModel):
    id: int
    key: str
    value: str
    layer: str
    score: float
    source: str
