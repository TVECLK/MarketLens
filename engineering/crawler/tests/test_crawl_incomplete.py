from unittest.mock import AsyncMock, MagicMock

import httpx
import pytest

from crawlers import (
    BaseJobCrawler,
    CrawlIncompleteError,
    CrawlResult,
    GovernmentJobsCrawler,
    IkmanCrawler,
    RoosterCrawler,
    TopJobsCrawler,
    XpressJobsCrawler,
)
from models import RawJobInput
from parsers import BaseJobParser

ROOSTER_URL = "https://api.rooster.jobs/jobSearch/jobs/search"


class _FakeParser(BaseJobParser):
    def parse_rule_based_fields(self, *args, **kwargs):
        raise NotImplementedError


class _Crawler(BaseJobCrawler):
    SOURCE_NAME = "Fake"

    def __init__(self):
        super().__init__()
        self._parser = _FakeParser(self.SOURCE_NAME)

    async def crawl_jobs(self, crawler_run_id, async_client, thunder_auth):
        raise NotImplementedError


def make_job(job_role="Engineer", description="Builds things") -> RawJobInput:
    return RawJobInput(
        job_role=job_role,
        description=description,
        crawler_run_id=1,
        source="Fake",
    )


# ---------- BaseJobCrawler rules ----------


def test_clean_crawl_with_usable_jobs_does_not_raise():
    _Crawler()._finish_crawl(usable_jobs=3)


def test_finish_crawl_returns_a_result_with_the_crawl_counts():
    crawler = _Crawler()
    crawler._count_jobs(attempted=10, failed=1)

    result = crawler._finish_crawl(usable_jobs=9)

    assert result == CrawlResult(
        source_name="Fake", jobs_attempted=10, jobs_failed=1, usable_jobs=9
    )


def test_zero_usable_jobs_raises():
    with pytest.raises(CrawlIncompleteError, match="no usable jobs"):
        _Crawler()._finish_crawl(usable_jobs=0)


def test_page_failure_raises_even_when_jobs_were_fetched():
    crawler = _Crawler()
    crawler._mark_page_failure("page 7 failed")

    with pytest.raises(CrawlIncompleteError, match="page 7 failed"):
        crawler._finish_crawl(usable_jobs=50)


def test_all_page_failures_are_reported_together():
    crawler = _Crawler()
    crawler._mark_page_failure("page 2 failed")
    crawler._mark_page_failure("page 5 failed")

    with pytest.raises(CrawlIncompleteError) as exc:
        crawler._finish_crawl(usable_jobs=10)

    assert "page 2 failed" in str(exc.value)
    assert "page 5 failed" in str(exc.value)


def test_error_message_names_the_source():
    with pytest.raises(CrawlIncompleteError, match="^Fake:"):
        _Crawler()._finish_crawl(usable_jobs=0)


@pytest.mark.parametrize(
    "job_role, description, expected",
    [
        ("Engineer", "Builds things", True),
        ("", "Builds things", False),
        ("Engineer", "", False),
        ("   ", "   ", False),
    ],
)
def test_is_usable(job_role, description, expected):
    assert BaseJobCrawler._is_usable(make_job(job_role, description)) is expected


def test_crawlers_do_not_share_error_state():
    first, second = _Crawler(), _Crawler()
    first._mark_page_failure("boom")

    second._finish_crawl(usable_jobs=1)


# ---------- Rooster end to end (fake HTTP client) ----------


def rooster_page(count: int, jobs: list[dict]):
    response = MagicMock(spec=httpx.Response)
    response.status_code = 200
    response.raise_for_status.return_value = None
    response.json.return_value = {"body": {"count": count, "data": jobs}}
    return response


def rooster_job(n: int, **overrides) -> dict:
    job = {
        "company_name": f"Acme {n}",
        "title": f"Engineer {n}",
        "location": "Colombo",
        "description": f"Builds thing {n}",
    }
    job.update(overrides)
    return job


def backend_ok():
    response = MagicMock(spec=httpx.Response)
    response.status_code = 200
    response.raise_for_status.return_value = None
    response.json.return_value = {"failed_jobs": []}
    return response


