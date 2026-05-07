from __future__ import annotations

import logging
from abc import ABC, abstractmethod
from typing import Any

from .client import ClawMemoryClient
from .session import SessionManager

logger = logging.getLogger("clawmemory_hermes")


class MemoryProvider(ABC):
    @abstractmethod
    def save(self, key: str, value: str, **kwargs: Any) -> None:
        ...

    @abstractmethod
    def recall(self, query: str, limit: int = 5) -> list[dict]:
        ...

    @abstractmethod
    def get_context(self, query: str, limit: int = 5) -> str:
        ...

    @abstractmethod
    def on_user_message(self, content: str) -> None:
        ...

    @abstractmethod
    def on_assistant_message(self, content: str) -> None:
        ...

    @abstractmethod
    def flush(self) -> None:
        ...


class ClawMemoryProvider(MemoryProvider):
    def __init__(
        self,
        base_url: str = "http://localhost:8765",
        api_key: str = "",
        session_strategy: str = "per-directory",
        enable_reasoning: bool = False,
        reasoning_interval: int = 5,
        reasoning_depth: int = 1,
        reasoning_level: str = "low",
    ) -> None:
        self.client = ClawMemoryClient(
            base_url=base_url,
            api_key=api_key,
            platform="hermes",
        )
        self.session = SessionManager(self.client, strategy=session_strategy)
        self.enable_reasoning = enable_reasoning
        self.reasoning_interval = reasoning_interval
        self.reasoning_depth = reasoning_depth
        self.reasoning_level = reasoning_level

    def save(self, key: str, value: str, **kwargs: Any) -> None:
        self.client.save_memory(
            key=key,
            value=value,
            layer=kwargs.get("layer", "episodic"),
            source=kwargs.get("source", "hermes"),
            memory_type=kwargs.get("memory_type", "knowledge"),
        )

    def recall(self, query: str, limit: int = 5) -> list[dict]:
        return self.client.search_memories(query, limit)

    def get_context(self, query: str, limit: int = 5) -> str:
        return self.session.get_context(query, limit)

    def on_user_message(self, content: str) -> None:
        self.session.on_user_message(content)

    def on_assistant_message(self, content: str) -> None:
        self.session.on_assistant_message(content)
        self.session.increment_turn()

        if (
            self.enable_reasoning
            and self.session._turn_count % self.reasoning_interval == 0
        ):
            try:
                self.session.reason(
                    f"Auto reasoning at turn {self.session._turn_count}",
                    depth=self.reasoning_depth,
                    level=self.reasoning_level,
                )
            except Exception as e:
                logger.warning("Dialectic reasoning failed: %s", e)

    def flush(self) -> None:
        self.session.flush()

    def close(self) -> None:
        self.flush()
        self.client.close()


def create_provider(config: dict | None = None) -> ClawMemoryProvider:
    cfg = config or {}
    return ClawMemoryProvider(
        base_url=cfg.get("base_url", "http://localhost:8765"),
        api_key=cfg.get("api_key", ""),
        session_strategy=cfg.get("session_strategy", "per-directory"),
        enable_reasoning=cfg.get("enable_reasoning", False),
        reasoning_interval=cfg.get("reasoning_interval", 5),
        reasoning_depth=cfg.get("reasoning_depth", 1),
        reasoning_level=cfg.get("reasoning_level", "low"),
    )
