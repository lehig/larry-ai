import json
import logging
import os
import yfinance as yf
from fastapi import FastAPI, Header, HTTPException, Query
from pydantic import BaseModel
import time
import psycopg2
from psycopg2 import extras

# Configure logging to write to ingestion.log inside LOG_DIR
log_dir = os.getenv("LOG_DIR", ".")
os.makedirs(log_dir, exist_ok=True)
logging.basicConfig(
    level=logging.INFO,
    format="%(message)s",
    handlers=[logging.FileHandler(os.path.join(log_dir, "ingestion.log"), mode="a")]
)
logger = logging.getLogger(__name__)

app = FastAPI()


class IngestRequest(BaseModel):
    ticker: str


class IngestResponse(BaseModel):
    ticker: str
    status: str
    rows_ingested: int


def log_event(
    request_id: str,
    ticker: str,
    status_code: int,
    rows_ingested: int,
    path: str,
    method: str,
    latency_ms: int,
) -> None:
    event = {
        "request_id": request_id,
        "ticker": ticker,
        "status_code": status_code,
        "rows_ingested": rows_ingested,
        "path": path,
        "method": method,
        "service": "ingestion",
    }
    logger.info(json.dumps(event))


@app.get("/health")
def health() -> dict:
    return {"status": "ok"}


@app.get("/ready")
def ready() -> dict:
    return {"status": "ready"}


@app.post("/ingest", response_model=IngestResponse)
def ingest(
    payload: IngestRequest,
    x_request_id: str = Header(default=""),
) -> IngestResponse:
    start = time.time()
    request_id = x_request_id or f"req-{time.time_ns()}"
    ticker = payload.ticker.strip().upper()
    if not ticker:
        raise HTTPException(status_code=400, detail="ticker is required")

    try:
        # Download historical data
        data = yf.download(ticker, period="5y", interval="1d")
        if data.empty:
            raise HTTPException(status_code=404, detail="No data found for ticker")

        # Convert to expected format
        # INSERT INTO raw_prices (ticker, date, open, high, low, close, volume)
        data = data.reset_index()
        data.columns = ["date", "open", "high", "low", "close", "volume"]
        data["ticker"] = ticker

        # Ensure column order matches the INSERT statement exactly
        df_for_db = data[["ticker", "date", "open", "high", "low", "close", "volume"]]
        data_tuples = [tuple(x) for x in df_for_db.to_numpy()]

        # Insert into database using a single %s
        insert_query = """
            INSERT INTO raw_prices (ticker, date, open, high, low, close, volume)
            VALUES %s
            ON CONFLICT (ticker, date) DO NOTHING;
        """
        conn = psycopg2.connect(
            host=os.getenv("HOST"),
            user=os.getenv("POSTGRES_USER"),
            password=os.getenv("POSTGRES_PASSWORD"),
            dbname=os.getenv("POSTGRES_DB"),
            port=os.getenv("POSTGRES_PORT"),
        )
        with conn.cursor() as cur:
            extras.execute_values(cur, insert_query, data_tuples)
            conn.commit()
        
        rows_ingested = len(data)

        response = IngestResponse(
            ticker=ticker,
            status="success",
            rows_ingested=rows_ingested,
        )

    except Exception as e:
        logger.error(f"Ingestion failed for {ticker}: {e}")
        raise HTTPException(status_code=500, detail=str(e))

    latency_ms = int((time.time() - start) * 1000)
    log_event(
        request_id=request_id,
        ticker=ticker,
        status_code=200,
        rows_ingested=rows_ingested,
        path="/ingest",
        method="POST",
        latency_ms=latency_ms,
    )
    return response