"""Provider-neutral token usage and per-call USD cost calculation."""

from dataclasses import dataclass
from decimal import Decimal, localcontext


@dataclass(frozen=True)
class TokenUsage:
    input_tokens: int
    output_tokens: int

    def __post_init__(self) -> None:
        for name in ("input_tokens", "output_tokens"):
            value = getattr(self, name)
            if type(value) is not int or value < 0:
                raise ValueError(f"{name} must be a non-negative integer")

    @property
    def total_tokens(self) -> int:
        return self.input_tokens + self.output_tokens


@dataclass(frozen=True)
class ModelPricing:
    input_usd_per_million_tokens: Decimal = Decimal("0")
    output_usd_per_million_tokens: Decimal = Decimal("0")

    def __post_init__(self) -> None:
        for name in ("input_usd_per_million_tokens", "output_usd_per_million_tokens"):
            value = getattr(self, name)
            if not isinstance(value, Decimal) or not value.is_finite() or value < 0:
                raise ValueError(f"{name} must be a finite non-negative Decimal")

    def calculate_cost(self, usage: TokenUsage) -> Decimal:
        input_count = Decimal(usage.input_tokens)
        output_count = Decimal(usage.output_tokens)
        input_rate = self.input_usd_per_million_tokens
        output_rate = self.output_usd_per_million_tokens
        # Retain exact products and their sum regardless of the caller's context.
        precision = (
                max(len(input_count.as_tuple().digits), len(output_count.as_tuple().digits))
                + max(len(input_rate.as_tuple().digits), len(output_rate.as_tuple().digits))
                + abs(
            (input_rate.adjusted() - len(input_rate.as_tuple().digits))
            - (output_rate.adjusted() - len(output_rate.as_tuple().digits))
        )
                + 2
        )
        with localcontext() as context:
            context.prec = max(28, precision)
            return (input_count * input_rate + output_count * output_rate) / Decimal(
                "1000000"
            )
