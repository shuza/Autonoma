"""Minimal AI runtime entrypoint for Phase 0."""

from dataclasses import dataclass
from typing import Any, Final, Mapping

from .provider import LLMProvider, MockProvider, ProviderRequest, ProviderResponse
from .structured_output import StructuredOutputResult, StructuredOutputGenerator

READY_STATUS: Final[str] = "ready"
RUNTIME_NAME: Final[str] = "autonoma-ai-runtime"


@dataclass(frozen=True)
class RuntimeApp:
    provider: LLMProvider
    name: str = RUNTIME_NAME
    status: str = READY_STATUS


def create_app(provider: LLMProvider | None = None) -> RuntimeApp:
    return RuntimeApp(provider=provider or MockProvider())


def generate_response(
        prompt: str, *, provider: LLMProvider | None = None
) -> ProviderResponse:
    app = create_app(provider)
    return app.provider.generate(ProviderRequest(prompt=prompt))


def generate_structured_response(
        prompt: str,
        response_schema: Mapping[str, Any],
        *,
        provider: LLMProvider | None = None
) -> StructuredOutputResult:
    app = create_app(provider)
    return StructuredOutputGenerator(app.provider).generate(prompt, response_schema)


def main() -> None:
    response = generate_response("runtime-healthcheck")
    print(f"{RUNTIME_NAME}:{READY_STATUS}:{response.provider_name}:{response.content}")


if __name__ == "__main__":
    main()
