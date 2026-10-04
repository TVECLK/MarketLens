import asyncio
import io
import logging
import re
import pytesseract
import httpx

from typing import List
from bs4 import BeautifulSoup
from PIL import Image
from playwright.async_api import async_playwright
from pydantic import ValidationError
from crawlers.base_crawler import BaseJobCrawler
from utils.thunder_id_client import ThunderAuth
from parsers.topjobs_parser import TopJobsParser
from models.raw_job import RawJobInput
from config import BATCH_SIZE


# pytesseract path setup in docker container
pytesseract.pytesseract.tesseract_cmd = "/usr/bin/tesseract"

# Configuration variable setup
LISTING_URL = "https://www.topjobs.lk/applicant/vacancybyfunctionalarea.jsp?FA=&jst=OPEN&sQut=&txtKeyWord=&chkGovt=&chkParttime=&chkWalkin=&chkNGO="
TESSERACT_LANG = "eng+sin+tam"
POPUP_WAIT_TIMEOUT = 15000
USER_AGENT = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"

logger = logging.getLogger(__name__)


class TopJobsCrawler(BaseJobCrawler):
    def __init__(self):
        self._parser = TopJobsParser()

    async def _get_total_pages(self, html: str) -> int:
        soup = BeautifulSoup(html, "html.parser")
        pagination_div = soup.find("div", class_="pagin-block page-show")

        if pagination_div:
            text = pagination_div.get_text(strip=True)
            match = re.search(r"(\d+)\s*page\(s\)", text)
            if match:
                return int(match.group(1))

        return 1

    async def _parse_listing_html(self, html: str) -> list[dict]:
        soup = BeautifulSoup(html, "html.parser")
        jobs = []
        rows = soup.find_all("tr", attrs={"onclick": re.compile(r"createAlert")})

        for row in rows:
            row_id = row.get("id")
            tds = row.find_all("td", recursive=False)
            if len(tds) < 6:
                continue

            jobs.append(
                {
                    "row_id": row_id,
                    "title": tds[2].find("h2").text.strip()
                    if tds[2].find("h2")
                    else "N/A",
                    "employer": tds[2].find("h1").text.strip()
                    if tds[2].find("h1")
                    else "N/A",
                    "location": tds[6].text.strip()
                    if len(tds) > 6
                    else (tds[5].text.strip() if len(tds) > 5 else "N/A"),
                    "ocr_text": None,
                    "error": None,
                }
            )
        return jobs

    async def _extract_complete_jobs_details(self):
        async with async_playwright() as p:
            browser = await p.chromium.launch(headless=True)
            context = await browser.new_context(user_agent=USER_AGENT)
            page = await context.new_page()

            await page.goto(f"{LISTING_URL}&pageNo=1", wait_until="networkidle")
            total_pages = await self._get_total_pages(await page.content())

            logger.info(f"TopJobs: Total pages detected: {total_pages}")

            all_jobs = []

            for page_num in range(1, total_pages + 1):
                logger.info(f"TopJobs: Scraping page {page_num} of {total_pages}...")
                if page_num > 1:
                    await page.goto(
                        f"{LISTING_URL}&pageNo={page_num}", wait_until="networkidle"
                    )

                content = await page.content()
                jobs = await self._parse_listing_html(content)
                logger.info(
                    f"TopJobs: Found {len(jobs)} jobs. Starting popup processing..."
                )

                for i, job in enumerate(jobs):
                    try:
                        await asyncio.sleep(3)

                        async with context.expect_page(
                            timeout=POPUP_WAIT_TIMEOUT
                        ) as popup_info:
                            await page.evaluate(
                                f"document.getElementById('{job['row_id']}').click()"
                            )

                        popup = await popup_info.value
                        await popup.wait_for_load_state("networkidle")

                        all_images = popup.locator("img")
                        img_locator = None

                        for count in range(await all_images.count()):
                            candidate = all_images.nth(count)
                            box = await candidate.bounding_box()
                            if box and box["width"] > 500:
                                img_locator = candidate
                                break

                        if not img_locator:
                            logger.error(
                                "TopJobs: Could not find a large advertisement image."
                            )
                            await popup.close()
                            continue

                        await img_locator.wait_for(
                            state="visible", timeout=POPUP_WAIT_TIMEOUT
                        )
                        screenshot_bytes = await img_locator.screenshot()
                        await popup.close()

                        # OCR Processing
                        image = Image.open(io.BytesIO(screenshot_bytes)).convert("L")
                        if image.width < 1400:
                            scale = 1400 / image.width
                            image = image.resize(
                                (int(image.width * scale), int(image.height * scale))
                            )

                        image = image.point(lambda x: 0 if x < 180 else 255, "1")
                        job["ocr_text"] = " ".join(
                            pytesseract.image_to_string(
                                image, lang=TESSERACT_LANG
                            ).split()
                        )

                        all_jobs.append(job)

                    except Exception as e:
                        logger.error(f"TopJobs: Failed to process {job['row_id']}: {e}")
                        job["error"] = str(e)

            await browser.close()

            return all_jobs

    async def crawl_jobs(
        self,
        crawler_run_id: int,
        async_client: httpx.AsyncClient,
        thunder_auth: ThunderAuth,
    ) -> None:

        logger.info("TopJobs: Top jobs crawl started.")
 
        job_data_list = await self._extract_complete_jobs_details()

        job_batch: List[RawJobInput] = []

        for result in job_data_list:
            try:
                job_input = self._parser.parse_rule_based_fields(result, crawler_run_id)
            except ValidationError as e:
                logger.warning(f"TopJobs: Skipping malformed job: {e}")
                continue

            job_batch.append(job_input)

            if len(job_batch) >= BATCH_SIZE:
                logger.info(f"TopJobs: Flushing full batch of {len(job_batch)} job records to backend.")
                await self._flush_batch(async_client, thunder_auth, job_batch)
 
        if job_batch:
            logger.info(f"TopJobs: Flushing remaining {len(job_batch)} job records to backend.")
            await self._flush_batch(async_client, thunder_auth, job_batch)
 
        logger.info("TopJobs: Top jobs crawl pass concluded.")
