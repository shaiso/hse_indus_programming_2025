// Package load_test — простой нагрузочный замер сервера.
//
// Запуск (mock-режим, server overhead):
//
//	go test ./tests/load -run TestLoad -v -count=1 -timeout=120s -args \
//	    -url http://localhost:8080/api/v1/analyze -rps 100 -duration 30s
//
// По умолчанию (без флагов) тест пропускается, чтобы не запускаться вместе
// с обычным `go test ./...`.
package loadtest

import (
	"bytes"
	"encoding/json"
	"flag"
	"io"
	"net/http"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var (
	flagURL      = flag.String("url", "", "target URL (e.g. http://localhost:8080/api/v1/analyze)")
	flagRPS      = flag.Int("rps", 50, "target requests per second")
	flagDuration = flag.Duration("duration", 30*time.Second, "test duration")
	flagWorkers  = flag.Int("workers", 16, "concurrent workers")
	flagPayload  = flag.String("payload", "", "path to JSON request body (optional, defaults to a small Go file)")
)

func TestLoad(t *testing.T) {
	if *flagURL == "" {
		t.Skip("set -url to run load test")
	}

	body := defaultPayload()
	if *flagPayload != "" {
		b, err := os.ReadFile(*flagPayload)
		if err != nil {
			t.Fatalf("read payload: %v", err)
		}
		body = b
	}

	report := runLoad(*flagURL, body, *flagRPS, *flagDuration, *flagWorkers)
	report.Print(t)
}

func runLoad(url string, body []byte, rps int, dur time.Duration, workers int) *Report {
	jobs := make(chan struct{}, 1024)

	var (
		ok        atomic.Int64
		errs      atomic.Int64
		latencyMu sync.Mutex
		latencies = make([]float64, 0, rps*int(dur.Seconds()))
		statusMu  sync.Mutex
		statuses  = make(map[int]int)
	)

	client := &http.Client{Timeout: 30 * time.Second}

	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range jobs {
				start := time.Now()
				req, _ := http.NewRequest("POST", url, bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				resp, err := client.Do(req)
				lat := time.Since(start).Seconds() * 1000
				if err != nil {
					errs.Add(1)
					continue
				}
				_, _ = io.Copy(io.Discard, resp.Body)
				_ = resp.Body.Close()

				latencyMu.Lock()
				latencies = append(latencies, lat)
				latencyMu.Unlock()

				statusMu.Lock()
				statuses[resp.StatusCode]++
				statusMu.Unlock()

				if resp.StatusCode >= 200 && resp.StatusCode < 300 {
					ok.Add(1)
				} else {
					errs.Add(1)
				}
			}
		}()
	}

	interval := time.Second / time.Duration(rps)
	if interval == 0 {
		interval = time.Microsecond
	}
	deadline := time.Now().Add(dur)
	ticker := time.NewTicker(interval)
	for time.Now().Before(deadline) {
		<-ticker.C
		select {
		case jobs <- struct{}{}:
		default:
			// workers saturated; drop tick to avoid unbounded queue
		}
	}
	ticker.Stop()
	close(jobs)
	wg.Wait()

	return &Report{
		RPS:       rps,
		Duration:  dur,
		Workers:   workers,
		OK:        ok.Load(),
		Errs:      errs.Load(),
		Latencies: latencies,
		Statuses:  statuses,
	}
}

type Report struct {
	RPS       int
	Duration  time.Duration
	Workers   int
	OK        int64
	Errs      int64
	Latencies []float64
	Statuses  map[int]int
}

func (r *Report) Print(t *testing.T) {
	t.Helper()
	if len(r.Latencies) == 0 {
		t.Fatalf("no completed requests: errs=%d", r.Errs)
	}
	sort.Float64s(r.Latencies)
	p := func(q float64) float64 {
		idx := int(float64(len(r.Latencies)-1) * q)
		return r.Latencies[idx]
	}
	total := r.OK + r.Errs
	actualRPS := float64(total) / r.Duration.Seconds()

	t.Logf("=== Load Report ===")
	t.Logf("target_rps: %d", r.RPS)
	t.Logf("duration:   %s", r.Duration)
	t.Logf("workers:    %d", r.Workers)
	t.Logf("requests:   %d (ok=%d, err=%d)", total, r.OK, r.Errs)
	t.Logf("actual_rps: %.1f", actualRPS)
	t.Logf("latency_ms p50=%.1f p95=%.1f p99=%.1f max=%.1f",
		p(0.50), p(0.95), p(0.99), r.Latencies[len(r.Latencies)-1])
	t.Logf("statuses:   %v", r.Statuses)
}

func defaultPayload() []byte {
	body := map[string]any{
		"language": "go",
		"format":   "yaml",
		"files": []map[string]string{
			{
				"name":    "main.go",
				"content": "package main\nimport \"github.com/gin-gonic/gin\"\nfunc main() {\n  r := gin.Default()\n  r.GET(\"/ping\", func(c *gin.Context) { c.JSON(200, gin.H{\"ok\": true}) })\n  r.Run()\n}\n",
			},
		},
	}
	b, _ := json.Marshal(body)
	return b
}
