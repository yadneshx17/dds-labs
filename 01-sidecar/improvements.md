**Concurrency (core to this lab)**
- One-goroutine-per-request logging (main.go:80). Learning: worker pool + buffered channel (fan-out/fan-in), backpressure, and graceful drain on shutdown. This is the classic sidecar motivation.

- Log writes are synchronous I/O per request with no batching. Learning: buffered writer + periodic flush, or batch writes.

**Robustness (servers, not scripts)**
[x] log.Fatalf inside sidecar handlers (cmd/sidecar/main.go:38,42,55,74) kills the whole process on one bad event. Learning: handlers return HTTP errors; only main() should decide process death.

- No timeouts on http.Server / http.DefaultClient (main.go:34,100). Learning: context + timeouts, why servers need read/write/idle limits.

**Design / reuse**
[x] LogEvent, doRequest, middleware duplicated across 3 services. Learning: extract internal/shared package, or a service.New(port) helper — show DRY without over-abstracting.

- Sidecar path + port hardcoded (cmd/sidecar/main.go:14). Learning: config via flags/env; 12-factor style. Parametrized config

- Plain fmt.Sprintf log lines vs structured JSON logs. Learning: why JSON lines / key-value logs are better for parsing.

- No graceful shutdown (signal.NotifyContext). Learning: cleanly draining in-flight requests + flushing the worker pool on SIGINT.
