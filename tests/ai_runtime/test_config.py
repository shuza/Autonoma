import pytest

from autonoma_runtime.app import create_app
from autonoma_runtime.config import ProviderConfigurationError, create_provider
from autonoma_runtime.openai_compatible import OpenAICompatibleProvider
from autonoma_runtime.provider import MockProvider


SETTINGS = {
    "AUTONOMA_LLM_PROVIDER": "openai_compatible",
    "AUTONOMA_LLM_BASE_URL": "http://localhost:11434/v1",
    "AUTONOMA_LLM_MODEL": "local-model",
}


def test_default_and_explicit_mock() -> None:
    assert isinstance(create_provider({}), MockProvider)
    assert isinstance(create_provider({"AUTONOMA_LLM_PROVIDER": "mock"}), MockProvider)


def test_local_provider_without_credentials_or_pricing() -> None:
    assert isinstance(create_provider(SETTINGS), OpenAICompatibleProvider)


@pytest.mark.parametrize(
    "changes",
    [
        {"AUTONOMA_LLM_PROVIDER": ""},
        {"AUTONOMA_LLM_PROVIDER": "unknown"},
        {"AUTONOMA_LLM_BASE_URL": ""},
        {"AUTONOMA_LLM_MODEL": " "},
        {"AUTONOMA_LLM_TIMEOUT_SECONDS": ""},
        {"AUTONOMA_LLM_TIMEOUT_SECONDS": "invalid"},
        {"AUTONOMA_LLM_TIMEOUT_SECONDS": "0"},
        {"AUTONOMA_LLM_TIMEOUT_SECONDS": "-1"},
        {"AUTONOMA_LLM_TIMEOUT_SECONDS": "NaN"},
        {"AUTONOMA_LLM_TIMEOUT_SECONDS": "Infinity"},
        {"AUTONOMA_LLM_BASE_URL": "http://remote.example/v1"},
        {"AUTONOMA_LLM_BASE_URL": "https://user:secret@example.com/v1"},
        {"AUTONOMA_LLM_API_KEY": ""},
        {"AUTONOMA_LLM_API_KEY": "secret\nheader"},
        {"AUTONOMA_LLM_INPUT_USD_PER_MILLION_TOKENS": "1"},
        {"AUTONOMA_LLM_OUTPUT_USD_PER_MILLION_TOKENS": "1"},
    ],
)
def test_invalid_configuration_is_explicit(changes) -> None:
    with pytest.raises(ProviderConfigurationError) as caught:
        create_provider(SETTINGS | changes)
    assert "secret" not in str(caught.value)


@pytest.mark.parametrize("field", ["BASE_URL", "MODEL"])
def test_missing_required_configuration(field: str) -> None:
    settings = dict(SETTINGS)
    del settings[f"AUTONOMA_LLM_{field}"]
    with pytest.raises(ProviderConfigurationError, match="required"):
        create_provider(settings)


@pytest.mark.parametrize("rate", ["", "invalid", "-1", "NaN", "Infinity"])
@pytest.mark.parametrize("side", ["INPUT", "OUTPUT"])
def test_invalid_pricing(rate: str, side: str) -> None:
    settings = SETTINGS | {
        "AUTONOMA_LLM_INPUT_USD_PER_MILLION_TOKENS": "0",
        "AUTONOMA_LLM_OUTPUT_USD_PER_MILLION_TOKENS": "0",
        f"AUTONOMA_LLM_{side}_USD_PER_MILLION_TOKENS": rate,
    }
    with pytest.raises(ProviderConfigurationError, match="pricing"):
        create_provider(settings)


def test_injected_provider_bypass_environment_even_when_falsy(monkeypatch) -> None:
    class FalsyMock(MockProvider):
        def __bool__(self) -> bool:
            return False

    monkeypatch.setenv("AUTONOMA_LLM_PROVIDER", "invalid")
    provider = FalsyMock()
    assert create_app(provider).provider is provider


def test_invalid_environment_fails_during_app_construction(monkeypatch) -> None:
    monkeypatch.setenv("AUTONOMA_LLM_PROVIDER", "openai_compatible")
    with pytest.raises(ProviderConfigurationError, match="required"):
        create_app()