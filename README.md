# larry-ai

Phase 0 bootstrap for a production-style ML stack:
- `api-go`: Go API (`/health`, `/ready`, `/recommendation`)
- `model-py`: Python model service (`/health`, `/ready`, `/predict`, `/model-info`)
- `db`: Postgres with first-run seed dataset

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
