import asyncio
import logging
import time

import httpx

from config import (
    THUNDER_BASE_URL,
    THUNDER_CLIENT_ID,
    THUNDER_CLIENT_SECRET,
    THUNDER_RESOURCE,
    THUNDER_VERIFY_TLS,
)

logger = logging.getLogger(__name__)

DEFAULT_TOKEN_TTL_SECONDS = 3600  # fallback only, if a response ever omits expires_in
TOKEN_REFRESH_BUFFER_SECONDS = 60


class ThunderTokenError(Exception):
    """Raised when a ThunderID access token could not be obtained."""


class ThunderIDClient:
    def __init__(self):
        self._cached_token: str | None = None
        self._cached_token_expiry: float = 0
        self._lock = asyncio.Lock()

    async def get_access_token(self):
        if self._is_cached_token_valid():
            return self._cached_token

        async with self._lock:
            if self._is_cached_token_valid():
                return self._cached_token

            token, expires_in = await self._fetch_access_token()
            self._cached_token = token
            self._cached_token_expiry = time.monotonic() + max(
                expires_in - TOKEN_REFRESH_BUFFER_SECONDS, 0
            )
            return token

    def invalidate_token(self, token: str):
        """Forces the next get_access_token() call to fetch a fresh token,
        e.g. after the backend rejects the cached one with a 401 — our local
        TTL has no way of knowing the token died earlier than expected.

        Only clears the cache if it still holds the given (now-known-bad)
        token: two crawlers can share this client and race a 401 against a
        concurrent refresh, and we must not let a stale invalidation wipe
        out a token someone else already renewed."""
        if self._cached_token == token:
            self._cached_token = None
            self._cached_token_expiry = 0

    def _is_cached_token_valid(self):
        return self._cached_token is not None and time.monotonic() < self._cached_token_expiry

    async def _fetch_access_token(self):
        async with httpx.AsyncClient(verify=THUNDER_VERIFY_TLS) as client:
            response = await client.post(
                f"{THUNDER_BASE_URL}/oauth2/token",
                auth=(THUNDER_CLIENT_ID, THUNDER_CLIENT_SECRET),
                headers={"Content-Type": "application/x-www-form-urlencoded"},
                data={
                    "grant_type": "client_credentials",
                    "scope": "crawler:runs crawler:complete crawler:lookup crawler:batch-save crawler:batch-update crawler:reconcile",
                    "resource": THUNDER_RESOURCE,
                },
            )
            response.raise_for_status()
            body = response.json()
            token = body.get("access_token")
            if not token:
                raise ValueError(f"ThunderID token response missing 'access_token': {body}")

            expires_in = body.get("expires_in", DEFAULT_TOKEN_TTL_SECONDS)

            logger.info("Fetched new ThunderID access token")
            return token, expires_in


class ThunderAuth(httpx.Auth):
    """Attaches a ThunderID Bearer token to a request and, if the backend
    rejects it with a 401, invalidates the cached token and retries once
    with a fresh one — centralizes auth so call sites never need to touch
    headers or status codes themselves. Pass this as `auth=` on individual
    requests to BACKEND_BASE_URL only — never as a client-wide default,
    since the shared httpx.AsyncClient in this codebase is also used to
    call third-party job sites, which must never see this token."""

    def __init__(self, thunder_client: ThunderIDClient):
        self._thunder_client = thunder_client

    async def async_auth_flow(self, request):
        token = await self._get_token()
        request.headers["Authorization"] = f"Bearer {token}"
        response = yield request

        if response.status_code == 401:
            logger.warning(
                f"{request.method} {request.url} rejected with 401 — "
                "invalidating cached ThunderID token and retrying with a fresh one"
            )
            self._thunder_client.invalidate_token(token)
            token = await self._get_token()
            request.headers["Authorization"] = f"Bearer {token}"
            yield request

    async def _get_token(self) -> str:
        try:
            return await self._thunder_client.get_access_token()
        except Exception as e:
            logger.error(f"Failed to obtain ThunderID access token: {e}")
            raise ThunderTokenError(str(e)) from e
