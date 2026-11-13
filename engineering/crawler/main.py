import asyncio
import logging
import sys

from utils.crawler_run_manager import CrawlerManager
from utils.email_alert import EmailAlertService

logging.basicConfig(
    level=logging.INFO, format="%(asctime)s [%(levelname)s] %(name)s: %(message)s"
)
logger = logging.getLogger(__name__)


async def crawl_job():
    logger.info("\n--- Crawling Started ---")

    alert_service = EmailAlertService()

    try:
        manager = CrawlerManager(alert_service)
        succeeded_sources = await manager.run_all_crawlers(concurrent=True)
    except Exception as e:
        logger.exception("CRITICAL ERROR encountered during execution lifecycle")
        await alert_service.send_crash_alert(e)
        sys.exit(1)

    # No active source is not a failure, there was just nothing to crawl
    if succeeded_sources is None:
        logger.info("--- Crawling Skipped: no active sources ---")
        return

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
