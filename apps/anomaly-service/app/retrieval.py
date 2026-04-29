from typing import Any
from uuid import UUID


class RetrievalError(RuntimeError):
    def __init__(self, message: str = "retrieval_failed"):
        super().__init__(message)
        self.error_code = "retrieval_failed"


def fetch_case_context(case_id: UUID) -> dict[str, Any]:
    if str(case_id) == "00000000-0000-0000-0000-000000000000":
        raise RetrievalError()
    return {
        "case_id": str(case_id),
        "amount_original": 12000.0,
        "currency_original": "USD",
    }


def fetch_grounding_context(case_context: dict[str, Any]) -> list[dict[str, str]]:
    if not case_context:
        raise RetrievalError()
    return [
        {
            "source": "policy",
            "reference": "POL-12",
            "excerpt": "Maximum discount 10%",
        }
    ]
