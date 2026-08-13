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

## Data Transformation (`clean_prices`)
The `/transform` endpoint in the `api-go` service handles cleaning the `raw_prices` data and pushing it into the `clean_prices` table. This is purely an in-database forward-filling process executed via a SQL query.

Here is how the process works:
1. **Find Date Limits**: It finds the minimum and maximum dates for each ticker in `raw_prices`.
2. **Generate Calendar**: It generates a continuous calendar for every ticker using `generate_series`.
3. **Expose Missing Days**: It performs a `LEFT JOIN` on `raw_prices` to expose missing days with `NULL` values.
4. **Group Values**: It increments a group tracker (`COUNT` window function) every time it hits a real, non-null value.
5. **Forward-Fill**: It uses the `FIRST_VALUE` window function partitioned by this group to cascade the last known price forward over the missing sequence.
6. **Idempotent Upsert**: It performs an `INSERT INTO clean_prices ... ON CONFLICT DO UPDATE` which makes the operation 100% idempotent if run multiple times.

You can trigger this transformation process by sending a GET request to the Go API:
```powershell
curl.exe http://localhost:8080/transform
```
