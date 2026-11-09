from abc import ABC, abstractmethod
from dataclasses import dataclass
import logging
from typing import List

import httpx

from config import (
    BACKEND_BASE_URL,
    MAX_JOB_FAILURE_RATIO,
    MIN_JOBS_FOR_FAILURE_RATIO,
)
from models import RawJobInput
from parsers.base_parser import BaseJobParser
from utils import ThunderAuth, ThunderTokenError

logger = logging.getLogger(__name__)

MAX_RETRIES = 3
RETRYABLE_STATUS_CODES = {500, 502, 503, 504, 429, 408}


class CrawlIncompleteError(Exception):
    """Raised when a crawl can't be trusted to have seen every live job
    (failed pages, too many failed jobs, or no usable jobs). Jobs it did fetch
    are still saved, but its source must not be reconciled for this run."""


@dataclass(frozen=True)
class CrawlResult:
    """What a crawler reports when crawl_jobs finishes without being incomplete.
    The manager only counts a source as successful if it gets one of these."""

    source_name: str
    jobs_attempted: int
    jobs_failed: int
    usable_jobs: int


class BaseJobCrawler(ABC):
    # Every crawler must set this to the source name the backend stores in its
    # `source` table. It is passed on to the crawler's parser.
    SOURCE_NAME: str
    _parser: BaseJobParser

    def __init__(self):
        source_name = getattr(self, "SOURCE_NAME", None)
        if not isinstance(source_name, str) or not source_name.strip():
            raise ValueError(
                f"{type(self).__name__} must define a non-empty SOURCE_NAME"
            )
        self._page_errors: List[str] = []
        self._jobs_attempted = 0
        self._jobs_failed = 0

    # Records a page-level failure (a listing/API page that could not be read).
    # The crawl carries on; _finish_crawl raises once at the end.
    def _mark_page_failure(self, reason: str) -> None:
        self._page_errors.append(reason)
        logger.error(reason)

    # Counts jobs the crawler tried to obtain (attempted) and the ones that could
    # not be obtained or saved (failed). One failed job never fails the crawl:
    # only the failure ratio checked in _finish_crawl does.
    def _count_jobs(self, attempted: int = 1, failed: int = 0) -> None:
        self._jobs_attempted += attempted
        self._jobs_failed += failed

    # A job is usable if it has the fields a real posting always has. Empty
    # fields mean the site layout changed under the parser.
    @staticmethod
    def _is_usable(job: RawJobInput) -> bool:
        return bool(job.job_role.strip() and job.description.strip())

    # Call at the end of crawl_jobs with the number of usable jobs parsed and
    # return its result. A crawl is incomplete on any failed page, when too large
    # a share of the attempted jobs failed, or when no usable jobs were fetched
    # at all; it then raises CrawlIncompleteError instead of returning.
    def _finish_crawl(self, usable_jobs: int) -> CrawlResult:
        reasons = list(self._page_errors)
        if (
            self._jobs_attempted >= MIN_JOBS_FOR_FAILURE_RATIO
            and self._jobs_failed / self._jobs_attempted > MAX_JOB_FAILURE_RATIO
        ):
            reasons.append(
                f"{self._jobs_failed} of {self._jobs_attempted} jobs failed "
                f"(limit {MAX_JOB_FAILURE_RATIO:.0%})"
            )
        if usable_jobs == 0:
            reasons.append("no usable jobs were fetched")
        if reasons:
            raise CrawlIncompleteError(
                f"{self.source_name}: crawl incomplete: {'; '.join(reasons)}"
            )
        return CrawlResult(
            source_name=self.source_name,
            jobs_attempted=self._jobs_attempted,
            jobs_failed=self._jobs_failed,
            usable_jobs=usable_jobs,
        )

    @property
    def source_name(self) -> str:
        return self.SOURCE_NAME

    @abstractmethod
    async def crawl_jobs(
        self,
        crawler_run_id: int,
        async_client: httpx.AsyncClient,
        thunder_auth: ThunderAuth,
    ) -> CrawlResult:
        pass

    async def _flush_batch(
        self,
        async_client: httpx.AsyncClient,
        thunder_auth: ThunderAuth,
        job_batch: List[RawJobInput],
    ) -> None:

        try:
            pending = list(job_batch)
            attempt = 0

            while pending and attempt < MAX_RETRIES:
                attempt += 1

                try:
                    response = await async_client.post(
                        f"{BACKEND_BASE_URL}/jobs/batch-save",
                        json=[job.model_dump() for job in pending],
                        auth=thunder_auth,
                    )
                    response.raise_for_status()
                    body = response.json()

                    failed_jobs = {
                        f["job_id"]: f
                        for f in body.get("failed_jobs", [])
                        if "job_id" in f
                    }

                    if not failed_jobs:
                        pending = []
                        break

                    next_pending = []
                    for job in pending:
                        failure = failed_jobs.get(job.job_id)
                        if failure is None:
                            continue

                        status_code = failure.get("status_code")
                        if status_code and status_code not in RETRYABLE_STATUS_CODES:
                            logger.error(
                                f"Dropping job {job.job_id} — non-retryable status "
                                f"{failure['status_code']}: {failure.get('error')}"
                            )
                            self._count_jobs(attempted=0, failed=1)
                            continue

                        next_pending.append(job)

                    pending = next_pending
                    if pending:
                        logger.info(
                            f"Retrying {len(pending)} job(s) after attempt {attempt}/{MAX_RETRIES}"
                        )

                except ThunderTokenError as e:
                    logger.error(
                        f"Batch POST could not obtain a ThunderID token on attempt "
                        f"{attempt}/{MAX_RETRIES} ({len(pending)} jobs pending): {e}"
                    )
                    continue
                except httpx.TimeoutException as e:
                    logger.error(
                        f"Batch POST timed out on attempt {attempt}/{MAX_RETRIES} "
                        f"({len(pending)} jobs pending): {e}"
                    )
                    continue
                except httpx.RequestError as e:
                    logger.error(
                        f"Batch POST network error on attempt {attempt}/{MAX_RETRIES} "
                        f"({len(pending)} jobs pending): {e}"
                    )
                    continue
                except httpx.HTTPStatusError as e:
                    status = e.response.status_code
                    if status >= 500 or status == 429:
                        logger.error(
                            f"Batch POST rejected by backend with server error {status} "
                            f"on attempt {attempt}/{MAX_RETRIES} ({len(pending)} jobs "
                            f"pending): {e.response.text}"
                        )
                        continue
                    logger.error(
                        f"Batch POST rejected by backend with client error {status} — "
                        f"dropping this batch of {len(pending)} jobs, NOT retrying "
                        f"(same payload would fail again): {e.response.text}"
                    )
                    self._count_jobs(attempted=0, failed=len(pending))
                    pending = []
                    break
                except ValueError as e:
                    logger.error(
                        f"Batch POST returned 200 but body wasn't valid JSON on "
                        f"attempt {attempt}/{MAX_RETRIES} ({len(pending)} jobs pending): {e}"
                    )
                    continue

            if pending:
                logger.error(
                    f"Giving up on {len(pending)} job(s) after {attempt} attempt(s): "
                    f"{[job.job_id for job in pending]}"
                )
                self._count_jobs(attempted=0, failed=len(pending))

        finally:
            job_batch.clear()
