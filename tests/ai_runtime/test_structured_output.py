import re

import pytest

from autonoma_runtime.app import generate_structured_response
from autonoma_runtime.provider import MockProvider
from autonoma_runtime.structured_output import StructuredOutputError


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