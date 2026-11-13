import logging
from unittest.mock import AsyncMock, MagicMock

import httpx
import pytest

import main
from crawlers.registry import CRAWLER_REGISTRY
from utils.crawler_run_manager import CrawlerManager
from crawlers.source_names import IKMAN, ROOSTER, TOPJOBS


def http_response(status_code: int) -> httpx.Response:
    return httpx.Response(
        status_code, request=httpx.Request("POST", "http://backend.test")
    )


def completions(client) -> list:
    """The /complete requests the client received."""
    return [call for call in client.calls if call.path.endswith("/complete")]


# ---------- _run_crawler ----------


async def test_run_crawler_returns_source_name_on_success(manager):
    result = await manager._run_crawler(ROOSTER, 1, MagicMock())
    assert result == ROOSTER


@pytest.mark.parametrize(
    "returned", [None, ROOSTER, True], ids=["nothing", "a string", "a bool"]
)
async def test_run_crawler_returns_none_when_crawler_returns_no_crawl_result(
    manager, returned, make_crawler_class
):
    manager._registry = {ROOSTER: make_crawler_class(ROOSTER, result=returned)}

    result = await manager._run_crawler(ROOSTER, 1, MagicMock())

    assert result is None


async def test_run_crawler_returns_none_when_crawler_raises(manager):
    result = await manager._run_crawler(TOPJOBS, 1, MagicMock())
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

    assert sorted(result) == [IKMAN, ROOSTER]
    manager._finalize_run.assert_awaited_once()
    _, run_id, sources = manager._finalize_run.await_args.args
    assert run_id == 7
    assert sorted(sources) == [IKMAN, ROOSTER]


