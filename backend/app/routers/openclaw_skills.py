from fastapi import APIRouter, Depends, HTTPException, Query
from app.middleware.auth import get_current_user
from app.services.openclaw_skill_scanner import (
    scan_all_skills,
    scan_global_skills,
    scan_workspace_skills,
    get_skill_detail,
)

router = APIRouter(prefix="/api/v1/openclaw-skills", tags=["openclaw-skills"])


@router.get("/scan")
def scan_skills(
    workspace_path: str | None = Query(None, description="Workspace path for local skills"),
    _=Depends(get_current_user),
):
    """Scan both global and workspace OpenClaw skills."""
    return scan_all_skills(workspace_path)


@router.get("/global")
def list_global_skills(_=Depends(get_current_user)):
    """List globally installed OpenClaw skills."""
    skills = scan_global_skills()
    return {"skills": skills, "total": len(skills)}


@router.get("/workspace")
def list_workspace_skills(
    workspace_path: str | None = Query(None),
    _=Depends(get_current_user),
):
    """List workspace-local OpenClaw skills."""
    skills = scan_workspace_skills(workspace_path)
    return {"skills": skills, "total": len(skills)}


@router.get("/detail")
def skill_detail(
    skill_dir: str = Query(..., description="Path to skill directory"),
    scope: str = Query("global", description="global or workspace"),
    _=Depends(get_current_user),
):
    """Get detailed info about a specific skill."""
    detail = get_skill_detail(skill_dir, scope)
    if not detail:
        raise HTTPException(status_code=404, detail="Skill not found")
    return detail
