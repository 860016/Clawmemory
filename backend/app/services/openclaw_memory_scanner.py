"""OpenClaw memory scanner — reads existing OpenClaw memory files from disk.

Supports all known OpenClaw directory layouts:
  v4 (current): ~/.openclaw/workspace/memory/          (single agent, default)
                ~/.openclaw/workspace-<profile>/memory/ (multi agent/profile)
                Also reads workspace/MEMORY.md (long-term memory)
  v1/v3 (old):  ~/.openclaw/workspace/{agent}/memory/
  v2 (old):     ~/.openclaw/agents/{agent}/memory/

Also handles:
  - Pure Markdown files (key = filename, value = full content)
  - JSON files (list of {key, value, ...} dicts)
  - Markdown with YAML front-matter (---\\nkey: value\\n---)
"""

import json
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


def _find_agent_dirs(openclaw_dir: Path) -> list[dict]:
    """Find all agent directories across known layouts.

    Returns list of {"name": str, "memory_dir": Path, "layout": str}
    
    Layout priority:
    v4 (current): workspace/memory/ (default agent)
                  workspace-*/memory/ (named profiles)
                  workspace/MEMORY.md (long-term)
    v1/v3 (old):  workspace/{agent}/memory/
    v2 (old):     agents/{agent}/memory/
    """
    agents = []

    # === v4 (current): ~/.openclaw/workspace/memory/ ===
    workspace = openclaw_dir / "workspace"
    if workspace.exists():
        # Default workspace has memory/ directly
        mem_dir = workspace / "memory"
        if mem_dir.exists():
            agents.append({
                "name": "main",
                "memory_dir": mem_dir,
                "layout": "workspace-v4",
            })
        # Also scan MEMORY.md in workspace root (long-term memory)
        memory_md = workspace / "MEMORY.md"
        if memory_md.exists() and not any(a["name"] == "main" for a in agents):
            agents.append({
                "name": "main",
                "memory_dir": workspace,
                "layout": "workspace-v4",
            })

        # Fallback: v1/v3 layout — workspace/{agent}/memory/
        if not agents:
            for d in sorted(workspace.iterdir()):
                if d.is_dir():
                    mem_dir = d / "memory"
                    if mem_dir.exists() or (d / "memories.json").exists():
                        agents.append({
                            "name": d.name,
                            "memory_dir": mem_dir if mem_dir.exists() else d,
                            "layout": "workspace",
                        })

    # v4: workspace-<profile>/memory/ (multi-profile)
    for d in sorted(openclaw_dir.iterdir()):
        if d.is_dir() and d.name.startswith("workspace-"):
            profile_name = d.name[len("workspace-"):]
            mem_dir = d / "memory"
            if mem_dir.exists():
                agents.append({
                    "name": profile_name,
                    "memory_dir": mem_dir,
                    "layout": "workspace-v4",
                })

    # v2 (old): agents/{agent}/memory/
    agents_dir = openclaw_dir / "agents"
    if agents_dir.exists():
        for d in sorted(agents_dir.iterdir()):
            if d.is_dir():
                # Skip sessions dir (not memory)
                if d.name == "sessions":
                    continue
                mem_dir = d / "memory"
                if mem_dir.exists() or (d / "memories.json").exists():
                    agents.append({
                        "name": d.name,
                        "memory_dir": mem_dir if mem_dir.exists() else d,
                        "layout": "agents",
                    })

    return agents


def _parse_frontmatter(text: str) -> tuple[dict, str]:
    """Parse YAML front-matter from Markdown. Returns (meta, body)."""
    if not text.startswith("---"):
        return {}, text
    match = re.match(r"^---\s*\n(.*?)\n---\s*\n(.*)", text, re.DOTALL)
    if not match:
        return {}, text
    meta = {}
    for line in match.group(1).splitlines():
        if ":" in line:
            k, v = line.split(":", 1)
            meta[k.strip()] = v.strip().strip('"').strip("'")
    return meta, match.group(2)


def _parse_md_file(path: Path) -> dict:
    """Parse a single Markdown memory file."""
    try:
        text = path.read_text(encoding="utf-8")
    except Exception as e:
        logger.warning(f"Cannot read {path}: {e}")
        return {}

    meta, body = _parse_frontmatter(text)

    # If front-matter has key/value, use them; otherwise key=filename
    key = meta.get("key", path.stem)
    value = meta.get("value", body.strip())

    # Detect layer from front-matter or heading
    layer = meta.get("layer", "knowledge")
    source = meta.get("source", "openclaw_import")

    # Try to extract importance
    importance = 0.5
    if "importance" in meta:
        try:
            importance = float(meta["importance"])
        except ValueError:
            pass

    # Extract tags from front-matter
    tags = []
    if "tags" in meta:
        tags = [t.strip() for t in meta["tags"].split(",") if t.strip()]

    # Parse date from filename (YYYY-MM-DD.md)
    date_match = re.match(r"(\d{4}-\d{2}-\d{2})", path.stem)
    if date_match and not meta.get("key"):
        key = date_match.group(1)

    return {
        "key": key,
        "value": value,
        "layer": layer,
        "importance": importance,
        "tags": tags,
        "source": source,
        "file_path": str(path),
    }


