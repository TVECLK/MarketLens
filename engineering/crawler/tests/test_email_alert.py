import logging
import ssl
from email.utils import parsedate_to_datetime
from unittest.mock import AsyncMock, MagicMock

import httpx
import pytest

from utils import email_alert
from utils.email_alert import FailedCrawler


def sent_message(smtp):
    smtp.send_message.assert_called_once()
    return smtp.send_message.call_args.args[0]


def html_of(message):
    return message.get_body(("html",)).get_content()


async def test_no_email_when_no_crawler_failed(smtp, service):
    await service.send_failure_alert([], 7)
    smtp.send_message.assert_not_called()


async def test_sends_one_email_listing_the_failed_crawlers(smtp, service):
    await service.send_failure_alert(
        [
            FailedCrawler("ikman", "RuntimeError: site down"),
            FailedCrawler("rooster", "x"),
        ],
        7,
    )

    message = sent_message(smtp)
    assert message["To"] == "a@test.lk, b@test.lk"
    assert message["Subject"].startswith("[Test App] 2 crawlers failed")
    assert "Test App" in message.get_body(("plain",)).get_content()
    html = html_of(message)
    assert "ikman" in html and "rooster" in html and "RuntimeError: site down" in html
    assert "Some crawlers did not finish" in html
    assert "deactivate the failed sources" in html and "issues are fixed" in html
    assert "failed crawlers below" in html
    assert 'href="http://app.test/data-sources"' in html
    assert "ikman" in message.get_body(("plain",)).get_content()
    smtp.starttls.assert_called_once()
    # The TLS connection must verify the certificate and hostname of the server
    context = smtp.starttls.call_args.kwargs["context"]
    assert context.verify_mode == ssl.CERT_REQUIRED and context.check_hostname
    smtp.login.assert_called_once_with("bot@test.lk", "secret")


async def test_the_email_has_a_date_and_a_message_id(smtp, make_service):
    service = make_service(from_email="MarketLens <bot@test.lk>")

    await service.send_failure_alert([FailedCrawler("ikman", "down")], 1)

    message = sent_message(smtp)
    assert message["Date"]
    assert parsedate_to_datetime(message["Date"]).tzinfo is not None
    assert message["Message-ID"].startswith("<") and message["Message-ID"].endswith(
        "@test.lk>"
    )


async def test_every_email_gets_its_own_message_id(smtp, service):
    await service.send_failure_alert([FailedCrawler("ikman", "down")], 1)
    first = sent_message(smtp)["Message-ID"]
    smtp.send_message.reset_mock()

    await service.send_failure_alert([FailedCrawler("ikman", "down")], 1)

    assert sent_message(smtp)["Message-ID"] != first


async def test_the_wording_is_singular_for_one_failed_crawler(smtp, service):
    await service.send_failure_alert([FailedCrawler("ikman", "down")], 1)

    message = sent_message(smtp)
    html = html_of(message)
    assert message["Subject"].startswith("[Test App] 1 crawler failed")
    assert "A crawler did not finish" in html
    assert "deactivate the failed source in the dashboard" in html
    assert "the issue is fixed, then activate it again" in html
    assert "failed crawler below" in html
    for plural in ("Some crawlers", "failed sources", "issues", "them", "(s)"):
        assert plural not in html


async def test_escapes_html_and_truncates_long_reasons(smtp, service):
    await service.send_failure_alert(
        [FailedCrawler("<b>x</b>", "<script>" + "a" * 500)], 1
    )

    html = html_of(sent_message(smtp))
    assert "<script>" not in html and "&lt;b&gt;x&lt;/b&gt;" in html
    assert "a" * 250 not in html and "..." in html


async def test_logs_the_recipients_the_server_refused(smtp, service, caplog):
    smtp.send_message.return_value = {"b@test.lk": (550, b"No such user")}

    with caplog.at_level(logging.INFO):
        await service.send_failure_alert([FailedCrawler("ikman", "down")], 1)

    warnings = [r for r in caplog.records if r.levelno == logging.WARNING]
    assert len(warnings) == 1
    assert "1 of 2" in warnings[0].message
    assert "b@test.lk (550 No such user)" in warnings[0].message
    # Only the accepted recipient is counted as sent
    assert "to 1 admin(s)" in caplog.text


async def test_logs_no_warning_when_every_recipient_is_accepted(smtp, service, caplog):
    with caplog.at_level(logging.INFO):
        await service.send_failure_alert([FailedCrawler("ikman", "down")], 1)

    assert not [r for r in caplog.records if r.levelno >= logging.WARNING]
    assert "to 2 admin(s)" in caplog.text


