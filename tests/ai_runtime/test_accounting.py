import pytest
from autonoma_runtime.accounting import ModelPricing, TokenUsage
from autonoma_runtime.app import generate_response, generate_structured_response
from autonoma_runtime.provider import MockProvider, ProviderRequest, ProviderResponse
from autonoma_runtime.structured_output import (
    StructuredOutputError,
    StructuredOutputSchemaError,
)
from decimal import Decimal, localcontext
from typing import Any, cast

PRICING = ModelPricing(Decimal("1.25"), Decimal("5"))


@pytest.mark.parametrize("value", [-1, True, False, 1.5, "1", None])
@pytest.mark.parametrize("field", ["input_tokens", "output_tokens"])
def test_invalid_usage_is_rejected(field: str, value: Any) -> None:
    counts = {"input_tokens": 0, "output_tokens": 0, field: value}
    with pytest.raises(ValueError, match=field):
        TokenUsage(**counts)


@pytest.mark.parametrize(
    "value", [Decimal("-1"), Decimal("NaN"), Decimal("Infinity"), 1.0, "1", None]
)
@pytest.mark.parametrize(
    "field", ["input_usd_per_million_tokens", "output_usd_per_million_tokens"]
)
def test_invalid_pricing_is_rejected(field: str, value: Any) -> None:
    with pytest.raises(ValueError, match=field):
        ModelPricing(**{field: value})


def test_usage_total_and_separate_input_output_cost() -> None:
    usage = TokenUsage(1000, 2000)

    assert usage.total_tokens == 3000
    assert PRICING.calculate_cost(usage) == Decimal("0.01125")


def test_cost_is_exact_even_with_low_decimal_precision() -> None:
    with localcontext() as context:
        context.prec = 2
        assert PRICING.calculate_cost(TokenUsage(12345, 6789)) == Decimal("0.04937625")


def test_cost_retains_widely_separated_rates_and_large_counts() -> None:
    pricing = ModelPricing(Decimal("1e30"), Decimal("1e-30"))
    usage = TokenUsage(10 ** 40, 1)
    with localcontext() as context:
        context.prec = 2
        cost = pricing.calculate_cost(usage)

    assert cost == Decimal("1" + "0" * 64 + "." + "0" * 35 + "1")


@pytest.mark.parametrize("value", [None, {}, 1])
def test_mock_rejects_invalid_usage_fixture(value: Any) -> None:
    with pytest.raises(ValueError, match="usage fixtures"):
        MockProvider(usage={"probe": value})


@pytest.mark.parametrize("value", [None, {}, Decimal("0")])
def test_mock_rejects_invalid_pricing_configuration(value: Any) -> None:
    with pytest.raises(ValueError, match="mock pricing"):
        MockProvider(pricing=value)


@pytest.mark.parametrize(
    ("usage", "pricing", "expected"),
    [
        (None, ModelPricing(), None),
        (TokenUsage(0, 0), ModelPricing(), Decimal("0")),
        (TokenUsage(10, 20), ModelPricing(), Decimal("0")),
        (TokenUsage(10, 20), None, None)
    ],
)
def test_unknown_accounting_is_distinct_from_zero(
        usage: TokenUsage | None,
        pricing: ModelPricing | None,
        expected: Decimal | None
) -> None:
    response = ProviderResponse("content", "mock", usage, pricing)

    assert response.cost_usd == expected


def test_plain_generation_carries_mock_usage_and_cost() -> None:
    provider = MockProvider(
        usage={"probe": TokenUsage(1000, 2000)}, pricing=PRICING
    )

    response = generate_response(" probe ", provider=provider)

    assert response.content == "mock-response:probe"
    assert response.usage is not None
    assert response.usage.total_tokens == 3000
    assert response.cost_usd == Decimal("0.01125")
    assert generate_response("unconfigured", provider=provider).usage is None
    assert generate_response("unconfigured", provider=provider).cost_usd is None


def test_structured_generation_carries_accounting() -> None:
    provider = MockProvider(
        {"probe": "true"}, usage={"probe": TokenUsage(1000, 2000)}, pricing=PRICING
    )

    result = generate_structured_response("probe", {"type": "boolean"}, provider=provider)

    assert result.data is True
    assert result.provider_response.usage == TokenUsage(1000, 2000)
    assert result.provider_response.cost_usd == Decimal("0.01125")


@pytest.mark.parametrize("content", ["not json", "NaN", "1e400", '"wrong type"'])
def test_output_failures_retain_usage_and_cost(content: str) -> None:
    provider = MockProvider(
        {"probe": content}, usage={"probe": TokenUsage(1000, 2000)}, pricing=PRICING
    )

    with pytest.raises(StructuredOutputError) as output:
        generate_structured_response("probe", {"type": "boolean"}, provider=provider)

    response = caught.value.provider_response
    assert response is not None
    assert response.usage == TokenUsage(1000, 2000)
    assert response.cost_usd == Decimal("0.01125")
    assert content not in str(caught.value)


def test_schema_failure_still_precedes_accounted_provider_call() -> None:
    class UncalledProvider:
        def generate(self, request: ProviderRequest) -> ProviderResponse:
            pytest.fail("invalid schema must not call provider")

    with pytest.raises(StructuredOutputSchemaError):
        generate_structured_response(
            "probe", {"unsupported": 1}, provider=UncalledProvider()
        )