def make_client(rooster_responses: list):
    """Serves the given responses (or raises them, if exceptions) to Rooster
    API calls in order, and accepts every batch-save call."""
    responses = iter(rooster_responses)
    client = MagicMock()
    client.batches = []

    async def post(url, **kwargs):
        if url == ROOSTER_URL:
            item = next(responses)
            if isinstance(item, Exception):
                raise item
            return item
        client.batches.append(kwargs["json"])
        return backend_ok()

    client.post = AsyncMock(side_effect=post)
    return client


@pytest.fixture(autouse=True)
def no_sleep(monkeypatch):
    monkeypatch.setattr("crawlers.rooster_crawler.asyncio.sleep", AsyncMock())


async def crawl_rooster(client):
    return await RoosterCrawler().crawl_jobs(
        crawler_run_id=1, async_client=client, thunder_auth=MagicMock()
    )


async def test_rooster_clean_crawl_succeeds_and_saves_jobs():
    jobs = [rooster_job(1), rooster_job(2)]
    # initial call, then page 1
    client = make_client([rooster_page(2, jobs), rooster_page(2, jobs)])

    result = await crawl_rooster(client)

    assert sum(len(batch) for batch in client.batches) == 2
    assert result == CrawlResult(
        source_name="Rooster", jobs_attempted=2, jobs_failed=0, usable_jobs=2
    )


async def test_rooster_site_down_fails_instead_of_succeeding_with_zero_jobs():
    client = make_client([httpx.ConnectError("site down")])

    with pytest.raises(CrawlIncompleteError):
        await crawl_rooster(client)

    assert client.batches == []


async def test_rooster_every_page_failing_fails():
    client = make_client(
        [rooster_page(40, []), httpx.ReadTimeout("slow"), httpx.ReadTimeout("slow")]
    )

    with pytest.raises(CrawlIncompleteError):
        await crawl_rooster(client)


async def test_rooster_one_failed_page_fails_but_other_pages_are_still_saved():
    page_1_jobs = [rooster_job(1), rooster_job(2)]
    client = make_client(
        [
            rooster_page(40, []),  # initial call: 2 pages
            rooster_page(40, page_1_jobs),  # page 1 ok
            httpx.ReadTimeout("slow"),  # page 2 fails
        ]
    )

    with pytest.raises(CrawlIncompleteError, match="page 2"):
        await crawl_rooster(client)

    assert sum(len(batch) for batch in client.batches) == 2


async def test_rooster_layout_change_with_empty_fields_fails():
    renamed = [{"name": "x"}, {"name": "y"}]  # keys the parser no longer finds
    client = make_client([rooster_page(2, renamed), rooster_page(2, renamed)])

    with pytest.raises(CrawlIncompleteError, match="no usable jobs"):
        await crawl_rooster(client)


async def test_rooster_a_few_empty_jobs_do_not_fail_the_crawl():
    jobs = [rooster_job(1), rooster_job(2, description=""), rooster_job(3)]
    client = make_client([rooster_page(3, jobs), rooster_page(3, jobs)])

    await crawl_rooster(client)

    assert sum(len(batch) for batch in client.batches) == 3


# ---------- per-job failure threshold ----------


@pytest.fixture
def limits(monkeypatch):
    """Pin the thresholds so the tests don't depend on the configured values."""
    monkeypatch.setattr("crawlers.base_crawler.MAX_JOB_FAILURE_RATIO", 0.2)
    monkeypatch.setattr("crawlers.base_crawler.MIN_JOBS_FOR_FAILURE_RATIO", 20)


def test_failures_under_the_ratio_do_not_raise(limits):
    crawler = _Crawler()
    crawler._count_jobs(attempted=100, failed=20)  # exactly 20%: not over the limit

    crawler._finish_crawl(usable_jobs=80)


def test_failures_over_the_ratio_raise(limits):
    crawler = _Crawler()
    crawler._count_jobs(attempted=100, failed=21)

    with pytest.raises(CrawlIncompleteError, match="21 of 100 jobs failed"):
        crawler._finish_crawl(usable_jobs=79)


