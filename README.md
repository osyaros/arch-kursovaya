# Курсовая работа — сервис управления пользователями

Распределённое приложение на Go, объединяющее результаты лабораторных работ 1–4 по дисциплине «Архитектура программного обеспечения»:

| Лаба | Тема | Что реализовано |
|------|------|-----------------|
| 1 | Docker | Контейнеризация всех сервисов через Docker Compose, multi-stage Dockerfile |
| 2 | Redis/KeyDB | Кеширование пользователей (Cache-Aside), структуры данных Redis |
| 3 | RabbitMQ | Event-driven: `user-service` → exchange → `notification-service` |
| 4 | Observability | Prometheus + Grafana + Alertmanager + Telegram-бот |

**Стек:** Go 1.23, Gin, PostgreSQL 16, KeyDB 6.3, RabbitMQ 3, Prometheus 2.44, Grafana, Alertmanager 0.24, Docker Compose.

Текст курсовой: [`kursovaya.md`](kursovaya.md) · DOCX: `kursovaya.docx` (генерация: `python3 scripts/generate_kursovaya_docx.py`).

---

## Архитектура

```
Клиент (Postman/curl)
        │
        ▼
  user-service ──────► PostgreSQL (хранение)
        │                    ▲
        ├──────► KeyDB (кеш Cache-Aside)
        │
        └──────► RabbitMQ (exchange user.exchange)
                        │
                        ▼
              notification-service (consumer, лог уведомлений)

  user-service ──► /metrics ◄── Prometheus ──► Grafana
                        │
                        └── Alertmanager ──► telegram-bot ──► Telegram
```

**Поток создания пользователя:** `POST /api/users` → запись в PostgreSQL → публикация `UserCreated` в RabbitMQ → consumer логирует уведомление → инкремент метрики `users_new_total`.

---

## Структура проекта

```
arch-oyu-lab4/
├── cmd/
│   ├── server/              # user-service (REST API)
│   ├── notification/        # notification-service (consumer)
│   └── telegram-bot/        # webhook Alertmanager → Telegram
├── internal/
│   ├── cache/               # Cache-Aside (go-redis)
│   ├── handler/             # HTTP handlers (Gin)
│   ├── metrics/             # Prometheus counter + histogram
│   ├── mq/                  # RabbitMQ publish/consume
│   ├── repository/          # PostgreSQL (pgx)
│   ├── service/             # бизнес-логика
│   └── telegram/            # Telegram Bot API
├── dockercompose/           # prometheus, grafana, alertmanager
├── scripts/                 # вспомогательные скрипты
├── docker-compose.yaml
└── Makefile
```

---

## Требования

- Docker и Docker Compose
- Go 1.23+ (для локального запуска без Docker)
- `curl` (для тестов)

`redis-cli` на хосте **не обязателен** — скрипты лабы 2 работают через `docker exec` в контейнер KeyDB.

---

## Быстрый старт

### 1. Клонировать и настроить окружение

```bash
cp .env.example .env
```

Заполните `TELEGRAM_BOT_TOKEN` и `TELEGRAM_CHAT_ID` (нужны для алертов в лабе 4, без них стек тоже поднимется).

### 2. Запустить полный стек

```bash
make up-lab4
```

Подождите ~30 секунд, пока поднимутся все контейнеры.

### 3. Проверить API

```bash
curl http://127.0.0.1:8080/health
curl -X POST http://127.0.0.1:8080/api/users \
  -H 'Content-Type: application/json' \
  -d '{"name":"Иван Иванов","email":"ivan@example.com"}'
```

> Для API используйте `127.0.0.1`, а не `localhost` — порт `:8080` может быть занят другим процессом.

### 4. Остановить

```bash
make down-lab4
```

---

## Сервисы и URL

| Сервис | URL / порт | Учётные данные |
|--------|------------|----------------|
| user-service (API) | http://127.0.0.1:8080 | — |
| Метрики | http://127.0.0.1:8080/metrics | — |
| telegram-bot | http://127.0.0.1:8081/health | — |
| PostgreSQL | localhost:5433 | app / app, БД `app` |
| KeyDB | localhost:6379 | — |
| RabbitMQ Management | http://localhost:15672 | guest / guest |
| RabbitMQ Prometheus | localhost:15692 | — |
| Prometheus | http://localhost:9090 | — |
| Grafana | http://localhost:3000 | admin / admin |
| Alertmanager | http://localhost:9093 | — |

