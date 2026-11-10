import asyncio
import logging
import sys

from utils.crawler_run_manager import CrawlerManager

logging.basicConfig(
    level=logging.INFO, format="%(asctime)s [%(levelname)s] %(name)s: %(message)s"
)
logger = logging.getLogger(__name__)


async def crawl_job():
    logger.info("\n--- Crawling Started ---")

    try:
        manager = CrawlerManager()
        succeeded_sources = await manager.run_all_crawlers(concurrent=True)
    except Exception:
        logger.exception("CRITICAL ERROR encountered during execution lifecycle")
        sys.exit(1)

    # A run where no crawler succeeded must not look like a success to the scheduler
    if not succeeded_sources:
        logger.error("--- Crawling Failed: no crawler completed successfully ---")
        sys.exit(1)

    logger.info("--- Crawling Completed ---")


if __name__ == "__main__":
    try:
        asyncio.run(crawl_job())
    except KeyboardInterrupt:
        logger.info("\nCrawling is stopped by user.")
        sys.exit(1)
