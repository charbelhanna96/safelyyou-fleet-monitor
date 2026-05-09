# SafelyYou Fleet Monitor

## How to Run

```bash
CSV_PATH=devices.csv go run ./cmd/server
```

The service listens on port `6733` by default. After starting the server, run the simulator against port `6733`.

## How to Run Unit Tests

```bash
go test -race ./...
```

## Time Spent and Most Difficult Part

I spent around 5-6 non-consecutive hours on this challenge.

The most difficult part was handling edge cases in the uptime calculation. For example, the uptime window needs to be inclusive: a device with heartbeats at minutes 0, 1, 2, 3, and 4 has 5 possible heartbeat slots, not 4. That is why the calculation includes the final minute.

I also considered duplicate heartbeats within the same minute. Without handling that case, uptime could exceed 100%, which is not meaningful. I chose to cap uptime at 100% to keep the value valid while keeping the implementation simple for the challenge.

## Extending the Data Model

The current implementation keeps aggregate state per device rather than storing every raw event in memory. It stores a sum and count, which allows average upload time to be calculated in constant time.

To support more metrics, I would follow the same pattern. For example, CPU usage could be tracked with a CPU usage sum and count, exposed through a new method and endpoint.

For production scale historical metrics, I would not keep all raw events in memory. I would keep only active state in memory and store historical metrics in a database. A database backed store could replace the current memory store without changing the HTTP handlers significantly.

## Runtime Complexity

- `AddHeartbeat`: O(1)
- `AddUploadTime`: O(1)
- `GetStats`: O(1)
- `CSV device loading`: O(n), performed once at startup

The main request paths use simple map lookups and aggregate updates. There is no per request scan over historical data.

## Architecture

The project uses a small layered structure:

- `cmd/server`: application entrypoint
- `internal/api/handlers`: HTTP handlers
- `internal/api/web`: request and response data transfer objects
- `internal/device`: device related types and CSV loading
- `internal/store`: in memory device state and statistics
- `internal/config`: environment based configuration

The storage layer is separated behind an interface used by the handlers. This keeps the HTTP layer separate from storage details and makes it easier to replace the memory store with a database later.

## Dependencies

- `github.com/joho/godotenv`: loads `.env` files for local development. Environment variables would normally be injected directly in deployment environments.

No router framework was added. Go's standard `net/http` ServeMux supports method based routes and path parameters natively, which covers all required endpoints without extra dependencies.

## Security

- `upload_time` must not be negative.
- HTTP server timeouts are configured for read, write, and idle connections.
- No authentication was added because it is not part of the challenge requirements.

## Known Limitations

1. **No persistence**: All device state is stored in memory. If the service restarts, all data is lost.

2. **Timestamp validation**: Strict future timestamp validation was intentionally not enforced. In a production system, I would validate unreasonable future timestamps while still allowing delayed or buffered device metrics.

3. **Uptime deduplication**: The uptime formula counts raw heartbeats rather than distinct minutes. Two heartbeats in the same minute count as two, but the window only has one slot. Uptime is capped at 100% to avoid impossible values, but the calculation is not a true "minutes with at least one heartbeat" count. A stricter implementation would store a `map[int64]struct{}` keyed by `timestamp.Unix() / 60` one entry per unique minute bucket. This was intentionally kept simple for the challenge, but would be the first thing I'd change in a production system.

4. **No authentication**: Any client that can reach the server can send data for any device.

5. **Upload stat `sent_at` not validated**: The OpenAPI spec marks `sent_at` as required in the upload stat request, but the device simulator does not include it or sends it at `0`. Validating its presence would reject all simulator requests. In a production system with real devices, `sent_at` should be validated and used for analysis of upload performance over time.

6. **No metrics or tracing**: The service includes basic request logging, but it does not expose Prometheus metrics or distributed traces. In production, I would add metrics for request counts, latency, heartbeat ingestion, upload stat ingestion, and error rates.

## API Contract Notes

The provided OpenAPI spec does not document `400 Bad Request` responses for the POST endpoints. In practice, the handlers return `400` for malformed JSON or missing required fields. The `openapi.yml` in this repo has been updated to reflect this.

The spec marks `sent_at` as required in the upload stat request body, but the device simulator sends `0` for this field. Validating `sent_at` as non zero would reject all simulator requests, so this validation was intentionally omitted for upload stats.

## Performance & Production Safety

- All request paths are O(1). no request loops or scans.
- Device `sync.RWMutex` allows concurrent reads across devices without blocking.
- HTTP server timeouts are configured to prevent slow client attacks.
- Graceful shutdown gives in flight requests time to complete.
- The service keeps only aggregate device state in memory and does not depend on external infrastructure such as a database or message broker.

## Extensibility

- The `DataAccess` interface decouples handlers from storage. Swapping the in-memory store for a database requires no handler changes.
- Each handler is in its own file, making it easy for multiple engineers to work without conflicts.
- New metrics follow the same pattern: add fields to `DeviceState`, add a method to `DataAccess`, add a handler.
- Unit tests cover all store edge cases and can be run with `go test -race ./...`.
- The storage implementation is isolated behind an interface, allowing unit tests to mock the persistence layer independently from HTTP handlers.

## AI Usage

I used AI during the challenge mainly as a sparring partner and code reviewer. I used it to challenge design decisions, identify edge cases, and improve the README. I wrote and reviewed the implementation myself rather than relying on AI to generate the solution end-to-end.
