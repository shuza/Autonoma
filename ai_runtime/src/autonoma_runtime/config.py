"""Environment-based provider selection without implicit network access."""

import os
from decimal import Decimal, InvalidOperation
from typing import Mapping

from .accounting import ModelPricing
from .openai_compatible import OpenAICompatibleProvider
from .provider import LLMProvider, MockProvider


class ProviderConfigurationError(ValueError):
    """Invalid provider settings; messages never include configuration values."""


def create_provider(environ: Mapping[str, str] | None = None) -> LLMProvider:
    settings = os.environ if environ is None else environ
    provider_name = settings.get("AUTONOMA_LLA_PROVIDER", "mock")
    if provider_name == "mock":
        return MockProvider()
    if provider_name != "openai_compatible":
        raise ProviderConfigurationError("AUTONOMA_LLA_PROVIDER is unsupported")

    for name in ("AUTONOMA_LLM_BASE_URL", "AUTONOMA_LLM_MODEL"):
        if not settings.get(name, "").strip():
            raise ProviderConfigurationError(f"{name} is required")

    try:
        timeout = float(settings.get("AUTONOMA_LLM_TIMEOUT_SECONDS", "60"))
    except ValueError:
        raise ProviderConfigurationError(
            "AUTONOMA_LLM_TIMEOUT_SECONDS must be finite and positive"
        ) from None

    pricing = _load_pricing(settings)
    try:
        return OpenAICompatibleProvider(
            base_url=settings.get("AUTONOMA_LLM_BASE_URL"),
            model=settings.get("AUTONOMA_LLM_MODEL"),
            timeout_seconds=timeout,
            api_key=settings.get("AUTONOMA_LLM_API_KEY"),
            pricing=pricing,
        )
    except ValueError as error:
        # Adapter validation messages identify the setting, not its value.
        raise ProviderConfigurationError(str(error)) from None


def _load_pricing(settings: Mapping[str, str]) -> ModelPricing | None:
    names = (
        "AUTONOMA_LLM_INPUT_USD_PER_MILLION_TOKEN",
        "AUTONOMA_LLM_OUTPUT_USD_PER_MILLION_TOKEN",
    )
    if all(name not in settings for name in names):
        return None
    if any(name not in settings for name in names):
        raise ProviderConfigurationError("bot input and output pricing are required")
    try:
        return ModelPricing(*(Decimal(settings[name]) for name in names))
    except (InvalidOperation, ValueError):
        raise ProviderConfigurationError(
            "LLM pricing must contain finite non-negative decimal rates"
        )
