"""Minimal AI runtime entrypoint for Phase 0."""

from dataclasses import dataclass
from typing import Final

from .provider import LLMProvider, MockProvider, ProviderRequest, ProviderResponse

READY_STATUS: Final[str] = "ready"
RUNTIME_NAME: Final[str] = "autonoma-ai-runtime"


@dataclass(frozen=True)
class RuntimeApp:
    provider: LLMProvider
    name: str = RUNTIME_NAME
    status: str = READY_STATUS


def create_app() -> RuntimeApp:
    return RuntimeApp(provider=MockProvider())


def generate_response(prompt: str) -> ProviderResponse:
    app = create_app()
    return app.provider.generate(ProviderRequest(prompt=prompt))


def main() -> None:
    response = generate_response("runtime-healthcheck")
    print(f"{RUNTIME_NAME}:{READY_STATUS}:{response.provider_name}:{response.content}")


if __name__ == "__main__":
    main()
