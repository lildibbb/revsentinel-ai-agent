from fastapi import FastAPI

from app.schemas import (
    Citation,
    GenerateReasoningRequest,
    GenerateReasoningResponse,
    ScoreRequest,
)

app = FastAPI(title="LeakGuard Anomaly Service", version="0.1.0")


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
    return GenerateReasoningResponse(
        status="success",
        prompt_version="v1",
        response_schema_version="v1",
        summary="Case requires analyst review due to policy mismatch.",
        recommended_action="Validate discount exception approval path.",
        citations=[
            Citation(
                source="policy",
                reference="POL-12",
                excerpt="Maximum discount 10%",
            )
        ],
        amount_original=12000.0,
        currency_original="USD",
        amount_myr_normalized=56400.0,
        fx_rate_to_myr=4.7,
    )
