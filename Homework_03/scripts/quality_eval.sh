#!/usr/bin/env bash
# Прогоняет качественный тест на 3 эталонных проектах через РЕАЛЬНУЮ LLM.
# Требует OPENAI_API_KEY в окружении.
set -euo pipefail

PORT="${PORT:-8089}"

if [[ -z "${OPENAI_API_KEY:-}" ]]; then
  echo "[quality] OPENAI_API_KEY is required" >&2
  exit 1
fi

echo "[quality] starting server on :$PORT (real LLM)..."
PORT="$PORT" LOG_LEVEL=warn go run ./cmd/server &
PID=$!
trap 'kill "$PID" 2>/dev/null || true' EXIT

for _ in {1..30}; do
  if curl -sf "http://localhost:$PORT/api/v1/health" >/dev/null; then break; fi
  sleep 0.5
done

go test ./tests/quality -run TestQuality -v -count=1 -timeout=600s -args \
  -url "http://localhost:$PORT/api/v1/analyze"
