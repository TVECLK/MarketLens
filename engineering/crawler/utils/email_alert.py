import asyncio
import logging
import smtplib
import ssl
from dataclasses import dataclass
from datetime import datetime
from email.message import EmailMessage
from email.utils import formatdate, make_msgid, parseaddr
from html import escape
from typing import Callable, Dict, List, Optional, Tuple

from config import (
    APP_NAME,
    ALERT_EMAILS_ENABLED,
    ALERT_FROM_EMAIL,
    ALERT_RECIPIENTS,
    DASHBOARD_URL,
    SMTP_HOST,
    SMTP_PASSWORD,
    SMTP_PORT,
    SMTP_SECURITY,
    SMTP_USERNAME,
)

logger = logging.getLogger(__name__)

MAX_REASON_LENGTH = 200
SMTP_TIMEOUT_SECONDS = 30
# The port used when SMTP_PORT is not set, by the way the connection is secured
DEFAULT_SMTP_PORTS = {"starttls": 587, "ssl": 465, "none": 25}

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
    app_name: str,
    title: str,
    intro: str,
    body: str,
    note: str,
    dashboard_url: str,
) -> str:
    return f"""\
<!DOCTYPE html>
<html>
<body style="margin:0;padding:0;background-color:#f9fafb;font-family:Arial,Helvetica,sans-serif;color:#111827;">
  <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="background-color:#f9fafb;padding:24px 0;">
    <tr><td align="center">
      <table role="presentation" width="600" cellpadding="0" cellspacing="0" style="max-width:600px;width:100%;background-color:#ffffff;border:1px solid #e5e7eb;border-radius:12px;overflow:hidden;">
        <tr><td style="background-color:{COLOR_DARK};padding:20px 28px;">
          <span style="color:#ffffff;font-size:18px;font-weight:bold;">{escape(app_name)}</span>
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
          This is an automated message from the {escape(app_name)} crawler. Sent at {escape(_timestamp())}.
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
<p style="font-size:14px;color:#374151;margin:0 0 12px 0;">Crawler run <strong>#{escape(str(run_id))}</strong> finished with the failed {"crawlers" if len(failed) > 1 else "crawler"} below.</p>
<table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="border:1px solid #e5e7eb;border-radius:8px;border-collapse:separate;">
  <tr style="background-color:#f9fafb;">
    <th align="left" style="padding:10px 12px;font-size:12px;color:#6b7280;text-transform:uppercase;letter-spacing:0.05em;">Source</th>
    <th align="left" style="padding:10px 12px;font-size:12px;color:#6b7280;text-transform:uppercase;letter-spacing:0.05em;">Reason</th>
  </tr>
  {rows}
</table>"""


