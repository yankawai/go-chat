# go-chat

Small anonymous WebSocket chat backend in Go.

## Architecture

- `cmd/server` wires configuration, logging, HTTP routing, websocket transport, and shutdown.
- `internal/config` owns environment parsing and runtime defaults.
- `internal/chat` owns message validation, event creation, and the concurrency-safe room abstraction.
- `internal/transport/http` owns HTTP routes, health checks, static compatibility serving, and security headers.
- `internal/transport/websocket` owns Gorilla websocket upgrade, deadlines, origin checks, client queues, and JSON IO.

The backend no longer depends on Gin. It uses the standard `net/http` server plus Gorilla WebSocket, which keeps the runtime small and makes transport concerns explicit.

## Run

```sh
go run ./cmd/server
```

The default HTTP address is `:8080`.

## Configuration

Configuration is provided through environment variables. See `.env.example` for supported values.

Important websocket controls:

- `WS_ALLOWED_ORIGINS`: comma-separated explicit origin allow-list. If empty, same-host and loopback origins are allowed.
- `WS_READ_LIMIT_BYTES`: maximum inbound websocket message size.
- `WS_SEND_QUEUE_SIZE`: per-client outbound queue size. Slow clients are disconnected instead of blocking the room.

## Checks

```sh
go test ./...
go vet ./...
```
