# go-structured-logs

Go structured JSON logging library. Module: `github.com/jnpt-golangs/logging` (v1.0.0, Go 1.23.2).

## Build & Run

```bash
go build ./...
go vet ./...
cd examples/gin-example && go run main.go
```

No test files exist yet.

## Package Structure

```
pkg/
  logger/      # Core: AppLogger, Config, LogEntry, MaskingUtil, correlation ID
  middleware/  # Gin middleware (LoggingMiddleware)
  httpclient/  # LoggingTransport (http.RoundTripper) + NewClient factory
examples/
  gin-example/ # Full Gin server demo
```

## Architecture

All logs emit as single JSON line to stdout via `json.Marshal(LogEntry)`.

Two log types (`type` field in JSON):
- `"application"` — from `Info`, `Debug`, `Warn`, `Error`, `ErrorWithErr`, `WarnError`
- `"request"` — from middleware and httpclient transport

Data flow:
1. `logger.New("name")` creates `AppLogger` using global `defaultConfig`
2. Incoming: `LoggingMiddleware` extracts/generates `X-Correlation-ID`, injects into context, logs request before `c.Next()`, captures response via wrapped `responseWriter`, logs response
3. Outgoing: `httpclient.NewClient` returns `*http.Client` with `LoggingTransport` that propagates correlation ID header, buffers bodies, logs both sides
4. HTTP status → level: `<400` INFO, `400–499` WARN, `>=500` ERROR

## Key Conventions

**Config is always optional** — all constructors accept `nil` and fall back to `DefaultConfig()`. Set global default once at startup with `logger.ConfigureDefaults(cfg)`.

**Extra fields** — logger methods accept `extra ...any`. Can be `map[string]interface{}` or any struct (JSON-marshalled). Multiple extras merge into one `extra` map in the log entry.

**Correlation ID** — always pass `context.Context`. Stored with typed key `contextKey("correlation_id")`. Never use a plain string key. Retrieve with `logger.GetCorrelationID(ctx)`.

**Masking** — `MaskingUtil` masks headers and body fields case-insensitively, recursively for nested maps/slices.
- Default masked headers: `Authorization`, `X-Api-Key`, `Cookie`, `Set-Cookie`
- Default masked fields: `password`, `secret`, `token`, `creditCard`, `ssn`
- Extend via `Config.MaskedHeaders` / `Config.MaskedFields`

**Path patterns** — `ExcludePatterns`/`IncludePatterns` support exact match, `/*` (prefix), `/**` (prefix) only — no full glob or regex.

**Stack traces** — captured on `ErrorWithErr`/`WarnError` when `Config.ApplicationLogging.IncludeStackTrace = true` (default true, depth 50).

## LogEntry JSON Fields

| Field | Notes |
|---|---|
| `@timestamp` | ISO 8601 with timezone |
| `@version` | always `"1"` |
| `application` | from `Config.ApplicationName` |
| `logger_name` | name passed to `New()` |
| `thread_name` | goroutine ID (`goroutine-N`) |
| `level` / `level_value` | TRACE=5000 DEBUG=10000 INFO=20000 WARN=30000 ERROR=40000 |
| `type` | `"application"` or `"request"` |
| `correlation_id` | from context |
| `method`, `uri` | request type only |
| `status_code`, `duration_ms` | response type only |
| `request_body`, `response_body` | masked if configured |
| `extra` | merged extra fields |
| `error` | class, message, stack_trace (optional) |

## Default Config Values

| Property | Default |
|---|---|
| `ApplicationName` | `"application"` |
| `Enabled` | `true` |
| `MinLevel` | `"DEBUG"` |
| `RequestLogging.LogHeaders` | `true` |
| `RequestLogging.LogBody` | `true` |
| `RequestLogging.LogResponseBody` | `true` |
| `RequestLogging.MaxBodySize` | `10240` bytes |
| `RequestLogging.ExcludePatterns` | `/health`, `/metrics`, `/favicon.ico` |
| `ApplicationLogging.IncludeStackTrace` | `true` |
| `ApplicationLogging.MaxStackTraceDepth` | `50` |