@pytest.mark.parametrize("concurrent", [True, False])
async def test_run_all_crawlers_returns_an_empty_list_when_all_fail(
    manager, concurrent, make_crawler_class
):
    manager.fetch_active_crawlers.return_value = {
        TOPJOBS: make_crawler_class(TOPJOBS, error=RuntimeError("down")),
        IKMAN: make_crawler_class(IKMAN, error=ValueError("layout changed")),
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


async def test_run_all_crawlers_runs_only_the_active_crawlers(
    manager, make_crawler_class
):
    manager.fetch_active_crawlers.return_value = {IKMAN: make_crawler_class(IKMAN)}
    manager._start_run = AsyncMock(return_value=7)
    manager._finalize_run = AsyncMock()

    result = await manager.run_all_crawlers(concurrent=True)

    assert result == [IKMAN]


@pytest.mark.parametrize("concurrent", [True, False])
async def test_run_all_crawlers_creates_no_run_when_no_source_is_active(
    manager, concurrent
):
    manager.fetch_active_crawlers.return_value = {}
    manager._start_run = AsyncMock(return_value=7)
    manager._finalize_run = AsyncMock()

    result = await manager.run_all_crawlers(concurrent=concurrent)

    assert result is None
    manager._start_run.assert_not_awaited()
    manager._finalize_run.assert_not_awaited()


async def test_run_all_crawlers_creates_no_run_when_the_sources_cannot_be_fetched(
    manager,
):
    manager.fetch_active_crawlers.side_effect = httpx.ConnectError("backend down")
    manager._start_run = AsyncMock(return_value=7)

    with pytest.raises(httpx.ConnectError, match="backend down"):
        await manager.run_all_crawlers(concurrent=True)

    manager._start_run.assert_not_awaited()


# ---------- fetch_active_crawlers ----------

SOURCES_URL = "http://backend.test/api/v1/crawler/sources"


# The `manager` fixture mocks fetch_active_crawlers, so these use the real method
@pytest.fixture
def real_manager(backend_url):
    return CrawlerManager()


def make_sources_client(names=None, error=None, status_code=200):
    """A fake HTTP client that answers the sources request with `names`, or
    raises `error`."""
    body = {
        "count": len(names or []),
        "sources": [{"id": i, "name": n} for i, n in enumerate(names or [])],
    }
    client = MagicMock()
    client.get = AsyncMock(
        side_effect=error,
        return_value=httpx.Response(
            status_code, json=body, request=httpx.Request("GET", SOURCES_URL)
        ),
    )
    return client


async def test_fetch_active_crawlers_returns_only_the_crawlers_of_active_sources(
    real_manager,
):
    client = make_sources_client([ROOSTER, IKMAN])

    result = await real_manager.fetch_active_crawlers(client)

    assert result == {
        ROOSTER: CRAWLER_REGISTRY[ROOSTER],
        IKMAN: CRAWLER_REGISTRY[IKMAN],
    }


async def test_fetch_active_crawlers_normalizes_case_and_whitespace(real_manager):
    client = make_sources_client([f"  {ROOSTER.title()} ", IKMAN.upper()])

    result = await real_manager.fetch_active_crawlers(client)

    assert set(result) == {ROOSTER, IKMAN}


async def test_fetch_active_crawlers_ignores_active_sources_without_a_crawler(
    real_manager, caplog
):
    client = make_sources_client([ROOSTER, "newspaper"])

    with caplog.at_level(logging.WARNING):
        result = await real_manager.fetch_active_crawlers(client)

    assert set(result) == {ROOSTER}
    assert "newspaper" in caplog.text


async def test_fetch_active_crawlers_returns_nothing_when_no_source_is_active(
    real_manager,
):
    assert await real_manager.fetch_active_crawlers(make_sources_client([])) == {}


async def test_fetch_active_crawlers_requests_the_sources_endpoint_with_auth(
    real_manager,
):
    client = make_sources_client([ROOSTER])

    await real_manager.fetch_active_crawlers(client)

    client.get.assert_awaited_once_with(SOURCES_URL, auth=real_manager._thunder_auth)


async def test_fetch_active_crawlers_raises_when_the_backend_returns_an_error(
    real_manager,
):
    client = make_sources_client(status_code=500)

    with pytest.raises(httpx.HTTPStatusError):
        await real_manager.fetch_active_crawlers(client)


async def test_fetch_active_crawlers_raises_when_the_backend_cannot_be_reached(
    real_manager,
):
    client = make_sources_client(error=httpx.ConnectError("down"))

    with pytest.raises(httpx.ConnectError):
        await real_manager.fetch_active_crawlers(client)


# ---------- _finalize_run ----------


async def test_finalize_run_reconciles_then_completes_when_a_source_succeeded(
    manager, make_client
):
    client = make_client()

    await manager._finalize_run(client, 7, [ROOSTER, IKMAN])

    assert [call.path for call in client.calls] == [
        "/jobs/reconcile",
        "/runs/7/complete",
    ]
    assert client.calls[0].body == {
        "crawler_run_id": 7,
        "sources": [ROOSTER, IKMAN],
    }
    assert client.calls[1].body["status"] == "COMPLETED"


async def test_finalize_run_records_failed_without_reconciling_when_no_source_succeeded(
    manager, make_client
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
    manager, reconcile_failure, make_client
):
    client = make_client({"/jobs/reconcile": reconcile_failure()})

    with pytest.raises((httpx.ConnectError, httpx.HTTPStatusError)):
        await manager._finalize_run(client, 7, [ROOSTER])

    sent = completions(client)
    assert len(sent) == 1
    assert sent[0].body["status"] == "FAILED"
    assert not any(call.body.get("status") == "COMPLETED" for call in sent)


async def test_finalize_run_raises_the_reconcile_error_when_marking_failed_also_fails(
    manager, caplog, make_client
):
    client = make_client(
        {
            "/jobs/reconcile": httpx.ConnectError("reconcile down"),
            "/complete": httpx.ReadTimeout("complete down"),
        }
    )

    with caplog.at_level(logging.ERROR, logger="utils.crawler_run_manager"):
        with pytest.raises(httpx.ConnectError, match="reconcile down"):
            await manager._finalize_run(client, 7, [ROOSTER])

    # the second failure is not raised, but it is logged
    assert "Could not mark crawler run 7 as FAILED" in caplog.text
    assert "complete down" in caplog.text
    assert len(completions(client)) == 1


async def test_finalize_run_raises_when_completing_the_run_fails(
    manager, caplog, make_client
):
    client = make_client({"/complete": httpx.ConnectError("complete down")})

    with caplog.at_level(logging.ERROR, logger="utils.crawler_run_manager"):
        with pytest.raises(httpx.ConnectError, match="complete down"):
            await manager._finalize_run(client, 7, [ROOSTER])

    assert "Failed to finalize crawler run 7" in caplog.text
    reconciles = [call for call in client.calls if call.path == "/jobs/reconcile"]
    assert len(reconciles) == 1
    # reconcile already worked, so the run is not retried as FAILED
    assert len(completions(client)) == 1


@pytest.mark.parametrize(
    "succeeded_sources, expected_status",
    [([ROOSTER], "COMPLETED"), ([], "FAILED")],
    ids=["COMPLETED", "FAILED"],
)
async def test_finalize_run_sends_the_expected_complete_request(
    manager, succeeded_sources, expected_status, make_client
):
    client = make_client()

    await manager._finalize_run(client, 7, succeeded_sources)

    (sent,) = completions(client)
    assert sent.path == "/runs/7/complete"
    assert sent.body == {"id": 7, "status": expected_status}
    assert sent.auth is manager._thunder_auth


async def test_finalize_run_sends_the_thunder_auth_with_the_reconcile_request(
    manager, make_client
):
    client = make_client()

    await manager._finalize_run(client, 7, [ROOSTER])

    assert client.calls[0].auth is manager._thunder_auth


# ---------- main.crawl_job ----------


def patch_manager(monkeypatch, returned=None, error=None):
    """Replaces the CrawlerManager main.py creates with a fake whose
    run_all_crawlers returns `returned` (None means no active source) or raises
    `error`."""
    fake_manager = MagicMock()
    fake_manager.run_all_crawlers = AsyncMock(return_value=returned, side_effect=error)
    monkeypatch.setattr(main, "CrawlerManager", lambda: fake_manager)
    return fake_manager


async def test_crawl_job_does_not_exit_when_a_source_succeeded(monkeypatch):
    fake_manager = patch_manager(monkeypatch, returned=[ROOSTER])

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


async def test_crawl_job_does_not_exit_when_no_source_is_active(monkeypatch):
    patch_manager(monkeypatch, returned=None)

    await main.crawl_job()  # no active source is not a failure
