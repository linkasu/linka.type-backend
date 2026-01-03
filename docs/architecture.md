# Architecture

## Goals
- Run on Yandex Cloud serverless containers with serverless YDB.
- Preserve Firebase client behavior and IDs while enabling a gradual migration.
- Provide realtime updates via long polling and WebSocket.

## Components
### core-api
- REST API for CRUD, onboarding, global import, admin checks, and account deletion.
- Dual-writes to YDB and Firebase RTDB.
- Writes change events to the `changes` table for realtime consumption.

### realtime
- Long polling endpoint (`/v1/changes`).
- WebSocket streaming (`/v1/stream`).
- Reads from the `changes` table and streams per-user events.

### sync-worker
- Consumes Firebase RTDB changes and applies them to YDB.
- Backfills missing YDB data on demand (read-through seeding).
- Keeps `admins`, `global` data, and `factory/questions` in sync.

## Data flow
1. Client calls API with a bearer token (default from `/v1/auth`).
2. core-api verifies token and resolves user_id.
3. core-api reads from YDB; if missing, it seeds from Firebase.
4. Mutations are written to YDB, then mirrored to Firebase.
5. A `changes` row is written for realtime delivery.

## Auth and authorization
- All endpoints except `/v1/auth` require a bearer token.
- Users can only access their own data.
- Global and factory write operations require admin (via `admins` table).

## Observability
- Structured logs with request_id + user_id.
- Yandex Cloud Logging for aggregation.
- Metrics: request rate, latency, error rate, change lag, sync lag.
