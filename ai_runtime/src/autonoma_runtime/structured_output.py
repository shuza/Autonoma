"""Validation of untrusted JSON returned by LLM providers."""

import json
import math
from dataclasses import dataclass
from typing import Any, Mapping, NoReturn

from .provider import LLMProvider, ProviderRequest, ProviderResponse


class StructuredOutputError(Exception):
    """Raised when provider output is not valid for its requested schema."""


class StructuredOutputSchemaError(ValueError):
    """Raised before generation when a schema is invalid or unsupported."""


@dataclass(frozen=True)
class StructuredOutputResult:
    data: Any
    provider_response: ProviderResponse


class StructuredOutputGenerator:
    """Generates and validates JSON without exposing provider-specific types."""

    def __init__(self, provider: LLMProvider) -> None:
        self._provider = provider

    def generate(
        self,
        prompt: str,
        response_schema: Mapping[str, Any],
    ) -> StructuredOutputResult:
        _validate_schema(response_schema, path="$")
        response = self._provider.generate(
            ProviderRequest(prompt=prompt, response_schema=response_schema)
        )
        try:
            data = json.loads(
                response.content,
                parse_constant=_reject_constant,
                parse_float=_parse_finite_float,
            )
        except json.JSONDecodeError as error:
            raise StructuredOutputError("provider response is not valid JSON") from error

        _validate(data, response_schema, path="$")
        return StructuredOutputResult(data=data, provider_response=response)


def _reject_constant(value: str) -> NoReturn:
    raise StructuredOutputError("provider response contains a non-standard JSON number")


def _parse_finite_float(value: str) -> float:
    number = float(value)
    if not math.isfinite(number):
        raise StructuredOutputError("provider response contains a non-finite number")
    return number


def _validate_schema(schema: Any, *, path: str) -> None:
    if not isinstance(schema, Mapping):
        raise StructuredOutputSchemaError(f"{path} schema must be an object")

    supported = {
        "type", "enum", "properties", "required", "additionalProperties",
        "items", "minItems", "minLength", "minimum", "maximum",
    }
    for keyword in schema:
        if keyword not in supported:
            raise StructuredOutputSchemaError(
                f"{path} uses unsupported schema keyword {keyword!r}"
            )

    if "type" in schema and schema["type"] not in (
        "object", "array", "string", "number", "integer", "boolean", "null"
    ):
        raise StructuredOutputSchemaError(f"{path}.type must be a supported type name")

    if "enum" in schema:
        values = schema["enum"]
        if not isinstance(values, list) or not values:
            raise StructuredOutputSchemaError(f"{path}.enum must be a non-empty list")
        if not all(_is_json_value(value) for value in values):
            raise StructuredOutputSchemaError(f"{path}.enum must contain finite JSON values")
        if any(
            _json_equal(value, previous)
            for index, value in enumerate(values)
            for previous in values[:index]
        ):
            raise StructuredOutputSchemaError(f"{path}.enum must contain unique values")

    if "properties" in schema:
        properties = schema["properties"]
        if not isinstance(properties, Mapping):
            raise StructuredOutputSchemaError(f"{path}.properties must be an object")
        for name, property_schema in properties.items():
            if not isinstance(name, str):
                raise StructuredOutputSchemaError(f"{path}.properties keys must be strings")
            _validate_schema(property_schema, path=f"{path}.properties.{name}")

    if "required" in schema:
        required = schema["required"]
        if (
            not isinstance(required, list)
            or not all(isinstance(name, str) for name in required)
            or len(set(required)) != len(required)
        ):
            raise StructuredOutputSchemaError(f"{path}.required must be unique property names")

    if "additionalProperties" in schema and not isinstance(
        schema["additionalProperties"], bool
    ):
            raise StructuredOutputSchemaError(f"{path}.additionalProperties must be a boolean")

    if "items" in schema:
        _validate_schema(schema["items"], path=f"{path}.items")

    for keyword in ("minItems", "minLength"):
        if keyword in schema:
            value = schema[keyword]
            if type(value) is not int or value < 0:
                raise StructuredOutputSchemaError(
                    f"{path}.{keyword} must be a non-negative integer"
                )

    for keyword in ("minimum", "maximum"):
        if keyword in schema and not _is_number(schema[keyword]):
            raise StructuredOutputSchemaError(f"{path}.{keyword} must be a finite number")
    if (
        "minimum" in schema
        and "maximum" in schema
        and schema["minimum"] > schema["maximum"]
    ):
        raise StructuredOutputSchemaError(f"{path}.minimum must not exceed maximum")


def _is_number(value: Any) -> bool:
    return type(value) is int or (type(value) is float and math.isfinite(value))


def _is_json_value(value: Any) -> bool:
    if value is None or isinstance(value, (str, bool)) or _is_number(value):
        return True
    if isinstance(value, list):
        return all(_is_json_value(item) for item in value)
    if isinstance(value, dict):
        return all(
            isinstance(key, str) and _is_json_value(item)
            for key, item in value.items()
        )
    return False


def _json_equal(left: Any, right: Any) -> bool:
    # Python considers True == 1; JSON Schema treats booleans and numbers separately.
    if _is_number(left) and _is_number(right):
        return left == right
    if type(left) is not type(right):
        return False
    if isinstance(left, dict):
        return left.keys() == right.keys() and all(
            _json_equal(value, right[key]) for key, value in left.items()
        )
    if isinstance(left, list):
        return len(left) == len(right) and all(
            _json_equal(a, b) for a, b in zip(left, right)
        )
    return left == right


def _validate(value: Any, schema: Mapping[str, Any], *, path: str) -> None:
    expected_type = schema.get("type")
    if expected_type is not None:
        _validate_type(value, expected_type, path=path)

    if "enum" in schema and not any(
        _json_equal(value, allowed) for allowed in schema["enum"]
    ):
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
        "number": _is_number,
        "integer": lambda item: _is_number(item) and (
            isinstance(item, int) or item.is_integer()
        ),
        "boolean": lambda item: isinstance(item, bool),
        "null": lambda item: item is None,
    }
    if not type_checks[expected_type](value):
        raise StructuredOutputError(f"{path} must be a {expected_type!r}")


def _validate_object(value: dict[str, Any], schema: Mapping[str, Any], path: str) -> None:
    properties = schema.get("properties", {})
    required = schema.get("required", [])
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
            _validate(property_value, property_schema, path=f"{path}.{property_name}")


def _validate_array(value: list[Any], schema: Mapping[str, Any], path: str) -> None:
    minimum = schema.get("minItems")
    if isinstance(minimum, int) and len(value) < minimum:
        raise StructuredOutputError(f"{path} must contain at least {minimum} items")

    item_schema = schema.get("items")
    if item_schema is not None:
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
