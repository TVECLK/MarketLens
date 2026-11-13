from typing import Dict, Type

from .base_crawler import BaseJobCrawler
from .governmentjobs_crawler import GovernmentJobsCrawler
from .ikman_crawler import IkmanCrawler
from .rooster_crawler import RoosterCrawler
from .source_names import GOVERNMENTJOBS, IKMAN, ROOSTER, TOPJOBS, XPRESSJOBS
from .topjobs_crawler import TopJobsCrawler
from .xpressjobs_crawler import XpressJobsCrawler

# Every crawler that exists, keyed by its source name
CRAWLER_REGISTRY: Dict[str, Type[BaseJobCrawler]] = {
    ROOSTER: RoosterCrawler,
    XPRESSJOBS: XpressJobsCrawler,
    TOPJOBS: TopJobsCrawler,
    GOVERNMENTJOBS: GovernmentJobsCrawler,
    IKMAN: IkmanCrawler,
}
