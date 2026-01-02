# Realtime

## Changes table
- Every mutation writes a row in `changes` keyed by (`user_id`, `cursor`).
- `cursor` is an opaque, monotonically increasing value (ULID or counter).
- `payload` includes minimal entity data needed for clients to update local state.

## Long polling
- `GET /v1/changes?cursor=...&timeout=25s&limit=100`
- If no new changes, the server waits up to `timeout`.
- Response includes the newest cursor and change list.

## WebSocket
- `WS /v1/stream?cursor=...`
- On connect, the server streams backlog then pushes new changes.
- Heartbeats are sent every 25s when idle.

## Ordering guarantees
- Per-user order follows `cursor`.
- Multiple entities may appear in a single batch.
