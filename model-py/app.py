import json
import logging
import os
import time
from typing import Dict, List

from fastapi import FastAPI, Header, HTTPException, Query
from pydantic import BaseModel

# Configure logging to write to model.log
logging.basicConfig(
    level=logging.INFO,
    format="%(message)s",
    handlers=[logging.FileHandler("model.log", mode="a")]
)
logger = logging.getLogger(__name__)

app = FastAPI()


class PredictRequest(BaseModel):
    ticker: str


class PredictResponse(BaseModel):
    ticker: str
    model_version: str
    regime_probabilities: Dict[str, float]
    confidence: float
    risk_signal: str
    top_features: List[str]


def log_event(
    request_id: str,
    ticker: str,
    model_version: str,
    latency_ms: int,
    status_code: int,
    path: str,
    method: str,
) -> None:
    event = {
        "request_id": request_id,
        "ticker": ticker,
        "model_version": model_version,
        "latency_ms": latency_ms,
        "status_code": status_code,
        "path": path,
        "method": method,
        "service": "model",
    }
    logger.info(json.dumps(event))


@app.get("/health")
def health() -> dict:
    return {"status": "ok"}


@app.get("/ready")
def ready() -> dict:
    return {"status": "ready"}


@app.get("/model-info")
def model_info() -> dict:
    return {
        "model_name": "baseline-regime-classifier",
        "active_model_version": os.getenv("MODEL_VERSION", "v0.1.0"),
    }


@app.post("/predict", response_model=PredictResponse)
def predict(
    payload: PredictRequest,
    model_version: str = Query(default=os.getenv("MODEL_VERSION", "v0.1.0")),
    x_request_id: str = Header(default=""),
) -> PredictResponse:
    start = time.time()
    request_id = x_request_id or f"req-{time.time_ns()}"
    ticker = payload.ticker.strip().upper()
    if not ticker:
        raise HTTPException(status_code=400, detail="ticker is required")

    response = PredictResponse(
        ticker=ticker,
        model_version=model_version,
        regime_probabilities={"bull": 0.61, "bear": 0.19, "sideways": 0.20},
        confidence=0.61,
        risk_signal="moderate",
        top_features=["return_5d", "volatility_20d", "momentum_14d"],
    )
    latency_ms = int((time.time() - start) * 1000)
    log_event(
        request_id=request_id,
        ticker=ticker,
        model_version=model_version,
        latency_ms=latency_ms,
        status_code=200,
        path="/predict",
        method="POST",
    )
    return response
