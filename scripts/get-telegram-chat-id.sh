#!/usr/bin/env bash
# Узнать chat_id после того, как вы написали боту /start в Telegram.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
# shellcheck disable=SC1091
source "$ROOT_DIR/.env"

if [ -z "${TELEGRAM_BOT_TOKEN:-}" ]; then
  echo "Заполните TELEGRAM_BOT_TOKEN в .env" >&2
  exit 1
fi

echo "Напишите боту /start в Telegram (бот ответит chat_id), затем при необходимости обновите .env:"

response=$(curl -s "https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/getUpdates")
chat_id=$(echo "$response" | python3 -c "
import json, sys
data = json.load(sys.stdin)
results = data.get('result', [])
if not results:
    print('', end='')
    sys.exit(0)
last = results[-1]
chat = last.get('message', {}).get('chat') or last.get('my_chat_member', {}).get('chat', {})
print(chat.get('id', ''), end='')
")

if [ -z "$chat_id" ]; then
  echo "chat_id не найден. Убедитесь, что вы написали боту /start." >&2
  echo "$response"
  exit 1
fi

if grep -q '^TELEGRAM_CHAT_ID=' "$ROOT_DIR/.env"; then
  sed -i '' "s/^TELEGRAM_CHAT_ID=.*/TELEGRAM_CHAT_ID=${chat_id}/" "$ROOT_DIR/.env"
else
  echo "TELEGRAM_CHAT_ID=${chat_id}" >> "$ROOT_DIR/.env"
fi

echo "TELEGRAM_CHAT_ID=${chat_id} записан в .env"
echo "Перезапустите alertmanager: docker compose up -d alertmanager"
