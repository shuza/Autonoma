import pytest
import re
from autonoma_runtime.app import generate_structured_response
from autonoma_runtime.provider import MockProvider, ProviderRequest, ProviderResponse
from autonoma_runtime.structured_output import (
    StructuredOutputError,
    StructuredOutputSchemaError,
)
from typing import Any

QUALIFICATION_SCHEMA = {
    "type": "object",
    "properties": {
        "qualified": {"type": "boolean"},
        "reason": {"type": "string", "minLength": 1},
        "score": {"type": "number", "minimum": 0, "maximum": 100},
    },
    "required": ["qualified", "reason", "score"],
    "additionalProperties": False,
}


def test_generate_structured_response_validation_mock_provider_output() -> None:
    provider = MockProvider(
        {
            "qualify": (
                '{"qualified": true, "reason": "Matches target market", "score": 85}'
            )
        }
    )

    result = generate_structured_response(
        " qualify ",
        QUALIFICATION_SCHEMA,
        provider=provider,
    )

    assert result.data == {
        "qualified": True,
        "reason": "Matches target market",
        "score": 85,
    }
    assert result.provider_response.provider_name == "mock-provider"


@pytest.mark.parametrize(
    ("content", "error"),
    [
        ("not json", "provider response is not valid JSON"),
        ('{"qualified": true, "reason": "Good fit"}', "$.score is required"),
        (
                '{"qualified": true, "reason": "Good fit", "score": 101}',
                "$.score must be at most 100",
        ),
        (
                '{"qualified": true, "reason": "Good fit", "score": 50, "action": "send"}',
                "$ contains unexpected properties: action",
        ),
    ],
)
def test_generate_structured_response_rejects_invalid_provider_output(
        content: str,
        error: str,
) -> None:
    provider = MockProvider({"qualify": content})

    with pytest.raises(StructuredOutputError, match=re.escape(error)):
        generate_structured_response(
            " qualify ",
            QUALIFICATION_SCHEMA,
            provider=provider,
        )


@pytest.mark.parametrize(
    "content",
    [
        "NaN", "Infinity", "-Infinity", "1e400", "-1e400",
        '{"score": NaN}', '{"ignored": [Infinity]}', '[{"score": 1e400}]',
    ],
)
def test_rejects_non_finite_number_everywhere(content: str) -> None:
    with pytest.raises(StructuredOutputError, match="number"):
        generate_structured_response(
            "probe", {}, provider=MockProvider({"probe": content})
        )


class RecordingProvider(MockProvider):
    def __init__(self) -> None:
        super().__init__({"probe": "{}"})
        self.requests: list[ProviderRequest] = []

    def generate(self, request: ProviderRequest) -> ProviderResponse:
        self.requests.append(request)
        return super().generate(request)


@pytest.mark.parametrize(
    "schema",
    [
        None, [], True,
        {"type": None}, {"type": []}, {"type": "unknown"},
        {"type": ["string", "null"]},
        {"maxLength": 3}, {"$ref": "#/definitions/item"},
        {"properties": []}, {"properties": {"unused": {"maxLength": 3}}},
        {"properties": {"unused": None}}, {"properties": {1: {}}},
        {"items": []}, {"items": {"pattern": "^a"}},
        {"required": "name"}, {"required": [1]}, {"required": ["a", "a"]},
        {"additionalProperties": {}}, {"additionalProperties": None},
        {"minItems": True}, {"minItems": -1}, {"minItems": 1.5},
        {"minLength": "1"}, {"minLength": -1},
        {"minLength": True}, {"minLength": 1.5}, {"minLength": None},
        {"minimum": True}, {"minimum": "0"},
        {"minimum": float("nan")}, {"maximum": float("inf")},
        {"maximum": False}, {"maximum": "-1"},
        {"minimum": 2, "maximum": 1},
        {"enum": []}, {"enum": "a"}, {"enum": [float("nan")]},
        {"enum": [{"nested": float("inf")}]}, {"enum": [(1, 2)]},
        {"enum": ["a", "a"]}, {"enum": [1, 1.0]},
        {"enum": [[1], [1.0]]}, {"enum": [{"x": 1}, {"x": 1.0}]},
    ],
)
def test_invalid_schemas_fail_before_provider_call(schema: Any) -> None:
    provider = RecordingProvider()

    with pytest.raises(StructuredOutputSchemaError):
        generate_structured_response("probe", schema, provider=provider)

    assert provider.requests == []


def test_schema_is_forwarded_to_provider() -> None:
    provider = RecordingProvider()
    schema = {"type": "object"}

    generate_structured_response("probe", schema, provider=provider)

    assert provider.requests == [ProviderRequest("probe", schema)]


