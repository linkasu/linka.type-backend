#!/usr/bin/env bash
set -euo pipefail

FOLDER_ID="${FOLDER_ID:?FOLDER_ID is required}"
SERVICE_ACCOUNT_ID="${SERVICE_ACCOUNT_ID:?SERVICE_ACCOUNT_ID is required}"
REGISTRY_ID="${REGISTRY_ID:?REGISTRY_ID is required}"
YDB_ENDPOINT="${YDB_ENDPOINT:?YDB_ENDPOINT is required}"
YDB_DATABASE="${YDB_DATABASE:?YDB_DATABASE is required}"
DIALOG_HELPER_URL="${DIALOG_HELPER_URL:?DIALOG_HELPER_URL is required}"
DIALOG_HELPER_API_KEY="${DIALOG_HELPER_API_KEY:?DIALOG_HELPER_API_KEY is required}"

IMAGE_TAG="${IMAGE_TAG:-latest}"
CONTAINER_NAME="${CONTAINER_NAME:-linka-dialog-worker}"

image="cr.yandex/${REGISTRY_ID}/linka-dialog-worker:${IMAGE_TAG}"

yc serverless container create --name "${CONTAINER_NAME}" --folder-id "${FOLDER_ID}" || true

envs="YDB_ENDPOINT=${YDB_ENDPOINT},YDB_DATABASE=${YDB_DATABASE},DIALOG_HELPER_URL=${DIALOG_HELPER_URL},DIALOG_HELPER_API_KEY=${DIALOG_HELPER_API_KEY}"

optional_vars=(
  YDB_TOKEN
  DIALOG_HELPER_TIMEOUT
  DIALOG_HELPER_MAX_AUDIO_BYTES
  DIALOG_WORKER_INTERVAL
  YC_FOLDER_ID
  YC_GPT_MODEL_URI
  ENV
  LOG_LEVEL
)

# Pass FOLDER_ID as YC_FOLDER_ID for GPT client if not explicitly set
if [[ -z "${YC_FOLDER_ID:-}" ]]; then
  export YC_FOLDER_ID="${FOLDER_ID}"
fi

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

echo "dialog-worker deployed."

yc serverless container get --name "${CONTAINER_NAME}" --folder-id "${FOLDER_ID}" --format json
