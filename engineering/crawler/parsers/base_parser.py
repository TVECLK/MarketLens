from abc import ABC, abstractmethod
from models import RawJobInput


class BaseJobParser(ABC):
    def __init__(self, source_name: str):
        self._source_name = source_name

    @abstractmethod
    def parse_rule_based_fields(self, *args, **kwargs) -> RawJobInput:
        pass
