import logging
from collections import namedtuple
from unittest.mock import AsyncMock, MagicMock

import httpx
import pytest

import main
from config import BACKEND_BASE_URL
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


def http_response(status_code: int) -> httpx.Response:
    return httpx.Response(
        status_code, request=httpx.Request("POST", "http://backend.test")
    )


Call = namedtuple("Call", ["path", "body", "auth"])


def make_client(behaviors: dict | None = None):
    """A fake HTTP client that records every POST it receives as a Call.

    `behaviors` maps the end of a request path (for example "/jobs/reconcile" or
    "/complete") to an exception to raise, or an httpx.Response to return.
    Any other request gets a 200."""
    behaviors = behaviors or {}
    client = MagicMock()
    client.calls = []

    async def post(url, json=None, auth=None, **kwargs):
        path = url.removeprefix(BACKEND_BASE_URL)
        client.calls.append(Call(path, json, auth))
        for suffix, behavior in behaviors.items():
            if path.endswith(suffix):
                if isinstance(behavior, Exception):
                    raise behavior
                return behavior
        return http_response(200)

    client.post = AsyncMock(side_effect=post)
    return client


def completions(client) -> list:
    """The /complete requests the client received."""
    return [call for call in client.calls if call.path.endswith("/complete")]


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
async def test_run_all_crawlers_returns_only_the_succeeded_sources(manager, concurrent):
    manager._start_run = AsyncMock(return_value=7)
    manager._finalize_run = AsyncMock()

    result = await manager.run_all_crawlers(concurrent=concurrent)

    assert sorted(result) == ["Ikman", "Rooster"]
    manager._finalize_run.assert_awaited_once()
    _, run_id, sources = manager._finalize_run.await_args.args
    assert run_id == 7
    assert sorted(sources) == ["Ikman", "Rooster"]


@pytest.mark.parametrize("concurrent", [True, False])
async def test_run_all_crawlers_returns_an_empty_list_when_all_fail(
    manager, concurrent
):
    manager._registry = {
        "topjobs": make_crawler_class("TopJobs", error=RuntimeError("down")),
        "ikman": make_crawler_class("Ikman", error=ValueError("layout changed")),
    }
    manager._start_run = AsyncMock(return_value=7)
    manager._finalize_run = AsyncMock()

    result = await manager.run_all_crawlers(concurrent=concurrent)

    assert result == []
    _, _, sources = manager._finalize_run.await_args.args
    assert sources == []


async def test_run_all_crawlers_propagates_a_finalize_failure(manager):
    manager._start_run = AsyncMock(return_value=7)
    manager._finalize_run = AsyncMock(side_effect=httpx.ConnectError("backend down"))

    with pytest.raises(httpx.ConnectError, match="backend down"):
        await manager.run_all_crawlers(concurrent=True)


@pytest.mark.parametrize("concurrent", [True, False])
async def test_run_all_crawlers_runs_nothing_when_the_run_cannot_be_started(
    manager, concurrent
):
    manager._start_run = AsyncMock(side_effect=httpx.ConnectError("backend down"))
    manager._run_crawler = AsyncMock()
    manager._finalize_run = AsyncMock()

    with pytest.raises(httpx.ConnectError, match="backend down"):
        await manager.run_all_crawlers(concurrent=concurrent)

    manager._run_crawler.assert_not_awaited()
    manager._finalize_run.assert_not_awaited()


# ---------- _finalize_run ----------


async def test_finalize_run_reconciles_then_completes_when_a_source_succeeded(manager):
    client = make_client()

    await manager._finalize_run(client, 7, ["Rooster", "Ikman"])

    assert [call.path for call in client.calls] == [
        "/jobs/reconcile",
        "/runs/7/complete",
    ]
    assert client.calls[0].body == {
        "crawler_run_id": 7,
        "sources": ["Rooster", "Ikman"],
    }
    assert client.calls[1].body["status"] == "COMPLETED"


async def test_finalize_run_records_failed_without_reconciling_when_no_source_succeeded(
    manager,
):
    client = make_client()

    await manager._finalize_run(client, 7, [])

    assert [call.path for call in client.calls] == ["/runs/7/complete"]
    assert client.calls[0].body["status"] == "FAILED"


