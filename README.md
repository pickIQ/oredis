<p align="center">
  <img src="assets/logo.png" width="420" alt="O'Redis">
</p>

<p align="center">
  <strong>An in-memory key-value store with TTL, served over HTTP.</strong><br>
  Written in plain Go — no frameworks, no third-party dependencies.
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.22%2B-00ADD8?logo=go&logoColor=white" alt="Go 1.22+">
  <img src="https://img.shields.io/badge/dependencies-none-success" alt="No dependencies">
  <img src="https://img.shields.io/badge/license-MIT-blue" alt="MIT License">
</p>

---

## About

O'Redis is a minimal Redis-flavoured key-value store exposed through a JSON HTTP API. Keys can be stored
indefinitely or with a time-to-live, after which they stop being visible to clients.

This is a **learning project**. It was built deliberately from the ground up — no router libraries, no logging
libraries, no DI containers — to understand how a Go web service actually works underneath the abstractions
most projects reach for by default. Everything here is the standard library.

It is not intended for production use. If you need a real key-value store, use [Redis](https://redis.io).

### Design notes

A few decisions worth calling out, since they are the point of the exercise:

- **Routing** uses `net/http.ServeMux` with the method-and-pattern syntax introduced in Go 1.22
  (`GET /keys/{key}`), which also gives correct `405 Method Not Allowed` handling for free.
- **Concurrency** follows the idiomatic Go model: one goroutine per request, shared state guarded by a
  `sync.RWMutex`. This differs from real Redis, which is single-threaded and gets atomicity for free.
- **TTL** is stored as an absolute expiry instant, not a duration. Expiry is currently lazy — an expired key
  becomes invisible to clients on read, but is not yet reclaimed from memory.
- **Middleware** is plain `func(http.Handler) http.Handler` composition. No framework required.
- **Dependencies are injected explicitly** through constructors. `cmd/server` is the only place in the
  codebase that knows which `Store` implementation is in use.

---

## Quick start

```bash
git clone https://github.com/pickIQ/oredis.git
cd oredis
go run ./cmd/server
```

The server listens on `:8080` by default.

```bash
# store a key that lives forever
curl -X PUT localhost:8080/keys/greeting \
  -H "Content-Type: application/json" \
  -d '{"value":"hello"}'

# read it back
curl localhost:8080/keys/greeting
# {"value":"hello"}

# store a key that expires in 30 seconds
curl -X PUT localhost:8080/keys/session \
  -H "Content-Type: application/json" \
  -d '{"value":"abc123","ttl_seconds":30}'

# delete a key
curl -X DELETE localhost:8080/keys/greeting
```

---

## API

| Method   | Path           | Request body                              | Success          | Errors |
| -------- | -------------- | ----------------------------------------- | ---------------- | ------ |
| `GET`    | `/keys/{key}`  | —                                         | `200` + value    | `404`  |
| `PUT`    | `/keys/{key}`  | `{"value": string, "ttl_seconds": int}`   | `204`            | `400`  |
| `DELETE` | `/keys/{key}`  | —                                         | `204`            | `404`  |
| `GET`    | `/healthz`     | —                                         | `200`            | —      |
| `GET`    | `/version`     | —                                         | `200`            | —      |

### Semantics

`PUT` is an **upsert** and is idempotent: it creates the key if absent and fully replaces it if present,
including its TTL. Sending the same request twice leaves the store in the same state.

`ttl_seconds` is optional. Omitting it, or sending `0`, stores the key **without expiry** — matching the
behaviour of `SET` without `EX` in Redis.

`GET` and `DELETE` agree on what "exists" means: an expired key is reported as missing by both, even though
it may still occupy memory until it is reclaimed.

### Response shapes

```jsonc
// 200 OK
{ "value": "hello" }

// 404 Not Found / 400 Bad Request
{ "error": "key not found" }
```

`204 No Content` responses carry no body, as the status implies.

---

## Configuration

Configuration is read from command-line flags, falling back to environment variables, then to defaults.
Invalid values fail fast at startup rather than being silently ignored.

| Flag     | Environment variable | Default | Description      |
| -------- | -------------------- | ------- | ---------------- |
| `-port`  | `OREDIS_PORT`        | `8080`  | HTTP listen port |

```bash
go run ./cmd/server -port=9090
OREDIS_PORT=9090 go run ./cmd/server
```

---

## Project layout

```
cmd/
  server/           entry point — wires dependencies together, nothing else
internal/
  config/           configuration loading and validation
  store/            the key-value store: interface, in-memory implementation, TTL
  httpapi/          HTTP handlers and route registration
  middleware/       recovery and request logging
  response/         JSON response helpers shared across layers
```

The dependency direction is strictly one-way: `httpapi` depends on the `store` *interface*, never on a
concrete implementation. `middleware` knows nothing about the application's domain at all.

---

## Development

```bash
go test ./...            # run the test suite
go test -race ./...      # run it with the race detector
go test -v ./...         # verbose, with subtest names
go vet ./...             # static analysis
gofmt -l .               # list files that need formatting
```

The test suite covers the store (unit tests, TTL behaviour, and a concurrent stress test under `-race`) and
the HTTP layer (table-driven tests through `net/http/httptest`, exercising routing, decoding, status codes,
and store side effects).

> **Note:** `-race` requires cgo and a working C toolchain. On Windows this is easiest under WSL.

---

## Roadmap

Planned, in rough order:

- [ ] Active TTL eviction — a background goroutine that reclaims expired keys, cancellable via `context`
- [ ] Graceful shutdown on `SIGINT`/`SIGTERM`, draining in-flight requests
- [ ] Benchmarks and `pprof` profiling; sharded map to reduce mutex contention
- [ ] Snapshot persistence, so data survives a restart
- [ ] Container image and CI pipeline

---

## License

MIT. See [LICENSE](LICENSE).
