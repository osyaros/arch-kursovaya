#!/usr/bin/env bash
# Общая настройка redis-cli: локально или через docker exec (macOS без brew install redis).

redis_cli_setup() {
  REDIS_HOST="${REDIS_HOST:-localhost}"
  REDIS_PORT="${REDIS_PORT:-6379}"
  KEYDB_CONTAINER="${KEYDB_CONTAINER:-keydb-archapp-lab4}"

  if command -v redis-cli >/dev/null 2>&1; then
    REDIS_CLI=(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT")
    return 0
  fi

  if command -v docker >/dev/null 2>&1; then
    if ! docker ps --format '{{.Names}}' 2>/dev/null | grep -qx "$KEYDB_CONTAINER"; then
      detected=$(docker ps --format '{{.Names}}' 2>/dev/null | grep -E '^keydb-archapp-lab' | head -1)
      if [[ -n "$detected" ]]; then
        KEYDB_CONTAINER="$detected"
      fi
    fi
    if docker ps --format '{{.Names}}' 2>/dev/null | grep -qx "$KEYDB_CONTAINER"; then
      REDIS_CLI=(docker exec "$KEYDB_CONTAINER" redis-cli)
      return 0
    fi
  fi

  echo "redis-cli не найден." >&2
  echo "  brew install redis   # или" >&2
  echo "  make up-lab4         # KeyDB в Docker ($KEYDB_CONTAINER)" >&2
  return 1
}
