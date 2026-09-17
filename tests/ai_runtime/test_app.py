import os
import subprocess
import sys
from autonoma_runtime.app import create_app
from pathlib import Path

PROJECT_ROOT = Path(__file__).resolve().parents[2]


def test_create_app_returns_ready_runtime() -> None:
    app = create_app()

    assert app.name == "autonoma-ai-runtime"
    assert app.status == "ready"


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

    assert result.stdout.strip() == "autonoma-ai-runtime:ready"
