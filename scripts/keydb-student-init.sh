#!/usr/bin/env bash
# Запуск: ./scripts/keydb-student-init.sh  (нужен make up)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=_redis_cli.sh
source "$SCRIPT_DIR/_redis_cli.sh"
redis_cli_setup

STUDENT_GROUP="${STUDENT_GROUP:-BIVT-23-SP-2}"
STUDENT_NO="${STUDENT_NO:-17}"
FULL_NAME="${STUDENT_NAME:-Осокин Ярослав Юрьевич}"
AGE="${STUDENT_AGE:-20}"
EMAIL="${STUDENT_EMAIL:-m2308752@edu.misis.ru}"
BASE="student:${STUDENT_GROUP}:${STUDENT_NO}"

"${REDIS_CLI[@]}" SET "$BASE" "$FULL_NAME"
"${REDIS_CLI[@]}" HSET "${BASE}:info" name "$FULL_NAME" age "$AGE" email "$EMAIL"
"${REDIS_CLI[@]}" DEL "${BASE}:timetable" 2>/dev/null || true
"${REDIS_CLI[@]}" RPUSH "${BASE}:timetable" "Математический анализ" "Архитектура ПО" "Базы данных"
"${REDIS_CLI[@]}" DEL "${BASE}:skills" 2>/dev/null || true
"${REDIS_CLI[@]}" SADD "${BASE}:skills" "Go" "Docker" "PostgreSQL" "Redis"
"${REDIS_CLI[@]}" DEL "${BASE}:tasks_w_priority" 2>/dev/null || true
"${REDIS_CLI[@]}" ZADD "${BASE}:tasks_w_priority" 100 "Сделать лабу 1" 150 "Сделать лабу 2" 80 "Повторить лекцию"

echo "Готово. Ключи с префиксом: $BASE"
