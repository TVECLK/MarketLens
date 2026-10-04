from unittest.mock import MagicMock

import pytest

from crawlers.base_crawler import BaseJobCrawler
from utils.thunder_id_client import ThunderIDClient


class _ConcreteCrawler(BaseJobCrawler):
    """BaseJobCrawler is abstract; _flush_batch is what we're testing and
    doesn't need a real crawl_jobs implementation behind it."""

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
