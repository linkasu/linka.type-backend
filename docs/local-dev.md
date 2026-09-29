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
- `TTS_SERVICE_TOKEN` - optional token sent upstream as `X-TTS-Service-Token`
- `TTS_SERVICE_JWT_PRIVATE_KEY_BASE64` and `TTS_SERVICE_JWT_KEY_ID` - optional Ed25519 service JWT signing pair; both must be set. The private key is the base64-encoded raw 64-byte key and is sent upstream only as a signature.
- `TTS_SERVICE_JWT_TTL` - optional JWT lifetime, defaults to and cannot exceed `5m`.
- `TTS_MAX_AUDIO_BYTES` - max proxied TTS response size (default `50MiB`)
- `TTS_TIMEOUT` - TTS upstream request timeout (default `120s`)
- `TTS_CONTROL_PLANE_ENABLED` - enables the isolated TTS control-plane runtime (default `false`); it does not expose routes.
- `TTS_CONTROL_POSTGRES_DSN` - PostgreSQL DSN, required only when the control plane is enabled.
- `TTS_CONTROL_REDIS_ADDR`, `TTS_CONTROL_REDIS_USERNAME`, `TTS_CONTROL_REDIS_PASSWORD`, `TTS_CONTROL_REDIS_DB` - Redis hot quota counters, required only when enabled.
- `TTS_CONTROL_TOKEN_SIGNING_KEY`, `TTS_CONTROL_TOKEN_PREVIOUS_SIGNING_KEY`, `TTS_CONTROL_IP_HASH_KEY` - secrets; signing and IP hash keys must be at least 32 bytes when enabled.
- `TTS_CONTROL_ANONYMOUS_DAILY_CHUNKS` / `TTS_CONTROL_AUTH_DAILY_CHUNKS` - defaults `30` / `200`.
- `TTS_CONTROL_ANONYMOUS_MAX_CHUNKS` / `TTS_CONTROL_AUTH_MAX_CHUNKS` - defaults `5` / `21`.
- `TTS_CONTROL_ANONYMOUS_GLOBAL_DAILY_BUDGET` / `TTS_CONTROL_ANONYMOUS_GLOBAL_MONTHLY_BUDGET` - reserved global-budget placeholders, default `0`.
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
- `TTS_CONTROL_POSTGRES_DSN=... go run ./cmd/tts-migrate` applies embedded TTS-only migrations manually. It is never run by application startup or deployment.
