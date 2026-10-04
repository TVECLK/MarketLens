from abc import ABC, abstractmethod
import logging
from typing import List

import httpx

from config import BACKEND_BASE_URL
from models.raw_job import RawJobInput
from utils.thunder_id_client import ThunderAuth, ThunderTokenError

logger = logging.getLogger(__name__)

MAX_RETRIES = 3
RETRYABLE_STATUS_CODES = {500, 502, 503, 504, 429, 408}


class BaseJobCrawler(ABC):
    @abstractmethod
    async def crawl_jobs(
        self,
        crawler_run_id: int,
        async_client: httpx.AsyncClient,
        thunder_auth: ThunderAuth,
    ) -> None:
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

        finally:
            job_batch.clear()
