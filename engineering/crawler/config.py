import os

from dotenv import load_dotenv

load_dotenv()


def _getenv(name, default=None):
    """The value without the whitespace around it, or `default` when nothing is left.
    A value of only spaces is easy to leave behind in a .env file or a Kubernetes
    Secret, and it would count as set."""
    return (os.getenv(name) or "").strip() or default


BACKEND_BASE_URL = _getenv("BACKEND_URL", "http://backend:8080/api/v1/crawler")
BATCH_SIZE = 10

MAX_JOB_FAILURE_RATIO = float(os.getenv("MAX_JOB_FAILURE_RATIO") or "0.2")
MIN_JOBS_FOR_FAILURE = int(os.getenv("MIN_JOBS_FOR_FAILURE") or "20")

THUNDER_BASE_URL = os.getenv("THUNDER_BASE_URL")
THUNDER_CLIENT_ID = os.getenv("THUNDER_CLIENT_ID")
THUNDER_CLIENT_SECRET = os.getenv("THUNDER_CLIENT_SECRET")
THUNDER_RESOURCE = os.getenv("THUNDER_RESOURCE")
THUNDER_VERIFY_TLS = os.getenv("THUNDER_INSECURE_TLS", "true").lower() != "true"


# Email alerts for failed crawlers
def _parse_smtp_port(value):
    """None for an empty value. Raises for anything that is not a valid port."""
    value = (value or "").strip()
    if not value:
        return None
    try:
        port = int(value)
    except ValueError:
        port = 0
    if not 1 <= port <= 65535:
        raise ValueError(f"SMTP_PORT must be a number from 1 to 65535, got '{value}'")
    return port


ALERT_EMAILS_ENABLED = os.getenv("ALERT_EMAILS_ENABLED", "true").lower() == "true"
SMTP_HOST = _getenv("SMTP_HOST")
# How the connection to the SMTP server is secured: "starttls" (upgrade a plain
# connection, usually port 587), "ssl" (TLS from the start, usually port 465) or
# "none" (no encryption, only for a trusted relay or a test server)
SMTP_SECURITY = (os.getenv("SMTP_SECURITY") or "starttls").strip().lower()
# Left empty, the port depends on SMTP_SECURITY (587, 465 or 25). It is only parsed
# when the alerts are enabled, so a wrong value cannot stop a crawler that does not
# send any email
SMTP_PORT = _parse_smtp_port(os.getenv("SMTP_PORT")) if ALERT_EMAILS_ENABLED else None
SMTP_USERNAME = _getenv("SMTP_USERNAME")
# Unlike the other settings it is not stripped: every character of a password counts,
# so only a value of nothing but whitespace is treated as not set
_smtp_password = os.getenv("SMTP_PASSWORD")
SMTP_PASSWORD = _smtp_password if _smtp_password and _smtp_password.strip() else None
# Left empty, the email is sent from SMTP_USERNAME
ALERT_FROM_EMAIL = _getenv("ALERT_FROM_EMAIL")
ALERT_RECIPIENTS = [
    email.strip()
    for email in os.getenv("ALERT_RECIPIENTS", "").split(",")
    if email.strip()
]
APP_NAME = _getenv("APP_NAME", "Labour Market Analysis System")
# The page the "Open Data Sources Dashboard" button of the email opens
DASHBOARD_URL = _getenv("DASHBOARD_URL", "http://localhost:3000/data-sources")
