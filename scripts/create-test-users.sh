#!/usr/bin/env bash
# Создаёт 12 пользователей для проверки алерта HighUserCreationRate.

set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
COUNT="${COUNT:-12}"

for i in $(seq 1 "$COUNT"); do
  code=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/api/users" \
    -H 'Content-Type: application/json' \
    -d "{\"name\":\"NewUser_$i\",\"email\":\"newuser_$i@example.com\"}")
  echo "POST /api/users #$i -> HTTP $code"
  if [ "$code" != "201" ]; then
    echo "Ошибка: ожидался 201, получен $code" >&2
    exit 1
  fi
done

echo "Готово: создано $COUNT пользователей"
