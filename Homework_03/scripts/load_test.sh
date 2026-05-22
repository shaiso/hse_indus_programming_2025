#!/usr/bin/env bash
# Запускает сервис в mock-режиме и прогоняет нагрузочные тесты.
# Использование: ./scripts/load_test.sh [rps] [duration]
set -euo pipefail

RPS="${1:-100}"
DURATION="${2:-30s}"
PORT="${PORT:-8089}"

echo "[load] starting server on :$PORT (mock LLM)..."
USE_MOCK_LLM=true PORT="$PORT" LOG_LEVEL=warn RATE_LIMIT_PER_MIN=1000000 \
  go run ./cmd/server &
PID=$!

cleanup() { kill "$PID" 2>/dev/null || true; }
trap cleanup EXIT

# wait for server
for _ in {1..30}; do
  if curl -sf "http://localhost:$PORT/api/v1/health" >/dev/null; then break; fi
  sleep 0.5
done

echo "[load] running test: rps=$RPS duration=$DURATION"
go test ./tests/load -run TestLoad -v -count=1 -timeout=300s -- \
  -url "http://localhost:$PORT/api/v1/analyze" \
  -rps "$RPS" -duration "$DURATION" -workers 32

cleanup
