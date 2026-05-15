from __future__ import annotations

import os
import uuid
from datetime import datetime
from pathlib import Path
from typing import Any

from .client import ClawMemoryClient

logger = __import__("logging").getLogger("clawmemory_hermes.session")

AUTO_FLUSH_THRESHOLD = 10


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
        self._auto_flush_check()

    def on_assistant_message(self, content: str) -> None:
        self._pending_messages.append({"role": "assistant", "content": content})
        self._auto_flush_check()

    def _auto_flush_check(self) -> None:
        if len(self._pending_messages) >= AUTO_FLUSH_THRESHOLD:
            logger.info("Auto-flushing %d pending messages", len(self._pending_messages))
            self.flush()

    def flush(self) -> None:
        if not self._pending_messages:
            return

        try:
            if len(self._pending_messages) == 1:
                m = self._pending_messages[0]
                key = f"{self.session_id}_{m['role']}_{int(datetime.now().timestamp())}"
                self.client.save_memory(
                    key=key,
                    value=f"[{m['role']}] {m['content']}",
                    layer="episodic",
                    source="hermes",
                    memory_type="conversation",
                )
            else:
                items = []
                for m in self._pending_messages:
                    key = f"{self.session_id}_{m['role']}_{int(datetime.now().timestamp())}_{hash(m['content'][:50])}"
                    items.append({
                        "key": key,
                        "value": f"[{m['role']}] {m['content']}",
                        "layer": "episodic",
                        "source": "hermes",
                        "memory_type": "conversation",
                    })
                self.client.batch_save_memories(items)

            logger.info("Flushed %d messages for session %s", len(self._pending_messages), self.session_id)
        except Exception as e:
            logger.error("Flush failed, attempting individual writes: %s", e)
            for m in self._pending_messages:
                try:
                    key = f"{self.session_id}_{m['role']}_{int(datetime.now().timestamp())}"
                    self.client.save_memory(
                        key=key,
                        value=f"[{m['role']}] {m['content']}",
                        layer="episodic",
                        source="hermes",
                        memory_type="conversation",
                    )
                except Exception as inner_err:
                    logger.error("Individual write also failed: %s", inner_err)

        try:
            self.client.track_session(
                self.session_id,
                metadata=f"turns:{self._turn_count} last_active:{datetime.now().isoformat()}",
            )
        except Exception as e:
            logger.error("Session track failed: %s", e)

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
