#!/usr/bin/env bash
set -euo pipefail

REGISTRY_ID="${REGISTRY_ID:?REGISTRY_ID is required}"
IMAGE_TAG="${IMAGE_TAG:-latest}"

core_image="cr.yandex/${REGISTRY_ID}/linka-core-api:${IMAGE_TAG}"
realtime_image="cr.yandex/${REGISTRY_ID}/linka-realtime:${IMAGE_TAG}"
sync_image="cr.yandex/${REGISTRY_ID}/linka-sync-worker:${IMAGE_TAG}"
dialog_image="cr.yandex/${REGISTRY_ID}/linka-dialog-worker:${IMAGE_TAG}"

docker buildx build --platform linux/amd64 -f Dockerfile.core-api -t "${core_image}" . --push
docker buildx build --platform linux/amd64 -f Dockerfile.realtime -t "${realtime_image}" . --push
docker buildx build --platform linux/amd64 -f Dockerfile.sync-worker -t "${sync_image}" . --push
docker buildx build --platform linux/amd64 -f Dockerfile.dialog-worker -t "${dialog_image}" . --push

echo "Pushed:"
echo "  ${core_image}"
echo "  ${realtime_image}"
echo "  ${sync_image}"
echo "  ${dialog_image}"
