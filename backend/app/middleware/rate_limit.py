import time
from collections import defaultdict
from fastapi import Request
from fastapi.responses import JSONResponse
from starlette.middleware.base import BaseHTTPMiddleware


class RateLimitMiddleware(BaseHTTPMiddleware):
    """Simple in-memory rate limiter per IP.

    Uses a sliding window algorithm. For production, replace with Redis-backed store.
    """

    def __init__(self, app, requests_per_minute: int = 60, burst: int = 10):
        super().__init__(app)
        self.requests_per_minute = requests_per_minute
        self.burst = burst
        self._requests: dict[str, list[float]] = defaultdict(list)

    def _cleanup(self, ip: str, now: float) -> None:
        """Remove requests older than 1 minute."""
        cutoff = now - 60
        self._requests[ip] = [t for t in self._requests[ip] if t > cutoff]

    async def dispatch(self, request: Request, call_next):
        # Skip rate limiting for static files and health checks
        path = request.url.path
        if path.startswith("/assets") or path == "/api/v1/health" or path == "/":
            return await call_next(request)

        ip = request.client.host if request.client else "unknown"
        now = time.monotonic()

        self._cleanup(ip, now)

        requests = self._requests[ip]

        # Check burst limit (requests in last second)
        recent = [t for t in requests if t > now - 1]
        if len(recent) >= self.burst:
            return JSONResponse(
                status_code=429,
                content={"error": "RATE_LIMITED", "message": "Too many requests. Please slow down."},
            )

        # Check per-minute limit
        if len(requests) >= self.requests_per_minute:
            return JSONResponse(
                status_code=429,
                content={"error": "RATE_LIMITED", "message": "Rate limit exceeded. Try again later."},
            )

        self._requests[ip].append(now)
        return await call_next(request)
