from .base_parser import BaseJobParser
from models import RawJobInput


class RoosterParser(BaseJobParser):
    SOURCE_NAME = "Rooster"

    def parse_rule_based_fields(self, job: dict, crawler_run_id: int) -> RawJobInput:
        return RawJobInput(
            employer=job.get("company_name") or "",
            job_role=job.get("title") or "",
            location=job.get("location") or "",
            description=job.get("description") or "",
            crawler_run_id=crawler_run_id,
            source=self.SOURCE_NAME,
        )
