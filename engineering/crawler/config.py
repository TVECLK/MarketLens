import os

from dotenv import load_dotenv

load_dotenv()

BACKEND_BASE_URL = os.getenv("BACKEND_URL", "http://backend:8080/api/v1/crawler")
BATCH_SIZE = 10

MAX_JOB_FAILURE_RATIO = float(os.getenv("MAX_JOB_FAILURE_RATIO") or "0.2")
MIN_JOBS_FOR_FAILURE = int(os.getenv("MIN_JOBS_FOR_FAILURE") or "20")

THUNDER_BASE_URL = os.getenv("THUNDER_BASE_URL")
THUNDER_CLIENT_ID = os.getenv("THUNDER_CLIENT_ID")
THUNDER_CLIENT_SECRET = os.getenv("THUNDER_CLIENT_SECRET")
THUNDER_RESOURCE = os.getenv("THUNDER_RESOURCE")
THUNDER_VERIFY_TLS = os.getenv("THUNDER_INSECURE_TLS", "true").lower() != "true"


# Email alerts for failed crawlers
ALERT_EMAILS_ENABLED = os.getenv("ALERT_EMAILS_ENABLED", "true").lower() == "true"
SMTP_HOST = os.getenv("SMTP_HOST")
# How the connection to the SMTP server is secured: "starttls" (upgrade a plain
# connection, usually port 587), "ssl" (TLS from the start, usually port 465) or
# "none" (no encryption, only for a trusted relay or a test server)
SMTP_SECURITY = (os.getenv("SMTP_SECURITY") or "starttls").strip().lower()
# Left empty, the port depends on SMTP_SECURITY (587, 465 or 25)
SMTP_PORT = int(os.getenv("SMTP_PORT") or 0) or None
SMTP_USERNAME = os.getenv("SMTP_USERNAME")
SMTP_PASSWORD = os.getenv("SMTP_PASSWORD")
ALERT_FROM_EMAIL = os.getenv("ALERT_FROM_EMAIL") or SMTP_USERNAME
ALERT_RECIPIENTS = [
    email.strip()
    for email in os.getenv("ALERT_RECIPIENTS", "").split(",")
    if email.strip()
]
APP_NAME = os.getenv("APP_NAME") or "Labour Market Analysis System"
# The page the "Open Data Sources Dashboard" button of the email opens
DASHBOARD_URL = os.getenv("DASHBOARD_URL", "http://localhost:3000/data-sources")
