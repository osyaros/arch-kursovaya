#!/usr/bin/env bash
# SCAN по KeyDB/Redis (совместимо с bash 3.2 на macOS).
# ./scripts/redis-scan-dump.sh | tee scan-output.txt

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=_redis_cli.sh
source "$SCRIPT_DIR/_redis_cli.sh"
redis_cli_setup

CURSOR=0
while true; do
  lines=()
  while IFS= read -r line; do
    lines+=("$line")
  done < <("${REDIS_CLI[@]}" --raw SCAN "$CURSOR")

  CURSOR="${lines[0]:-0}"
  if [[ "${#lines[@]}" -gt 1 ]]; then
    keys=("${lines[@]:1}")
  else
    keys=()
  fi

  for key in "${keys[@]:-}"; do
    [[ -z "${key:-}" ]] && continue
    typ=$("${REDIS_CLI[@]}" TYPE "$key")
    echo "Key: $key, Type: $typ"
    case "$typ" in
      string) "${REDIS_CLI[@]}" GET "$key" ;;
      hash) "${REDIS_CLI[@]}" HGETALL "$key" ;;
      list) "${REDIS_CLI[@]}" LRANGE "$key" 0 -1 ;;
      set) "${REDIS_CLI[@]}" SMEMBERS "$key" ;;
      zset) "${REDIS_CLI[@]}" ZRANGE "$key" 0 -1 WITHSCORES ;;
    esac
  done

  [[ "$CURSOR" == "0" ]] && break
done
