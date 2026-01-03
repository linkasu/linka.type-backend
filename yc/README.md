# Yandex Cloud deployment (CLI outline)

This is a minimal outline for deploying the three services to Yandex Cloud Serverless Containers. Adjust names, service accounts, and networks as needed.

## Prerequisites
- `yc` CLI configured (`yc init`).
- Container Registry created.
- Service account with roles: `serverless.containers.editor`, `container-registry.images.pusher`, `logging.writer`, `lockbox.payloadViewer` (as needed).

## Build and push images
```bash
REGISTRY_ID=your-registry-id

# core-api
DOCKER_BUILDKIT=1 docker build -f Dockerfile.core-api -t cr.yandex/${REGISTRY_ID}/linka-core-api:latest .
docker push cr.yandex/${REGISTRY_ID}/linka-core-api:latest

# realtime
DOCKER_BUILDKIT=1 docker build -f Dockerfile.realtime -t cr.yandex/${REGISTRY_ID}/linka-realtime:latest .
docker push cr.yandex/${REGISTRY_ID}/linka-realtime:latest

# sync-worker
DOCKER_BUILDKIT=1 docker build -f Dockerfile.sync-worker -t cr.yandex/${REGISTRY_ID}/linka-sync-worker:latest .
docker push cr.yandex/${REGISTRY_ID}/linka-sync-worker:latest
```

## Create containers (example)
```bash
SERVICE_ACCOUNT_ID=your-service-account-id

# core-api
yc serverless container create --name linka-core-api

yc serverless container revision deploy \
  --container-name linka-core-api \
  --image cr.yandex/${REGISTRY_ID}/linka-core-api:latest \
  --service-account-id ${SERVICE_ACCOUNT_ID} \
  --memory 512M \
  --metadata-options gce-http-endpoint=enabled \
  --environment PORT=8080,FIREBASE_PROJECT_ID=...,FIREBASE_DATABASE_URL=...,YDB_ENDPOINT=...,YDB_DATABASE=...

# realtime
yc serverless container create --name linka-realtime

yc serverless container revision deploy \
  --container-name linka-realtime \
  --image cr.yandex/${REGISTRY_ID}/linka-realtime:latest \
  --service-account-id ${SERVICE_ACCOUNT_ID} \
  --memory 512M \
  --metadata-options gce-http-endpoint=enabled \
  --environment PORT=8080,FIREBASE_PROJECT_ID=...,YDB_ENDPOINT=...,YDB_DATABASE=...

# sync-worker
yc serverless container create --name linka-sync-worker

yc serverless container revision deploy \
  --container-name linka-sync-worker \
  --image cr.yandex/${REGISTRY_ID}/linka-sync-worker:latest \
  --service-account-id ${SERVICE_ACCOUNT_ID} \
  --memory 512M \
  --metadata-options gce-http-endpoint=enabled \
  --environment FIREBASE_PROJECT_ID=...,FIREBASE_DATABASE_URL=...,YDB_ENDPOINT=...,YDB_DATABASE=...,SYNC_POLL_INTERVAL=5s
```

## Routing
- Use API Gateway or Application Load Balancer to expose `/v1` routes.
- Enable WebSocket routing for `/v1/stream`.

## Secrets
- Store Firebase Admin key in Lockbox and inject via `FIREBASE_CREDENTIALS_JSON` or `FIREBASE_CREDENTIALS_FILE`.
- Optionally store YDB token in Lockbox (`YDB_TOKEN`) if metadata is not available.
