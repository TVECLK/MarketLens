import json
from unittest.mock import AsyncMock, MagicMock

import httpx
import pytest

from crawlers.base_crawler import MAX_RETRIES, RETRYABLE_STATUS_CODES
from models.raw_job import RawJobInput
from utils.thunder_id_client import ThunderTokenError


def make_job(job_id: str = "job-1", crawler_run_id: int = 1) -> RawJobInput:
    return RawJobInput(
        job_id=job_id,
        employer="Acme",
        job_role="Engineer",
        location="Colombo",
        description="desc",
        crawler_run_id=crawler_run_id,
        source="ikman",
    )


def make_ok_response(failed: list | None = None):
    """A 200 response reporting the given failures (job_id/status_code/error
    dicts). An empty/None list means every job in the request succeeded."""
    response = MagicMock(spec=httpx.Response)
    response.status_code = 200
    response.raise_for_status.return_value = None
    response.json.return_value = {"failed_jobs": failed or []}
    return response


def make_bad_json_response():
    """A 200 whose body isn't valid JSON (e.g. a proxy error page)."""
    response = MagicMock(spec=httpx.Response)
    response.status_code = 200
    response.raise_for_status.return_value = None
    response.json.side_effect = json.JSONDecodeError("bad json", "", 0)
    return response


def make_error_response(status_code: int, text: str = "error"):
    """A response whose raise_for_status() raises, mirroring a real
    non-2xx httpx.Response."""
    response = MagicMock(spec=httpx.Response)
    response.status_code = status_code
    response.text = text
    request = httpx.Request("POST", "https://api.example.com/jobs/batch-save")
    response.raise_for_status.side_effect = httpx.HTTPStatusError(
        f"{status_code} error", request=request, response=response
    )
    return response


def failure(job_id: str, status_code: int = 503, error: str = "error"):
    return {"job_id": job_id, "status_code": status_code, "error": error}


# ---------------------------------------------------------------------------
# All jobs saved successfully — single attempt, no retries needed
# ---------------------------------------------------------------------------


class TestAllSucceeded:
    @pytest.mark.asyncio
    async def test_empty_failed_jobs_list_clears_batch_in_one_call(
        self, crawler, thunder_auth
    ):
        job_batch = [make_job("a"), make_job("b")]
        client = AsyncMock()
        client.post.return_value = make_ok_response()

        await crawler._flush_batch(client, thunder_auth, job_batch)

        assert job_batch == []
        assert client.post.await_count == 1

    @pytest.mark.asyncio
    async def test_missing_failed_jobs_key_clears_batch(self, crawler, thunder_auth):
        """The backend contract says absence of the key means no failures,
        same as an explicit empty list — .get() must treat them the same."""
        job_batch = [make_job("a")]
        client = AsyncMock()
        response = MagicMock(spec=httpx.Response)
        response.status_code = 200
        response.raise_for_status.return_value = None
        response.json.return_value = {"message": "ok"}  # no failed_jobs key
        client.post.return_value = response

        await crawler._flush_batch(client, thunder_auth, job_batch)

        assert job_batch == []
        assert client.post.await_count == 1

    @pytest.mark.asyncio
    async def test_sends_correct_payload_shape(self, crawler, thunder_auth):
        """retry_count must never be sent to the backend. Auth is handled by
        ThunderAuth itself now, so we only check it's passed through as
        `auth=`, not that any header content is built here."""
        job = make_job("a")
        job_batch = [job]
        client = AsyncMock()
        client.post.return_value = make_ok_response()

        await crawler._flush_batch(client, thunder_auth, job_batch)

        client.post.assert_awaited_once()
        _, kwargs = client.post.call_args
        assert kwargs["auth"] is thunder_auth
        sent_payload = kwargs["json"][0]
        assert isinstance(RawJobInput.model_validate(sent_payload), RawJobInput)
        assert sent_payload["job_id"] == "a"


# ---------------------------------------------------------------------------
# Internal retries: a failure on attempt N is retried within the SAME call
# ---------------------------------------------------------------------------


