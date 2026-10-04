import asyncio
import httpx
import logging
import math
from typing import List
from pydantic import ValidationError
from utils.thunder_id_client import ThunderAuth
from config import BATCH_SIZE

from crawlers.base_crawler import BaseJobCrawler
from parsers.rooster_parser import RoosterParser
from models.raw_job import RawJobInput

logger = logging.getLogger(__name__)


class RoosterCrawler(BaseJobCrawler):
    def __init__(self):
        self._parser = RoosterParser()

    async def _fetch_all_jobs(self, async_client: httpx.AsyncClient):
        base_url = "https://api.rooster.jobs/jobSearch/jobs/search"
        limit = 20
        all_jobs = []

        # Initial call to get total count
        payload = {
            "query": [],
            "limit": limit,
            "page": 1,
            "filters": {"country": "Sri Lanka"},
        }

        try:
            response = await async_client.post(base_url, json=payload)
            response.raise_for_status()
            response_json = response.json()
        except httpx.RequestError as e:
            logger.error(
                f"Rooster: Request failed while fetching initial job page: {e}"
            )
            return all_jobs
        except httpx.HTTPStatusError as e:
            logger.error(
                f"Rooster: Unexpected status {e.response.status_code} while fetching initial job page: {e}"
            )
            return all_jobs
        except ValueError as e:
            logger.error(
                f"Rooster: Failed to decode JSON from initial job page response: {e}"
            )
            return all_jobs

        try:
            total_jobs = response_json["body"]["count"]
        except (KeyError, TypeError) as e:
            logger.error(
                f"Rooster: Unexpected response structure, missing 'body.count': {e}"
            )
            return all_jobs

        total_pages = math.ceil(total_jobs / limit)
        logger.info(
            f"Rooster: Total jobs to fetch: {total_jobs} over {total_pages} pages."
        )

        for page in range(1, total_pages + 1):
            payload["page"] = page

            try:
                response = await async_client.post(base_url, json=payload)
                response.raise_for_status()
                response_json = response.json()
            except httpx.RequestError as e:
                logger.error(f"Rooster: Request failed on page {page}: {e}")
                continue
            except httpx.HTTPStatusError as e:
                logger.error(
                    f"Rooster: Unexpected status {e.response.status_code} on page {page}: {e}"
                )
                continue
            except ValueError as e:
                logger.error(f"Rooster: Failed to decode JSON on page {page}: {e}")
                continue

            try:
                page_jobs = response_json["body"]["data"]
            except (KeyError, TypeError) as e:
                logger.warning(
                    f"Rooster: Missing 'body.data' on page {page}, skipping: {e}"
                )
                continue

            for job in page_jobs:
                all_jobs.append(job)

            await asyncio.sleep(1)

        return all_jobs

    async def crawl_jobs(
        self,
        crawler_run_id: int,
        async_client: httpx.AsyncClient,
        thunder_auth: ThunderAuth,
    ) -> None:

        logger.info("Rooster: Rooster crawl started.")
 
        job_data_list = await self._fetch_all_jobs(async_client)

        job_batch: List[RawJobInput] = []

        for result in job_data_list:
            try:
                job_input = self._parser.parse_rule_based_fields(result, crawler_run_id)
            except ValidationError as e:
                logger.warning(f"Rooster: Skipping malformed job: {e}")
                continue

            job_batch.append(job_input)

            if len(job_batch) >= BATCH_SIZE:
                logger.info(f"Rooster: Flushing full batch of {len(job_batch)} job records to backend.")
                await self._flush_batch(async_client, thunder_auth, job_batch)
 
        if job_batch:
            logger.info(f"Rooster: Flushing remaining {len(job_batch)} job records to backend.")
            await self._flush_batch(async_client, thunder_auth, job_batch)
 
        logger.info("Rooster: Rooster crawl pass concluded.")
