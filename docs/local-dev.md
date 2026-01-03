# Local Development

## Requirements
- Go 1.22+
- Firebase Admin key JSON (for token verification and RTDB access)
- YDB endpoint and credentials (dev database)

## Environment
- `PORT` - HTTP port (default 8080)
- `HTTP_ADDR` - override listen address (default `:8080`)
- `FIREBASE_PROJECT_ID`
- `FIREBASE_DATABASE_URL`
- `FIREBASE_CREDENTIALS_JSON` or `FIREBASE_CREDENTIALS_FILE`
- `YDB_ENDPOINT`
- `YDB_DATABASE`
- `YDB_TOKEN`
- `FEATURE_READ_SOURCE` - `firebase_only`, `ydb_primary`, or `cohort`
- `FEATURE_COHORT_PERCENT` - 0-100 for `cohort` mode
- `TTS_PROXY_ENABLED` - enable `/v1/tts` and `/v1/voices`
- `TTS_BASE_URL` - defaults to `https://tts.linka.su`
- `SYNC_POLL_INTERVAL` - sync-worker interval (default `5s`)

## Running (placeholder)
- `go run ./cmd/core-api`
- `go run ./cmd/realtime`
- `go run ./cmd/sync-worker`