---

## REST API

| Метод | Путь | Описание |
|-------|------|----------|
| GET | `/health` | Проверка работоспособности |
| GET | `/metrics` | Метрики Prometheus |
| GET | `/api/users` | Список пользователей |
| POST | `/api/users` | Создание пользователя |
| GET | `/api/users/:id` | Получение по ID (с кешем) |
| PATCH | `/api/users/:id` | Частичное обновление |
| DELETE | `/api/users/:id` | Удаление |

**Тело создания:**
```json
{"name": "Иван Иванов", "email": "ivan@example.com"}
```

**Тело обновления:**
```json
{"name": "Новое имя"}
```

---

## Команды Makefile

| Команда | Описание |
|---------|----------|
| `make up-lab4` | Полный стек в Docker (сборка + запуск) |
| `make down-lab4` | Остановка стека |
| `make build-lab4` | Только сборка образов |
| `make up` | Только инфра: postgres, keydb, rabbitmq |
| `make down` | Остановка инфра |
| `make build` | Сборка Go-бинарников в `bin/` |
| `make run` | user-service локально (нужен `make up`) |
| `make run-notification` | notification-service локально |
| `make run-telegram-bot` | telegram-bot локально |
| `make init-student` | Заполнить KeyDB данными студента (лаба 2) |
| `make scan` | SCAN по ключам KeyDB (лаба 2) |
| `make create-test-users` | 12 POST-запросов для проверки алертов |
| `make telegram-chat-id` | Подсказка по получению chat_id |

---

## Лабораторная 1 — Docker

Все компоненты описаны в `docker-compose.yaml` и запускаются одной командой:

```bash
make up-lab4
```

Go-сервисы собираются multi-stage Dockerfile (`Dockerfile`, `Dockerfile.notification`, `Dockerfile.telegram-bot`).

---

## Лабораторная 2 — KeyDB / Redis

### Кеширование в приложении

- Стратегия **Cache-Aside** для `GET /api/users/:id`
- Ключ: `user:v1:{uuid}`, TTL 60 сек
- Инвалидация при `PATCH` и `DELETE`

Проверка: дважды запросить одного пользователя — в логах `user-service` первый раз `cache miss`, второй — `cache hit`:

```bash
docker logs user-service --tail 20
```

### Структуры данных Redis (скрипты)

```bash
make up-lab4          # KeyDB должен быть запущен
make init-student     # создать ключи student:BIVT-23-SP-2:17
make scan             # обход SCAN, вывод типов и значений
make scan | tee scan-output.txt   # сохранить для скриншота
```

Переменные для `init-student` (опционально):

```bash
STUDENT_GROUP=BIVT-23-SP-2 STUDENT_NO=17 \
STUDENT_NAME="ФИО" STUDENT_EMAIL=email@edu.misis.ru \
make init-student
```

---

## Лабораторная 3 — RabbitMQ

### Топология в приложении

| Сущность | Имя |
|----------|-----|
| Exchange | `user.exchange` (topic) |
| Queue | `user.events` |
| Binding | `user.#` |
| Routing key при публикации | `user.created` |

### Проверка event-driven

1. Создать пользователя:
   ```bash
   curl -X POST http://127.0.0.1:8080/api/users \
     -H 'Content-Type: application/json' \
     -d '{"name":"Test","email":"test_'$(date +%s)'@example.com"}'
   ```

2. Логи consumer:
   ```bash
   docker logs notification-service --tail 10
   ```
   Ожидается: `уведомление: создан новый пользователь`.

3. RabbitMQ UI: http://localhost:15672 (guest / guest)
   - **Queues** → `user.events` → Ready = 0 (сообщение уже обработано)
   - **Exchanges** → `user.exchange`

Чтобы увидеть сообщение в очереди: `docker stop notification-service`, создайте пользователя, сделайте скриншот (Ready = 1), затем `docker start notification-service`.

---

## Лабораторная 4 — Observability

### Кастомные метрики