class TestInternalRetrySucceeds:
    @pytest.mark.asyncio
    async def test_retryable_failure_recovers_on_second_attempt(
        self, crawler, thunder_auth
    ):
        """A job that fails with a retryable status on attempt 1 and
        succeeds on attempt 2 should end up saved, with no external call
        needed — both attempts happen inside this one _flush_batch call."""
        job = make_job("a")
        job_batch = [job]
        client = AsyncMock()
        client.post.side_effect = [
            make_ok_response([failure("a", 503)]),
            make_ok_response([]),  # succeeds on retry
        ]

        await crawler._flush_batch(client, thunder_auth, job_batch)

        assert job_batch == []
        assert client.post.await_count == 2

    @pytest.mark.asyncio
    async def test_shrinking_batch_matches_the_documented_scenario(
        self, crawler, thunder_auth
    ):
        """Mirrors the exact walkthrough this design was built for: send
        10, 5 fail -> retry those 5, 3 fail -> retry those 3, and confirm
        each retry only re-sends the jobs still outstanding, not the ones
        that already succeeded."""
        jobs = [make_job(f"job-{i}") for i in range(10)]
        job_batch = list(jobs)
        client = AsyncMock()
        client.post.side_effect = [
            make_ok_response([failure(f"job-{i}") for i in range(5)]),  # 10 -> 5 fail
            make_ok_response([failure(f"job-{i}") for i in range(3)]),  # 5 -> 3 fail
            make_ok_response([]),  # 3 -> all succeed
        ]

        await crawler._flush_batch(client, thunder_auth, job_batch)

        assert job_batch == []
        assert client.post.await_count == 3

        sent_ids_per_call = [
            [j["job_id"] for j in call.kwargs["json"]]
            for call in client.post.call_args_list
        ]
        assert len(sent_ids_per_call[0]) == 10
        assert sent_ids_per_call[1] == [f"job-{i}" for i in range(5)]
        assert sent_ids_per_call[2] == [f"job-{i}" for i in range(3)]

    @pytest.mark.asyncio
    async def test_jobs_that_already_succeeded_are_not_resent_on_retry(
        self, crawler, thunder_auth
    ):
        saved = make_job("saved")
        retryable = make_job("retryable")
        job_batch = [saved, retryable]
        client = AsyncMock()
        client.post.side_effect = [
            make_ok_response([failure("retryable", 503)]),
            make_ok_response([]),
        ]

        await crawler._flush_batch(client, thunder_auth, job_batch)

        second_call_ids = [
            j["job_id"] for j in client.post.call_args_list[1].kwargs["json"]
        ]
        assert second_call_ids == ["retryable"]
        assert job_batch == []


# ---------------------------------------------------------------------------
# Non-retryable failures: dropped immediately, no extra attempts spent
# ---------------------------------------------------------------------------


