"""OpenClaw skill scanner — scans installed OpenClaw skills from disk.

Scans skills from multiple directories (priority high→low):
  1. <workspace>/skills/                  — workspace skills (highest priority)
  2. <workspace>/.agents/skills/          — project agent skills
  3. ~/.agents/skills/                    — personal agent skills (cross-workspace)
  4. ~/.openclaw/skills/                  — managed/local skills (all agents)
  5. bundled (npm/app)                    — built-in skills (not scanned here)

Skill file: SKILL.md (uppercase) with YAML frontmatter.
Also checks skill.md (lowercase) for backward compatibility on case-insensitive filesystems.
"""

import logging
import os
import re
from pathlib import Path

logger = logging.getLogger(__name__)

# SKILL.md filenames to try (uppercase first, then lowercase fallback)
_SKILL_FILENAMES = ["SKILL.md", "skill.md"]


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


def _find_skill_md(skill_dir: Path) -> Path | None:
    """Find SKILL.md or skill.md in a skill directory."""
    for filename in _SKILL_FILENAMES:
        path = skill_dir / filename
        if path.exists():
            return path
    return None


def _parse_skill_md(path: Path) -> dict | None:
    """Parse a SKILL.md file and extract metadata.

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


def _scan_skills_dir(skills_dir: Path, scope: str) -> list[dict]:
    """Scan a single skills directory for skill folders."""
    if not skills_dir.exists():
        return []

    results = []
    for d in sorted(skills_dir.iterdir()):
        if not d.is_dir() or d.name.startswith("."):
            continue
        skill_md = _find_skill_md(d)
        if skill_md:
            parsed = _parse_skill_md(skill_md)
            if parsed:
                parsed["scope"] = scope
                results.append(parsed)

    return results


def scan_global_skills() -> list[dict]:
    """Scan globally installed OpenClaw skills.

    Checks (in priority order):
      ~/.openclaw/skills/
      ~/.agents/skills/
    """
    results = []

    # ~/.openclaw/skills/
    openclaw_dir = _detect_openclaw_dir()
    if openclaw_dir:
        results.extend(_scan_skills_dir(openclaw_dir / "skills", scope="global"))

    # ~/.agents/skills/ (personal agent skills, cross-workspace)
    agents_dir = Path.home() / ".agents" / "skills"
    if agents_dir.exists():
        results.extend(_scan_skills_dir(agents_dir, scope="agents"))

    return results


def scan_workspace_skills(workspace_path: str | Path | None = None) -> list[dict]:
    """Scan project-local OpenClaw skills.

    Checks (in priority order):
      <workspace>/skills/
      <workspace>/.agents/skills/
      <workspace>/.openclaw/skills/  (legacy)
    """
    if workspace_path is None:
        workspace_path = Path.cwd()
    else:
        workspace_path = Path(workspace_path)

    results = []

    # <workspace>/skills/ (highest priority)
    results.extend(_scan_skills_dir(workspace_path / "skills", scope="workspace"))

    # <workspace>/.agents/skills/
    results.extend(_scan_skills_dir(workspace_path / ".agents" / "skills", scope="workspace-agents"))

    # <workspace>/.openclaw/skills/ (legacy path)
    results.extend(_scan_skills_dir(workspace_path / ".openclaw" / "skills", scope="workspace-legacy"))

    return results


def scan_all_skills(workspace_path: str | Path | None = None) -> dict:
    """Scan all skill sources and return combined results.

    Returns dict with global_skills, workspace_skills, total.
    """
    global_skills = scan_global_skills()
    workspace_skills = scan_workspace_skills(workspace_path)

    return {
        "global_skills": global_skills,
        "workspace_skills": workspace_skills,
        "total": len(global_skills) + len(workspace_skills),
    }


def get_skill_detail(skill_dir: str, scope: str = "global") -> dict | None:
    """Get detailed info about a specific skill."""
    path = Path(skill_dir)
    skill_md = _find_skill_md(path)
    if not skill_md:
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
