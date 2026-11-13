import logging
import asyncio
import httpx
from datetime import datetime, timezone
from typing import Dict, List, Optional, Type

from config import BACKEND_BASE_URL

from crawlers import BaseJobCrawler, CrawlResult
from crawlers.registry import CRAWLER_REGISTRY
from .email_alert import EmailAlertService, FailedCrawler
from .thunder_id_client import ThunderAuth, ThunderIDClient

logger = logging.getLogger(__name__)

HTTP_CLIENT_TIMEOUT_SECONDS = 300.0


class CrawlerManager:
    def __init__(self, alert_service: EmailAlertService):
        self._alert_service = alert_service
        self._registry: Dict[str, Type[BaseJobCrawler]] = {}
        self._failure_reasons: Dict[str, str] = {}
        self._thunder_client = ThunderIDClient()
        self._thunder_auth = ThunderAuth(self._thunder_client)

    # This function asks the backend for the active sources and returns the crawlers
    # of those sources, keyed by source name. It raises if the backend cannot be
    # reached, so a failure never turns into crawling sources that were disabled
    async def fetch_active_crawlers(
        self, client: httpx.AsyncClient
    ) -> Dict[str, Type[BaseJobCrawler]]:
        response = await client.get(
            f"{BACKEND_BASE_URL}/sources", auth=self._thunder_auth
        )
        response.raise_for_status()

        active_names = {
            source["name"].strip().lower() for source in response.json()["sources"]
        }

        for name in sorted(active_names - CRAWLER_REGISTRY.keys()):
            logger.warning(f"Crawler: Active source '{name}' has no crawler, skipping.")
        for name in sorted(CRAWLER_REGISTRY.keys() - active_names):
            logger.info(f"Crawler: Source '{name}' is not active, skipping.")

        return {
            name: crawler_class
            for name, crawler_class in CRAWLER_REGISTRY.items()
            if name in active_names
        }

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
            self._failure_reasons[name] = "No crawler found for this source"
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
                self._failure_reasons[name] = (
                    "Crawler finished without reporting a result"
                )
                return None
            logger.info(
                f"--- Finished crawler: {name} ({result.jobs_attempted} jobs attempted, "
                f"{result.jobs_failed} failed, {result.usable_jobs} usable) ---"
            )
            return result.source_name
        except Exception as e:
            logger.error(f"Crawler '{name}' failed: {e}", exc_info=True)
            self._failure_reasons[name] = f"{type(e).__name__}: {e}"
            return None

    # This function runs the crawlers of the active sources: it fetches them from the
    # backend, starts a run, sends each crawler to _run_crawler (concurrently or one by
    # one) and finalizes the run. It returns the sources that crawled successfully, or
    # None if no source is active (nothing to crawl, so no run is created)
    async def run_all_crawlers(self, concurrent: bool = False) -> Optional[List[str]]:
        async with httpx.AsyncClient(timeout=HTTP_CLIENT_TIMEOUT_SECONDS) as client:
            # Fetched before the run starts, so a failure leaves no half-created run
            self._registry = await self.fetch_active_crawlers(client)
            if not self._registry:
                logger.warning("Crawler: No active sources, nothing to crawl.")
                return None

            crawler_run_id = await self._start_run(client)
            self._failure_reasons = {}

            tasks = [
                self._run_crawler(name, crawler_run_id, client)
                for name in self._registry
            ]

            if concurrent:
                results = await asyncio.gather(*tasks, return_exceptions=True)
            else:
                results = [await task for task in tasks]

            succeeded_sources = [r for r in results if isinstance(r, str)]

            await self._finalize_run(client, crawler_run_id, succeeded_sources)

            # The active crawlers that did not succeed. The email is only sent when
            # there is at least one, and after the run is recorded
            failed_crawlers = [
                FailedCrawler(
                    name, self._failure_reasons.get(name, "Crawler did not succeed")
                )
                for name in self._registry
                if name not in succeeded_sources
            ]
            await self._alert_service.send_failure_alert(
                failed_crawlers, crawler_run_id
            )

            logger.info("Scraper execution pipeline concluded.")
            return succeeded_sources