class TestNonRetryableFailures:
    @pytest.mark.asyncio
    async def test_non_retryable_failure_dropped_without_retrying(
        self, crawler, thunder_auth
    ):
        job_batch = [make_job("a")]
        client = AsyncMock()
        client.post.return_value = make_ok_response([failure("a", 422, "bad province")])

        await crawler._flush_batch(client, thunder_auth, job_batch)

        assert job_batch == []
        assert client.post.await_count == 1  # no point retrying a 422

    @pytest.mark.asyncio
    async def test_mixed_batch_only_retries_the_retryable_one(
        self, crawler, thunder_auth
    ):
        saved = make_job("saved")
        retryable = make_job("retryable")
        permanent = make_job("permanent")
        job_batch = [saved, retryable, permanent]

        client = AsyncMock()
        client.post.side_effect = [
            make_ok_response([failure("retryable", 503), failure("permanent", 422)]),
            make_ok_response([]),  # retryable succeeds on 2nd attempt
        ]

        await crawler._flush_batch(client, thunder_auth, job_batch)

        assert job_batch == []
        assert client.post.await_count == 2
        second_call_ids = [
            j["job_id"] for j in client.post.call_args_list[1].kwargs["json"]
        ]
        assert second_call_ids == ["retryable"]  # permanent was dropped, not retried

    @pytest.mark.asyncio
    async def test_matching_is_by_job_id_not_position(self, crawler, thunder_auth):
        """Guards against a regression back to positional zip() matching:
        the failure list order doesn't match job_batch order, and the
        right job must still be the one retried."""
        first = make_job("first")
        second = make_job("second")
        job_batch = [first, second]

        client = AsyncMock()
        client.post.side_effect = [
            make_ok_response([failure("second", 503)]),
            make_ok_response([]),
        ]

        await crawler._flush_batch(client, thunder_auth, job_batch)

        second_call_ids = [
            j["job_id"] for j in client.post.call_args_list[1].kwargs["json"]
        ]
        assert second_call_ids == ["second"]
        assert job_batch == []

    @pytest.mark.asyncio
    async def test_unknown_status_code_treated_as_non_retryable(
        self, crawler, thunder_auth
    ):
        """Any status code outside RETRYABLE_STATUS_CODES should be dropped
        immediately, not just 422 specifically — the safer default for a
        code the crawler doesn't recognize."""
        assert 418 not in RETRYABLE_STATUS_CODES  # sanity check on the fixture itself
        job_batch = [make_job("a")]
        client = AsyncMock()
        client.post.return_value = make_ok_response([failure("a", 418, "teapot")])

        await crawler._flush_batch(client, thunder_auth, job_batch)

        assert job_batch == []
        assert client.post.await_count == 1

    @pytest.mark.asyncio
    async def test_429_rate_limit_is_treated_as_retryable(self, crawler, thunder_auth):
        """429 is in RETRYABLE_STATUS_CODES — a rate-limited job is the
        backend's problem, not the job's, so it should get another chance
        rather than being dropped."""
        job_batch = [make_job("a")]
        client = AsyncMock()
        client.post.side_effect = [
            make_ok_response([failure("a", 429, "rate limited")]),
            make_ok_response([]),  # succeeds once the limit clears
        ]

        await crawler._flush_batch(client, thunder_auth, job_batch)

        assert client.post.await_count == 2
        assert job_batch == []

    @pytest.mark.asyncio
    async def test_408_request_timeout_is_treated_as_retryable(self, crawler, thunder_auth):
        """408 is in RETRYABLE_STATUS_CODES — a job failing because the
        request timed out server-side should get another chance rather
        than being dropped."""
        job_batch = [make_job("a")]
        client = AsyncMock()
        client.post.side_effect = [
            make_ok_response([failure("a", 408, "request timeout")]),
            make_ok_response([]),  # succeeds on retry
        ]

        await crawler._flush_batch(client, thunder_auth, job_batch)

        assert client.post.await_count == 2
        assert job_batch == []


# ---------------------------------------------------------------------------
# Retry exhaustion (poison-pill protection) — all within one call now
# ---------------------------------------------------------------------------


class TestRetryExhaustion:
    @pytest.mark.asyncio
    async def test_job_dropped_after_max_retries_and_batch_still_clears(
        self, crawler, thunder_auth
    ):
        """A job that fails on every attempt should be tried exactly
        MAX_RETRIES times, then dropped and logged — job_batch must still
        end up empty rather than holding onto it forever."""
        job_batch = [make_job("a")]
        client = AsyncMock()
        client.post.return_value = make_ok_response([failure("a", 503, "still down")])

        await crawler._flush_batch(client, thunder_auth, job_batch)

        assert job_batch == []
        assert client.post.await_count == MAX_RETRIES

    @pytest.mark.asyncio
    async def test_does_not_exceed_max_retries_attempts(self, crawler, thunder_auth):
        """Even if the backend would keep failing forever, _flush_batch
        must not call the backend more than MAX_RETRIES times in one go."""
        job_batch = [make_job("a")]
        client = AsyncMock()
        client.post.return_value = make_ok_response([failure("a", 503)])

        await crawler._flush_batch(client, thunder_auth, job_batch)

        assert client.post.await_count == MAX_RETRIES
        assert job_batch == []

    @pytest.mark.asyncio
    async def test_logs_the_dropped_job_ids_on_final_giveup(
        self, crawler, thunder_auth, caplog
    ):
        job_batch = [make_job("stubborn-job")]
        client = AsyncMock()
        client.post.return_value = make_ok_response([failure("stubborn-job", 503)])

        with caplog.at_level("ERROR"):
            await crawler._flush_batch(client, thunder_auth, job_batch)

        assert any("stubborn-job" in record.message for record in caplog.records)


