from unittest.mock import AsyncMock, MagicMock

import httpx
import pytest

from crawlers import BaseJobCrawler
from parsers import BaseJobParser
from utils.crawler_run_manager import CrawlerManager


class _FakeParser(BaseJobParser):
    def parse_rule_based_fields(self, *args, **kwargs):
        raise NotImplementedError


_REAL_RESULT = object()


def make_crawler_class(
    source_name: str, error: Exception | None = None, result=_REAL_RESULT
):
    """A BaseJobCrawler whose crawl_jobs raises `error`, or returns `result`
    (by default the CrawlResult a finished crawl returns)."""

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


@pytest.fixture
def manager():
    manager = CrawlerManager()
    manager._registry = {
        "rooster": make_crawler_class("Rooster"),
        "ikman": make_crawler_class("Ikman"),
        "topjobs": make_crawler_class("TopJobs", error=RuntimeError("site down")),
    }
    return manager


def ok_response():
    response = MagicMock(spec=httpx.Response)
    response.raise_for_status.return_value = None
    return response


# ---------- _run_crawler ----------


async def test_run_crawler_returns_source_name_on_success(manager):
    result = await manager._run_crawler("rooster", 1, MagicMock())
    assert result == "Rooster"


@pytest.mark.parametrize(
    "returned", [None, "Rooster", True], ids=["nothing", "a string", "a bool"]
)
async def test_run_crawler_returns_none_when_crawler_returns_no_crawl_result(
    manager, returned
):
    manager._registry = {"rooster": make_crawler_class("Rooster", result=returned)}

    result = await manager._run_crawler("rooster", 1, MagicMock())

    assert result is None


async def test_run_crawler_returns_none_when_crawler_raises(manager):
    result = await manager._run_crawler("topjobs", 1, MagicMock())
    assert result is None


async def test_run_crawler_returns_none_for_unknown_name(manager):
    result = await manager._run_crawler("nope", 1, MagicMock())
    assert result is None


# ---------- run_all_crawlers ----------


@pytest.mark.parametrize("concurrent", [True, False])
async def test_run_all_crawlers_reconciles_only_successful_sources(manager, concurrent):
    manager._start_run = AsyncMock(return_value=7)
    manager._finalize_run = AsyncMock()

    await manager.run_all_crawlers(concurrent=concurrent)

    manager._finalize_run.assert_awaited_once()
    _, run_id, sources = manager._finalize_run.await_args.args
    assert run_id == 7
    assert sorted(sources) == ["Ikman", "Rooster"]


@pytest.mark.parametrize("concurrent", [True, False])
async def test_run_all_crawlers_passes_empty_list_when_all_fail(manager, concurrent):
    manager._registry = {
        "topjobs": make_crawler_class("TopJobs", error=RuntimeError("down")),
        "ikman": make_crawler_class("Ikman", error=ValueError("layout changed")),
    }
    manager._start_run = AsyncMock(return_value=7)
    manager._finalize_run = AsyncMock()

    await manager.run_all_crawlers(concurrent=concurrent)

    _, _, sources = manager._finalize_run.await_args.args
    assert sources == []


# ---------- _finalize_run ----------


async def test_finalize_run_sends_sources_to_reconcile(manager):
    client = MagicMock()
    client.post = AsyncMock(return_value=ok_response())

    await manager._finalize_run(client, 7, ["Rooster", "Ikman"])

    calls = client.post.await_args_list
    assert calls[0].args[0].endswith("/jobs/reconcile")
    assert calls[0].kwargs["json"] == {
        "crawler_run_id": 7,
        "sources": ["Rooster", "Ikman"],
    }
    assert calls[1].args[0].endswith("/runs/7/complete")


async def test_finalize_run_skips_reconcile_when_no_source_succeeded(manager):
    client = MagicMock()
    client.post = AsyncMock(return_value=ok_response())

    await manager._finalize_run(client, 7, [])

    urls = [call.args[0] for call in client.post.await_args_list]
    assert not any(url.endswith("/jobs/reconcile") for url in urls)
    assert urls[-1].endswith("/runs/7/complete")


async def test_finalize_run_reraises_when_reconcile_fails(manager):
    client = MagicMock()
    client.post = AsyncMock(side_effect=httpx.ConnectError("backend down"))

    with pytest.raises(httpx.ConnectError):
        await manager._finalize_run(client, 7, ["Rooster"])
