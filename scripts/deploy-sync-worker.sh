#!/usr/bin/env bash
set -euo pipefail

FOLDER_ID="${FOLDER_ID:?FOLDER_ID is required}"
SERVICE_ACCOUNT_ID="${SERVICE_ACCOUNT_ID:?SERVICE_ACCOUNT_ID is required}"
REGISTRY_ID="${REGISTRY_ID:?REGISTRY_ID is required}"
YDB_ENDPOINT="${YDB_ENDPOINT:?YDB_ENDPOINT is required}"
YDB_DATABASE="${YDB_DATABASE:?YDB_DATABASE is required}"
FIREBASE_PROJECT_ID="${FIREBASE_PROJECT_ID:?FIREBASE_PROJECT_ID is required}"
FIREBASE_DATABASE_URL="${FIREBASE_DATABASE_URL:?FIREBASE_DATABASE_URL is required}"

IMAGE_TAG="${IMAGE_TAG:-latest}"
CONTAINER_NAME="${CONTAINER_NAME:-linka-sync-worker}"

image="cr.yandex/${REGISTRY_ID}/linka-sync-worker:${IMAGE_TAG}"

yc serverless container create --name "${CONTAINER_NAME}" --folder-id "${FOLDER_ID}" || true

envs="YDB_ENDPOINT=${YDB_ENDPOINT},YDB_DATABASE=${YDB_DATABASE},FIREBASE_PROJECT_ID=${FIREBASE_PROJECT_ID},FIREBASE_DATABASE_URL=${FIREBASE_DATABASE_URL}"

optional_vars=(
  FIREBASE_CREDENTIALS_JSON
  FIREBASE_CREDENTIALS_FILE
  YDB_TOKEN
  SYNC_POLL_INTERVAL
  SYNC_STREAM_ENABLED
  SYNC_STREAM_PATH
  SYNC_STREAM_RECONNECT
  ENV
  LOG_LEVEL
)

for var in "${optional_vars[@]}"; do
  value="${!var:-}"
  if [[ -n "${value}" ]]; then
    envs+="${envs:+,}${var}=${value}"
  fi
done

yc serverless container revision deploy \
  --container-name "${CONTAINER_NAME}" \
  --folder-id "${FOLDER_ID}" \
  --image "${image}" \
  --cores 1 \
  --memory 512MB \
  --core-fraction 100 \
  --execution-timeout 600s \
  --concurrency 1 \
  --service-account-id "${SERVICE_ACCOUNT_ID}" \
  --metadata-options gce-http-endpoint=enabled \
  --environment "${envs}"

echo "sync-worker deployed."

yc serverless container get --name "${CONTAINER_NAME}" --folder-id "${FOLDER_ID}" --format json
