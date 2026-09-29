# Yandex Cloud Deployment

## Resources
- Serverless Containers: `core-api`, `realtime`, `sync-worker`.
- Serverless YDB database.
- API Gateway or Application Load Balancer for HTTP + WebSocket routing.
- Lockbox secrets for Firebase Admin SDK and optional TTS credentials.
- Cloud Logging + metrics.

## Networking
- Public endpoints for HTTP + WebSocket.
- Outbound access for Firebase RTDB and TTS proxy (if enabled).

## Configuration
- Environment variables are injected per service (YDB endpoint, Firebase key, feature flag settings).
- Use Lockbox to mount Firebase Admin credentials as a file or env var.
- For YDB auth in Serverless Containers, enable metadata access and omit `YDB_TOKEN` so the service can fetch IAM tokens.

## Deployment layout
- `Dockerfile.core-api`, `Dockerfile.realtime`, `Dockerfile.sync-worker` build the service containers.
- `yc/` will contain deployment configs (TBD: Terraform or yc-serverless spec).

## Optional TTS Service JWT

Leave `TTS_SERVICE_JWT_PRIVATE_KEY_BASE64`, `TTS_SERVICE_JWT_KEY_ID`, and `TTS_SERVICE_JWT_TTL` unset to preserve the existing `X-TTS-Service-Token` behavior. Configure the private key as a deployment secret and the key ID and optional TTL (at most `5m`) as deployment variables. Both key settings are required together; invalid values make the new revision fail at startup.

# TTS Control Plane Foundation

The control plane is disabled by default. Leave `TTS_CONTROL_PLANE_ENABLED=false` until PostgreSQL has been provisioned, `go run ./cmd/tts-migrate` has completed manually, and Redis is reachable. The core API validates PostgreSQL, Redis, and the embedded migration ledger only when the flag is true; it never migrates automatically and currently registers no new HTTP routes.

Rolling invariant: old revisions ignore all TTS control-plane variables; new revisions with the flag off neither require nor connect to PostgreSQL or Redis.

Configure deployment variables: `TTS_CONTROL_PLANE_ENABLED` (keep `false`), `TTS_CONTROL_REDIS_USERNAME`, `TTS_CONTROL_REDIS_DB`, daily/max chunk limits, `TTS_CONTROL_TRUSTED_PROXY_HOPS` (default `1`), `TTS_CONTROL_CHUNK_CHARS` (default `240`), and optional anonymous global budget placeholders. Configure secrets: `TTS_CONTROL_POSTGRES_DSN`, `TTS_CONTROL_REDIS_ADDR`, `TTS_CONTROL_REDIS_PASSWORD`, `TTS_CONTROL_TOKEN_SIGNING_KEY`, optional `TTS_CONTROL_TOKEN_PREVIOUS_SIGNING_KEY`, and `TTS_CONTROL_IP_HASH_KEY`.

The gateway must overwrite or append `X-Forwarded-For` before forwarding to core-api. The API selects the configured rightmost trusted entry and rejects malformed selected addresses; it never trusts the leftmost client-supplied value by default.