def test_ratio_is_ignored_below_the_minimum_number_of_jobs(limits):
    crawler = _Crawler()
    crawler._count_jobs(attempted=19, failed=19)

    crawler._finish_crawl(usable_jobs=5)


def test_ratio_applies_from_the_minimum_number_of_jobs(limits):
    crawler = _Crawler()
    crawler._count_jobs(attempted=20, failed=10)

    with pytest.raises(CrawlIncompleteError, match="10 of 20 jobs failed"):
        crawler._finish_crawl(usable_jobs=10)


def test_count_jobs_accumulates_attempts_and_failures():
    crawler = _Crawler()
    crawler._count_jobs()
    crawler._count_jobs(failed=1)
    crawler._count_jobs(attempted=0, failed=2)

    assert (crawler._jobs_attempted, crawler._jobs_failed) == (2, 3)


def make_rejecting_client(rooster_responses: list, rejected: int):
    """Like make_client, but the backend permanently rejects the first
    `rejected` jobs it is sent (non-retryable status)."""
    responses = iter(rooster_responses)
    client = MagicMock()
    remaining = {"rejected": rejected}

    async def post(url, **kwargs):
        if url == ROOSTER_URL:
            return next(responses)
        failed = []
        for job in kwargs["json"]:
            if remaining["rejected"] > 0:
                remaining["rejected"] -= 1
                failed.append(
                    {"job_id": job["job_id"], "status_code": 400, "error": "bad"}
                )
        response = backend_ok()
        response.json.return_value = {"failed_jobs": failed}
        return response

    client.post = AsyncMock(side_effect=post)
    return client


def rooster_jobs(n: int) -> list[dict]:
    return [rooster_job(i) for i in range(n)]


async def test_rooster_few_rejected_jobs_do_not_fail_the_crawl(limits):
    # count=20 means one API page, which holds 25 jobs
    jobs = rooster_jobs(25)
    client = make_rejecting_client([rooster_page(20, jobs), rooster_page(20, jobs)], 2)

    await crawl_rooster(client)  # 2 of 25 = 8%


async def test_rooster_many_rejected_jobs_fail_the_crawl(limits):
    jobs = rooster_jobs(25)
    client = make_rejecting_client([rooster_page(20, jobs), rooster_page(20, jobs)], 8)

    with pytest.raises(CrawlIncompleteError, match="8 of 25 jobs failed"):
        await crawl_rooster(client)  # 8 of 25 = 32%


async def test_rooster_small_crawl_is_not_judged_by_the_ratio(limits):
    jobs = rooster_jobs(10)
    client = make_rejecting_client([rooster_page(10, jobs), rooster_page(10, jobs)], 10)

    await crawl_rooster(client)  # every job rejected, but only 10 attempted


# ---------- source name ----------


class _NoSourceCrawler(BaseJobCrawler):
    async def crawl_jobs(self, crawler_run_id, async_client, thunder_auth):
        raise NotImplementedError


def test_crawler_without_a_source_name_raises():
    with pytest.raises(ValueError, match="_NoSourceCrawler must define"):
        _NoSourceCrawler()


@pytest.mark.parametrize("bad_name", ["", "   ", None, 5])
def test_crawler_with_an_invalid_source_name_raises(bad_name):
    class _BadCrawler(_NoSourceCrawler):
        SOURCE_NAME = bad_name

    with pytest.raises(ValueError, match="SOURCE_NAME"):
        _BadCrawler()


# These names must match the rows in the backend's `source` table
@pytest.mark.parametrize(
    "crawler_class, expected",
    [
        (RoosterCrawler, "Rooster"),
        (IkmanCrawler, "Ikman"),
        (TopJobsCrawler, "TopJobs"),
        (GovernmentJobsCrawler, "GovernmentJobs"),
        (XpressJobsCrawler, "XpressJobs"),
    ],
)
def test_crawlers_pass_their_source_name_to_their_parser(crawler_class, expected):
    crawler = crawler_class()

    assert crawler.source_name == expected
    assert crawler._parser._source_name == expected
