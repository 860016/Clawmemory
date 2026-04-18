from pydantic import BaseModel


class EntityCreate(BaseModel):
    name: str
    entity_type: str
    description: str | None = None
    properties: dict | None = None


class EntityUpdate(BaseModel):
    name: str | None = None
    entity_type: str | None = None
    description: str | None = None
    properties: dict | None = None


class RelationCreate(BaseModel):
    source_id: int
    target_id: int
    relation_type: str
    description: str | None = None


class EntityResponse(BaseModel):
    id: int
    name: str
    entity_type: str
    description: str | None
    properties: dict | None


class RelationResponse(BaseModel):
    id: int
    source_id: int
    target_id: int
    relation_type: str
    description: str | None
