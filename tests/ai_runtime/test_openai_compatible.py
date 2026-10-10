import json
import pytest
import threading
import time
from autonoma_runtime.accounting import ModelPricing, TokenUsage
from autonoma_runtime.app import generate_response, generate_structured_response
from autonoma_runtime.openai_compatible import (
    OpenAICompatibleProvider,
    ProviderError,
    ProviderHTTPError,
    ProviderResponseError,
    ProviderTimeoutError,
)
from autonoma_runtime.structured_output import StructuredOutputError
from decimal import Decimal
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from typing import Any


@pytest.fixture
def endpoint():
    state = {
        "status": 200,
        "body": {
            "choices": [{"message": {"content": "true"}, "finish_reason": "stop"}],
            "usage": {"prompt_tokens": 10, "completion_tokens": 20, "total_tokens": 30},
        },
        "delay": 0,
        "requests": []
    }

    class Handler(BaseHTTPRequestHandler):
        def do_POST(self) -> None:
            body = self.rfile.read(int(self.headers["Content-Length"]))
            state["requests"].append((self.path, dict(self.headers), json.loads(body)))
            time.sleep(state["delay"])
            self.send_response(state["status"])
            self.send_header("Location", "/redirect-target")
            self.end_headers()
            response = state["body"]
            data = response if isinstance(response, bytes) else json.dumps(response).encode()
            try:
                self.wfile.write(data)
            except (BrokenPipeError, ConnectionResetError):
                pass

        def log_message(self, format, *args) -> None:
            pass

    server = ThreadingHTTPServer(("127.0.0.1", 0), Handler)
    thread = threading.Thread(target=server.serve_forever)
    thread.start()
    try:
        yield f"http://127.0.0.1:{server.server_port}/v1", state
    finally:
        server.shutdown()
        server.server_close()
        thread.join()


def provider(url: str, **kwargs) -> OpenAICompatibleProvider:
    return OpenAICompatibleProvider(
        base_url=url, model="test-model", timeout_seconds=1, **kwargs
    )


def test_plan_generation_and_wire_contact(endpoint) -> None:
    url, state = endpoint
    response = generate_response(
        "hello",
        provider=provider(
            url, api_key="test-key",
            pricing=ModelPricing(Decimal("1"), Decimal("2")),
        )
    )
    assert response.content == "true"
    assert response.usage == TokenUsage(10, 20)
    assert response.cost_usd == Decimal("0.00005")
    path, headers, payload = state["requests"][0]
    assert path == "/v1/chat/completions"
    assert headers["Authorization"] == "Bearer test-key"
    assert payload == {
        "model": "test-model", "stream": False,
        "messages": [{"role": "user", "content": "hello"}],
    }


def test_structured_generation_schema_and_accounting(endpoint) -> None:
    url, state = endpoint
    schema = {"type": "boolean"}
    result = generate_structured_response("probe", schema, provider=provider(url))
    assert result.data is True
    assert result.provider_response.usage.total_tokens == 30
    assert result.provider_response.cost_usd is None
    _, headers, payload = state["requests"][0]
    assert "Authorization" not in headers
    assert payload["response_format"]["json_schema"]["schema"] == schema
    assert payload["response_format"]["type"] == "json_schema"


@pytest.mark.parametrize("usage", [None, "missing"])
def test_missing_usage_is_unknow(endpoint, usage) -> None:
    url, state = endpoint
    if usage == "missing":
        del state["body"]["usage"]
    else:
        state["body"]["usage"] = usage
    response = generate_response("probe", provider=provider(url, pricing=ModelPricing()))
    assert response.usage is None
    assert response.cost_usd is None


def test_reported_zero_usage_and_free_pricing(endpoint) -> None:
    url, state = endpoint
    state["body"]["usage"] = {"prompt_tokens": 0, "completion_tokens": 0}
    response = generate_response("probe", provider=provider(url, pricing=ModelPricing()))
    assert response.usage == TokenUsage(0, 0)
    assert response.cost_usd == Decimal("0")


