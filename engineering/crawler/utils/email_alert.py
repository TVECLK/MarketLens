import asyncio
import logging
import smtplib
from dataclasses import dataclass
from datetime import datetime
from email.message import EmailMessage
from html import escape
from typing import List, Optional

from config import (
    ALERT_EMAILS_ENABLED,
    ALERT_FROM_EMAIL,
    ALERT_RECIPIENTS,
    DASHBOARD_URL,
    SMTP_HOST,
    SMTP_PASSWORD,
    SMTP_PORT,
    SMTP_USERNAME,
)

logger = logging.getLogger(__name__)

MAX_REASON_LENGTH = 200
SMTP_TIMEOUT_SECONDS = 30

# Colors of the frontend theme (frontend/app/globals.css)
COLOR_DARK = "#1e1e2e"
COLOR_PRIMARY = "#005A9C"


@dataclass(frozen=True)
class FailedCrawler:
    name: str
    reason: str


def _truncate(text: str) -> str:
    text = " ".join(str(text).split())
    if len(text) <= MAX_REASON_LENGTH:
        return text
    return text[: MAX_REASON_LENGTH - 3] + "..."


def _timestamp() -> str:
    return datetime.now().astimezone().strftime("%Y-%m-%d %H:%M %Z")


# Wraps the content in the shared layout. Only inline CSS and tables are used,
# because email clients ignore <style> blocks and have poor support for flexbox/grid.
def _render_html(
    title: str, intro: str, body: str, note: str, dashboard_url: str
) -> str:
    return f"""\
<!DOCTYPE html>
<html>
<body style="margin:0;padding:0;background-color:#f9fafb;font-family:Arial,Helvetica,sans-serif;color:#111827;">
  <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="background-color:#f9fafb;padding:24px 0;">
    <tr><td align="center">
      <table role="presentation" width="600" cellpadding="0" cellspacing="0" style="max-width:600px;width:100%;background-color:#ffffff;border:1px solid #e5e7eb;border-radius:12px;overflow:hidden;">
        <tr><td style="background-color:{COLOR_DARK};padding:20px 28px;">
          <span style="color:#ffffff;font-size:18px;font-weight:bold;">Labour Market Demand Application</span>
          <span style="color:#93c5fd;font-size:13px;margin-left:8px;">Crawler Alert</span>
        </td></tr>
        <tr><td style="background-color:#fef2f2;border-bottom:1px solid #fecaca;padding:18px 28px;">
          <div style="color:#991b1b;font-size:20px;font-weight:bold;">{escape(title)}</div>
          <div style="color:#7f1d1d;font-size:14px;margin-top:6px;">{escape(intro)}</div>
        </td></tr>
        <tr><td style="padding:24px 28px;">
          {body}
          <p style="font-size:14px;line-height:1.6;color:#374151;margin:24px 0 20px 0;">{escape(note)}</p>
          <table role="presentation" cellpadding="0" cellspacing="0"><tr>
            <td style="background-color:{COLOR_PRIMARY};border-radius:8px;">
              <a href="{escape(dashboard_url, quote=True)}" style="display:inline-block;padding:12px 24px;color:#ffffff;font-size:14px;font-weight:bold;text-decoration:none;">Open Data Sources Dashboard</a>
            </td>
          </tr></table>
        </td></tr>
        <tr><td style="background-color:#f3f4f6;padding:14px 28px;color:#6b7280;font-size:12px;">
          This is an automated message from the Labour Market Analysis crawler. Sent at {escape(_timestamp())}.
        </td></tr>
      </table>
    </td></tr>
  </table>
</body>
</html>
"""


def _failure_body(failed: List[FailedCrawler], run_id: int) -> str:
    rows = "".join(
        f"""\
<tr>
  <td style="padding:10px 12px;border-bottom:1px solid #e5e7eb;font-size:14px;font-weight:bold;color:#111827;">{escape(f.name)}</td>
  <td style="padding:10px 12px;border-bottom:1px solid #e5e7eb;font-size:13px;color:#4b5563;">{escape(_truncate(f.reason))}</td>
</tr>"""
        for f in failed
    )
    return f"""\
<p style="font-size:14px;color:#374151;margin:0 0 12px 0;">Crawler run <strong>#{escape(str(run_id))}</strong> finished with the failed crawlers below.</p>
<table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="border:1px solid #e5e7eb;border-radius:8px;border-collapse:separate;">
  <tr style="background-color:#f9fafb;">
    <th align="left" style="padding:10px 12px;font-size:12px;color:#6b7280;text-transform:uppercase;letter-spacing:0.05em;">Source</th>
    <th align="left" style="padding:10px 12px;font-size:12px;color:#6b7280;text-transform:uppercase;letter-spacing:0.05em;">Reason</th>
  </tr>
  {rows}
</table>"""


