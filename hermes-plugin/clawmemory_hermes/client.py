from __future__ import annotations

import json
from typing import Any

import httpx


class ClawMemoryClient:
    def __init__(
        self,
        base_url: str = "http://localhost:8765",
        api_key: str = "",
        platform: str = "hermes",
        timeout: float = 10.0,
    ) -> None:
        self.base_url = base_url.rstrip("/")
        self.api_key = api_key
        self.platform = platform
        self._client = httpx.Client(
            timeout=timeout,
            headers={
                "Content-Type": "application/json",
                "X-API-Key": self.api_key,
                "X-Platform": self.platform,
            },
        )

    def _url(self, path: str) -> str:
        return f"{self.base_url}/api/v1/external{path}"

    def _request(self, method: str, path: str, body: dict | None = None) -> Any:
        url = self._url(path)
        max_retries = 2
        last_err: Exception | None = None
        for attempt in range(max_retries + 1):
            try:
                if method == "GET":
                    resp = self._client.get(url)
                else:
                    resp = self._client.post(url, json=body or {})
                resp.raise_for_status()
                return resp.json()
            except Exception as e:
                last_err = e
                logger.error("ClawMemory API request failed (attempt %d/%d): %s %s - %s",
                             attempt + 1, max_retries + 1, method, path, e)
                if attempt < max_retries:
                    import time
                    time.sleep((attempt + 1) * 1.0)
        raise last_err

    def save_memory(
        self,
        key: str,
        value: str,
        layer: str = "episodic",
        source: str = "",
        memory_type: str = "conversation",
    ) -> dict:
        return self._request(
            "POST",
            "/memories",
            {
                "key": key,
                "value": value,
                "layer": layer,
                "source": source or self.platform,
                "memory_type": memory_type,
            },
        )

    def batch_save_memories(self, memories: list[dict]) -> dict:
        items = []
        for m in memories:
            items.append(
                {
                    "key": m["key"],
                    "value": m["value"],
                    "layer": m.get("layer", "episodic"),
                    "source": m.get("source", self.platform),
                    "memory_type": m.get("memory_type", "conversation"),
                }
            )
        return self._request("POST", "/memories/batch", {"memories": items})

    def search_memories(self, query: str, limit: int = 10) -> list[dict]:
        result = self._request("GET", f"/memories/search?q={query}&limit={limit}")
        return result.get("items", [])

    def get_context(self, query: str, limit: int = 5) -> dict:
        return self._request("GET", f"/memories/context?q={query}&limit={limit}")

    def reason(
        self,
        query: str,
        depth: int = 1,
        level: str = "medium",
        session_id: str = "",
    ) -> dict:
        return self._request(
            "POST",
            "/reason",
            {
                "query": query,
                "depth": depth,
                "level": level,
                "session_id": session_id,
            },
        )

    def push_conversation(
        self,
        session_id: str,
        messages: list[dict] | None = None,
        summary: str = "",
        title: str = "",
        agent_name: str = "",
    ) -> dict:
        return self._request(
            "POST",
            "/conversations",
            {
                "session_id": session_id,
                "messages": messages or [],
                "summary": summary,
                "title": title,
                "agent_name": agent_name or self.platform,
            },
        )

    def track_session(self, session_id: str, metadata: str = "") -> dict:
        return self._request(
            "POST",
            "/sessions/track",
            {"session_id": session_id, "metadata": metadata},
        )

    def close(self) -> None:
        self._client.close()
