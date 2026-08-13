CREATE TABLE IF NOT EXISTS raw_prices (
  ticker TEXT NOT NULL,
  date DATE NOT NULL,
  open NUMERIC(12, 4) NOT NULL,
  high NUMERIC(12, 4) NOT NULL,
  low NUMERIC(12, 4) NOT NULL,
  close NUMERIC(12, 4) NOT NULL,
  volume BIGINT NOT NULL,
  PRIMARY KEY (ticker, date),
  CHECK (open > 0 AND high > 0 AND low > 0 AND close > 0 AND volume > 0 AND high >= low)
);

INSERT INTO raw_prices (ticker, date, open, high, low, close, volume)
VALUES
  ('AAPL', DATE '2025-01-02', 190.00, 193.50, 189.50, 192.30, 55000000),
  ('AAPL', DATE '2025-01-03', 192.30, 194.10, 191.20, 193.60, 42000000),
  ('MSFT', DATE '2025-01-02', 375.00, 379.40, 373.10, 378.90, 21000000)
ON CONFLICT (ticker, date) DO NOTHING;


