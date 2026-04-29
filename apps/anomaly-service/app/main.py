import os
from uuid import uuid4

import httpx
from fastapi import FastAPI
from fastapi.responses import JSONResponse
from pydantic import ValidationError

from app.fx import FXError, normalize_to_myr
from app.retrieval import RetrievalError, fetch_case_context, fetch_grounding_context
from app.schemas import (
    Citation,
    GenerateReasoningRequest,
    GenerateReasoningResponse,
    ScoreRequest,
)
from app.vertex_client import VertexCallError, generate_grounded_reasoning

app = FastAPI(title="LeakGuard Anomaly Service", version="0.1.0")
PROMPT_VERSION = "v1"
RESPONSE_SCHEMA_VERSION = "v1"
MODEL_PROVIDER = "vertex-ai"
MODEL_NAME = "gemini-2.5-pro"
MODEL_VERSION = "v1"
CASE_SERVICE_URL = os.getenv("CASE_SERVICE_URL", "http://case-service:8083")


class ReasoningPersistenceError(RuntimeError):
    pass


def persist_reasoning_artifact(
    case_id: str,
    case_context: dict,
    fx_result: dict,
    reasoning: dict,
):
    payload = {
        "status": "success",
        "model_provider": MODEL_PROVIDER,
        "model_name": MODEL_NAME,
        "model_version": MODEL_VERSION,
        "prompt_version": PROMPT_VERSION,
        "response_schema_version": RESPONSE_SCHEMA_VERSION,
        "summary": reasoning["summary"],
        "recommended_action": reasoning["recommended_action"],
        "citations": reasoning["citations"],
        "amount_original": case_context["amount_original"],
        "currency_original": case_context["currency_original"],
        "amount_myr_normalized": fx_result["amount_myr_normalized"],
        "fx_rate_to_myr": fx_result["fx_rate_to_myr"],
        "trace_id": str(uuid4()),
    }
    response = httpx.post(
        f"{CASE_SERVICE_URL}/cases/{case_id}/reasoning",
        json=payload,
        timeout=10.0,
    )
    if response.status_code >= 300:
        raise ReasoningPersistenceError(
            f"case_service_persist_failed: status={response.status_code}"
        )


@app.get("/healthz")
def healthz():
    return {"ok": True}


@app.post("/score")
def score(req: ScoreRequest):
    # MVP scaffold: deterministic placeholder score.
    # Real implementation will compute features + run an anomaly model.
    return {"anomaly_score": 0.42, "model_version": "stub-v0"}


@app.post("/reasoning/generate", response_model=GenerateReasoningResponse)
def generate_reasoning(req: GenerateReasoningRequest):
    def _error(status_code: int, error_code: str, error_message: str):
        return JSONResponse(
            status_code=status_code,
            content={
                "status": "failed",
                "error_code": error_code,
                "error_message": error_message,
            },
        )

    try:
        case_context = fetch_case_context(req.case_id)
        grounding_context = fetch_grounding_context(case_context)
        fx_result = normalize_to_myr(
            case_context["amount_original"],
            case_context["currency_original"],
        )
        reasoning = generate_grounded_reasoning(case_context, grounding_context)
        persist_reasoning_artifact(str(req.case_id), case_context, fx_result, reasoning)
        return GenerateReasoningResponse(
            status="success",
            prompt_version=PROMPT_VERSION,
            response_schema_version=RESPONSE_SCHEMA_VERSION,
            summary=reasoning["summary"],
            recommended_action=reasoning["recommended_action"],
            citations=[Citation.model_validate(c) for c in reasoning["citations"]],
            amount_original=case_context["amount_original"],
            currency_original=case_context["currency_original"],
            amount_myr_normalized=fx_result["amount_myr_normalized"],
            fx_rate_to_myr=fx_result["fx_rate_to_myr"],
        )
    except RetrievalError as err:
        return _error(502, "retrieval_failed", str(err))
    except FXError as err:
        return _error(502, "fx_failed", str(err))
    except VertexCallError as err:
        return _error(502, "vertex_call_failed", str(err))
    except ReasoningPersistenceError as err:
        return _error(502, "reasoning_persistence_failed", str(err))
    except ValidationError as err:
        return _error(500, "schema_validation_failed", str(err))
    except (KeyError, TypeError) as err:
        return _error(500, "schema_validation_failed", str(err))
