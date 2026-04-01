package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type App struct {
	db                  *sql.DB
	modelBaseURL        string
	defaultModelVersion string
}

type predictRequest struct {
	Ticker string `json:"ticker"`
}

type predictResponse struct {
	Ticker              string             `json:"ticker"`
	ModelVersion        string             `json:"model_version"`
	RegimeProbabilities map[string]float64 `json:"regime_probabilities"`
	Confidence          float64            `json:"confidence"`
	RiskSignal          string             `json:"risk_signal"`
	TopFeatures         []string           `json:"top_features"`
}

func main() {
	logDir := envOr("LOG_DIR", ".")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Fatalf("failed to create log dir: %v", err)
	}
	logFile, err := os.OpenFile(logDir+"/api.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	port := envOr("PORT", "8080")
	dsn := envOr("DB_DSN", "postgres://market_user:market_pass@localhost:5432/market?sslmode=disable")
	modelBaseURL := strings.TrimRight(envOr("MODEL_BASE_URL", "http://localhost:8000"), "/")
	defaultModelVersion := envOr("DEFAULT_MODEL_VERSION", "v0.1.0")

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	app := &App{
		db:                  db,
		modelBaseURL:        modelBaseURL,
		defaultModelVersion: defaultModelVersion,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", app.handleHealth)
	mux.HandleFunc("/ready", app.handleReady)
	mux.HandleFunc("/recommendation", app.withRequestLog(app.handleRecommendation))
	mux.HandleFunc("/transform", app.withRequestLog(app.handleTransform))

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	log.Printf("api listening on :%s", port)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server error: %v", err)
	}
}

func (a *App) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *App) handleReady(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := a.db.PingContext(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready", "reason": "db_unreachable"})
		return
	}

	readyReq, err := http.NewRequestWithContext(ctx, http.MethodGet, a.modelBaseURL+"/ready", nil)
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready", "reason": "model_unreachable"})
		return
	}
	readyResp, err := http.DefaultClient.Do(readyReq)
	if err != nil || readyResp.StatusCode >= 300 {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready", "reason": "model_unreachable"})
		return
	}
	defer readyResp.Body.Close()

	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (a *App) handleRecommendation(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	ticker := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("ticker")))
	if ticker == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "ticker is required"})
		return
	}

	var exists bool
	err := a.db.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM raw_prices WHERE ticker = $1)", ticker).Scan(&exists)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "db query failed"})
		return
	}
	if !exists {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "ticker not found in seeded dataset"})
		return
	}

	body, _ := json.Marshal(predictRequest{Ticker: ticker})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.modelBaseURL+"/predict?model_version="+a.defaultModelVersion, strings.NewReader(string(body)))
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "model request creation failed"})
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", requestIDFromHeader(r))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "model request failed"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "model service returned non-2xx"})
		return
	}

	var prediction predictResponse
	if err := json.NewDecoder(resp.Body).Decode(&prediction); err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "invalid model response"})
		return
	}
	writeJSON(w, http.StatusOK, prediction)
}

func (a *App) withRequestLog(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID := requestIDFromHeader(r)
		r.Header.Set("X-Request-ID", requestID)
		ticker := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("ticker")))
		modelVersion := a.defaultModelVersion

		rec := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next(rec, r)

		entry := map[string]any{
			"request_id":    requestID,
			"ticker":        ticker,
			"model_version": modelVersion,
			"latency_ms":    time.Since(start).Milliseconds(),
			"status_code":   rec.statusCode,
			"path":          r.URL.Path,
			"method":        r.Method,
			"service":       "api",
		}
		b, _ := json.Marshal(entry)
		log.Println(string(b))
	}
}

func (a *App) handleTransform(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second) // Extended timeout for batch processing
	defer cancel()

	// This query does pure in-database forward filling for ALL tickers:
	// 1. Find min and max dates for each ticker
	// 2. Generate a continuous calendar for EACH ticker
	// 3. LEFT JOIN raw_prices to expose missing days as NULL
	// 4. Increment a "grp" tracker every time we hit a real, non-null value
	// 5. FIRST_VALUE uses that "grp" to cascade the last known price forward
	// 6. ON CONFLICT DO UPDATE makes it 100% idempotent if ran twice.
	query := `
		WITH ticker_limits AS (
			SELECT ticker, MIN(date) AS min_date, MAX(date) AS max_date 
			FROM raw_prices 
			GROUP BY ticker
		),
		monotonic_calendar AS (
			SELECT ticker, generate_series(min_date, max_date, '1 day'::interval)::date AS date
			FROM ticker_limits
		),
		joined_data AS (
			SELECT 
				cal.ticker,
				cal.date,
				rp.open, rp.high, rp.low, rp.close, rp.volume,
				COUNT(rp.close) OVER (PARTITION BY cal.ticker ORDER BY cal.date) AS grp
			FROM monotonic_calendar cal
			LEFT JOIN raw_prices rp ON cal.date = rp.date AND cal.ticker = rp.ticker
		),
		forward_filled AS (
			SELECT 
				ticker,
				date,
				FIRST_VALUE(open) OVER (PARTITION BY ticker, grp ORDER BY date) AS open,
				FIRST_VALUE(high) OVER (PARTITION BY ticker, grp ORDER BY date) AS high,
				FIRST_VALUE(low) OVER (PARTITION BY ticker, grp ORDER BY date) AS low,
				FIRST_VALUE(close) OVER (PARTITION BY ticker, grp ORDER BY date) AS close,
				FIRST_VALUE(volume) OVER (PARTITION BY ticker, grp ORDER BY date) AS volume
			FROM joined_data
		)
		INSERT INTO clean_prices (ticker, date, open, high, low, close, volume)
		SELECT * FROM forward_filled
		ON CONFLICT (ticker, date) DO UPDATE SET
			open = EXCLUDED.open,
			high = EXCLUDED.high,
			low = EXCLUDED.low,
			close = EXCLUDED.close,
			volume = EXCLUDED.volume;
	`

	res, err := a.db.ExecContext(ctx, query)
	if err != nil {
		log.Printf("Transform query failed: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to transform data"})
		return
	}

	rowsAffected, _ := res.RowsAffected()

	writeJSON(w, http.StatusOK, map[string]any{
		"status":        "success",
		"rows_upserted": rowsAffected,
	})
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func writeJSON(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(data)
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func requestIDFromHeader(r *http.Request) string {
	v := strings.TrimSpace(r.Header.Get("X-Request-ID"))
	if v == "" {
		return fmt.Sprintf("req-%d", time.Now().UnixNano())
	}
	return v
}
