from datetime import UTC, datetime


class FXError(RuntimeError):
    def __init__(self, message: str = "fx_failed"):
        super().__init__(message)
        self.error_code = "fx_failed"


def normalize_to_myr(amount: float, currency: str) -> dict[str, float | str]:
    rates = {"MYR": 1.0, "USD": 4.7, "SGD": 3.45}
    code = currency.upper()
    if code not in rates:
        raise FXError()
    rate = rates[code]
    return {
        "amount_myr_normalized": round(amount * rate, 2),
        "fx_rate_to_myr": rate,
        "fx_rate_timestamp": datetime.now(UTC).isoformat(),
    }
