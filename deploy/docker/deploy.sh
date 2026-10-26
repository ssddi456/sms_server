#!/usr/bin/env bash

set -euo pipefail

IMAGE_REF=""
HEALTH_URL="http://127.0.0.1:8322/sms/healthz"
HEALTH_TIMEOUT=60
COMPOSE_FILE="deploy/docker/docker-compose.prod.yml"
STATE_DIR=".deploy_state"
IMAGE_TAR="sms-server-image.tar"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --image)
      IMAGE_REF="${2:-}"
      shift 2
      ;;
    --health-url)
      HEALTH_URL="${2:-}"
      shift 2
      ;;
    --health-timeout)
      HEALTH_TIMEOUT="${2:-}"
      shift 2
      ;;
    --compose-file)
      COMPOSE_FILE="${2:-}"
      shift 2
      ;;
    --image-tar)
      IMAGE_TAR="${2:-}"
      shift 2
      ;;
    *)
      echo "Unknown argument: $1" >&2
      exit 1
      ;;
  esac
done

if ! command -v docker >/dev/null 2>&1; then
  echo "docker command not found" >&2
  exit 1
fi

if ! docker compose version >/dev/null 2>&1; then
  echo "docker compose command not found" >&2
  exit 1
fi

if ! command -v curl >/dev/null 2>&1; then
  echo "curl command not found" >&2
  exit 1
fi

if [[ ! -f "$COMPOSE_FILE" ]]; then
  echo "Compose file not found: $COMPOSE_FILE" >&2
  exit 1
fi

if [[ ! -f "$IMAGE_TAR" ]]; then
  echo "Docker image tar not found: $IMAGE_TAR" >&2
  exit 1
fi

if [[ -z "$IMAGE_REF" ]] && [[ -f "manifest.txt" ]]; then
  IMAGE_REF="$(grep -E '^image=' manifest.txt | head -n1 | cut -d'=' -f2- || true)"
fi

if [[ -z "$IMAGE_REF" ]]; then
  echo "Image reference is required. Use --image or include image= in manifest.txt." >&2
  exit 1
fi

if [[ ! -f ".env" ]] && [[ -f ".env.example" ]]; then
  cp ".env.example" ".env"
  echo "Created .env from .env.example"
fi

mkdir -p data "$STATE_DIR"

PRE_DEPLOY_IMAGE_FILE="$STATE_DIR/pre_deploy_image.txt"
LAST_SUCCESS_IMAGE_FILE="$STATE_DIR/last_success_image.txt"

previous_image_id="$(docker inspect --format='{{.Image}}' sms-server 2>/dev/null || true)"
if [[ -n "$previous_image_id" ]]; then
  printf '%s\n' "$previous_image_id" > "$PRE_DEPLOY_IMAGE_FILE"
  echo "Captured previous container image: $previous_image_id"
fi

echo "Loading image tar: $IMAGE_TAR"
docker load -i "$IMAGE_TAR" >/dev/null

export SMS_SERVER_IMAGE="$IMAGE_REF"

echo "Deploying image: $SMS_SERVER_IMAGE"
docker compose -f "$COMPOSE_FILE" up -d --remove-orphans --force-recreate

echo "Waiting for health check: $HEALTH_URL (timeout ${HEALTH_TIMEOUT}s)"
start_ts="$(date +%s)"
healthy=0
while true; do
  if curl -fsS --max-time 3 "$HEALTH_URL" >/dev/null 2>&1; then
    healthy=1
    break
  fi

  now_ts="$(date +%s)"
  elapsed=$((now_ts - start_ts))
  if (( elapsed >= HEALTH_TIMEOUT )); then
    break
  fi
  sleep 2
done

if (( healthy == 1 )); then
  active_image_id="$(docker inspect --format='{{.Image}}' sms-server 2>/dev/null || true)"
  if [[ -n "$active_image_id" ]]; then
    printf '%s\n' "$active_image_id" > "$LAST_SUCCESS_IMAGE_FILE"
  fi
  echo "Deployment successful."
  echo "Active image id: ${active_image_id:-unknown}"
  exit 0
fi

echo "Health check failed after ${HEALTH_TIMEOUT}s."

if [[ -f "$PRE_DEPLOY_IMAGE_FILE" ]]; then
  rollback_image_id="$(cat "$PRE_DEPLOY_IMAGE_FILE")"
  if [[ -n "$rollback_image_id" ]]; then
    echo "Rolling back to previous image id: $rollback_image_id"
    export SMS_SERVER_IMAGE="$rollback_image_id"
    docker compose -f "$COMPOSE_FILE" up -d --remove-orphans --force-recreate
    exit 1
  fi
fi

echo "Rollback not available; previous image id was not captured."
exit 1
