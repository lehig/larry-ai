# Fintech-Grade Production ML Capstone Checklist
## Financial Market Decision Support System

---

## Project Goal

Build a production-style financial ML decision-support platform that demonstrates:

- Probabilistic time-series modeling
- Bias-safe financial data processing
- Reproducible ML pipelines
- Containerized deployment
- Observability and traceability
- Model versioning and rollback capability
- CI/CD artifact discipline

This project is designed to reflect **real fintech ML system expectations**, not notebook-only ML workflows.

---

# Phase 0 — Reproducible Production Dev Environment

## Environment & Bootstrapping
- [X] Docker Desktop installed and Linux engine verified
- [X] One command to run full stack (`make up` or script) {with `start.bat`}
- [X] One command to shut down stack
- [X] One command to run tests
- [X] Automatic seed of minimal dataset on first run
- [X] `.env.example` committed

## Service Health & Readiness
- [X] Go API exposes:
  - [X] `/health`
  - [X] `/ready`
- [X] Python model service exposes:
  - [X] `/health`
  - [X] `/ready`
- [X] Go `/ready` fails if:
  - [X] DB unreachable
  - [X] Model service unreachable

## Logging & Traceability
- [X] Structured JSON logs (Go + Python)
- [X] Request correlation ID passed:
  - Web → Go → Model → DB logs
- [X] Logs include:
  - request_id
  - ticker
  - model_version
  - latency_ms
  - status_code

---

# Phase 1 — Financial Data Engineering (Integrity First)

## Ingestion
- [ ] Select data source (API or historical dataset)
- [ ] Create `raw_prices` table
- [ ] Unique constraint `(ticker, date)`
- [ ] Idempotent upsert ingestion logic

## Data Validation
- [ ] OHLC consistency checks
- [ ] Null validation
- [ ] Monotonic time enforcement
- [ ] Volume sanity checks

## Transform Layer
- [ ] Create `features` table
- [ ] Rolling returns
- [ ] Volatility metrics
- [ ] Momentum indicators

## Backfill + Incremental Updates
- [ ] Full historical backfill job
- [ ] Daily incremental updater
- [ ] Re-runnable without duplication

## Bias Safety (High Signal)
- [ ] No lookahead bias
- [ ] No data leakage
- [ ] Strict train/test time separation

## Documentation
- [ ] `docs/data-pipeline.md`
Includes:
- Data lineage
- Bias prevention explanation
- Failure modes
- Recovery procedures

---

# Phase 2 — ML / Algorithms (Production ML)

## Baseline Signal
- [ ] Deterministic baseline (e.g., MA crossover)

## Probabilistic Model (Choose One)
- [ ] Markov Regime Model
- [ ] Hidden Markov Model
- [ ] Logistic / Tree Model using regime features

## Model Artifact Management
- [ ] Model artifacts versioned
- [ ] Model metadata stored:
  - Training window
  - Feature version
  - Metrics snapshot
  - Data snapshot reference

## Backtesting System
- [ ] Walk-forward validation
- [ ] Immutable backtest result storage
- [ ] Metrics captured:
  - Calibration
  - Drawdown
  - Hit rate
  - Regime classification accuracy

## Inference Output (Decision Support)
Return:
- Regime probabilities
- Confidence score
- Risk signal
- Top contributing features

---

# Phase 3 — Production ML Serving

## Model Service
- [ ] Dockerized Python inference service
- [ ] `/predict`
- [ ] `/metrics`
- [ ] `/model-info`

## Go API Layer
- [ ] Typed API contracts
- [ ] Timeout + retry logic
- [ ] Graceful degradation if model unavailable
- [ ] Input validation and schema enforcement

## Model Versioning
- [ ] Default model = “active”
- [ ] Allow explicit version selection via query
- [ ] Ability to rollback active model pointer

---

# Phase 4 — Observability (Fintech-Level Signal)

## Metrics
Track:
- [ ] Request latency (p50 / p95)
- [ ] Error rate
- [ ] Prediction distribution tracking
- [ ] Regime distribution drift

## Logging Queries Should Answer
- Which model served this decision?
- What data window was used?
- What was inference latency?

---

# Phase 5 — CI/CD (Artifact Discipline)

## Continuous Integration
- [ ] Unit tests (model + feature code)
- [ ] Lint + type checks
- [ ] Container builds
- [ ] Images tagged with commit SHA

## Continuous Deployment
- [ ] Deploy only pre-built images
- [ ] No rebuild during deploy
- [ ] Environment-specific config separation

---

# Phase 6 — Production Deployment

## Infrastructure Requirements
- [ ] External container hosting
- [ ] Persistent Postgres database
- [ ] Container registry
- [ ] Secret management solution

---

# Phase 7 — Resume-Level Stretch Features (Choose 1–2)

- [ ] Shadow model deployment
- [ ] Feature drift monitoring
- [ ] Online retraining pipeline
- [ ] Streaming inference (Kafka or queue system)
- [ ] Feature store architecture pattern

---

# Smoke Tests

## Model Service

`curl /health`

`curl /model-info`

## API

`GET /recommendation?ticker=APPL`

## End-to-End

`Web => API => Model => DB`

---

# Definition of Done (Capstone Ready)

- [ ] System deployable with one command
- [ ] All services containerized
- [ ] Data pipeline documented and bias-safe
- [ ] Model versioned and reproducible
- [ ] Observability implemented
- [ ] CI builds and deploys production artifacts
- [ ] Live deployment accessible externally

---

# Optional Deployment Stack (Non-AWS Example)

Frontend: Vercel  
API + Model Containers: Render or Fly.io  
Database: Neon Postgres  
CI/CD: GitHub Actions  

---

# Resume Narrative Target

> Built a containerized financial ML decision-support platform using Go, Python, and Docker, implementing probabilistic regime modeling, bias-safe backtesting, model versioning, and production observability; deployed via CI/CD to managed container infrastructure with persistent Postgres storage.