@pytest.mark.parametrize("status", [401, 429, 500, 302, 307])
def test_http_failure_are_sanitized_without_retry_or_redirect(endpoint, status) -> None:
    url, state = endpoint
    state["status"] = status = status
    state["body"] = b"sensitive-provider-body"
    with pytest.raises(ProviderHTTPError) as caught:
        generate_response("secret-prompt", provider=provider(url, api_key="secret-key"))
    assert caught.value.status_code == status
    assert str(caught.value) == f"provider returned HTTP {status}"
    assert len(state["requests"]) == 1


@pytest.mark.parametrize(
    "body",
    [
        b"not json", b"\xff", b"NaN", [], {},
        {"choices": []}, {"choices": [None]},
        {"choices": [{"finish_reason": "length", "message": {"content": "partial"}}]},
        {"choices": [{"finish_reason": "stop", "message": {"content": None}}]},
        {"choices": [{"finish_reason": "stop", "message": {"content": []}}]},
        {"choices": [{"finish_reason": "stop", "message": {
            "content": "text", "refusal": "refused",
        }}]},
        {"choices": [{"finish_reason": "stop", "message": {
            "content": "text", "tool_calls": [{"function": {}}]
        }}]}
    ],
)
def test_malformed_envelopes_fail_explicitly(endpoint, body: Any) -> None:
    url, state = endpoint
    state["body"] = body
    with pytest.raises(ProviderResponseError):
        generate_response("probe", provider=provider(url))


@pytest.mark.parametrize(
    "usage",
    [
        [], {}, {"prompt_tokens": 1},
        {"prompt_tokens": True, "completion_tokens": 0},
        {"prompt_tokens": -1, "completion_tokens": 0},
        {"prompt_tokens": 1.5, "completion_tokens": 0},
        {"prompt_tokens": 1, "completion_tokens": "2"},
        {"prompt_tokens": 1, "completion_tokens": 2, "total_tokens": 4},
        {"prompt_tokens": 0, "completion_tokens": 0, "total_tokens": False},
    ],
)
def test_invalid_usage_fails(endpoint, usage) -> None:
    url, state = endpoint
    state["body"]["usage"] = usage
    with pytest.raises(ProviderResponseError):
        generate_response("probe", provider=provider(url))


@pytest.mark.parametrize("content", ["not json", '"wrong type"', "NaN"])
def test_structured_output_failure_preserves_accounting(endpoint, content) -> None:
    url, state = endpoint
    state["body"]["choices"][0]["message"]["content"] = content
    with pytest.raises(StructuredOutputError) as caught:
        generate_structured_response(
            "probe", {"type": "boolean"},
            provider=provider(url, pricing=ModelPricing()),
        )
    assert caught.value.provider_response.usage == TokenUsage(10, 20)
    assert caught.value.provider_response.cost_usd == Decimal("0")


def test_timeout(endpoint) -> None:
    url, state = endpoint
    state["delay"] = 0.1
    adapter = OpenAICompatibleProvider(
        base_url=url, model="test", timeout_seconds=0.01,
    )
    with pytest.raises(ProviderTimeoutError):
        generate_response("probe", provider=adapter)
    assert len(state["requests"]) == 1


def test_connection_failure(endpoint) -> None:
    server = ThreadingHTTPServer(("127.0.0.1", 0), BaseHTTPRequestHandler)
    port = server.server_port
    server.server_close()
    with pytest.raises(ProviderError, match="connection failed"):
        generate_response("probe", provider=provider(f"http://127.0.0.1:{port}/v1"))


@pytest.mark.parametrize(
    "config",
    [
        {"base_url": "file:///tmp/model"}, {"base_url": "http://remote.example/v1"},
        {"base_url": "https://user:password@example.com/v1"},
        {"base_url": "https://example.com/v1?key=secret"},
        {"base_url": "https://example.com:invalid/v1"},
        {"base_url": "https://example.com/v1\n"}, {"base_url": ""},
        {"model": ""}, {"model": None},
        {"timeout_seconds": 0}, {"timeout_seconds": -1},
        {"timeout_seconds": True}, {"timeout_seconds": float("inf")},
        {"api_key": ""}, {"api_key": "key\ninjection"},
        {"pricing": {}},
    ],
)
def test_invalid_config_fails_before_network(config) -> None:
    settings = {"base_url": "https://example.com/v1", "model": "test", "timeout_seconds": 1}
    settings.update(config)
    with pytest.raises(ValueError):
        OpenAICompatibleProvider(**settings)