# ---------------------------------------------------------------------------
# Transport-level failures (no usable response at all)
# ---------------------------------------------------------------------------


class TestTransportFailures:
    @pytest.mark.asyncio
    async def test_timeout_is_retried_internally_then_dropped(
        self, crawler, thunder_auth
    ):
        """A persistent timeout should consume retry attempts like any
        other retryable failure, and the batch must still end up empty
        rather than being left buffered forever for an external caller."""
        job_batch = [make_job("a"), make_job("b")]
        client = AsyncMock()
        client.post.side_effect = httpx.TimeoutException("timed out")

        await crawler._flush_batch(client, thunder_auth, job_batch)

        assert client.post.await_count == MAX_RETRIES
        assert job_batch == []

    @pytest.mark.asyncio
    async def test_timeout_recovers_if_a_later_attempt_succeeds(
        self, crawler, thunder_auth
    ):
        job_batch = [make_job("a")]
        client = AsyncMock()
        client.post.side_effect = [
            httpx.TimeoutException("timed out"),
            make_ok_response([]),
        ]

        await crawler._flush_batch(client, thunder_auth, job_batch)

        assert client.post.await_count == 2
        assert job_batch == []

    @pytest.mark.asyncio
    async def test_persistent_token_fetch_failure_is_retried_then_dropped(
        self, crawler, thunder_auth
    ):
        """ThunderAuth raises ThunderTokenError (via the mocked client.post,
        standing in for the real auth flow) when it can't obtain a token at
        all — e.g. Thunder ID itself is down. That should be treated like
        any other transient transport failure: consume retry attempts, then
        give up gracefully, not crash out of _flush_batch unhandled."""
        job_batch = [make_job("a"), make_job("b")]
        client = AsyncMock()
        client.post.side_effect = ThunderTokenError("Thunder ID unreachable")

        await crawler._flush_batch(client, thunder_auth, job_batch)

        assert client.post.await_count == MAX_RETRIES
        assert job_batch == []

    @pytest.mark.asyncio
    async def test_token_fetch_failure_recovers_if_a_later_attempt_succeeds(
        self, crawler, thunder_auth
    ):
        job_batch = [make_job("a")]
        client = AsyncMock()
        client.post.side_effect = [
            ThunderTokenError("Thunder ID unreachable"),
            make_ok_response([]),
        ]

        await crawler._flush_batch(client, thunder_auth, job_batch)

        assert client.post.await_count == 2
        assert job_batch == []

    @pytest.mark.asyncio
    async def test_connection_error_is_retried_internally_then_dropped(
        self, crawler, thunder_auth
    ):
        job_batch = [make_job("a")]
        client = AsyncMock()
        client.post.side_effect = httpx.ConnectError("connection refused")

        await crawler._flush_batch(client, thunder_auth, job_batch)

        assert client.post.await_count == MAX_RETRIES
        assert job_batch == []

    @pytest.mark.asyncio
    async def test_server_error_status_is_retried_internally_then_dropped(
        self, crawler, thunder_auth
    ):
        job_batch = [make_job("a"), make_job("b")]
        client = AsyncMock()
        client.post.return_value = make_error_response(503, text="upstream down")

        await crawler._flush_batch(client, thunder_auth, job_batch)

        assert client.post.await_count == MAX_RETRIES
        assert job_batch == []

    @pytest.mark.asyncio
    async def test_server_error_recovers_if_a_later_attempt_succeeds(
        self, crawler, thunder_auth
    ):
        job_batch = [make_job("a")]
        client = AsyncMock()
        client.post.side_effect = [
            make_error_response(503),
            make_ok_response([]),
        ]

        await crawler._flush_batch(client, thunder_auth, job_batch)

        assert client.post.await_count == 2
        assert job_batch == []

    @pytest.mark.asyncio
    async def test_client_error_status_drops_entire_batch_on_first_attempt(
        self, crawler, thunder_auth
    ):
        """A 4xx at the whole-request level (malformed array, bad resource,
        or even a 401) is not a per-job failure — the identical request
        would fail again, so the whole batch is dropped on the first
        attempt, no retries spent. 401 specifically behaves exactly like any
        other client error at this layer now: ThunderAuth already tries a
        refreshed token once, transparently, before _flush_batch ever sees
        the response — see TestThunderAuth in test_thunder_id_client.py for
        that retry behavior."""
        job_batch = [make_job("a"), make_job("b")]
        client = AsyncMock()
        client.post.return_value = make_error_response(400, text="malformed request")

        await crawler._flush_batch(client, thunder_auth, job_batch)

        assert job_batch == []
        assert client.post.await_count == 1

    @pytest.mark.asyncio
    async def test_malformed_json_body_on_200_is_retried_then_dropped(
        self, crawler, thunder_auth
    ):
        """A 200 with a body that isn't valid JSON (e.g. a proxy error page)
        must not crash the crawl loop. It should be treated as a retryable
        failure, and eventually dropped if it never becomes valid JSON."""
        job_batch = [make_job("a")]
        client = AsyncMock()
        client.post.return_value = make_bad_json_response()

        await crawler._flush_batch(client, thunder_auth, job_batch)

        assert client.post.await_count == MAX_RETRIES
        assert job_batch == []

    @pytest.mark.asyncio
    async def test_malformed_json_recovers_if_a_later_attempt_is_valid(
        self, crawler, thunder_auth
    ):
        job_batch = [make_job("a")]
        client = AsyncMock()
        client.post.side_effect = [
            make_bad_json_response(),
            make_ok_response([]),
        ]

        await crawler._flush_batch(client, thunder_auth, job_batch)

        assert client.post.await_count == 2
        assert job_batch == []

    @pytest.mark.asyncio
    async def test_failed_job_entry_missing_job_id_is_ignored_not_crashed(
        self, crawler, thunder_auth
    ):
        """Defensive contract check: a malformed failed_jobs entry from the
        backend (missing job_id) should be skipped, not raise KeyError.
        Since it can't be matched to any job, that job is treated as if it
        succeeded (not present in the usable failures) and the batch clears
        on the very first attempt."""
        job_batch = [make_job("a")]
        client = AsyncMock()
        client.post.return_value = make_ok_response(
            [{"status_code": 503, "error": "missing job_id field"}]
        )

        await crawler._flush_batch(client, thunder_auth, job_batch)

        assert job_batch == []
        assert client.post.await_count == 1

    @pytest.mark.asyncio
    async def test_whole_request_429_is_retried_internally_then_dropped(
        self, crawler, thunder_auth
    ):
        """A 429 at the whole-request level (the backend itself is
        rate-limiting the batch endpoint, not rejecting a specific job)
        must be retried like a 5xx, not dropped like a generic 4xx."""
        job_batch = [make_job("a"), make_job("b")]
        client = AsyncMock()
        client.post.return_value = make_error_response(429, text="rate limited")

        await crawler._flush_batch(client, thunder_auth, job_batch)

        assert client.post.await_count == MAX_RETRIES
        assert job_batch == []

    @pytest.mark.asyncio
    async def test_whole_request_429_recovers_if_a_later_attempt_succeeds(
        self, crawler, thunder_auth
    ):
        job_batch = [make_job("a")]
        client = AsyncMock()
        client.post.side_effect = [
            make_error_response(429, text="rate limited"),
            make_ok_response([]),
        ]

        await crawler._flush_batch(client, thunder_auth, job_batch)

        assert client.post.await_count == 2
        assert job_batch == []
