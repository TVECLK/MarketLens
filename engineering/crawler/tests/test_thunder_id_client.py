import asyncio
from unittest.mock import AsyncMock, MagicMock, patch

import httpx
import pytest

from utils.thunder_id_client import (
    DEFAULT_TOKEN_TTL_SECONDS,
    TOKEN_REFRESH_BUFFER_SECONDS,
    ThunderAuth,
    ThunderTokenError,
)

DEFAULT_EXPIRES_IN = 3600
EFFECTIVE_TTL = DEFAULT_EXPIRES_IN - TOKEN_REFRESH_BUFFER_SECONDS


def make_token_response(token: str = "test-token", expires_in: int = DEFAULT_EXPIRES_IN):
    response = MagicMock()
    response.json.return_value = {"access_token": token, "expires_in": expires_in}
    return response


def patch_async_client(post_return_value=None, post_side_effect=None):
    """Patches httpx.AsyncClient as used inside ThunderIDClient._fetch_access_token,
    so `async with httpx.AsyncClient(...) as client: await client.post(...)` resolves
    to a controllable mock. Returns (patcher, mock_client) — enter the patcher as a
    context manager, and inspect calls via mock_client.post."""
    mock_client = AsyncMock()
    if post_side_effect is not None:
        mock_client.post.side_effect = post_side_effect
    else:
        mock_client.post.return_value = post_return_value or make_token_response()

    mock_cm = MagicMock()
    mock_cm.__aenter__ = AsyncMock(return_value=mock_client)
    mock_cm.__aexit__ = AsyncMock(return_value=False)

    patcher = patch("utils.thunder_id_client.httpx.AsyncClient", return_value=mock_cm)
    return patcher, mock_client


class FakeClock:
    """A controllable stand-in for time.monotonic()."""

    def __init__(self, start: float = 0.0):
        self.now = start

    def __call__(self):
        return self.now

    def advance(self, seconds: float):
        self.now += seconds


# ---------------------------------------------------------------------------
# _fetch_access_token: the raw HTTP call to Thunder ID
# ---------------------------------------------------------------------------

class TestFetchAccessToken:

    @pytest.mark.asyncio
    async def test_returns_access_token_and_expires_in_from_response(self, client):
        patcher, _ = patch_async_client(
            post_return_value=make_token_response("abc123", expires_in=1800)
        )

        with patcher:
            token, expires_in = await client._fetch_access_token()

        assert token == "abc123"
        assert expires_in == 1800

    @pytest.mark.asyncio
    async def test_missing_expires_in_falls_back_to_default_ttl(self, client):
        """The real Thunder ID response always includes expires_in, but if
        it's ever absent we shouldn't crash — fall back to a sane default
        instead of caching the token with an unknown lifetime."""
        response = MagicMock()
        response.json.return_value = {"access_token": "abc123"}
        patcher, _ = patch_async_client(post_return_value=response)

        with patcher:
            token, expires_in = await client._fetch_access_token()

        assert token == "abc123"
        assert expires_in == DEFAULT_TOKEN_TTL_SECONDS

    @pytest.mark.asyncio
    async def test_posts_with_correct_url_auth_headers_and_payload(self, client, monkeypatch):
        monkeypatch.setattr("utils.thunder_id_client.THUNDER_BASE_URL", "https://thunder.test")
        monkeypatch.setattr("utils.thunder_id_client.THUNDER_CLIENT_ID", "client-id")
        monkeypatch.setattr("utils.thunder_id_client.THUNDER_CLIENT_SECRET", "client-secret")
        monkeypatch.setattr("utils.thunder_id_client.THUNDER_RESOURCE", "https://resource.test")
        patcher, mock_client = patch_async_client()

        with patcher:
            await client._fetch_access_token()

        mock_client.post.assert_awaited_once()
        args, kwargs = mock_client.post.call_args
        assert args[0] == "https://thunder.test/oauth2/token"
        assert kwargs["auth"] == ("client-id", "client-secret")
        assert kwargs["headers"] == {"Content-Type": "application/x-www-form-urlencoded"}
        assert kwargs["data"] == {
            "grant_type": "client_credentials",
            "scope": (
                "crawler:runs crawler:complete crawler:lookup "
                "crawler:batch-save crawler:batch-update crawler:reconcile"
            ),
            "resource": "https://resource.test",
        }

    @pytest.mark.asyncio
    async def test_passes_verify_tls_flag_through_to_async_client(self, client, monkeypatch):
        monkeypatch.setattr("utils.thunder_id_client.THUNDER_VERIFY_TLS", False)
        patcher, _ = patch_async_client()

        with patcher as mock_async_client_cls:
            await client._fetch_access_token()

        mock_async_client_cls.assert_called_once_with(verify=False)

    @pytest.mark.asyncio
    async def test_malformed_response_missing_access_token_raises(self, client):
        """A 200 with no 'access_token' field must fail loudly and clearly —
        not with a None token that gets cached and silently sent as
        'Bearer None' on every subsequent request."""
        bad_response = MagicMock()
        bad_response.json.return_value = {"unexpected": "shape"}
        patcher, _ = patch_async_client(post_return_value=bad_response)

        with patcher:
            with pytest.raises(ValueError, match="access_token"):
                await client._fetch_access_token()

    @pytest.mark.asyncio
    async def test_non_2xx_status_raises_before_reading_body(self, client):
        """A rejected request (bad credentials, 5xx, ...) must surface as an
        HTTPStatusError with the real status/body, not a confusing KeyError
        from trying to parse an error payload as a token response."""
        error_response = MagicMock(spec=httpx.Response)
        error_response.status_code = 401
        error_response.text = "invalid_client"
        request = httpx.Request("POST", "https://thunder.test/oauth2/token")
        error_response.raise_for_status.side_effect = httpx.HTTPStatusError(
            "401 error", request=request, response=error_response
        )
        patcher, _ = patch_async_client(post_return_value=error_response)

        with patcher:
            with pytest.raises(httpx.HTTPStatusError):
                await client._fetch_access_token()

    @pytest.mark.asyncio
    async def test_network_error_propagates(self, client):
        patcher, _ = patch_async_client(post_side_effect=httpx.ConnectError("refused"))

        with patcher:
            with pytest.raises(httpx.ConnectError):
                await client._fetch_access_token()