@pytest.mark.parametrize(
    ("content", "schema"),
    [
        ("true", {"type": "number"}),
        ("true", {"type": "integer"}),
        ("1", {"type": "boolean"}),
        ("1.5", {"type": "integer"}),
        ("null", {"type": "object"}),
        ('"1"', {"type": "number"}),
        ("{}", {"type": "array"}),
        ("[]", {"type": "object"}),
        ("0", {"type": "null"}),
        ('""', {"type": "string", "minLength": 1}),
        ("[]", {"type": "array", "minItems": 1}),
        ("-1", {"type": "number", "minimum": 0}),
        ("true", {"enum": [1]}),
        ("1", {"enum": [True]}),
        ('{"value": true}', {"enum": [{"value": 1}]}),
        ("[true]", {"enum": [[1]]}),
        ('"other"', {"enum": ["allowed"]}),
    ],
)
def test_rejects_wrong_types_and_constraint_violations(
        content: str, schema: dict[str, Any]
) -> None:
    with pytest.raises(StructuredOutputError):
        generate_structured_response(
            "probe", schema, provider=MockProvider({"probe": content})
        )


@pytest.mark.parametrize(
    ("content", "schema"),
    [
        ("1.0", {"type": "integer"}),
        ("1.5", {"type": "number"}),
        ("1e308", {"type": "number"}),
        ("null", {"type": "null"}),
        ("false", {"type": "boolean"}),
        ('"ok"', {"type": "string", "minLength": 2}),
        ('[]', {"type": "array", "minItems": 0}),
        ("0", {"type": "number", "minimum": 0, "maximum": 0}),
        ("1.0", {"enum": [1]}),
        ("true", {"enum": [1, True]}),
        ('{"value": [1.0]}', {"enum": [{"value": [1]}]}),
        ("[1, 0]", {"enum": [[1], [1, 0]]}),
        ("[1, 2]", {"enum": [[1, 2], [1, 3]]}),
        ('{"extra": 1}', {"type": "object", "properties": {"optional": {}}}),
        ("{}", {"type": "object", "additionalProperties": False}),
        ('"not a number"', {"minimum": 1}),
    ],
)
def test_accepts_supported_types_and_constraint_boundaries(
        content: str, schema: dict[str, Any],
) -> None:
    result = generate_structured_response(
        "probe", schema, provider=MockProvider({"probe": content})
    )

    assert result.provider_response.content == content


NESTED_SCHEMA = {
    "type": "object",
    "properties": {
        "leads": {
            "type": "array",
            "minItems": 1,
            "items": QUALIFICATION_SCHEMA,
        },
    },
    "required": ["leads"],
    "additionalProperties": False,
}


def test_accepts_nested_objects_and_arrays() -> None:
    content = '{"leads": [{"qualified": true, "reason": "Fit", "score": 100}]}'

    result = generate_structured_response(
        "probe", NESTED_SCHEMA, provider=MockProvider({"probe": content})
    )

    assert result.provider_response.content == content
    assert result.data["leads"][0]["score"] == 100


@pytest.mark.parametrize(
    ("content", "error"),
    [
        ('{"leads": []}', "$.leads must contain at least 1 items"),
        (
                '{"leads": [{"qualified": true, "reason": "Fit"}]}',
                "$.leads[0].score is required",
        ),
        (
                '{"leads": [{"qualified": true, "reason": "Fit", "score": true}]}',
                "$.leads[0].score must be a 'number'",
        ),
        (
                '{"leads": [{"qualified": true, "reason": "Fit", "score": 1, "extra": 1}]}',
                "$.leads[0] contains unexpected properties: extra",
        ),
    ],
)
def test_nested_validation_reports_failure_path(content: str, error: str) -> None:
    with pytest.raises(StructuredOutputError, match=re.escape(error)):
        generate_structured_response(
            "probe", NESTED_SCHEMA, provider=MockProvider({"probe": content})
        )


def test_provider_failure_is_not_hidden() -> None:
    class TimeoutProvider:
        def generate(self, request: ProviderRequest) -> ProviderResponse:
            raise TimeoutError("provider timed out")

    with pytest.raises(TimeoutError, match="provider timed out"):
        generate_structured_response("probe", {}, provider=TimeoutProvider())


def test_structured_generation_accepts_prompt_keyword() -> None:
    result = generate_structured_response(
        prompt="probe",
        response_schema={"type": "boolean"},
        provider=MockProvider({"probe": "true"}),
    )

    assert result.data is True