@pytest.mark.parametrize(
    "reconcile_failure",
    [
        lambda: httpx.ConnectError("reconcile down"),
        lambda: http_response(500),
        lambda: http_response(503),
    ],
    ids=["network error", "HTTP 500", "HTTP 503"],
)
async def test_finalize_run_marks_the_run_failed_once_when_reconcile_fails(
    manager, reconcile_failure
):
    client = make_client({"/jobs/reconcile": reconcile_failure()})

    with pytest.raises((httpx.ConnectError, httpx.HTTPStatusError)):
        await manager._finalize_run(client, 7, ["Rooster"])

    sent = completions(client)
    assert len(sent) == 1
    assert sent[0].body["status"] == "FAILED"
    assert not any(call.body.get("status") == "COMPLETED" for call in sent)


async def test_finalize_run_raises_the_reconcile_error_when_marking_failed_also_fails(
    manager, caplog
):
    client = make_client(
        {
            "/jobs/reconcile": httpx.ConnectError("reconcile down"),
            "/complete": httpx.ReadTimeout("complete down"),
        }
    )

    with caplog.at_level(logging.ERROR, logger="utils.crawler_run_manager"):
        with pytest.raises(httpx.ConnectError, match="reconcile down"):
            await manager._finalize_run(client, 7, ["Rooster"])

    # the second failure is not raised, but it is logged
    assert "Could not mark crawler run 7 as FAILED" in caplog.text
    assert "complete down" in caplog.text
    assert len(completions(client)) == 1


async def test_finalize_run_raises_when_completing_the_run_fails(manager, caplog):
    client = make_client({"/complete": httpx.ConnectError("complete down")})

    with caplog.at_level(logging.ERROR, logger="utils.crawler_run_manager"):
        with pytest.raises(httpx.ConnectError, match="complete down"):
            await manager._finalize_run(client, 7, ["Rooster"])

    assert "Failed to finalize crawler run 7" in caplog.text
    reconciles = [call for call in client.calls if call.path == "/jobs/reconcile"]
    assert len(reconciles) == 1
    # reconcile already worked, so the run is not retried as FAILED
    assert len(completions(client)) == 1


@pytest.mark.parametrize(
    "succeeded_sources, expected_status",
    [(["Rooster"], "COMPLETED"), ([], "FAILED")],
    ids=["COMPLETED", "FAILED"],
)
async def test_finalize_run_sends_the_expected_complete_request(
    manager, succeeded_sources, expected_status
):
    client = make_client()

    await manager._finalize_run(client, 7, succeeded_sources)

    (sent,) = completions(client)
    assert sent.path == "/runs/7/complete"
    assert sent.body == {"id": 7, "status": expected_status}
    assert sent.auth is manager._thunder_auth


async def test_finalize_run_sends_the_thunder_auth_with_the_reconcile_request(manager):
    client = make_client()

    await manager._finalize_run(client, 7, ["Rooster"])

    assert client.calls[0].auth is manager._thunder_auth


# ---------- main.crawl_job ----------


def patch_manager(monkeypatch, returned=None, error=None):
    """Replaces the CrawlerManager main.py creates with a fake whose
    run_all_crawlers returns `returned` or raises `error`."""
    fake_manager = MagicMock()
    fake_manager.run_all_crawlers = AsyncMock(return_value=returned, side_effect=error)
    monkeypatch.setattr(main, "CrawlerManager", lambda: fake_manager)
    return fake_manager


async def test_crawl_job_does_not_exit_when_a_source_succeeded(monkeypatch):
    fake_manager = patch_manager(monkeypatch, returned=["Rooster"])

    await main.crawl_job()  # a SystemExit here would fail the test

    fake_manager.run_all_crawlers.assert_awaited_once_with(concurrent=True)


async def test_crawl_job_exits_with_1_when_no_source_succeeded(monkeypatch):
    patch_manager(monkeypatch, returned=[])

    with pytest.raises(SystemExit) as exit_info:
        await main.crawl_job()

    assert exit_info.value.code == 1


async def test_crawl_job_exits_with_1_when_the_run_raises(monkeypatch):
    patch_manager(monkeypatch, error=RuntimeError("backend down"))

    with pytest.raises(SystemExit) as exit_info:
        await main.crawl_job()

    assert exit_info.value.code == 1
