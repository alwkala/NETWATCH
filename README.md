# NETWATCH — Local Network Intelligence

Privacy-first Windows desktop app: local network discovery, device inventory,
diagnostics and event history. No account, cloud or telemetry; the UI makes no
external requests (fonts are bundled).

```
React UI  ──HTTP+SSE──▶  Go engine (127.0.0.1:random, bearer token)  ──▶  SQLite
 (Wails/WebView2)         discovery · scan · diff/events · fingerprint
```

* `internal/netenv` – OS access behind an interface (Windows: iphlpapi/ARP/ICMP; Linux for dev)
* `internal/engine` – scans, offline detection (2 consecutive misses), events, ping, WoL, port probe
* `internal/store`  – SQLite inventory, history, events
* `internal/api`    – REST + SSE; Host/Origin checks + per-session token
* `frontend/`       – the UI; `HttpNetworkService` replaces `MockNetworkService` automatically when an engine is present
* `main.go`, `app.go` – Wails host (Windows only)

## Develop (any OS)

    go run ./cmd/netwatchd -origin http://localhost:3000     # prints {baseUrl, token}
    cd frontend && npm install && npm run dev
    # open http://localhost:3000/?api=<baseUrl>&token=<token>   (without ?api= you get the mock prototype)

## Build the desktop app (Windows)

    go build -tags desktop,production -ldflags "-w -s -H windowsgui" -o build/bin/netwatch.exe .

Data lives in `%LOCALAPPDATA%\NetWatch\data\network.db` and logs in `%LOCALAPPDATA%\NetWatch\data\netwatch.log`.

## Tests

`go test -race ./...` (the engine runs against a simulated network).
