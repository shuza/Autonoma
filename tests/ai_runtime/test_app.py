import os
import subprocess
import sys
import pytest
from autonoma_runtime.app import create_app, generate_response, generate_structured_response
from pathlib import Path

PROJECT_ROOT = Path(__file__).resolve().parents[2]


def test_create_app_returns_ready_runtime() -> None:
    app = create_app()

    assert app.name == "autonoma-ai-runtime"
    assert app.status == "ready"
    assert app.provider.name == "mock-provider"


def test_generate_response_uses_mock_provider() -> None:
    response = generate_response(" hello runtime")

    assert response.provider_name == "mock-provider"
    assert response.content == "mock-response:hello runtime"


def test_module_entrypoint_reports_ready_status() -> None:
    python_path = str(PROJECT_ROOT / "ai_runtime" / "src")
    env = os.environ.copy()
    env["PYTHONPATH"] = python_path

    result = subprocess.run(
        [sys.executable, "-m", "autonoma_runtime"],
        check=True,
        capture_output=True,
        text=True,
        env=env
    )

    assert result.stdout.strip() == "autonoma-ai-runtime:ready:mock-provider:mock-response:runtime-healthcheck"


@pytest.mark.parametrize("structured", [False, True])
@pytest.mark.parametrize("prompt", [None, True, 123, {}, [], "", " \n "])
def test_malformed_prompt_fails_before_generation(prompt, structured):
    class UncalledProvider:
        def generate(self, request):
            pytest.fail("invalid prompt must not call provider")

    with pytest.raises(ValueError, match="prompt must be a non-empty string"):
        if structured:
            generate_structured_response(prompt, {}, provider=UncalledProvider())
        else:
            generate_response(prompt, provider=UncalledProvider())