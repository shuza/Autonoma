"""Provider abstractions for the AI runtime"""

from dataclasses import dataclass
from decimal import Decimal
from typing import Any, Mapping, Protocol

from .accounting import ModelPricing, ProviderAccounting, TokenUsage


class ProviderError(Exception):
    """Sanitized provider transport or response failure."""


class ProviderResponseError(ProviderError):
    """Rejected completion, optionally carrying content-free accounting."""

    def __init__(
            self, message: str, *, accounting: ProviderAccounting | None = None
    ):
        super().__init__(message)
        self.accounting = accounting


@dataclass(frozen=True)
class ProviderRequest:
    prompt: str
    response_schema: Mapping[str, Any] | None = None

    def __post_init__(self) -> None:
        if not isinstance(self.prompt, str) or not self.prompt.strip():
            raise ValueError("prompt must be a non-empty string")


@dataclass(frozen=True)
class ProviderResponse:
    content: str
    provider_name: str
    usage: TokenUsage | None = None
    pricing: ModelPricing | None = None

    @property
    def cost_usd(self) -> Decimal | None:
        if self.usage is None or self.pricing is None:
            return None
        return self.pricing.calculate_cost(self.usage)


class LLMProvider(Protocol):
    def generate(self, request: ProviderRequest) -> ProviderResponse:
        """Generate a provider response."""


class MockProvider:
    """Deterministic provider used for local development and tests"""

    name = "mock-provider"

    def __init__(
            self,
            responses: Mapping[str, str] | None = None,
            *,
            usage: Mapping[str, TokenUsage] | None = None,
            pricing: ModelPricing = ModelPricing()
    ) -> None:
        self._responses = dict(responses or {})
        self._usage = dict(usage or {})
        if not all(isinstance(value, TokenUsage) for value in self._usage.values()):
            raise ValueError("mock usage fixtures must be TokenUsage values")
        if not isinstance(pricing, ModelPricing):
            raise ValueError("mock pricing must be ModelPricing")
        self._pricing = pricing

    def generate(self, request: ProviderRequest) -> ProviderResponse:
        normalized_prompt = request.prompt.strip()
        return ProviderResponse(
            content=self._responses.get(
                normalized_prompt,
                f"mock-response:{normalized_prompt}",
            ),
            provider_name=self.name,
            usage=self._usage.get(normalized_prompt),
            pricing=self._pricing,
        )
