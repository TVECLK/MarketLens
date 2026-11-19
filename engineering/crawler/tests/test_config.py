import pytest


@pytest.mark.parametrize("port", ["abc", "0", "99999", "-1", "25.5"])
def test_an_invalid_port_does_not_stop_the_import_when_alerts_are_disabled(
    load_config, port
):
    assert load_config("false", port).SMTP_PORT is None


@pytest.mark.parametrize("port", ["abc", "0", "99999", "-1", "25.5"])
def test_an_invalid_port_fails_the_import_when_alerts_are_enabled(load_config, port):
    with pytest.raises(ValueError, match=f"SMTP_PORT must be .*'{port}'"):
        load_config("true", port)


@pytest.mark.parametrize("enabled", ["true", "false"])
def test_a_blank_port_means_the_default_of_the_security_mode(load_config, enabled):
    assert load_config(enabled, "").SMTP_PORT is None


def test_a_valid_port_is_used_when_alerts_are_enabled(load_config):
    assert load_config("true", "2525").SMTP_PORT == 2525


@pytest.mark.parametrize(
    "name, setting",
    [
        ("SMTP_HOST", "SMTP_HOST"),
        ("SMTP_USERNAME", "SMTP_USERNAME"),
        ("SMTP_PASSWORD", "SMTP_PASSWORD"),
        ("ALERT_FROM_EMAIL", "ALERT_FROM_EMAIL"),
    ],
)
@pytest.mark.parametrize("blank", ["   ", "\t", " \n"])
def test_a_setting_of_only_whitespace_counts_as_not_set(
    load_config, name, setting, blank
):
    assert getattr(load_config(**{name: blank}), setting) is None


def test_whitespace_around_a_setting_is_removed(load_config):
    loaded = load_config(
        SMTP_HOST=" smtp.test ",
        SMTP_USERNAME=" bot@test.lk\n",
        ALERT_FROM_EMAIL="  alerts@test.lk ",
    )

    assert loaded.SMTP_HOST == "smtp.test"
    assert loaded.SMTP_USERNAME == "bot@test.lk"
    assert loaded.ALERT_FROM_EMAIL == "alerts@test.lk"


def test_a_password_is_kept_exactly_as_written(load_config):
    assert load_config(SMTP_PASSWORD=" secret \n").SMTP_PASSWORD == " secret \n"


def test_the_name_and_the_dashboard_url_fall_back_when_they_are_blank(load_config):
    loaded = load_config(APP_NAME="  ", DASHBOARD_URL=" ")

    assert loaded.APP_NAME == "Labour Market Analysis System"
    assert loaded.DASHBOARD_URL == "http://localhost:3000/data-sources"
