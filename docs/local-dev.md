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
- `FIREBASE_API_KEY` - required for `POST /v1/auth`
- `YDB_ENDPOINT`
- `YDB_DATABASE`
- `YDB_TOKEN`
- `YDB_METADATA_URL` - override Yandex Cloud metadata token URL
- `YDB_METADATA_DISABLED` - set to disable metadata token fallback
- `FEATURE_READ_SOURCE` - `firebase_only`, `ydb_primary`, or `cohort`
- `FEATURE_COHORT_PERCENT` - 0-100 for `cohort` mode
- `TTS_PROXY_ENABLED` - enable `/v1/tts` and `/v1/voices`
- `TTS_BASE_URL` - defaults to `https://tts.linka.su`
- `DIALOG_HELPER_URL` - dialog-helper API base URL
- `DIALOG_HELPER_API_KEY` - API key for dialog-helper
- `DIALOG_HELPER_TIMEOUT` - dialog-helper request timeout (default `20s`)
- `DIALOG_HELPER_MAX_AUDIO_BYTES` - max audio upload size (default `8MB`)
- `DIALOG_WORKER_INTERVAL` - dialog suggestion worker interval (default `15s`)
- `SYNC_POLL_INTERVAL` - sync-worker interval (default `5s`)
- `SYNC_STREAM_ENABLED` - enable RTDB streaming (default `false`)
- `SYNC_STREAM_PATH` - RTDB path for streaming (default `users`)
- `SYNC_STREAM_RECONNECT` - reconnect delay (default `5s`)
- `FIREBASE_ACCESS_TOKEN` - optional OAuth token for RTDB streaming

## Running (placeholder)
- `go run ./cmd/core-api`
- `go run ./cmd/realtime`
- `go run ./cmd/sync-worker`
