import logging
import asyncio
import httpx
from datetime import datetime, timezone
from typing import Dict, List, Optional, Type

from config import BACKEND_BASE_URL

from crawlers.base_crawler import BaseJobCrawler
from crawlers.ikman_crawler import IkmanCrawler
from crawlers.xpressjobs_crawler import XpressJobsCrawler
from crawlers.topjobs_crawler import TopJobsCrawler
from crawlers.rooster_crawler import RoosterCrawler
from crawlers.governmentjobs_crawler import GovernmentJobsCrawler
from utils.thunder_id_client import ThunderAuth, ThunderIDClient

logger = logging.getLogger(__name__)

HTTP_CLIENT_TIMEOUT_SECONDS = 300.0


class CrawlerManager:
    def __init__(self):
        self._registry: Dict[str, Type[BaseJobCrawler]] = {
            "rooster": RoosterCrawler,
            "xpress": XpressJobsCrawler,
            "topjobs": TopJobsCrawler,
            "governmentjobs": GovernmentJobsCrawler,
            "ikman": IkmanCrawler,
        }
        self._thunder_client = ThunderIDClient()
        self._thunder_auth = ThunderAuth(self._thunder_client)

    # This function created the new crawler session and return the new crawler run id
    async def _start_run(self, client: httpx.AsyncClient) -> int:
        current_time_iso = datetime.now(timezone.utc).astimezone().isoformat()
        start_payload = {
            "started_at": current_time_iso,
            "finished_at": None,
            "status": "RUNNING",
        }
        try:
            init_res = await client.post(
                f"{BACKEND_BASE_URL}/runs",
                json=start_payload,
                auth=self._thunder_auth,
            )
            init_res.raise_for_status()
            response_data = init_res.json()
            crawler_run_id = response_data.get("id")

            if crawler_run_id is None:
                raise ValueError(f"'id' missing in /runs response: {response_data}")

            logger.info(
                f"Initialized Tracking Crawler Session Run ID: {crawler_run_id}"
            )
            return crawler_run_id
        except Exception as e:
            logger.warning(
                f"Crawler: Could not connect to tracking backend."
            )
            raise

    #This function sets the status of the current crawling session to "COMPLETED"
    #and sets the end date of the jobs that are not equal to current crawler run id
    async def _finalize_run(self, client: httpx.AsyncClient, crawler_run_id: int) -> None:
        try:
            logger.info("Crawler: Executing pipeline reconciliation.")

            reconcile_res = await client.post(
                f"{BACKEND_BASE_URL}/jobs/reconcile",
                json={"crawler_run_id": crawler_run_id},
                auth=self._thunder_auth,
            )
            reconcile_res.raise_for_status()

            complete_res = await client.post(
                f"{BACKEND_BASE_URL}/runs/{crawler_run_id}/complete",
                json={"id": crawler_run_id, "status": "COMPLETED"},
                auth=self._thunder_auth,
            )
            complete_res.raise_for_status()
        except Exception as e:
            logger.error(f"Crawler: Failed to finalize crawler run {crawler_run_id}: {e}")

    # This function calls the relevant crawlers
    async def _run_crawler(
        self,
        name: str,
        crawler_run_id: int,
        client: httpx.AsyncClient,
    ) -> None:
        crawler_class = self._registry.get(name)
        if not crawler_class:
            logger.warning(f"Crawler '{name}' not found in registry.")
            return

        try:
            logger.info(f"--- Starting crawler: {name} ---")
            crawler_instance = crawler_class()
            await crawler_instance.crawl_jobs(
                crawler_run_id=crawler_run_id,
                async_client=client,
                thunder_auth=self._thunder_auth
            )
            logger.info(f"--- Finished crawler: {name} ---")
        except Exception as e:
            logger.error(f"Crawler '{name}' failed: {e}", exc_info=True)

    # This function calls the all crawlers one by one and send it to _run_crawler function
    async def run_all_crawlers(
        self, crawler_names: Optional[List[str]] = None, concurrent: bool = False
    ) -> None:
        names = crawler_names or list(self._registry.keys())

        async with httpx.AsyncClient(timeout=HTTP_CLIENT_TIMEOUT_SECONDS) as client:
            crawler_run_id = await self._start_run(client)

            tasks = [
                self._run_crawler(name, crawler_run_id, client)
                for name in names
                if name in self._registry
            ]

            if concurrent:
                await asyncio.gather(*tasks, return_exceptions=True)
            else:
                for task in tasks:
                    await task

            await self._finalize_run(client, crawler_run_id)
            logger.info("Scraper execution pipeline concluded.")
