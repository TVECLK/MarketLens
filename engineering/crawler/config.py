import os

BACKEND_BASE_URL = os.getenv("BACKEND_URL", "http://backend:8080/api/v1/crawler")
BATCH_SIZE = 10

# A crawler is treated as incomplete when more than this fraction of the jobs it
# tried to fetch failed, but only once at least MIN_JOBS_FOR_FAILURE_RATIO jobs
# were attempted (a handful of jobs is too small a sample to judge).
MAX_JOB_FAILURE_RATIO = float(os.getenv("MAX_JOB_FAILURE_RATIO", "0.2"))
MIN_JOBS_FOR_FAILURE_RATIO = int(os.getenv("MIN_JOBS_FOR_FAILURE_RATIO", "20"))

THUNDER_BASE_URL = os.getenv("THUNDER_BASE_URL")
THUNDER_CLIENT_ID = os.getenv("THUNDER_CLIENT_ID")
THUNDER_CLIENT_SECRET = os.getenv("THUNDER_CLIENT_SECRET")
THUNDER_RESOURCE = os.getenv("THUNDER_RESOURCE")
THUNDER_VERIFY_TLS = os.getenv("THUNDER_INSECURE_TLS", "true").lower() != "true"
