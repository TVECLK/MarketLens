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
