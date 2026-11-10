from .base_parser import BaseJobParser
from models import RawJobInput


class XpressJobsParser(BaseJobParser):
    def parse_rule_based_fields(self, job: dict, crawler_run_id: int) -> RawJobInput:
        return RawJobInput(
            employer=job.get("employer") or "",
            job_role=job.get("job_title") or "",
            location=job.get("location") or "",
            description=job.get("description") or "",
            crawler_run_id=crawler_run_id,
            source=self._source_name,
        )
