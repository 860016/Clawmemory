import pytest
from unittest.mock import AsyncMock, patch
from app.services.openclaw_service import OpenClawService


@pytest.mark.asyncio
async def test_list_agents():
    svc = OpenClawService()
    with patch("httpx.AsyncClient.get", new_callable=AsyncMock) as mock_get:
        mock_get.return_value.status_code = 200
        mock_get.return_value.json.return_value = [{"id": "main", "name": "Main Agent"}]
        mock_get.return_value.raise_for_status = lambda: None
        result = await svc.list_agents()
        assert len(result) == 1


@pytest.mark.asyncio
async def test_send_message():
    svc = OpenClawService()
    with patch("httpx.AsyncClient.post", new_callable=AsyncMock) as mock_post:
        mock_post.return_value.status_code = 200
        mock_post.return_value.json.return_value = {"response": "Hello!"}
        mock_post.return_value.raise_for_status = lambda: None
        result = await svc.send_message("main", "Hi")
        assert result["response"] == "Hello!"
