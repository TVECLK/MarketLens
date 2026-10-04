import logging
import pytesseract
import httpx

from io import BytesIO
from typing import List
from crawl4ai import AsyncWebCrawler
from bs4 import BeautifulSoup
from PIL import Image
from crawlers.base_crawler import BaseJobCrawler
from utils.thunder_id_client import ThunderAuth
from parsers.governmentjobs_parser import GovernmentJobsParser
from config import BATCH_SIZE
from pydantic import ValidationError
from models.raw_job import RawJobInput

# pytesseract path setup in docker container
pytesseract.pytesseract.tesseract_cmd = "/usr/bin/tesseract"

# Configuration variable setup
TESSERACT_LANG = "eng+sin+tam"
IMAGE_URL = "amazonaws.com/mytutor.lk/vacancy"

logger = logging.getLogger(__name__)


class GovernmentJobsCrawler(BaseJobCrawler):
    def __init__(self):
        self._parser = GovernmentJobsParser()

    def _remove_sinhala_control_chars(self, text):
        cleaned_text = text.replace("\u200c", "").replace("\u200d", "")
        return cleaned_text

    async def _perform_ocr(self, async_client: httpx.AsyncClient, image_url):
        try:
            response = await async_client.get(image_url, timeout=10)
            img = Image.open(BytesIO(response.content)).convert(
                "L"
            )  # Convert to grayscale for better OCR
            return " ".join(
                pytesseract.image_to_string(img, lang=TESSERACT_LANG).split()
            )
        except Exception as e:
            return f"OCR Error: {e}"

    async def _fetch_job_details(self, async_client: httpx.AsyncClient):
        base_url = "https://governmentjobs.lk/index.php?page={}&ipp=25&"
        jobs_data = []

        # Open the crawler once for the whole process
        async with AsyncWebCrawler() as crawler:
            result = await crawler.arun(
                url="https://governmentjobs.lk/index.php?page=1&ipp=25&"
            )
            soup = BeautifulSoup(result.html, "html.parser")

            pagination_text = soup.select_one(".category-results .paginate").text
            total_pages = int(pagination_text.split()[-1])

            for page_num in range(1, total_pages + 1):
                current_url = base_url.format(page_num)

                # Fetch the specific page
                page_result = await crawler.arun(url=current_url)
                page_soup = BeautifulSoup(page_result.html, "html.parser")

                # Iterate through each job
                for job_div in page_soup.select(".grid-view.product"):
                    try:
                        title = job_div.select_one("h5 strong").text.strip()
                        employer = job_div.select_one('a[href*="vtag"]').text.strip()

                        # Get the link to the detailed image page
                        image_page_link = job_div.select_one(
                            'a[href*="image-view.php"]'
                        )["href"]
                        full_image_page_url = (
                            "https://governmentjobs.lk/" + image_page_link
                        )

                        # Visit the detail page using the same crawler session
                        # MUST use await and arun()
                        img_res = await crawler.arun(url=full_image_page_url)
                        img_soup = BeautifulSoup(img_res.html, "html.parser")

                        img_tags = img_soup.select(".page-content img")

                        description = ""
                        for img_tag in img_tags:
                            src = img_tag.get("src")
                            if src and IMAGE_URL in src:
                                description = description + await self._perform_ocr(
                                    async_client, src
                                )

                        description = self._remove_sinhala_control_chars(description)

                        jobs_data.append(
                            {
                                "title": title,
                                "employer": employer,
                                "location": "Sri Lanka",
                                "description": description.strip(),
                            }
                        )

                    except Exception as e:
                        logger.error(f"GovernmentJobs: Error parsing job: {e}")
                        continue

        return jobs_data

    async def crawl_jobs(
        self,
        crawler_run_id: int,
        async_client: httpx.AsyncClient,
        thunder_auth: ThunderAuth,
    ) -> None:

        logger.info("GovernmentJobs: Government jobs crawl started.")
 
        job_data_list = await self._fetch_job_details(async_client)

        job_batch: List[RawJobInput] = []

        for result in job_data_list:
            try:
                job_input = self._parser.parse_rule_based_fields(result, crawler_run_id)
            except ValidationError as e:
                logger.warning(f"GovernmentJobs: Skipping malformed job: {e}")
                continue

            job_batch.append(job_input)

            if len(job_batch) >= BATCH_SIZE:
                logger.info(f"GovernmentJobs: Flushing full batch of {len(job_batch)} job records to backend.")
                await self._flush_batch(async_client, thunder_auth, job_batch)
 
        if job_batch:
            logger.info(f"GovernmentJobs: Flushing remaining {len(job_batch)} job records to backend.")
            await self._flush_batch(async_client, thunder_auth, job_batch)
 
        logger.info("GovernmentJobs: Government jobs crawl pass concluded.")
