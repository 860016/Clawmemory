from __future__ import annotations

import os
import uuid
from datetime import datetime
from pathlib import Path
from typing import Any

from .client import ClawMemoryClient


class SessionManager:
    def __init__(self, client: ClawMemoryClient, strategy: str = "per-directory") -> None:
        self.client = client
        self.strategy = strategy
        self._session_id: str | None = None
        self._project_path: str | None = None
        self._turn_count: int = 0
        self._pending_messages: list[dict] = []

    @property
    def session_id(self) -> str:
        if self._session_id is None:
            self._session_id = self._resolve_session_id()
        return self._session_id

    def _resolve_session_id(self) -> str:
        if self.strategy == "global":
            return "hermes-global"
        if self.strategy == "per-repo":
            git_dir = self._find_git_dir()
            if git_dir:
                return f"hermes-repo-{hash(git_dir)}"
            return "hermes-global"
        if self.strategy == "per-directory":
            cwd = os.getcwd()
            return f"hermes-dir-{hash(cwd)}"
        return f"hermes-session-{uuid.uuid4().hex[:8]}"

    def _find_git_dir(self) -> str | None:
        current = Path(os.getcwd())
        while current != current.parent:
            if (current / ".git").exists():
                return str(current)
            current = current.parent
        return None

    def on_user_message(self, content: str) -> None:
        self._pending_messages.append({"role": "user", "content": content})

    def on_assistant_message(self, content: str) -> None:
        self._pending_messages.append({"role": "assistant", "content": content})

    def flush(self) -> None:
        if not self._pending_messages:
            return

        combined = "\n".join(
            f"[{m['role']}] {m['content']}" for m in self._pending_messages
        )
        key = f"{self.session_id}_turn_{self._turn_count}_{int(datetime.now().timestamp())}"

        self.client.save_memory(
            key=key,
            value=combined,
            layer="episodic",
            source="hermes",
            memory_type="conversation",
        )

        self.client.track_session(
            self.session_id,
            metadata=f"turns:{self._turn_count} last_active:{datetime.now().isoformat()}",
        )

        self._pending_messages.clear()

    def increment_turn(self) -> None:
        self._turn_count += 1

    def get_context(self, query: str, limit: int = 5) -> str:
        result = self.client.get_context(query, limit)
        return result.get("system_prompt_addition", "")

    def reason(self, query: str, depth: int = 1, level: str = "low") -> str:
        result = self.client.reason(query, depth, level, self.session_id)
        return result.get("reasoning", "")

    def reset(self) -> None:
        if self.strategy == "per-session":
            self._session_id = None
        self._turn_count = 0
        self._pending_messages.clear()
