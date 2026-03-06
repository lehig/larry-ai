# larry-ai

Phase 0 bootstrap for a production-style ML stack:
- `api-go`: Go API (`/health`, `/ready`, `/recommendation`)
- `model-py`: Python model service (`/health`, `/ready`, `/predict`, `/model-info`)
- `ingestion-py`: Python data ingestion service (`/health`, `/ready`, `/ingest`)
- `db`: Postgres with first-run seed dataset
- `logs/`: Subdirectories for each service to stream and persist app logs

## Prereqs
- Docker Desktop running
- Linux engine active (`docker info` shows `OSType: linux`)

## Quick start
```powershell
make up
```

## Stop
```powershell
make down
```

## Optional Make targets
```powershell
make test
make logs
```

## Notes
- DB seed SQL runs only on first initialization of the Postgres volume.
- To rerun seed from scratch, remove containers and volume:
```powershell
docker compose --env-file .env down -v
```

## Ingestion API usage
To ingest up to 5 years of historical data for a ticker into the Postgres database, send a POST request to localhost:
```powershell
curl.exe -X POST "http://localhost:8081/ingest" -H "Content-Type: application/json" -d '{"ticker": "AAPL"}'
```
Or use the Swagger UI at `http://localhost:8081/docs`
