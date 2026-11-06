from abc import ABC, abstractmethod
from models import RawJobInput


class BaseJobParser(ABC):
    # Must match the source name the backend stores in the `source` table
    SOURCE_NAME: str

    @abstractmethod
    def parse_rule_based_fields(self, *args, **kwargs) -> RawJobInput:
        pass
