from pydantic import BaseModel


class BackupCreate(BaseModel):
    notes: str | None = None


class BackupResponse(BaseModel):
    id: int
    filename: str
    file_size: int
    memory_count: int
    backup_type: str
    notes: str | None
    created_at: str