class EmailAlertService:
    """Emails the admins when crawlers fail. Every setting defaults to the value in
    config.py, which is read when the service is created."""

    def __init__(
        self,
        enabled: Optional[bool] = None,
        smtp_host: Optional[str] = None,
        smtp_port: Optional[int] = None,
        smtp_username: Optional[str] = None,
        smtp_password: Optional[str] = None,
        from_email: Optional[str] = None,
        recipients: Optional[List[str]] = None,
        dashboard_url: Optional[str] = None,
    ):
        self._enabled = ALERT_EMAILS_ENABLED if enabled is None else enabled
        self._smtp_host = smtp_host or SMTP_HOST
        self._smtp_port = smtp_port or SMTP_PORT
        self._smtp_username = smtp_username or SMTP_USERNAME
        self._smtp_password = smtp_password or SMTP_PASSWORD
        self._from_email = from_email or ALERT_FROM_EMAIL
        self._recipients = ALERT_RECIPIENTS if recipients is None else recipients
        self._dashboard_url = dashboard_url or DASHBOARD_URL

    def _is_ready(self) -> bool:
        if not self._enabled:
            logger.info("Email alert: Alert emails are disabled, skipping.")
            return False
        if not (
            self._smtp_host
            and self._smtp_username
            and self._smtp_password
            and self._recipients
        ):
            logger.warning(
                "Email alert: SMTP settings or recipients are missing, skipping."
            )
            return False
        return True

    # Blocking: it is run in a thread by _deliver
    def _send_message(self, message: EmailMessage) -> None:
        with smtplib.SMTP(
            self._smtp_host, self._smtp_port, timeout=SMTP_TIMEOUT_SECONDS
        ) as server:
            server.starttls()
            server.login(self._smtp_username, self._smtp_password)
            server.send_message(message)

    # Best effort: it never raises, so a mail problem cannot change or hide the
    # outcome of the crawl
    async def _deliver(self, subject: str, text: str, html: str) -> None:
        if not self._is_ready():
            return
        try:
            message = EmailMessage()
            message["Subject"] = subject
            message["From"] = self._from_email or self._smtp_username
            message["To"] = ", ".join(self._recipients)
            message.set_content(text)
            message.add_alternative(html, subtype="html")
            await asyncio.to_thread(self._send_message, message)
            logger.info(
                f"Email alert: Sent '{subject}' to {len(self._recipients)} admin(s)."
            )
        except Exception as e:
            logger.error(f"Email alert: Could not send '{subject}': {e}")

    # This function emails the admins about the crawlers that failed in a run. It
    # does nothing when no crawler failed
    async def send_failure_alert(
        self, failed: List[FailedCrawler], run_id: int
    ) -> None:
        if not failed:
            return

        count = len(failed)
        title = f"{count} crawler{'s' if count > 1 else ''} failed"
        note = (
            "Please deactivate the failed source(s) in the Data Sources dashboard "
            "until the issue is fixed, then activate them again."
        )
        text = "\n".join(
            [
                f"MarketLens: {title} in crawler run #{run_id}",
                "",
                *[f"- {f.name}: {_truncate(f.reason)}" for f in failed],
                "",
                note,
                f"Dashboard: {self._dashboard_url}",
            ]
        )
        html = _render_html(
            title,
            "Some crawlers did not finish successfully in the latest run.",
            _failure_body(failed, run_id),
            note,
            self._dashboard_url,
        )
        await self._deliver(
            f"[Labour Market Demand Application] {title} in crawler run #{run_id}",
            text,
            html,
        )

    # This function emails the admins that the whole crawler run crashed before the
    # crawlers could finish, so there is no list of failed crawlers
    async def send_crash_alert(self, error: Optional[BaseException]) -> None:
        reason = (
            _truncate(f"{type(error).__name__}: {error}") if error else "Unknown error"
        )
        title = "Crawler run crashed"
        note = (
            "No crawler results were recorded for this run. Please check the crawler "
            "logs and the backend, and review the sources in the Data Sources "
            "dashboard."
        )
        body = (
            '<p style="font-size:14px;color:#374151;margin:0 0 12px 0;">The crawler '
            "run stopped with an error:</p>"
            '<p style="font-size:13px;color:#991b1b;background-color:#fef2f2;'
            "border:1px solid #fecaca;"
            f'border-radius:8px;padding:12px;margin:0;">{escape(reason)}</p>'
        )
        text = "\n".join(
            [
                f"MarketLens: {title}",
                "",
                reason,
                "",
                note,
                f"Dashboard: {self._dashboard_url}",
            ]
        )
        html = _render_html(
            title,
            "The latest crawler run did not complete.",
            body,
            note,
            self._dashboard_url,
        )
        await self._deliver(f"[Labour Market Demand Application] {title}", text, html)
