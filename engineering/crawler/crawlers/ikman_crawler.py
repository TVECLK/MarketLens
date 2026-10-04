import httpx
import logging
import re
import math
from typing import List

from crawlers.base_crawler import BaseJobCrawler
from parsers.ikman_parser import IkmanParser
from utils.thunder_id_client import ThunderAuth
from pydantic import ValidationError
from models.raw_job import RawJobInput
from config import BATCH_SIZE

from crawl4ai import (
    AsyncWebCrawler,
    CrawlerRunConfig,
    BrowserConfig,
    MemoryAdaptiveDispatcher,
)

logger = logging.getLogger(__name__)


class IkmanCrawler(BaseJobCrawler):
    def __init__(self):
        self._parser = IkmanParser()

    # This function returns the last page number from the site
    async def _get_last_page_from_text(self) -> int:

        async with AsyncWebCrawler() as crawler:
            result = await crawler.arun(url="https://ikman.lk/en/ads/sri-lanka/jobs")
            html_content = result.html
            match = re.search(r"of ([\d,]+) ads", html_content)
            if match:
                total_ads = int(match.group(1).replace(",", ""))
                ads_per_page = 25
                last_page = math.ceil(total_ads / ads_per_page)
                logger.info(
                    f"Ikman: Total Ads: {total_ads}, Calculated Last Page: {last_page}"
                )
                return last_page
            else:
                logger.warning("Ikman: Could not find the total ad count text.")
                return 1

    async def crawl_jobs(
        self,
        crawler_run_id: int,
        async_client: httpx.AsyncClient,
        thunder_auth: ThunderAuth,
    ) -> None:

        logger.info("Ikman: Ikman crawl started.")

        max_pages = await self._get_last_page_from_text()

        job_batch: List[RawJobInput] = []

        browser_config = BrowserConfig(
            headless=True, extra_args=["--disable-gpu", "--no-sandbox"]
        )
        dispatcher = MemoryAdaptiveDispatcher(
            memory_threshold_percent=80.0, max_session_permit=10
        )

        async with AsyncWebCrawler(config=browser_config) as crawler:
            all_detail_urls = []

            for page in range(1, max_pages + 1):
                url = f"https://ikman.lk/en/ads/sri-lanka/jobs?page={page}"
                res = await crawler.arun(
                    url=url, config=CrawlerRunConfig(cache_mode="BYPASS")
                )
                if res.success:
                    links = [
                        f"https://ikman.lk{link['href']}"
                        if link["href"].startswith("/")
                        else link["href"]
                        for link in res.links.get("internal", [])
                        if "/en/ad/" in link["href"]
                    ]
                    all_detail_urls.extend(links)

            unique_urls = list(set(all_detail_urls))
            logger.info(
                f"Ikman: Processing structural extraction queue for {len(unique_urls)} links."
            )

            detail_config = CrawlerRunConfig(cache_mode="BYPASS", stream=True)
            results_generator = await crawler.arun_many(
                urls=unique_urls, config=detail_config, dispatcher=dispatcher
            )

            async for result in results_generator:
                if not result.success or not result.markdown:
                    continue

                try:
                    job_input = self._parser.parse_rule_based_fields(
                        markdown=result.markdown.raw_markdown,
                        crawler_run_id=crawler_run_id,
                    )
                except ValidationError as e:
                    logger.warning(
                        f"Ikman: Skipping malformed job at {result.url}: {e}"
                    )
                    continue

                job_batch.append(job_input)

                if len(job_batch) >= BATCH_SIZE:
                    await self._flush_batch(async_client, thunder_auth, job_batch)

            if job_batch:
                await self._flush_batch(async_client, thunder_auth, job_batch)

        logger.info("Ikman: Ikman crawl pass concluded.")
