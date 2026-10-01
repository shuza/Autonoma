"""Deterministic provider abstractions for the AI runtime"""

from dataclasses import dataclass
from typing import Protocol


@dataclass(frozen=True)
class ProviderRequest:
    prompt: str


@dataclass(frozen=True)
class ProviderResponse:
    content: str
    provider_name: str


class LLMProvider(Protocol):
    def generate(self, request: ProviderRequest) -> ProviderResponse:
        """Generate a deterministic provider response"""


class MockProvider:
    """Deterministic provider used for local development and tests"""

    name = "mock-provider"

    def generate(self, request: ProviderRequest) -> ProviderResponse:
        normalized_prompt = request.prompt.strip()
        return ProviderResponse(
            content=f"mock-response:{normalized_prompt}",
            provider_name=self.name
        )
