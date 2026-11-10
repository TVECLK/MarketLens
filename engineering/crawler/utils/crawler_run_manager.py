import logging
import asyncio
import httpx
from datetime import datetime, timezone
from typing import Dict, List, Optional, Type

from config import BACKEND_BASE_URL

from crawlers import (
    BaseJobCrawler,
    CrawlResult,
    IkmanCrawler,
    XpressJobsCrawler,
    TopJobsCrawler,
    RoosterCrawler,
    GovernmentJobsCrawler,
)
from .thunder_id_client import ThunderAuth, ThunderIDClient

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
            logger.warning(f"Crawler: Could not connect to tracking backend: {e}")
            raise

    # This function sets the final status of the crawling session
    async def _complete_run(
        self, client: httpx.AsyncClient, crawler_run_id: int, status: str
    ) -> None:
        complete_res = await client.post(
            f"{BACKEND_BASE_URL}/runs/{crawler_run_id}/complete",
            json={"id": crawler_run_id, "status": status},
            auth=self._thunder_auth,
        )
        complete_res.raise_for_status()

    # Best effort: records the run as FAILED so it is not left RUNNING. It never
    # raises, so it cannot hide the error that made the run fail.
    async def _mark_run_failed(
        self, client: httpx.AsyncClient, crawler_run_id: int
    ) -> None:
        try:
            await self._complete_run(client, crawler_run_id, "FAILED")
        except Exception as e:
            logger.error(
                f"Crawler: Could not mark crawler run {crawler_run_id} as FAILED: {e}"
            )

    # This function reconciles stale jobs of the sources that crawled successfully
    # and sets the final status of the current crawling session: "COMPLETED" if at
    # least one source crawled successfully, otherwise "FAILED"
    async def _finalize_run(
        self,
        client: httpx.AsyncClient,
        crawler_run_id: int,
        succeeded_sources: List[str],
    ) -> None:
        try:
            if succeeded_sources:
                logger.info(
                    f"Crawler: Executing pipeline reconciliation for sources: {succeeded_sources}"
                )
                reconcile_res = await client.post(
                    f"{BACKEND_BASE_URL}/jobs/reconcile",
                    json={
                        "crawler_run_id": crawler_run_id,
                        "sources": succeeded_sources,
                    },
                    auth=self._thunder_auth,
                )
                reconcile_res.raise_for_status()
            else:
                logger.warning(
                    "Crawler: No crawler completed successfully, skipping reconciliation."
                )
        except Exception as e:
            logger.error(
                f"Crawler: Reconciliation failed for crawler run {crawler_run_id}: {e}"
            )
            await self._mark_run_failed(client, crawler_run_id)
            raise

        status = "COMPLETED" if succeeded_sources else "FAILED"
        try:
            await self._complete_run(client, crawler_run_id, status)
        except Exception as e:
            logger.error(
                f"Crawler: Failed to finalize crawler run {crawler_run_id}: {e}"
            )
            raise

    # This function calls the relevant crawlers and returns the source name from the
    # crawler's CrawlResult, or None if it raised or did not return a CrawlResult
    async def _run_crawler(
        self,
        name: str,
        crawler_run_id: int,
        client: httpx.AsyncClient,
    ) -> Optional[str]:
        crawler_class = self._registry.get(name)
        if not crawler_class:
            logger.warning(f"Crawler '{name}' not found in registry.")
            return None

        try:
            logger.info(f"--- Starting crawler: {name} ---")
            crawler_instance = crawler_class()
            result = await crawler_instance.crawl_jobs(
                crawler_run_id=crawler_run_id,
                async_client=client,
                thunder_auth=self._thunder_auth,
            )
            # Only an explicit result counts as success: a crawler that returns
            # nothing never ran its completeness checks, so it must not be reconciled
            if not isinstance(result, CrawlResult):
                logger.error(
                    f"Crawler '{name}' returned {type(result).__name__} instead of a "
                    "CrawlResult, treating it as failed."
                )
                return None
            logger.info(
                f"--- Finished crawler: {name} ({result.jobs_attempted} jobs attempted, "
                f"{result.jobs_failed} failed, {result.usable_jobs} usable) ---"
            )
            return result.source_name
        except Exception as e:
            logger.error(f"Crawler '{name}' failed: {e}", exc_info=True)
            return None

    # This function calls the all crawlers one by one and send it to _run_crawler function.
    # It returns the sources that crawled successfully
    async def run_all_crawlers(
        self, crawler_names: Optional[List[str]] = None, concurrent: bool = False
    ) -> List[str]:
        names = crawler_names or list(self._registry.keys())

        async with httpx.AsyncClient(timeout=HTTP_CLIENT_TIMEOUT_SECONDS) as client:
            crawler_run_id = await self._start_run(client)

            tasks = [
                self._run_crawler(name, crawler_run_id, client)
                for name in names
                if name in self._registry
            ]

            if concurrent:
                results = await asyncio.gather(*tasks, return_exceptions=True)
            else:
                results = [await task for task in tasks]

            succeeded_sources = [r for r in results if isinstance(r, str)]

            await self._finalize_run(client, crawler_run_id, succeeded_sources)
            logger.info("Scraper execution pipeline concluded.")
            return succeeded_sources
