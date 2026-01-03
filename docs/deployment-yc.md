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

## Deployment layout
- `Dockerfile.core-api`, `Dockerfile.realtime`, `Dockerfile.sync-worker` build the service containers.
- `yc/` will contain deployment configs (TBD: Terraform or yc-serverless spec).