| Метрика | Тип | Описание |
|---------|-----|----------|
| `users_new_total` | Counter | Созданные пользователи |
| `api_user_request_duration_seconds` | Histogram | Длительность GET/POST, лейбл `method=get\|post` |

Внешние метрики: RabbitMQ (порт 15692, prometheus plugin).

### Grafana — дашборды (папка Lab4)

| Дашборд | Содержание |
|---------|------------|
| CPU Load | Загрузка CPU user-service |
| User Registrations | `increase(users_new_total[5m])` |
| RabbitMQ Queue Depth | глубина очереди, publish/deliver rate, consumers |

> **RabbitMQ «всегда 0»:** глубина очереди (`messages ready`) часто равна 0, потому что consumer успевает обработать сообщения. Смотрите панели **Publish rate** и **Totals** — они показывают активность.

После изменения дашбордов: `docker compose restart grafana`.

### Алерты (Prometheus → Alertmanager)

| Алерт | Условие | Severity |
|-------|---------|----------|
| HighSystemCPUUsage | CPU user-service > 90% за 2 мин | critical |
| HighUserCreationRate | > 10 пользователей за 5 мин | warning |
| HighGETRequestDuration | средняя длительность GET > 2 с | critical |

Проверка алерта на создание пользователей:

```bash
make create-test-users
```

Через 1–2 минуты сработает `HighUserCreationRate`. Проверка в UI: http://localhost:9093

### Настройка Telegram-бота

1. Создайте бота через [@BotFather](https://t.me/BotFather), токен — в `.env`:
   ```
   TELEGRAM_BOT_TOKEN=...
   ```

2. Запустите стек: `make up-lab4`

3. Напишите боту `/start` — он ответит `chat_id`.

4. Добавьте в `.env`:
   ```
   TELEGRAM_CHAT_ID=...
   ```

5. Перезапустите бота:
   ```bash
   docker compose up -d telegram-bot
   ```

Alertmanager шлёт webhook на `http://telegram-bot:8081/alerts` (см. `dockercompose/alertmanager.yml`).

---

## Локальная разработка (без Docker для Go-сервисов)

```bash
make up                  # postgres + keydb + rabbitmq
make run                 # user-service на :8080
make run-notification    # consumer в отдельном терминале
make run-telegram-bot    # бот в отдельном терминале (нужен .env)
```

---

## Конфигурация

| Файл | Назначение |
|------|------------|
| `.env` | `TELEGRAM_BOT_TOKEN`, `TELEGRAM_CHAT_ID` |
| `dockercompose/prometheus.yml` | scrape targets |
| `dockercompose/alert_rules.yml` | правила алертов |
| `dockercompose/alertmanager.yml` | маршрутизация в telegram-bot |
| `dockercompose/grafana/` | provisioning дашбордов и datasource |

Переменные окружения Go-сервисов (через `internal/config`):

| Переменная | По умолчанию |
|------------|--------------|
| `HTTP_ADDR` | `:8080` |
| `DATABASE_URL` | `postgres://app:app@localhost:5432/app?sslmode=disable` |
| `REDIS_ADDR` | `localhost:6379` |
| `RABBITMQ_URL` | `amqp://guest:guest@localhost:5672/` |

---

## Устранение неполадок

| Проблема | Решение |
|----------|---------|
| `redis-cli не найден` | Запустите `make up-lab4` — скрипты используют `docker exec keydb-archapp-lab4` |
| API не отвечает на localhost:8080 | Используйте `127.0.0.1:8080` |
| `make create-test-users` → 500 | Email уже существует в БД; используйте уникальные email или очистите postgres volume |
| Сеть не удаляется при `down` | `docker compose down --remove-orphans` (уже в Makefile) |
| Grafana RabbitMQ = 0 | Нормально для queue depth; смотрите publish rate, создайте пользователей |
| Алерт не приходит в Telegram | Проверьте `.env`, перезапустите `telegram-bot`, напишите боту `/start` |

Полезные команды:

```bash
docker compose ps
docker logs user-service --tail 50
docker logs notification-service --tail 50
docker logs telegram-bot --tail 50
```

---

## Автор

**Осокин Ярослав Юрьевич** · группа BIVT-23-SP-2, № 17