# ---------------------------------------------------------------------------
# _is_cached_token_valid: the cache-freshness check
# ---------------------------------------------------------------------------

class TestIsCachedTokenValid:

    def test_false_when_nothing_cached_yet(self, client):
        assert client._is_cached_token_valid() is False

    def test_true_when_token_cached_and_not_expired(self, client):
        clock = FakeClock(start=50.0)
        client._cached_token = "tok"
        client._cached_token_expiry = 100.0

        with patch("utils.thunder_id_client.time.monotonic", clock):
            assert client._is_cached_token_valid() is True

    def test_false_exactly_at_expiry_boundary(self, client):
        """Expiry check is a strict '<', so the exact expiry instant already
        counts as stale — avoids handing out a token the instant it dies."""
        clock = FakeClock(start=100.0)
        client._cached_token = "tok"
        client._cached_token_expiry = 100.0

        with patch("utils.thunder_id_client.time.monotonic", clock):
            assert client._is_cached_token_valid() is False

    def test_false_after_expiry_has_passed(self, client):
        clock = FakeClock(start=150.0)
        client._cached_token = "tok"
        client._cached_token_expiry = 100.0

        with patch("utils.thunder_id_client.time.monotonic", clock):
            assert client._is_cached_token_valid() is False


# ---------------------------------------------------------------------------
# invalidate_token: forces the next get_access_token() to refetch
# ---------------------------------------------------------------------------

class TestInvalidateToken:

    def test_clears_cached_token_and_expiry_when_token_matches(self, client):
        client._cached_token = "stale-token"
        client._cached_token_expiry = 9999999.0

        client.invalidate_token("stale-token")

        assert client._cached_token is None
        assert client._is_cached_token_valid() is False

    def test_does_not_clear_a_newer_token_that_no_longer_matches(self, client):
        """Race-safety: a caller holding a since-superseded token must not be
        able to wipe out a token someone else already refreshed to. Only an
        invalidate_token() call naming the *current* cached token takes
        effect."""
        client._cached_token = "fresh-token"
        client._cached_token_expiry = 9999999.0

        client.invalidate_token("stale-token")

        assert client._cached_token == "fresh-token"
        assert client._is_cached_token_valid() is True

    @pytest.mark.asyncio
    async def test_next_get_access_token_call_fetches_a_new_token(self, client):
        """The scenario this exists for: the backend rejects a cached token
        with a 401 well before our local TTL thinks it's expired — the next
        get_access_token() call must not just hand back the same stale
        token."""
        clock = FakeClock(start=0.0)
        patcher, mock_client = patch_async_client(
            post_side_effect=[
                make_token_response("stale-token"),
                make_token_response("fresh-token"),
            ]
        )

        with patcher, patch("utils.thunder_id_client.time.monotonic", clock):
            first = await client.get_access_token()
            client.invalidate_token(first)
            second = await client.get_access_token()

        assert first == "stale-token"
        assert second == "fresh-token"
        assert mock_client.post.await_count == 2


