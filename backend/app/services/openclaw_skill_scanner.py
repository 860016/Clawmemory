"""OpenClaw skill scanner — scans installed OpenClaw skills from disk.

Skills are stored as directories with skill.md files in:
  Global: ~/.openclaw/skills/
  Workspace: .openclaw/skills/ (project-local)
"""

import logging
import os
import re
from pathlib import Path
from typing import Generator

logger = logging.getLogger(__name__)


def _detect_openclaw_dir() -> Path | None:
    """Return the OpenClaw config directory if it exists.
    
    Checks: OPENCLAW_STATE_DIR env > OPENCLAW_HOME env > ~/.openclaw
    """
    candidates = []
    state_dir = os.environ.get("OPENCLAW_STATE_DIR", "")
    home_dir = os.environ.get("OPENCLAW_HOME", "")
    if state_dir:
        candidates.append(Path(state_dir))
    if home_dir:
        candidates.append(Path(home_dir) / ".openclaw")
    candidates.append(Path.home() / ".openclaw")

    for p in candidates:
        if p and p.exists():
            return p
    return None


def _parse_skill_md(path: Path) -> dict | None:
    """Parse a skill.md file and extract metadata.
    
    Expected format:
    ---
    name: Skill Name
    description: Skill description
    version: 1.0.0
    author: author-name
    tags: tag1, tag2
    ---
    
    Body content here...
    """
    try:
        text = path.read_text(encoding="utf-8")
    except Exception as e:
        logger.warning(f"Cannot read {path}: {e}")
        return None
    
    # Parse front-matter
    meta = {}
    body = text
    if text.startswith("---"):
        match = re.match(r"^---\s*\n(.*?)\n---\s*\n(.*)", text, re.DOTALL)
        if match:
            for line in match.group(1).splitlines():
                if ":" in line:
                    k, v = line.split(":", 1)
                    meta[k.strip().lower()] = v.strip().strip('"').strip("'")
            body = match.group(2)
    
    return {
        "name": meta.get("name", path.parent.name),
        "description": meta.get("description", ""),
        "version": meta.get("version", "unknown"),
        "author": meta.get("author", "unknown"),
        "tags": [t.strip() for t in meta.get("tags", "").split(",") if t.strip()],
        "skill_dir": str(path.parent),
        "skill_md": str(path),
        "body_preview": body.strip()[:500],
    }


def scan_global_skills() -> list[dict]:
    """Scan globally installed OpenClaw skills (~/.openclaw/skills/)."""
    openclaw_dir = _detect_openclaw_dir()
    if not openclaw_dir:
        return []
    
    skills_dir = openclaw_dir / "skills"
    if not skills_dir.exists():
        return []
    
    results = []
    for d in sorted(skills_dir.iterdir()):
        if not d.is_dir() or d.name.startswith("."):
            continue
        skill_md = d / "skill.md"
        if skill_md.exists():
            parsed = _parse_skill_md(skill_md)
            if parsed:
                parsed["scope"] = "global"
                results.append(parsed)
    
    return results


def scan_workspace_skills(workspace_path: str | Path | None = None) -> list[dict]:
    """Scan project-local OpenClaw skills (.openclaw/skills/)."""
    if workspace_path is None:
        workspace_path = Path.cwd()
    else:
        workspace_path = Path(workspace_path)
    
    skills_dir = workspace_path / ".openclaw" / "skills"
    if not skills_dir.exists():
        return []
    
    results = []
    for d in sorted(skills_dir.iterdir()):
        if not d.is_dir() or d.name.startswith("."):
            continue
        skill_md = d / "skill.md"
        if skill_md.exists():
            parsed = _parse_skill_md(skill_md)
            if parsed:
                parsed["scope"] = "workspace"
                results.append(parsed)
    
    return results


def get_skill_detail(skill_dir: str, scope: str = "global") -> dict | None:
    """Get detailed info about a specific skill."""
    path = Path(skill_dir)
    skill_md = path / "skill.md"
    if not skill_md.exists():
        return None
    
    parsed = _parse_skill_md(skill_md)
    if parsed:
        parsed["scope"] = scope
        # Read full body content
        try:
            text = skill_md.read_text(encoding="utf-8")
            if text.startswith("---"):
                match = re.match(r"^---\s*\n(.*?)\n---\s*\n(.*)", text, re.DOTALL)
                if match:
                    parsed["body_full"] = match.group(2).strip()
                else:
                    parsed["body_full"] = text
            else:
                parsed["body_full"] = text
        except Exception:
            parsed["body_full"] = parsed.get("body_preview", "")
        
        # List files in skill dir
        parsed["files"] = [f.name for f in sorted(path.iterdir()) if not f.name.startswith(".")]
    
    return parsed
