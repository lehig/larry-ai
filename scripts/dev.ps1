param(
  [Parameter(Mandatory = $true)]
  [ValidateSet("up", "down", "test", "logs")]
  [string]$Command
)

$ErrorActionPreference = "Stop"

if (-not (Test-Path ".env")) {
  Copy-Item ".env.example" ".env"
  Write-Host "Created .env from .env.example"
}

function Invoke-Compose {
  param([string[]]$ComposeArgs)
  if (-not $ComposeArgs -or $ComposeArgs.Count -eq 0) {
    throw "Compose args are required"
  }
  & docker compose --env-file .env @ComposeArgs
  if ($LASTEXITCODE -ne 0) {
    throw "docker compose failed: docker compose --env-file .env $($ComposeArgs -join ' ')"
  }
}

switch ($Command) {
  "up" {
    Invoke-Compose -ComposeArgs @("up", "--build", "-d")
    Write-Host "Stack started. API: http://localhost:8080, Model: http://localhost:8000"
  }
  "down" {
    Invoke-Compose -ComposeArgs @("down")
  }
  "logs" {
    Invoke-Compose -ComposeArgs @("logs", "-f")
  }
  "test" {
    $health = Invoke-RestMethod -Uri "http://localhost:8080/health"
    if ($health.status -ne "ok") { throw "API /health failed" }

    $ready = Invoke-RestMethod -Uri "http://localhost:8080/ready"
    if ($ready.status -ne "ready") { throw "API /ready failed" }

    $modelReady = Invoke-RestMethod -Uri "http://localhost:8000/ready"
    if ($modelReady.status -ne "ready") { throw "Model /ready failed" }

    $headers = @{ "X-Request-ID" = "smoke-test-001" }
    $rec = Invoke-RestMethod -Uri "http://localhost:8080/recommendation?ticker=AAPL" -Headers $headers
    if ($rec.ticker -ne "AAPL") { throw "Recommendation failed" }
    Write-Host "All smoke checks passed."
  }
}