# ---------------------------------------------------------------------------
# get_access_token: caching + refresh behavior
# ---------------------------------------------------------------------------

class TestGetAccessTokenCaching:

    @pytest.mark.asyncio
    async def test_fetches_and_caches_token_on_first_call(self, client):
        clock = FakeClock(start=0.0)
        patcher, mock_client = patch_async_client(
            post_return_value=make_token_response("first-token")
        )

        with patcher, patch("utils.thunder_id_client.time.monotonic", clock):
            token = await client.get_access_token()

        assert token == "first-token"
        assert client._cached_token == "first-token"
        assert client._cached_token_expiry == EFFECTIVE_TTL
        mock_client.post.assert_awaited_once()

    @pytest.mark.asyncio
    async def test_cache_expiry_is_derived_from_the_response_expires_in(self, client):
        """The whole point of reading expires_in: a token with a shorter (or
        longer) lifetime than the usual 3600s must be cached for exactly
        that long, not a hardcoded duration."""
        clock = FakeClock(start=0.0)
        patcher, _ = patch_async_client(
            post_return_value=make_token_response("short-lived", expires_in=120)
        )

        with patcher, patch("utils.thunder_id_client.time.monotonic", clock):
            await client.get_access_token()

        assert client._cached_token_expiry == 120 - TOKEN_REFRESH_BUFFER_SECONDS

    @pytest.mark.asyncio
    async def test_expires_in_shorter_than_buffer_clamps_to_zero(self, client):
        """A token whose expires_in is smaller than our refresh buffer must
        not produce a negative expiry — it should just be treated as due
        for an immediate refresh on the very next call."""
        clock = FakeClock(start=0.0)
        patcher, _ = patch_async_client(
            post_return_value=make_token_response("very-short-lived", expires_in=30)
        )

        with patcher, patch("utils.thunder_id_client.time.monotonic", clock):
            await client.get_access_token()

        assert client._cached_token_expiry == 0

    @pytest.mark.asyncio
    async def test_returns_cached_token_without_refetching_while_valid(self, client):
        clock = FakeClock(start=0.0)
        patcher, mock_client = patch_async_client(
            post_return_value=make_token_response("first-token")
        )

        with patcher, patch("utils.thunder_id_client.time.monotonic", clock):
            await client.get_access_token()
            clock.advance(EFFECTIVE_TTL - 1)
            token = await client.get_access_token()

        assert token == "first-token"
        mock_client.post.assert_awaited_once()

    @pytest.mark.asyncio
    async def test_refetches_once_cached_token_has_expired(self, client):
        clock = FakeClock(start=0.0)
        patcher, mock_client = patch_async_client(
            post_side_effect=[
                make_token_response("first-token"),
                make_token_response("second-token"),
            ]
        )

        with patcher, patch("utils.thunder_id_client.time.monotonic", clock):
            first = await client.get_access_token()
            clock.advance(EFFECTIVE_TTL + 1)
            second = await client.get_access_token()

        assert first == "first-token"
        assert second == "second-token"
        assert mock_client.post.await_count == 2

    @pytest.mark.asyncio
    async def test_failed_fetch_is_not_cached_and_propagates(self, client):
        patcher, _ = patch_async_client(post_side_effect=httpx.ConnectError("refused"))

        with patcher:
            with pytest.raises(httpx.ConnectError):
                await client.get_access_token()

        assert client._cached_token is None

    @pytest.mark.asyncio
    async def test_two_instances_do_not_share_cached_token(self, client, another_client):
        """Caching is intentionally instance-level, not class-level — the
        shared-across-crawlers benefit comes from CrawlerManager handing out
        one ThunderIDClient instance, not from global state on the class."""
        patcher, _ = patch_async_client(post_return_value=make_token_response("tok-a"))

        with patcher:
            await client.get_access_token()

        assert client._cached_token == "tok-a"
        assert another_client._cached_token is None


class TestGetAccessTokenConcurrency:

    @pytest.mark.asyncio
    async def test_concurrent_calls_share_a_single_fetch(self, client):
        """Guards against a token stampede: CrawlerManager shares one
        ThunderIDClient across crawlers that can run concurrently, so two
        callers racing in with no cached token yet must not both hit the
        Thunder ID endpoint — the lock should make the second one wait for
        the first's result instead."""

        async def slow_post(*args, **kwargs):
            await asyncio.sleep(0)  # yield control so both callers race in
            return make_token_response("shared-token")

        patcher, mock_client = patch_async_client(post_side_effect=slow_post)

        with patcher:
            token_a, token_b = await asyncio.gather(
                client.get_access_token(), client.get_access_token()
            )

        assert token_a == "shared-token"
        assert token_b == "shared-token"
        mock_client.post.assert_awaited_once()