def _parse_json_file(path: Path) -> list[dict]:
    """Parse a JSON memory file (list or dict)."""
    try:
        data = json.loads(path.read_text(encoding="utf-8"))
    except Exception as e:
        logger.warning(f"Cannot parse JSON {path}: {e}")
        return []

    items = []
    if isinstance(data, list):
        for item in data:
            if isinstance(item, dict):
                items.append(_normalize_json_item(item, path))
    elif isinstance(data, dict):
        # Could be a single memory or {memories: [...]}
        if "memories" in data and isinstance(data["memories"], list):
            for item in data["memories"]:
                if isinstance(item, dict):
                    items.append(_normalize_json_item(item, path))
        else:
            items.append(_normalize_json_item(data, path))

    return items


def _normalize_json_item(item: dict, path: Path) -> dict:
    """Normalize a JSON memory item to our schema."""
    # Handle different key names across OpenClaw versions
    key = item.get("key") or item.get("name") or item.get("title") or path.stem
    value = item.get("value") or item.get("content") or item.get("text") or ""
    layer = item.get("layer") or item.get("level") or "knowledge"

    # Map numeric layer to string
    layer_map = {"1": "preference", "2": "knowledge", "3": "short_term", "4": "private"}
    if str(layer) in layer_map:
        layer = layer_map[str(layer)]
    elif isinstance(layer, int) and 1 <= layer <= 4:
        layer = layer_map[str(layer)]

    importance = item.get("importance", 0.5)
    if isinstance(importance, int) and 1 <= importance <= 5:
        importance = importance / 5.0  # Normalize 1-5 → 0.2-1.0

    tags = item.get("tags", [])
    if isinstance(tags, str):
        tags = [t.strip() for t in tags.split(",") if t.strip()]

    source = item.get("source", "openclaw_import")

    return {
        "key": str(key),
        "value": str(value),
        "layer": layer,
        "importance": float(importance),
        "tags": tags,
        "source": source,
        "file_path": str(path),
    }


def scan_agent_memories(memory_dir: Path) -> list[dict]:
    """Scan all memory files in a directory, return list of parsed items.
    
    Also scans MEMORY.md in parent directory (workspace root) for long-term memory.
    """
    items = []
    if not memory_dir.exists():
        return items

    for f in sorted(memory_dir.iterdir()):
        if f.is_dir():
            continue
        if f.name.startswith("."):
            continue
        # Skip writing-in-progress markers
        if f.suffix == ".writing":
            continue

        if f.suffix == ".md":
            parsed = _parse_md_file(f)
            if parsed and parsed.get("value"):
                items.append(parsed)
        elif f.suffix == ".json":
            parsed = _parse_json_file(f)
            items.extend(parsed)
        elif f.suffix == ".txt":
            # Plain text: key=filename, value=content
            try:
                text = f.read_text(encoding="utf-8").strip()
                if text:
                    items.append({
                        "key": f.stem,
                        "value": text,
                        "layer": "knowledge",
                        "importance": 0.5,
                        "tags": [],
                        "source": "openclaw_import",
                        "file_path": str(f),
                    })
            except Exception:
                pass

    # Also check MEMORY.md in workspace root (parent of memory/ dir)
    memory_md = memory_dir.parent / "MEMORY.md"
    if memory_md.exists() and memory_md != memory_dir:
        parsed = _parse_md_file(memory_md)
        if parsed and parsed.get("value"):
            parsed["key"] = "MEMORY.md"
            parsed["layer"] = "preference"  # Long-term memory
            items.append(parsed)

    return items


def scan_openclaw_memories(openclaw_dir: Path | None = None) -> list[dict]:
    """Full scan: detect OpenClaw dir, find all agents, scan their memories.

    Returns list of {"agent_name": str, "layout": str, "files": int, "memories": list[dict]}
    """
    if openclaw_dir is None:
        openclaw_dir = _detect_openclaw_dir()
    if openclaw_dir is None or not openclaw_dir.exists():
        return []

    agents = _find_agent_dirs(openclaw_dir)
    results = []

    for agent_info in agents:
        mem_dir = agent_info["memory_dir"]
        memories = scan_agent_memories(mem_dir)

        # Also check for memories.json in parent dir
        json_file = mem_dir.parent / "memories.json"
        if json_file.exists():
            memories.extend(_parse_json_file(json_file))

        results.append({
            "agent_name": agent_info["name"],
            "layout": agent_info["layout"],
            "memory_dir": str(mem_dir),
            "files": len(memories),
            "memories": memories,
        })

    return results
