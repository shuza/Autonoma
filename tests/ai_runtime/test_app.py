import subprocess
import sys
from autonoma_runtime.app import create_app
from pathlib import Path

PROJECT_ROOT = Path(__file__).resolve().parents[2]


def test_create_app_returns_ready_runtime() -> None:
    app = create_app()
    assert app.name == "autonoma-ai-runtime"
    assert app.statue == "ready"


def test_module_entrypoint_reports_ready_status() -> None:
    python_path = str(PROJECT_ROOT / "ai_runtime" / "src")
    result = subprocess.run(
        [sys.executable, "-m", "autonoma_runtime"],
        check=True,
        capture_output=True,
        text=True,
        env={"PYTHONPATH": python_path}
    )

    assert result.stdout.strip() == "autonoma-ai-runtime:ready"