# ---------------------------------------------------------------------------
# ThunderAuth: attaches the Bearer token and retries once on a 401
#
# These drive a real httpx.AsyncClient over an httpx.MockTransport, rather
# than calling async_auth_flow() by hand, so the test exercises httpx's
# actual generator-based auth protocol (the thing that decides whether a
# second yielded request really gets sent) instead of just our own
# assumptions about it.
# ---------------------------------------------------------------------------

def make_thunder_client_mock(token: str = "test-token"):
    mock = MagicMock()
    mock.get_access_token = AsyncMock(return_value=token)
    return mock


class TestThunderAuth:

    @pytest.mark.asyncio
    async def test_attaches_bearer_token_to_the_request(self):
        thunder_client = make_thunder_client_mock("good-token")
        seen_requests = []

        async def handler(request):
            seen_requests.append(request)
            return httpx.Response(200, json={"ok": True})

        async with httpx.AsyncClient(transport=httpx.MockTransport(handler)) as http_client:
            response = await http_client.get(
                "https://backend.test/thing", auth=ThunderAuth(thunder_client)
            )

        assert response.status_code == 200
        assert seen_requests[0].headers["Authorization"] == "Bearer good-token"

    @pytest.mark.asyncio
    async def test_401_invalidates_the_stale_token_and_retries_once(self):
        """A 401 means the token was bad, not the request — it must
        invalidate exactly the token that was rejected, fetch a fresh one,
        and retry with that, rather than giving up."""
        thunder_client = make_thunder_client_mock()
        thunder_client.get_access_token.side_effect = ["stale-token", "fresh-token"]
        seen_auth_headers = []

        async def handler(request):
            # Snapshot the header value now — async_auth_flow mutates and
            # re-yields the SAME Request object on retry, so storing the
            # object itself would make every entry reflect its final state.
            seen_auth_headers.append(request.headers["Authorization"])
            if len(seen_auth_headers) == 1:
                return httpx.Response(401, json={"error": "invalid_token"})
            return httpx.Response(200, json={"ok": True})

        async with httpx.AsyncClient(transport=httpx.MockTransport(handler)) as http_client:
            response = await http_client.get(
                "https://backend.test/thing", auth=ThunderAuth(thunder_client)
            )

        assert response.status_code == 200
        assert seen_auth_headers == ["Bearer stale-token", "Bearer fresh-token"]
        thunder_client.invalidate_token.assert_called_once_with("stale-token")
        assert thunder_client.get_access_token.await_count == 2

    @pytest.mark.asyncio
    async def test_gives_up_after_one_retry_if_still_401(self):
        """Only one retry is attempted — a credential that's genuinely
        invalid (not just stale) must not loop forever."""
        thunder_client = make_thunder_client_mock("bad-token")
        call_count = 0

        async def handler(request):
            nonlocal call_count
            call_count += 1
            return httpx.Response(401, json={"error": "invalid_token"})

        async with httpx.AsyncClient(transport=httpx.MockTransport(handler)) as http_client:
            response = await http_client.get(
                "https://backend.test/thing", auth=ThunderAuth(thunder_client)
            )

        assert response.status_code == 401
        assert call_count == 2
        thunder_client.invalidate_token.assert_called_once_with("bad-token")

    @pytest.mark.asyncio
    async def test_does_not_retry_on_non_401_errors(self):
        thunder_client = make_thunder_client_mock()
        call_count = 0

        async def handler(request):
            nonlocal call_count
            call_count += 1
            return httpx.Response(500)

        async with httpx.AsyncClient(transport=httpx.MockTransport(handler)) as http_client:
            response = await http_client.get(
                "https://backend.test/thing", auth=ThunderAuth(thunder_client)
            )

        assert response.status_code == 500
        assert call_count == 1
        thunder_client.invalidate_token.assert_not_called()

    @pytest.mark.asyncio
    async def test_token_fetch_failure_raises_thunder_token_error(self):
        """If we can't get a token at all, the request must never even reach
        the transport, and the failure must be a ThunderTokenError — not an
        arbitrary exception a caller might not know to catch."""
        thunder_client = MagicMock()
        thunder_client.get_access_token = AsyncMock(
            side_effect=httpx.ConnectError("refused")
        )

        async def handler(request):
            raise AssertionError("should never reach the transport")

        async with httpx.AsyncClient(transport=httpx.MockTransport(handler)) as http_client:
            with pytest.raises(ThunderTokenError):
                await http_client.get(
                    "https://backend.test/thing", auth=ThunderAuth(thunder_client)
                )
