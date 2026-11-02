import pytest

from app import config

pytestmark = pytest.mark.unit


@pytest.fixture
def clean_env(monkeypatch):
    monkeypatch.setattr(config, "load_dotenv", lambda: None)
    for name in ("OPENROUTER_API_KEY", "OPENROUTER_MODEL", "OPENROUTER_RESPONSE_FORMAT",
                 "OPENROUTER_MAX_TOKENS", "LLM_TIMEOUT_SECONDS", "AI_SERVICE_KEY"):
        monkeypatch.delenv(name, raising=False)
    monkeypatch.setenv("OPENROUTER_API_KEY", "test-only-openrouter-api-key-123456789")
    monkeypatch.setenv("AI_SERVICE_KEY", "test-only-internal-service-key-123456789")
    monkeypatch.setenv("OPENROUTER_MODEL", "nvidia/fixture-nemotron:free")
    return monkeypatch


def test_defaults(clean_env):
    settings = config.Settings.from_env()
    assert settings.model == "nvidia/fixture-nemotron:free"
    assert settings.timeout_seconds == 120
    assert settings.response_format == "prompt"
    assert settings.max_tokens == 4096


def test_custom_values_normalized(clean_env):
    clean_env.setenv("OPENROUTER_MODEL", "  nvidia/another-nemotron:free  ")
    clean_env.setenv("OPENROUTER_RESPONSE_FORMAT", " JSON_SCHEMA ")
    clean_env.setenv("OPENROUTER_MAX_TOKENS", "2048")
    clean_env.setenv("LLM_TIMEOUT_SECONDS", "20")
    settings = config.Settings.from_env()
    assert settings.model == "nvidia/another-nemotron:free"
    assert settings.response_format == "json_schema"
    assert settings.max_tokens == 2048
    assert settings.timeout_seconds == 20


@pytest.mark.parametrize("secret", ["", "short", " " * 40, "REPLACE_WITH_A_RANDOM_SECRET_AT_LEAST_32_CHARACTERS"])
def test_invalid_service_secret_rejected(clean_env, secret):
    clean_env.setenv("AI_SERVICE_KEY", secret)
    with pytest.raises(ValueError, match="AI_SERVICE_KEY"):
        config.Settings.from_env()


@pytest.mark.parametrize("secret", ["", "short", "REPLACE_WITH_YOUR_OPENROUTER_API_KEY"])
def test_invalid_provider_key_rejected(clean_env, secret):
    clean_env.setenv("OPENROUTER_API_KEY", secret)
    with pytest.raises(ValueError, match="OPENROUTER_API_KEY"):
        config.Settings.from_env()


@pytest.mark.parametrize("model", ["", "qwen3.5:4b", "nvidia/nemotron-paid", "other/nemotron:free", "nvidia/unrelated:free"])
def test_invalid_or_paid_model_rejected(clean_env, model):
    clean_env.setenv("OPENROUTER_MODEL", model)
    with pytest.raises(ValueError, match="OPENROUTER_MODEL"):
        config.Settings.from_env()


@pytest.mark.parametrize("value", ["0", "301", "nan", "inf", "not-a-number"])
def test_invalid_timeout_rejected(clean_env, value):
    clean_env.setenv("LLM_TIMEOUT_SECONDS", value)
    with pytest.raises(ValueError):
        config.Settings.from_env()


@pytest.mark.parametrize("value", ["0", "16385", "not-an-integer"])
def test_invalid_token_budget_rejected(clean_env, value):
    clean_env.setenv("OPENROUTER_MAX_TOKENS", value)
    with pytest.raises(ValueError):
        config.Settings.from_env()


def test_invalid_format_rejected(clean_env):
    clean_env.setenv("OPENROUTER_RESPONSE_FORMAT", "unsupported")
    with pytest.raises(ValueError, match="OPENROUTER_RESPONSE_FORMAT"):
        config.Settings.from_env()


def test_provider_key_must_not_equal_internal_key(clean_env):
    clean_env.setenv("OPENROUTER_API_KEY", "test-only-internal-service-key-123456789")
    with pytest.raises(ValueError, match="must be different"):
        config.Settings.from_env()


def test_keys_not_exposed_by_settings_repr(clean_env):
    settings = config.Settings.from_env()
    assert settings.api_key not in repr(settings)
    assert settings.service_key not in repr(settings)