async def test_a_smtp_failure_does_not_raise(smtp, service):
    smtp.send_message.side_effect = OSError("smtp down")
    await service.send_failure_alert([FailedCrawler("ikman", "down")], 1)


async def test_skips_quietly_when_disabled(smtp, make_service, caplog):
    with caplog.at_level(logging.DEBUG):
        await make_service(enabled=False).send_failure_alert(
            [FailedCrawler("ikman", "down")], 1
        )

    smtp.send_message.assert_not_called()
    assert caplog.records == []


@pytest.mark.parametrize(
    "missing",
    [
        {"smtp_host": ""},
        {"smtp_password": ""},
        {"smtp_username": ""},
        {"recipients": []},
    ],
    ids=["host", "password-only-username", "username-only-password", "recipients"],
)
async def test_skips_when_the_settings_are_missing(smtp, make_service, missing):
    await make_service(**missing).send_failure_alert(
        [FailedCrawler("ikman", "down")], 1
    )
    smtp.send_message.assert_not_called()


async def test_ssl_security_connects_with_tls_from_the_start(smtp, make_service):
    service = make_service(smtp_security="ssl", smtp_port=None)

    await service.send_failure_alert([FailedCrawler("ikman", "down")], 1)

    # The default port of ssl, and a context that verifies the server
    args, kwargs = email_alert.smtplib.SMTP_SSL.call_args
    assert args == ("smtp.test", 465)
    assert kwargs["context"].check_hostname
    email_alert.smtplib.SMTP.assert_not_called()
    smtp.starttls.assert_not_called()
    smtp.login.assert_called_once_with("bot@test.lk", "secret")
    sent_message(smtp)


async def test_none_security_sends_without_encryption(smtp, make_service):
    service = make_service(smtp_security="none", smtp_port=None)

    await service.send_failure_alert([FailedCrawler("ikman", "down")], 1)

    assert email_alert.smtplib.SMTP.call_args.args == ("smtp.test", 25)
    smtp.starttls.assert_not_called()
    sent_message(smtp)


async def test_starttls_is_the_security_by_default(smtp, make_service):
    service = make_service(smtp_security="starttls", smtp_port=None)

    await service.send_failure_alert([FailedCrawler("ikman", "down")], 1)

    assert email_alert.smtplib.SMTP.call_args.args == ("smtp.test", 587)
    smtp.starttls.assert_called_once()


async def test_a_server_without_login_gets_the_email_without_logging_in(
    smtp, make_service
):
    service = make_service(smtp_username="", smtp_password="")

    await service.send_failure_alert([FailedCrawler("ikman", "down")], 1)

    smtp.login.assert_not_called()
    assert sent_message(smtp)["From"] == "bot@test.lk"


async def test_skips_an_unknown_security(smtp, make_service, caplog):
    service = make_service(smtp_security="tls")

    with caplog.at_level(logging.ERROR):
        await service.send_failure_alert([FailedCrawler("ikman", "down")], 1)

    smtp.send_message.assert_not_called()
    assert "'tls'" in caplog.text


async def test_skips_when_there_is_no_sender(smtp, make_service):
    service = make_service(smtp_username="", smtp_password="", from_email="")

    await service.send_failure_alert([FailedCrawler("ikman", "down")], 1)

    smtp.send_message.assert_not_called()


async def test_the_sender_falls_back_to_the_username(smtp, make_service):
    service = make_service(from_email=None, smtp_username="user@test.lk")

    await service.send_failure_alert([FailedCrawler("ikman", "down")], 1)

    assert sent_message(smtp)["From"] == "user@test.lk"


async def test_an_error_while_building_the_email_does_not_raise(
    smtp, service, monkeypatch, caplog
):
    def broken(*args, **kwargs):
        raise ValueError("template broke")

    monkeypatch.setattr("utils.email_alert._failure_body", broken)

    with caplog.at_level(logging.ERROR):
        await service.send_failure_alert([FailedCrawler("ikman", "down")], 1)

    smtp.send_message.assert_not_called()
    assert "template broke" in caplog.text


