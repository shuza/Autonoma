"""Explicitly configured, non-streaming Chat Completions provider."""

import json
import math
import socket
from http.client import HTTPException
from typing import Any, NoReturn
from urllib.error import HTTPError, URLError
from urllib.parse import urlsplit
from urllib.request import HTTPRedirectHandler, Request, build_opener

from .accounting import ModelPricing, TokenUsage
from .provider import ProviderRequest, ProviderResponse


class ProviderError(Exception):
    """Sanitized provider transport or response failure."""


class ProviderHTTPError(ProviderError):
    def __init__(self, status_code: int) -> None:
        super().__init__(f"provider returned HTTP {status_code}")
        self.status_code = status_code


class ProviderTimeoutError(ProviderError):
    """The provider request exceeded its socket timeout."""


class ProviderResponseError(ProviderError):
    """The provider returned an invalid completion envelope."""


class _NoRedirects(HTTPRedirectHandler):
    def redirect_request(self, req, fp, code, msg, headers, newurl):
        return None


def _reject_constant(value: str) -> NoReturn:
    raise ValueError("non-standard JSON number")


class OpenAICompatibleProvider:
    name = "openai-compatible"

    def __init__(
            self,
            *,
            base_url: str,
            model: str,
            timeout_seconds: float,
            api_key: str | None = None,
            pricing: ModelPricing | None = None,
    ) -> None:
        if not isinstance(base_url, str) or any(
                character.isspace() or ord(character) < 32 for character in base_url
        ):
            raise ValueError("base_url must be an HTTP(s) URL")
        try:
            parsed = urlsplit(base_url)
            port = parsed.port
        except ValueError:
            raise ValueError("base_url must be an HTTP(s) URL") from None
        if (
                parsed.scheme not in ("http", "https")
                or not parsed.hostname
                or parsed.username is not None
                or parsed.password is not None
                or parsed.query
                or parsed.fragment
                or port == 0
        ):
            raise ValueError("base_url must be an HTTP(s) URL without credentials or query")
        if parsed.scheme == "http" and parsed.hostname not in (
                "localhost", "127.0.0.1", "::1"
        ):
            raise ValueError("remote provider require HTTPS")
        if not isinstance(model, str) or not model.strip():
            raise ValueError("model must be non-empty")
        if (
                type(timeout_seconds) not in (int, float)
                or not math.isfinite(timeout_seconds)
                or timeout_seconds <= 0
        ):
            raise ValueError("timeout_seconds must be finite and positive")
        if api_key is not None and (
                not isinstance(api_key, str)
                or not api_key.strip()
                or any(ord(character) < 33 or ord(character) > 126 for character in api_key)
        ):
            raise ValueError("api_key must be a non-empty printable header value")
        if pricing is not None and not isinstance(pricing, ModelPricing):
            raise ValueError("pricing must be ModelPricing or None")
        self._endpoint = base_url.rstrip("/") + "/chat/completions"
        self._model = model
        self._timeout = timeout_seconds
        self._api_key = api_key
        self._pricing = pricing
        self._openai = build_opener(_NoRedirects())

    def generate(self, request: ProviderRequest) -> ProviderResponse:
        payload: dict[str, Any] = {
            "model": self._model,
            "messages": [{"role": "user", "content": request.prompt}],
            "stream": False,
        }
        if request.response_schema is not None:
            payload["response_format"] = {
                "type": "json_schema",
                "json_schema": {
                    "name": "autonoma_response",
                    "schema": request.response_schema,
                }
            }
        headers = {"Content-type": "application/json", "Accept": "application/json"}
        if self._api_key is not None:
            headers["Authorization"] = f"Bearer {self._api_key}"
        http_request = Request(
            self._endpoint,
            data=json.dumps(payload, allow_nan=False).encode("utf-8"),
            headers=headers,
            method="POST",
        )
        try:
            with self._openai.open(http_request, timeout=self._timeout) as response:
                body = response.read()
        except HTTPError as error:
            status = error.code
            error.close()
            raise ProviderHTTPError(status) from None
        except (TimeoutError, socket.timeout):
            raise ProviderTimeoutError("provider request timed out") from None
        except URLError as error:
            if isinstance(error.reason, TimeoutError):
                raise ProviderTimeoutError("provider request timed out") from None
            raise ProviderError("provider connection failed") from None
        except (OSError, HTTPException):
            raise ProviderError("provider transport failed") from None

        try:
            envelope = json.loads(body, parse_constant=_reject_constant)
        except (ValueError, UnicodeDecodeError):
            raise ProviderResponseError("provider response is not valid JSON") from None
        return self._parse_response(envelope)

    def _parse_response(self, envelope):
        if not isinstance(envelope, dict):
            raise ProviderResponseError("provider response must be an object")
        choices = envelope.get("choices")
        if not isinstance(choices, list) or len(choices) != 1:
            raise ProviderResponseError("provider response must contain one choice")
        choice = choices[0]
        if not isinstance(choice, dict) or choice.get("finish_reason") != "stop":
            raise ProviderResponseError("provider did not return a complete response")
        message = choice.get("message")
        if (
                not isinstance(message, dict)
                or message.get("refusal") is not None
                or message.get("tool_calls")
                or not isinstance(message.get("content"), str)
        ):
            raise ProviderResponseError("provider did not return text content")

        usage = None
        raw_usage = envelope.get("usage")
        if raw_usage is not None:
            if not isinstance(raw_usage, dict):
                raise ProviderResponseError("provider usage must be an object")
            try:
                usage = TokenUsage(
                    raw_usage.get("prompt_tokens"), raw_usage.get("completion_tokens")
                )
            except ValueError:
                raise ProviderResponseError("provider token counts are invalid") from None
            if "total_tokens" in raw_usage and (
                    type(raw_usage["total_tokens"]) is not int
                    or raw_usage["total_tokens"] != usage.total_tokens
            ):
                raise ProviderResponseError("provider total token count is inconsistent")

        return ProviderResponse(
            content=message["content"],
            provider_name=self.name,
            usage=usage,
            pricing=self._pricing,
        )
