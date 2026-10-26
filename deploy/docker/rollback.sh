#!/usr/bin/env bash

set -euo pipefail

IMAGE_REF="${SMS_SERVER_IMAGE:-}"
HEALTH_URL="http://127.0.0.1:8322/sms/healthz"
HEALTH_TIMEOUT=60
COMPOSE_FILE="deploy/docker/docker-compose.prod.yml"
STATE_DIR=".deploy_state"

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

if [[ -z "$IMAGE_REF" ]] && [[ -f "$STATE_DIR/last_success_image.txt" ]]; then
  IMAGE_REF="$(cat "$STATE_DIR/last_success_image.txt")"
fi

if [[ -z "$IMAGE_REF" ]]; then
  echo "Rollback image is required. Use --image or ensure $STATE_DIR/last_success_image.txt exists." >&2
  exit 1
fi

export SMS_SERVER_IMAGE="$IMAGE_REF"
echo "Rolling back to image: $SMS_SERVER_IMAGE"
docker compose -f "$COMPOSE_FILE" up -d --remove-orphans --force-recreate

echo "Waiting for health check: $HEALTH_URL (timeout ${HEALTH_TIMEOUT}s)"
start_ts="$(date +%s)"
while true; do
  if curl -fsS --max-time 3 "$HEALTH_URL" >/dev/null 2>&1; then
    echo "Rollback successful."
    exit 0
  fi

  now_ts="$(date +%s)"
  elapsed=$((now_ts - start_ts))
  if (( elapsed >= HEALTH_TIMEOUT )); then
    break
  fi
  sleep 2
done

echo "Rollback completed but health check still failing."
exit 1
