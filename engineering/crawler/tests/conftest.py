from collections import namedtuple
from unittest.mock import AsyncMock, MagicMock

import httpx
import pytest

from utils import ThunderIDClient
from utils.crawler_run_manager import CrawlerManager
from crawlers import BaseJobCrawler
from parsers import BaseJobParser


class _ConcreteCrawler(BaseJobCrawler):
    """BaseJobCrawler is abstract; _flush_batch is what we're testing and
    doesn't need a real crawl_jobs implementation behind it."""

    SOURCE_NAME = "Fake"

    async def crawl_jobs(self, crawler_run_id, async_client, thunder_auth):
        raise NotImplementedError


@pytest.fixture
def crawler():
    return _ConcreteCrawler()


@pytest.fixture
def thunder_auth():
    """A stand-in for the ThunderAuth passed into _flush_batch. Auth is now
    centralized inside ThunderAuth itself (token fetch, header attach, 401
    retry) — _flush_batch just forwards this opaquely as `auth=` on the
    request, so a plain sentinel is enough; it's never called directly."""
    return MagicMock()


@pytest.fixture
def client():
    return ThunderIDClient()


@pytest.fixture
def another_client():
    """A second, independent ThunderIDClient — for tests that must confirm
    two instances don't share cache state."""
    return ThunderIDClient()


# ---------- Crawler run manager ----------

# The manager tests use their own backend URL, so they don't depend on the
# BACKEND_URL of the environment they run in
TEST_BACKEND_URL = "http://backend.test/api/v1/crawler"

# A request the fake backend received
Call = namedtuple("Call", ["path", "body", "auth"])

_REAL_RESULT = object()


class _FakeParser(BaseJobParser):
    def parse_rule_based_fields(self, *args, **kwargs):
        raise NotImplementedError


@pytest.fixture
def backend_url(monkeypatch):
    """Points the manager at TEST_BACKEND_URL and returns it."""
    monkeypatch.setattr("utils.crawler_run_manager.BACKEND_BASE_URL", TEST_BACKEND_URL)
    return TEST_BACKEND_URL


@pytest.fixture
def make_crawler_class():
    """Returns a factory for a BaseJobCrawler whose crawl_jobs raises `error`, or
    returns `result` (by default the CrawlResult a finished crawl returns)."""

    def _make(source_name: str, error: Exception | None = None, result=_REAL_RESULT):
        class _Crawler(BaseJobCrawler):
            SOURCE_NAME = source_name

            def __init__(self):
                super().__init__()
                self._parser = _FakeParser(self.SOURCE_NAME)

            async def crawl_jobs(self, crawler_run_id, async_client, thunder_auth):
                if error is not None:
                    raise error
                if result is not _REAL_RESULT:
                    return result
                return self._finish_crawl(usable_jobs=1)

        return _Crawler

    return _make


@pytest.fixture
def manager(backend_url, make_crawler_class):
    """A CrawlerManager with two crawlers that succeed (Rooster, Ikman) and one
    that raises (TopJobs)."""
    manager = CrawlerManager()
    manager._registry = {
        "rooster": make_crawler_class("Rooster"),
        "ikman": make_crawler_class("Ikman"),
        "topjobs": make_crawler_class("TopJobs", error=RuntimeError("site down")),
    }
    return manager


@pytest.fixture
def make_client(backend_url):
    """Returns a factory for a fake HTTP client that records every POST it
    receives as a Call.

    `behaviors` maps the end of a request path (for example "/jobs/reconcile" or
    "/complete") to an exception to raise, or an httpx.Response to return.
    Any other request gets a 200."""

    def _make(behaviors: dict | None = None):
        behaviors = behaviors or {}
        client = MagicMock()
        client.calls = []

        async def post(url, json=None, auth=None, **kwargs):
            if not url.startswith(backend_url):
                pytest.fail(f"request sent outside the test backend: {url}")
            path = url.removeprefix(backend_url)
            client.calls.append(Call(path, json, auth))
            for suffix, behavior in behaviors.items():
                if path.endswith(suffix):
                    if isinstance(behavior, Exception):
                        raise behavior
                    return behavior
            return httpx.Response(200, request=httpx.Request("POST", url))

        client.post = AsyncMock(side_effect=post)
        return client

    return _make
