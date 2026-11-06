import re

from .base_parser import BaseJobParser
from models import RawJobInput


class IkmanParser(BaseJobParser):
    SOURCE_NAME = "Ikman"

    # This list includes the values that we need to remove in _clean_noise method
    __NOISE_MARKERS = [
        "Show more",
        "Call employer",
        "WhatsApp",
        "Boost this ad",
        "Report this ad",
        "Stay Alert",
        "Popular Jobs",
        "Trending Jobs",
        "Jobs by District",
        "Jobs by City",
        "More from ikman",
        "More ads from",
        "Help & Support",
        "About ikman",
        "Blog & Guides",
        "Download our app",
        "© 20",
    ]

    # This function removes unneccesary characters from the markdown
    def _clean_markdown_links(self, text: str) -> str:
        text = re.sub(r"\[([^\]]*)\]\([^\)]+\)", r"\1", text)
        text = re.sub(r"\]\([^\)]+\)", "", text)
        return text.strip()

    # This function cuts off noise in footer and description
    def _clean_noise(self, text: str) -> str:
        for marker in self.__NOISE_MARKERS:
            idx = text.find(marker)
            if idx != -1:
                text = text[:idx]
        text = re.sub(r"\n+", " ", text)
        text = re.sub(r"\s{2,}", " ", text)
        return text.strip()

    # This function extracts employer and job role from the given markdown and label
    def _extract_label(self, markdown: str, label: str) -> str:
        match = re.search(rf"{label}\s*[:\-]\s*([^\n]+)", markdown, re.IGNORECASE)
        return match.group(1).strip() if match else ""

    # This function extracts the location
    def _extract_location_from_posted_line(self, markdown: str) -> str:
        """Extract district (second span) from 'Posted on ... City, District' line."""
        match = re.search(
            r"Posted on [^,]+,\s*([^,\n]+),\s*([^,\n]+)", markdown, re.IGNORECASE
        )
        if match:
            # group(2) is the second span — the parent district
            location = match.group(2).strip().rstrip(".,")
            location = self._clean_markdown_links(location)
            return location
        return ""

    # This function extracts the description from the markdown
    def _extract_description(self, markdown: str) -> str:
        """Grab job body content — saved to key_responsibilities."""
        patterns = [
            r"About the role\s*\n+([\s\S]+?)(?=\n#{1,3}\s|\Z)",
            r"Key Responsibilities\s*\n+([\s\S]+?)(?=\n#{1,3}\s|\Z)",
            r"Job Description\s*\n+([\s\S]+?)(?=\n#{1,3}\s|\Z)",
        ]
        for pattern in patterns:
            match = re.search(pattern, markdown, re.IGNORECASE)
            if match:
                return match.group(1).strip()

        meta_end = re.search(
            r"(Application deadline|Required work experience)[^\n]*\n",
            markdown,
            re.IGNORECASE,
        )
        if meta_end:
            return markdown[meta_end.end() :].strip()

        return ""

    # This function returns the employer, job role, location and description of the job
    def parse_rule_based_fields(
        self, markdown: str, crawler_run_id: int
    ) -> RawJobInput:

        employer = self._extract_label(markdown, "Employer")
        job_role = self._extract_label(markdown, "Role")

        job_role = self._clean_markdown_links(job_role)

        if employer.lower() in ("private poster", "private", ""):
            employer = ""

        if not job_role:
            h1 = re.search(r"^#\s+(.+)$", markdown, re.MULTILINE)
            job_role = self._clean_markdown_links(h1.group(1).strip()) if h1 else ""

        location = self._extract_location_from_posted_line(markdown)
        if not location:
            location = self._extract_label(markdown, "Location")
        if not location and job_role:
            dash = re.search(r"-\s+([A-Za-z\s]+)$", job_role)
            if dash:
                location = dash.group(1).strip()

        location = location or ""

        description = self._extract_description(markdown)
        description = self._clean_noise(description)

        return RawJobInput(
            employer=employer,
            job_role=job_role,
            location=location,
            description=description,
            crawler_run_id=crawler_run_id,
            source=self.SOURCE_NAME,
        )