class EmailAlertService:
    """Send emails when crawlers fail. The settings come from config.py, which reads
    the environment once, when it is imported."""

    def __init__(self):
        self._enabled = ALERT_EMAILS_ENABLED
        self._smtp_host = SMTP_HOST
        self._smtp_security = SMTP_SECURITY
        # Without a port, the default one of the security mode is used
        self._smtp_port = SMTP_PORT or DEFAULT_SMTP_PORTS.get(SMTP_SECURITY)
        self._smtp_username = SMTP_USERNAME
        self._smtp_password = SMTP_PASSWORD
        # Without a sender, the emails are sent from the username
        self._from_email = ALERT_FROM_EMAIL or SMTP_USERNAME
        self._recipients = ALERT_RECIPIENTS
        self._dashboard_url = DASHBOARD_URL
        self._app_name = APP_NAME

    def _is_ready(self) -> bool:
        # Turning the alerts off is a choice, so it is not logged
        if not self._enabled:
            return False
        if self._smtp_security not in DEFAULT_SMTP_PORTS:
            logger.error(
                f"Email alert: SMTP_SECURITY '{self._smtp_security}' is not one of "
                f"{', '.join(DEFAULT_SMTP_PORTS)}, skipping."
            )
            return False
        if not (self._smtp_host and self._recipients and self._from_email):
            logger.warning(
                "Email alert: SMTP host, sender or recipients are missing, skipping."
            )
            return False
        # The login is optional, but a username without a password (or the other way
        # round) is a mistake
        if bool(self._smtp_username) != bool(self._smtp_password):
            logger.warning(
                "Email alert: SMTP_USERNAME and SMTP_PASSWORD must be set together, "
                "skipping."
            )
            return False
        return True

    # Blocking: it is run in a thread by _deliver. It returns the recipients the
    # server refused (address -> (code, reason)), which is empty when all were
    # accepted. smtplib only raises when every recipient is refused.
    def _send_message(self, message: EmailMessage) -> Dict[str, Tuple[int, bytes]]:
        # A default context verifies the server certificate and hostname, so the
        # credentials of login() are never sent to an impostor
        context = ssl.create_default_context()
        if self._smtp_security == "ssl":
            connection = smtplib.SMTP_SSL(
                self._smtp_host,
                self._smtp_port,
                timeout=SMTP_TIMEOUT_SECONDS,
                context=context,
            )
        else:
            connection = smtplib.SMTP(
                self._smtp_host, self._smtp_port, timeout=SMTP_TIMEOUT_SECONDS
            )
        with connection as server:
            if self._smtp_security == "starttls":
                server.starttls(context=context)
            if self._smtp_username:
                server.login(self._smtp_username, self._smtp_password)
            return server.send_message(message)

    # The domain of "Name <user@domain>" or "user@domain", None if it has none
    @staticmethod
    def _sender_domain(sender: str) -> Optional[str]:
        _, address = parseaddr(sender)
        return address.partition("@")[2] or None

    @staticmethod
    def _decode(reason: bytes | str) -> str:
        if isinstance(reason, bytes):
            return reason.decode("utf-8", errors="replace")
        return str(reason)

    # Best effort: nothing in here can raise, building the email included, so a mail
    # problem cannot change or hide the outcome of the crawl. build_email returns the
    # subject, the plain text and the HTML.
    async def _deliver(
        self, label: str, build_email: Callable[[], Tuple[str, str, str]]
    ) -> None:
        try:
            if not self._is_ready():
                return
            subject, text, html = build_email()
            message = EmailMessage()
            message["Subject"] = subject
            message["From"] = self._from_email
            message["To"] = ", ".join(self._recipients)
            # smtplib adds neither header and not every server fills them in. The
            # Message-ID uses the sender's domain, as the hostname of a pod is random
            message["Date"] = formatdate(localtime=True)
            message["Message-ID"] = make_msgid(
                domain=self._sender_domain(self._from_email)
            )
            message.set_content(text)
            message.add_alternative(html, subtype="html")
            refused = await asyncio.to_thread(self._send_message, message)
            if refused:
                details = ", ".join(
                    f"{address} ({code} {self._decode(reason)})"
                    for address, (code, reason) in refused.items()
                )
                logger.warning(
                    f"Email alert: The server refused {len(refused)} of "
                    f"{len(self._recipients)} recipient(s) of '{subject}': {details}"
                )
            accepted = len(self._recipients) - len(refused)
            if accepted > 0:
                logger.info(f"Email alert: Sent '{subject}' to {accepted} admin(s).")
        except Exception as e:
            logger.error(f"Email alert: Could not send the {label}: {e}")

    def _failure_email(
        self, failed: List[FailedCrawler], run_id: int
    ) -> Tuple[str, str, str]:
        if len(failed) > 1:
            title = f"{len(failed)} crawlers failed"
            intro = "Some crawlers did not finish successfully in the latest run."
            note = (
                "Please deactivate the failed sources in the dashboard until the "
                "issues are fixed, then activate them again."
            )
        else:
            title = "1 crawler failed"
            intro = "A crawler did not finish successfully in the latest run."
            note = (
                "Please deactivate the failed source in the dashboard until the "
                "issue is fixed, then activate it again."
            )
        text = "\n".join(
            [
                f"{self._app_name}: {title} in crawler run #{run_id}",
                "",
                *[f"- {f.name}: {_truncate(f.reason)}" for f in failed],
                "",
                note,
                f"Dashboard: {self._dashboard_url}",
            ]
        )
        html = _render_html(
            self._app_name,
            title,
            intro,
            _failure_body(failed, run_id),
            note,
            self._dashboard_url,
        )
        return f"[{self._app_name}] {title} in crawler run #{run_id}", text, html

    def _crash_email(self, error: Optional[BaseException]) -> Tuple[str, str, str]:
        reason = (
            _truncate(f"{type(error).__name__}: {error}") if error else "Unknown error"
        )
        title = "Crawler run crashed"
        note = (
            "Please check the crawler "
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
                f"{self._app_name}: {title}",
                "",
                reason,
                "",
                note,
                f"Dashboard: {self._dashboard_url}",
            ]
        )
        html = _render_html(
            self._app_name,
            title,
            "The latest crawler run did not complete.",
            body,
            note,
            self._dashboard_url,
        )
        return f"[{self._app_name}] {title}", text, html

    # This function emails the admins about the crawlers that failed in a run. It
    # does nothing when no crawler failed, and it never raises
    async def send_failure_alert(
        self, failed: List[FailedCrawler], run_id: int
    ) -> None:
        if not failed:
            return
        await self._deliver(
            "failure alert", lambda: self._failure_email(failed, run_id)
        )

    # This function sends an email that the whole crawler run crashed before the
    # crawlers could finish, so there is no list of failed crawlers. It never raises
    async def send_crash_alert(self, error: Optional[BaseException]) -> None:
        await self._deliver("crash alert", lambda: self._crash_email(error))
