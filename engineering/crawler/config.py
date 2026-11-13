import os

BACKEND_BASE_URL = os.getenv("BACKEND_URL", "http://backend:8080/api/v1/crawler")
BATCH_SIZE = 10

MAX_JOB_FAILURE_RATIO = float(os.getenv("MAX_JOB_FAILURE_RATIO", "0.2"))
MIN_JOBS_FOR_FAILURE = int(os.getenv("MIN_JOBS_FOR_FAILURE", "20"))

THUNDER_BASE_URL = os.getenv("THUNDER_BASE_URL")
THUNDER_CLIENT_ID = os.getenv("THUNDER_CLIENT_ID")
THUNDER_CLIENT_SECRET = os.getenv("THUNDER_CLIENT_SECRET")
THUNDER_RESOURCE = os.getenv("THUNDER_RESOURCE")
THUNDER_VERIFY_TLS = os.getenv("THUNDER_INSECURE_TLS", "true").lower() != "true"


# Email alerts for failed crawlers (Google SMTP: use an App Password, not the account password)
ALERT_EMAILS_ENABLED = os.getenv("ALERT_EMAILS_ENABLED", "true").lower() == "true"
SMTP_HOST = os.getenv("SMTP_HOST", "smtp.gmail.com")
SMTP_PORT = int(os.getenv("SMTP_PORT", "587"))
SMTP_USERNAME = os.getenv("SMTP_USERNAME")
SMTP_PASSWORD = os.getenv("SMTP_PASSWORD")
ALERT_FROM_EMAIL = os.getenv("ALERT_FROM_EMAIL") or SMTP_USERNAME
ALERT_RECIPIENTS = [
    email.strip()
    for email in os.getenv("ALERT_RECIPIENTS", "").split(",")
    if email.strip()
]
# The page the "Open Data Sources Dashboard" button of the email opens
DASHBOARD_URL = os.getenv("DASHBOARD_URL", "http://localhost:3000/")
