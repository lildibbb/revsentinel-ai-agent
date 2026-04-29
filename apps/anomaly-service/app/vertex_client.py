from typing import Any


class VertexCallError(RuntimeError):
    def __init__(self, message: str = "vertex_call_failed"):
        super().__init__(message)
        self.error_code = "vertex_call_failed"


def generate_grounded_reasoning(
    case_context: dict[str, Any],
    retrieved_context: list[dict[str, str]],
) -> dict[str, Any]:
    if not retrieved_context:
        raise VertexCallError()
    return {
        "summary": "Case requires analyst review due to policy mismatch.",
        "recommended_action": "Validate discount exception approval path.",
        "citations": retrieved_context,
    }
