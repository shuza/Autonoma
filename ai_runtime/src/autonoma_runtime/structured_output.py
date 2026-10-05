"""Validation of untrusted JSON returned by LLM providers."""

import json
from dataclasses import dataclass
from typing import Any, Mapping

from .provider import LLMProvider, ProviderRequest, ProviderResponse


class StructuredOutputError(Exception):
    """Raised when provider output is not valid for its requested schema."""


@dataclass(frozen=True)
class StructuredOutputResult:
    data: Any
    provider_response: ProviderResponse


class StructuredOutputGenerator:
    """Generates and validates JSON without exposing provider-specific types."""

    def __init__(self, provider: LLMProvider):
        self._provider = provider

    def generate(
            self,
            prompt: str,
            response_schema: Mapping[str, Any]
    ) -> StructuredOutputResult:
        response = self._provider.generate(
            ProviderRequest(prompt=prompt, response_schema=response_schema)
        )
        try:
            data = json.loads(response.content)
        except json.JSONDecodeError as error:
            raise StructuredOutputError("provider response is not valid JSON") from error

        _validate(data, response_schema, path="$")
        return StructuredOutputResult(data=data, provider_response=response)


def _validate(value: Any, schema: Mapping[str, Any], *, path: str) -> None:
    expected_type = schema.get("type")
    if expected_type is not None:
        _validate_type(value, expected_type, path=path)

    if "enum" in schema and value not in schema["enum"]:
        raise StructuredOutputError(f"{path} must be one of the allowed values")

    if isinstance(value, dict):
        _validate_object(value, schema, path)
    elif isinstance(value, list):
        _validate_array(value, schema, path)
    elif isinstance(value, str):
        _validate_string(value, schema, path)
    elif isinstance(value, (int, float)) and not isinstance(value, bool):
        _validate_number(value, schema, path)


def _validate_type(value: Any, expected_type: Any, path: str) -> None:
    type_checks = {
        "object": lambda item: isinstance(item, dict),
        "array": lambda item: isinstance(item, list),
        "string": lambda item: isinstance(item, str),
        "number": lambda item: isinstance(item, (int, float)) and not isinstance(item, bool),
        "integer": lambda item: isinstance(item, int) and not isinstance(item, bool),
        "boolean": lambda item: isinstance(item, bool),
        "null": lambda item: item is None,
    }
    if not isinstance(expected_type, str) or expected_type not in type_checks:
        raise StructuredOutputError(f"{path} uses unsupported schema type {expected_type!r}")
    if not type_checks[expected_type](value):
        raise StructuredOutputError(f"{path} must be a {expected_type!r}")


def _validate_object(value: dict[str, Any], schema: Mapping[str, Any], path: str) -> None:
    properties = schema.get("properties", {})
    if not isinstance(properties, Mapping):
        raise StructuredOutputError(f"{path}.properties must be an object")

    required = schema.get("required", [])
    if not isinstance(required, list) or not all(isinstance(item, str) for item in required):
        raise StructuredOutputError(f"{path}.properties must be a list of property names")
    for property_name in required:
        if property_name not in value:
            raise StructuredOutputError(f"{path}.{property_name} is required")

    if schema.get("additionalProperties") is False:
        unexpected = set(value) - set(properties)
        if unexpected:
            raise StructuredOutputError(
                f"{path} contains unexpected properties: {', '.join(sorted(unexpected))}"
            )

    for property_name, property_value in value.items():
        property_schema = properties.get(property_name)
        if property_schema is not None:
            if not isinstance(property_schema, Mapping):
                raise StructuredOutputError(
                    f"{path}.{property_name} schema must be an object"
                )
            _validate(property_value, property_schema, path=f"{path}.{property_name}")


def _validate_array(value: list[Any], schema: Mapping[str, Any], path: str) -> None:
    minimum = schema.get("minItems")
    if isinstance(minimum, int) and len(value) < minimum:
        raise StructuredOutputError(f"{path} must contain at least {minimum} items")

    item_schema = schema.get("items")
    if item_schema is not None:
        if not isinstance(item_schema, Mapping):
            raise StructuredOutputError(f"{path}.items schema must be an object")
        for index, item_value in enumerate(value):
            _validate(item_value, item_schema, path=f"{path}[{index}]")


def _validate_string(value: str, schema: Mapping[str, Any], path: str) -> None:
    minimum = schema.get("minLength")
    if isinstance(minimum, int) and len(value) < minimum:
        raise StructuredOutputError(f"{path} must be at least {minimum} characters")


def _validate_number(value: int | float, schema: Mapping[str, Any], path: str) -> None:
    minimum = schema.get("minimum")
    if isinstance(minimum, (int, float)) and value < minimum:
        raise StructuredOutputError(f"{path} must be at least {minimum}")

    maximum = schema.get("maximum")
    if isinstance(maximum, (int, float)) and value > maximum:
        raise StructuredOutputError(f"{path} must be at most {maximum}")
