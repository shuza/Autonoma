import os

import pytest


@pytest.fixture(autouse=True)
def isolate_llm_environment(monkeypatch):
    for name in os.environ:
        if name.startswith("AUTONOMA_LLM_"):
            monkeypatch.delenv(name)