async def test_the_run_completes_when_the_alert_cannot_be_built(
    manager, smtp, make_service, monkeypatch
):
    manager._start_run = AsyncMock(return_value=7)
    manager._finalize_run = AsyncMock()
    service = make_service()
    broken = MagicMock(side_effect=ValueError("template broke"))
    monkeypatch.setattr(service, "_failure_email", broken)
    manager._alert_service = service

    result = await manager.run_all_crawlers(concurrent=True)

    broken.assert_called_once()  # it was reached, and its error was contained
    smtp.send_message.assert_not_called()
    assert sorted(result) == ["source_a", "source_b"]


async def test_no_email_when_every_active_source_has_a_crawler(smtp, service):
    await service.send_missing_crawlers_alert([])
    smtp.send_message.assert_not_called()


async def test_the_missing_crawlers_email_lists_the_sources(smtp, service):
    await service.send_missing_crawlers_alert(["glassdoor", "<b>x</b>"])

    message = sent_message(smtp)
    html = html_of(message)
    assert message["Subject"] == "[Test App] 2 active sources have no crawler"
    assert "glassdoor" in html and "&lt;b&gt;x&lt;/b&gt;" in html
    assert "ask a developer to add their crawlers" in html
    assert 'href="http://app.test/data-sources"' in html
    assert "glassdoor" in message.get_body(("plain",)).get_content()


async def test_the_missing_crawlers_wording_is_singular_for_one_source(smtp, service):
    await service.send_missing_crawlers_alert(["glassdoor"])

    message = sent_message(smtp)
    html = html_of(message)
    assert message["Subject"] == "[Test App] 1 active source has no crawler"
    assert "The source below is active" in html
    assert "deactivate it in the dashboard" in html and "add its crawler" in html
    # The URL and the button also contain the words "sources" and "Sources"
    wording = html.replace("data-sources", "").replace("Data Sources", "")
    for plural in ("sources", "them", "their", "(s)"):
        assert plural not in wording


@pytest.mark.parametrize("concurrent", [True, False])
async def test_manager_alerts_about_the_active_sources_without_a_crawler(
    manager, alert_service, concurrent
):
    manager._sources_without_crawler = ["glassdoor"]
    manager._start_run = AsyncMock(return_value=7)
    manager._finalize_run = AsyncMock()

    await manager.run_all_crawlers(concurrent=concurrent)

    alert_service.send_missing_crawlers_alert.assert_awaited_once_with(["glassdoor"])


async def test_manager_alerts_even_when_no_active_source_has_a_crawler(
    manager, alert_service
):
    manager.fetch_active_crawlers.return_value = {}
    manager._sources_without_crawler = ["glassdoor"]
    manager._start_run = AsyncMock(return_value=7)

    result = await manager.run_all_crawlers()

    assert result is None
    manager._start_run.assert_not_awaited()
    alert_service.send_missing_crawlers_alert.assert_awaited_once_with(["glassdoor"])


async def test_crash_alert_describes_the_error(smtp, service):
    await service.send_crash_alert(RuntimeError("backend <down>"))

    message = sent_message(smtp)
    assert "crashed" in message["Subject"]
    assert "RuntimeError: backend &lt;down&gt;" in html_of(message)


# ---------- wiring in the manager ----------


@pytest.mark.parametrize("concurrent", [True, False])
async def test_manager_alerts_only_about_the_failed_active_crawlers(
    manager, alert_service, concurrent
):
    manager._start_run = AsyncMock(return_value=7)
    manager._finalize_run = AsyncMock()

    await manager.run_all_crawlers(concurrent=concurrent)

    failed, run_id = alert_service.send_failure_alert.await_args.args
    assert run_id == 7
    assert [(f.name, f.reason) for f in failed] == [
        ("source_c", "RuntimeError: site down")
    ]


async def test_manager_gives_an_empty_list_when_nothing_failed(
    manager, alert_service, make_crawler_class
):
    manager.fetch_active_crawlers.return_value = {
        "source_a": make_crawler_class("source_a")
    }
    manager._start_run = AsyncMock(return_value=7)
    manager._finalize_run = AsyncMock()

    await manager.run_all_crawlers(concurrent=True)

    assert alert_service.send_failure_alert.await_args.args[0] == []


async def test_manager_still_alerts_when_finalizing_the_run_fails(
    manager, alert_service
):
    manager._start_run = AsyncMock(return_value=7)
    manager._finalize_run = AsyncMock(side_effect=httpx.ConnectError("backend down"))

    with pytest.raises(httpx.ConnectError, match="backend down"):
        await manager.run_all_crawlers(concurrent=True)

    failed, run_id = alert_service.send_failure_alert.await_args.args
    assert run_id == 7
    assert [f.name for f in failed] == ["source_c"]
