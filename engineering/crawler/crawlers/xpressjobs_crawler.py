import asyncio
import httpx
import logging

from typing import List
from bs4 import BeautifulSoup
from pydantic import ValidationError
from crawlers.base_crawler import BaseJobCrawler
from utils.thunder_id_client import ThunderAuth
from parsers.xpressjobs_parser import XpressJobsParser
from models.raw_job import RawJobInput
from config import BATCH_SIZE

logger = logging.getLogger(__name__)


class XpressJobsCrawler(BaseJobCrawler):
    def __init__(self):
        self._parser = XpressJobsParser()

    def _clean_html(self, html_content):
        if not html_content:
            return ""
        soup = BeautifulSoup(html_content, "html.parser")
        return soup.get_text(separator=" ").strip()

    async def _fetch_job_details(self, async_client: httpx.AsyncClient, job_id):
        url = f"https://xpress.jobs/api/jobs/publishedJob?jobId={job_id}"

        try:
            response = await async_client.get(url)
            response.raise_for_status()

            if response.status_code != 200:
                logger.warning(
                    f"Unexpected status {response.status_code} while fetching job {job_id}"
                )
                return None

            data = response.json()

            job_item = data.get("jobItem") or {}
            return {
                "job_title": data.get("jobTitle"),
                "employer": job_item.get("organizationName"),
                "location": job_item.get("locations"),
                "description": self._clean_html(data.get("jobInfo", "")),
            }

        except httpx.HTTPStatusError as e:
            logger.warning(
                f"Unexpected status {e.response.status_code} while fetching job {job_id}"
            )
            return None
        except httpx.RequestError as e:
            logger.warning(f"Request failed while fetching job {job_id}: {e}")
            return None
        except ValueError as e:
            logger.warning(f"Failed to decode JSON for job {job_id}: {e}")
            return None
        except AttributeError as e:
            logger.warning(f"Unexpected response structure for job {job_id}: {e}")
            return None

    async def _process_all_jobs(self, async_client: httpx.AsyncClient):
        final_data = []
        page = 1

        while True:
            logger.info(f"XpressJobs: --- Fetching page {page} ---")

            # Build the URL with the current page
            list_url = f"https://xpress.jobs/api/jobs/searchJobs?page={page}&pageSize=20&keyword=&locations=&sectors=&jobTypes=&careerLevels=&sortBy=SortedCreateDate+DESC&byCVLess=false&byWalkIn=false"

            try:
                response = await async_client.get(list_url, timeout=10)
                jobs_list = response.json()
            except Exception as e:
                logger.error(f"XpressJobs: Error fetching page {page}: {e}")
                break

            # Break the loop if the list is empty
            if not jobs_list:
                logger.info("XpressJobs: No more jobs found. Finishing.")
                break

            # Process each job on the current page
            for job_summary in jobs_list:
                job_id = job_summary["jobId"]
                logger.info(
                    f"XpressJobs: Processing job {job_id}: {job_summary['jobTitle']}"
                )

                details = await self._fetch_job_details(async_client, job_id)
                if details:
                    final_data.append(details)

                await asyncio.sleep(10)

            # Move to next page
            page += 1

            await asyncio.sleep(10)

        return final_data

    async def crawl_jobs(
        self,
        crawler_run_id: int,
        async_client: httpx.AsyncClient,
        thunder_auth: ThunderAuth,
    ) -> None:

        logger.info("XpressJobs: Xpress jobs crawl started.")
        
        job_data_list = await self._process_all_jobs(async_client)

        job_batch: List[RawJobInput] = []

        for result in job_data_list:
            try:
                job_input = self._parser.parse_rule_based_fields(result, crawler_run_id)
            except ValidationError as e:
                logger.warning(f"XpressJobs: Skipping malformed job: {e}")
                continue

            job_batch.append(job_input)

            if len(job_batch) >= BATCH_SIZE:
                logger.info(f"XpressJobs: Flushing full batch of {len(job_batch)} job records to backend.")
                await self._flush_batch(async_client, thunder_auth, job_batch)
 
        if job_batch:
            logger.info(f"XpressJobs: Flushing remaining {len(job_batch)} job records to backend.")
            await self._flush_batch(async_client, thunder_auth, job_batch)
 
        logger.info("XpressJobs: Xpress jobs crawl pass concluded.")
