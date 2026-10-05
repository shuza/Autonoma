"""Provider abstractions for the AI runtime"""

from dataclasses import dataclass
from typing import Any, Mapping, Protocol


@dataclass(frozen=True)
class ProviderRequest:
    prompt: str
    response_schema: Mapping[str, Any] | None = None


@dataclass(frozen=True)
class ProviderResponse:
    content: str
    provider_name: str


class LLMProvider(Protocol):
    def generate(self, request: ProviderRequest) -> ProviderResponse:
        """Generate a provider response."""


class MockProvider:
    """Deterministic provider used for local development and tests"""

    name = "mock-provider"

    def __init__(self, responses: Mapping[str, str] | None = None) -> None:
        self._responses = dict(responses or {})

    def generate(self, request: ProviderRequest) -> ProviderResponse:
        normalized_prompt = request.prompt.strip()
        return ProviderResponse(
            content=self._responses.get(
                normalized_prompt,
                f"mock-response:{normalized_prompt}",
            ),
            provider_name=self.name
        )
