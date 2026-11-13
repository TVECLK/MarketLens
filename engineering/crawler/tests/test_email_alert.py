from unittest.mock import AsyncMock

import pytest

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
    assert "2 crawlers failed" in message["Subject"]
    html = html_of(message)
    assert "ikman" in html and "rooster" in html and "RuntimeError: site down" in html
    assert "deactivate the failed source(s)" in html
    assert 'href="http://app.test/data-sources"' in html
    assert "ikman" in message.get_body(("plain",)).get_content()
    smtp.starttls.assert_called_once()
    smtp.login.assert_called_once_with("bot@test.lk", "secret")


async def test_escapes_html_and_truncates_long_reasons(smtp, service):
    await service.send_failure_alert(
        [FailedCrawler("<b>x</b>", "<script>" + "a" * 500)], 1
    )

    html = html_of(sent_message(smtp))
    assert "<script>" not in html and "&lt;b&gt;x&lt;/b&gt;" in html
    assert "a" * 250 not in html and "..." in html


async def test_a_smtp_failure_does_not_raise(smtp, service):
    smtp.send_message.side_effect = OSError("smtp down")
    await service.send_failure_alert([FailedCrawler("ikman", "down")], 1)


async def test_skips_when_disabled(smtp, make_service):
    await make_service(enabled=False).send_failure_alert(
        [FailedCrawler("ikman", "down")], 1
    )
    smtp.send_message.assert_not_called()


@pytest.mark.parametrize(
    "missing",
    [{"smtp_password": ""}, {"smtp_username": ""}, {"recipients": []}],
    ids=["password", "username", "recipients"],
)
async def test_skips_when_the_settings_are_missing(
    smtp, make_service, monkeypatch, missing
):
    # The service falls back to config.py for an empty setting, so those are cleared
    monkeypatch.setattr("utils.email_alert.SMTP_PASSWORD", None)
    monkeypatch.setattr("utils.email_alert.SMTP_USERNAME", None)
    await make_service(**missing).send_failure_alert(
        [FailedCrawler("ikman", "down")], 1
    )
    smtp.send_message.assert_not_called()


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
