# Local Development

## Requirements
- Go 1.22+
- Firebase Admin key JSON (for token verification and RTDB access)
- YDB endpoint and credentials (dev database)

## Environment
- `PORT` - HTTP port (default 8080)
- `FIREBASE_PROJECT_ID`
- `FIREBASE_CREDENTIALS_JSON` or `FIREBASE_CREDENTIALS_FILE`
- `YDB_ENDPOINT`
- `YDB_DATABASE`
- `YDB_TOKEN`

## Running (placeholder)
- `go run ./cmd/core-api`
- `go run ./cmd/realtime`
- `go run ./cmd/sync-worker`
