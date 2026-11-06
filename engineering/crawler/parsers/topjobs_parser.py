from .base_parser import BaseJobParser
from models import RawJobInput


class TopJobsParser(BaseJobParser):
    SOURCE_NAME = "TopJobs"

    def parse_rule_based_fields(self, data: dict, crawler_run_id: int) -> RawJobInput:
        return RawJobInput(
            employer=data.get("employer", ""),
            job_role=data.get("title", ""),
            location=data.get("location", ""),
            description=data.get("ocr_text", ""),
            crawler_run_id=crawler_run_id,
            source=self.SOURCE_NAME,
        )
