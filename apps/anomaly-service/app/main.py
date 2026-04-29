from fastapi import FastAPI
from pydantic import BaseModel

app = FastAPI(title="LeakGuard Anomaly Service", version="0.1.0")


class ScoreRequest(BaseModel):
    feature_vector: dict


@app.get("/healthz")
def healthz():
    return {"ok": True}


@app.post("/score")
def score(req: ScoreRequest):
    # MVP scaffold: deterministic placeholder score.
    # Real implementation will compute features + run an anomaly model.
    return {"anomaly_score": 0.42, "model_version": "stub-v0"}
