# КУРСОВАЯ РАБОТА

**по дисциплине «Архитектура программного обеспечения»**

**Тема:** Разработка распределённого приложения управления пользователями с применением контейнеризации, кеширования, событийно-ориентированной архитектуры и системы наблюдаемости

**Выполнил:** студент группы BIVT-23-SP-2, № 17  
**Осокин Ярослав Юрьевич**

---

## СОДЕРЖАНИЕ

Введение  
1. Теоретическая часть  
2. Проектирование архитектуры  
3. Реализация  
4. Тестирование  
Заключение  
Список использованных источников  

---

## ВВЕДЕНИЕ

Современные информационные системы предъявляют высокие требования к производительности, отказоустойчивости и прозрачности работы. Монолитное приложение с синхронными вызовами и единственным хранилищем данных быстро становится узким местом при росте нагрузки: увеличивается время отклика, усложняется масштабирование отдельных компонентов, затрудняется диагностика инцидентов.

В рамках лабораторного практикума последовательно освоены четыре архитектурных направления: контейнеризация (лабораторная 1), кеширование данных (лабораторная 2), событийно-ориентированное взаимодействие (лабораторная 3) и наблюдаемость (лабораторная 4). Итоговый проект объединяет все четыре этапа в единое приложение — сервис управления пользователями с асинхронными уведомлениями и полным observability-стеком.

**Цель курсовой работы** — спроектировать и реализовать распределённое приложение, демонстрирующее применение контейнеризации, кеширования, event-driven архитектуры и системы мониторинга.

**Задачи:**

1. разработать REST API на Go для управления пользователями;
2. контейнеризировать приложение и инфраструктурные компоненты с помощью Docker Compose;
3. реализовать кеширование пользователей в KeyDB (Redis) по стратегии Cache-Aside;
4. организовать асинхронную обработку событий через RabbitMQ (producer/consumer);
5. настроить систему мониторинга Prometheus + Grafana + Alertmanager;
6. реализовать доставку алертов в Telegram через webhook.

**Технологический стек:** Go 1.23, Gin, PostgreSQL 16, KeyDB 6.3, RabbitMQ 3 (management + prometheus plugin), Prometheus 2.44, Grafana, Alertmanager 0.24, Docker Compose.

---

## 1 ТЕОРЕТИЧЕСКАЯ ЧАСТЬ

### 1.1 Контейнеризация с Docker

Контейнеризация — технология изоляции приложений и их зависимостей в облегчённых окружениях (контейнерах). В отличие от виртуальных машин, контейнеры совместно используют ядро операционной системы хоста, что снижает накладные расходы на запуск и обслуживание.

Docker — платформа контейнеризации, в которой образ (image) описывает файловую систему и конфигурацию приложения, а контейнер является запущенным экземпляром образа. Docker Compose позволяет декларативно описать многокомпонентное приложение: зависимости между сервисами, сетевые настройки, тома и переменные окружения.

В лабораторной работе 1 изучены основы Docker: создание Dockerfile, многоэтапная сборка (multi-stage build), запуск инфраструктурных сервисов (PostgreSQL, KeyDB, RabbitMQ) через `docker compose`. Каждый компонент системы изолирован в отдельном контейнере, что обеспечивает воспроизводимость окружения на любой машине разработчика.

### 1.2 Кеширование с Redis/KeyDB

Кеширование — механизм временного хранения часто запрашиваемых данных в быстром хранилище для снижения нагрузки на основную базу данных и уменьшения времени отклика.

Redis (Remote Dictionary Server) — in-memory хранилище типа «ключ–значение», поддерживающее строки, хеши, списки, множества и отсортированные множества. KeyDB — форк Redis с совместимым протоколом, использованный в проекте как сервер кеширования.

Стратегия **Cache-Aside (Lazy Loading)** — наиболее распространённый паттерн кеширования. При чтении приложение сначала обращается к кешу; при промахе (cache miss) данные загружаются из БД и записываются в кеш. При изменении или удалении данных соответствующий ключ инвалидируется.

В лабораторной работе 2 дополнительно изучены структуры данных Redis: строки (`SET`/`GET`), хеши (`HSET`/`HGETALL`), списки (`RPUSH`/`LRANGE`), множества (`SADD`/`SMEMBERS`), отсортированные множества (`ZADD`/`ZRANGE`), а также итерация по ключам командой `SCAN`.

### 1.3 Событийно-ориентированная архитектура и RabbitMQ

Событийно-ориентированная архитектура (Event-Driven Architecture, EDA) — архитектурный стиль, при котором компоненты взаимодействуют посредством обмена событиями. Отправитель (producer) публикует событие, не зная о получателях. Потребители (consumers) подписываются на события и обрабатывают их асинхронно.

Преимущества EDA: слабая связанность сервисов, независимое масштабирование producer и consumer, устойчивость к временной недоступности downstream-компонентов.

RabbitMQ — брокер сообщений, реализующий протокол AMQP 0-9-1. Ключевые сущности: exchange (точка маршрутизации), queue (очередь), binding (правило маршрутизации). RabbitMQ поддерживает типы exchange: fanout, direct, topic, headers. В лабораторной работе 3 изучена маршрутизация сообщений для всех четырёх типов через Management UI, а в приложении используется topic exchange `user.exchange` с binding `user.#`.

### 1.4 Observability: Prometheus, Grafana и Alertmanager

Observability (наблюдаемость) — способность понять внутреннее состояние системы по внешним сигналам. Три столпа наблюдаемости: метрики, логи и трассировки. В данной работе реализован уровень метрик и алертинга.

**Prometheus** — система мониторинга с pull-моделью сбора: Prometheus периодически обращается к HTTP-эндпоинту `/metrics` каждого сервиса. Метрики хранятся как временные ряды; запросы выполняются на языке PromQL.

**Grafana** — платформа визуализации метрик. Позволяет строить дашборды из различных источников данных. В проекте настроены три дашборда: загрузка CPU, регистрации пользователей и состояние очереди RabbitMQ.

**Alertmanager** — компонент, принимающий алерты от Prometheus, группирующий их и доставляющий получателям (email, webhook, Telegram и др.). В проекте Alertmanager отправляет webhook на сервис `telegram-bot`, который пересылает уведомления в Telegram.

---

## 2 ПРОЕКТИРОВАНИЕ АРХИТЕКТУРЫ

### 2.1 Общая архитектура системы

Система состоит из трёх прикладных сервисов и инфраструктурного слоя. Прикладные сервисы: `user-service` (REST API + producer), `notification-service` (consumer), `telegram-bot` (приём алертов). Инфраструктура: PostgreSQL, KeyDB, RabbitMQ, Prometheus, Grafana, Alertmanager.

| Компонент | Технология | Роль |
|-----------|------------|------|
| user-service | Go + Gin | REST API, кеширование, публикация событий, метрики `/metrics` |
| notification-service | Go | Consumer RabbitMQ, логирование уведомлений |
| telegram-bot | Go | Webhook Alertmanager → Telegram |
| postgres | PostgreSQL 16 | Основное хранилище пользователей |
| keydb | KeyDB 6.3 | Кеш пользователей (Cache-Aside) |
| rabbitmq | RabbitMQ 3 | Брокер событий |
| prometheus | Prometheus 2.44 | Сбор и хранение метрик |
| grafana | Grafana | Визуализация метрик |
| alertmanager | Alertmanager 0.24 | Маршрутизация алертов |

*Таблица 1 — Компоненты системы*

Поток создания пользователя:

1. HTTP `POST /api/users` → `UserService.Create`;
2. `INSERT` в PostgreSQL;
3. публикация события `UserCreated` в exchange `user.exchange` с routing key `user.created`;
4. `notification-service` получает сообщение из очереди `user.events` и логирует уведомление;
5. инкремент метрики `users_new_total`.

### 2.2 Бизнес-логика и доменная модель

Центральная сущность — **пользователь (User)**. Содержит: UUID-идентификатор, имя, email, дату создания.

REST API реализует полный CRUD:

| Метод | Путь | Описание |
|-------|------|----------|
| GET | `/api/users` | Список всех пользователей |
| POST | `/api/users` | Создание пользователя |
| GET | `/api/users/:id` | Получение по ID (с кешированием) |
| PATCH | `/api/users/:id` | Частичное обновление |
| DELETE | `/api/users/:id` | Удаление |
| GET | `/health` | Проверка работоспособности |
| GET | `/metrics` | Метрики Prometheus |

### 2.3 Архитектура кеширования (лабораторная 2)

Кеширование реализовано по стратегии **Cache-Aside** для операции `GetByID`:

1. запрос к KeyDB по ключу `user:v1:{uuid}`;
2. при cache hit — возврат данных из кеша без обращения к PostgreSQL;
3. при cache miss — чтение из БД и запись в кеш с TTL 60 секунд.

При `Update` и `Delete` ключ в кеше удаляется (инвалидация), чтобы следующий `GetByID` не вернул устаревшие данные. Список пользователей (`List`) не кешируется, так как результат меняется при каждом создании.

Дополнительно в лабораторной работе 2 продемонстрирована работа со структурами данных Redis через скрипт `keydb-student-init.sh`: для студента группы BIVT-23-SP-2, № 17 созданы ключи `student:BIVT-23-SP-2:17` (string), `:info` (hash), `:timetable` (list), `:skills` (set), `:tasks_w_priority` (sorted set). Обход ключей выполняется командой `SCAN` через скрипт `redis-scan-dump.sh`.

### 2.4 Архитектура брокера сообщений (лабораторная 3)

В RabbitMQ настроен **topic exchange** `user.exchange`, привязанный к очереди `user.events` с binding key `user.#`. При создании пользователя `user-service` публикует JSON-событие с routing key `user.created`.

`notification-service` подключается к очереди `user.events` и потребляет сообщения. Подтверждение (ack) отправляется только после успешной обработки; при ошибке сообщение возвращается в очередь (nack с requeue), что реализует семантику at-least-once delivery.

В рамках лабораторной работы 3 через Management UI дополнительно созданы и продемонстрированы обменники всех типов (fanout, direct, topic, headers) с очередью `queue.BIVT-23-SP-2.17` и соответствующими binding-правилами.

### 2.5 Архитектура мониторинга (лабораторная 4)

`user-service` экспортирует метрики на `/metrics`. Prometheus собирает их каждые 15 секунд, а также метрики RabbitMQ с порта 15692 (prometheus plugin).

| Метрика | Тип | Описание |
|---------|-----|----------|
| `users_new_total` | Counter | Количество созданных пользователей |
| `api_user_request_duration_seconds` | Histogram | Длительность GET/POST запросов, лейбл `method=get\|post` |

*Таблица 2 — Кастомные метрики Prometheus*

Настроены три алерта:

| Алерт | Условие | Severity |
|-------|---------|----------|
| HighSystemCPUUsage | CPU user-service > 90% за 2 мин | critical |
| HighUserCreationRate | > 10 пользователей за 5 мин | warning |
| HighGETRequestDuration | средняя длительность GET > 2 с | critical |

Alertmanager маршрутизирует все алерты на webhook `http://telegram-bot:8081/alerts`. Сервис `telegram-bot` форматирует сообщение и отправляет его в Telegram через Bot API.

В Grafana provisioned три дашборда: CPU Load, User Registrations, RabbitMQ Queue.

---

## 3 РЕАЛИЗАЦИЯ

### 3.1 Структура проекта

Проект реализован на Go 1.23. Ссылка на репозиторий: `https://github.com/kinso/arch-oyu-lab4`

```
arch-oyu-lab4/
├── cmd/
│   ├── server/           # user-service — REST API
│   ├── notification/     # notification-service — consumer
│   └── telegram-bot/     # приём webhook от Alertmanager
├── internal/
│   ├── cache/            # Cache-Aside через go-redis
│   ├── config/           # конфигурация из env
│   ├── events/           # модель UserCreated
│   ├── handler/          # HTTP handlers (Gin)
│   ├── metrics/          # Prometheus counter + histogram
│   ├── models/           # доменные модели
│   ├── mq/               # RabbitMQ client (publish/consume)
│   ├── repository/       # PostgreSQL (pgxpool)
│   ├── service/          # бизнес-логика
│   └── telegram/         # клиент Telegram API + webhook handler
├── dockercompose/
│   ├── prometheus.yml
│   ├── alert_rules.yml
│   ├── alertmanager.yml
│   └── grafana/          # provisioning дашбордов
├── scripts/
│   ├── keydb-student-init.sh
│   ├── redis-scan-dump.sh
│   ├── create-test-users.sh
│   └── get-telegram-chat-id.sh
├── Dockerfile
├── Dockerfile.notification
├── Dockerfile.telegram-bot
├── docker-compose.yaml
└── Makefile
```

### 3.2 Основные компоненты

**UserService** связывает три зависимости: репозиторий PostgreSQL, кеш KeyDB и клиент RabbitMQ. Метод `GetByID` реализует Cache-Aside. Метод `Create` сохраняет пользователя в БД, публикует событие `UserCreated` в RabbitMQ и инкрементирует метрику `users_new_total`. Методы `Update` и `Delete` инвалидируют кеш.

**MQ Client** (`internal/mq/rabbit.go`) объявляет топологию RabbitMQ при старте (идемпотентно): exchange `user.exchange` (topic), очередь `user.events`, binding `user.#`. Публикация использует routing key `user.created`. Consumer работает с `autoAck=false` для надёжной доставки.

**NotificationService** — consumer, который при получении события `UserCreated` записывает уведомление в структурированный лог (JSON через slog).

**Metrics** — пакет `internal/metrics` регистрирует Counter `users_new_total` и HistogramVec `api_user_request_duration_seconds` с лейблом `method`. Измерение длительности выполняется в HTTP-handler для GET и POST запросов.

**Telegram-bot** — отдельный сервис с эндпоинтами `/health` и `/alerts`. Принимает webhook от Alertmanager, форматирует алерты в HTML-сообщение и отправляет в Telegram. Поддерживает команды `/start` и `/chatid` для получения chat_id.

### 3.3 Контейнеризация (лабораторная 1)

Каждый Go-сервис собирается в Docker-образ с multi-stage build:

```dockerfile
FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /server ./cmd/server

FROM alpine:3.20
COPY --from=build /server /server
EXPOSE 8080
ENTRYPOINT ["/server"]
```

Файл `docker-compose.yaml` описывает 9 сервисов. Все объединены в сеть `arch-oyu-lab4_default`. Для PostgreSQL и RabbitMQ настроены тома для персистентности данных.

| Сервис | Порт (хост) | Назначение |
|--------|-------------|------------|
| user-service | 8080 | REST API + метрики |
| telegram-bot | 8081 | Webhook алертов |
| postgres | 5433 | PostgreSQL |
| keydb | 6379 | KeyDB/Redis |
| rabbitmq | 15672 | Management UI |
| prometheus | 9090 | Web UI Prometheus |
| grafana | 3000 | Дашборды (admin/admin) |
| alertmanager | 9093 | UI Alertmanager |

*Таблица 3 — Порты сервисов*

Запуск полного стека: `make up-lab4`. Остановка: `make down-lab4`.

### 3.4 Ссылка на репозиторий

Исходный код размещён в репозитории: `https://github.com/kinso/arch-oyu-lab4`

Для запуска:

```bash
cp .env.example .env
# заполнить TELEGRAM_BOT_TOKEN и TELEGRAM_CHAT_ID
make up-lab4
```

API доступен по адресу `http://127.0.0.1:8080/api/users`.

---

## 4 ТЕСТИРОВАНИЕ

### 4.1 Функциональное тестирование REST API

Функциональное тестирование выполнено с помощью Hoppscotch. Проверены все основные эндпоинты.

#### 4.1.1 Проверка работоспособности сервиса

Запрос `GET /health` возвращает статус 200 OK с телом `{"status":"ok"}`, подтверждая работоспособность API.

*[Рисунок 1 — Проверка работоспособности GET /health]*

#### 4.1.2 Создание пользователя

Запрос `POST /api/users` с телом `{"name": "Иван Иванов", "email": "ivan@example.com"}` возвращает статус 201 Created с данными созданного пользователя, включая сгенерированный UUID.

*[Рисунок 2 — Создание пользователя POST /api/users]*

#### 4.1.3 Получение пользователя по ID

Запрос `GET /api/users/{id}` возвращает пользователя. Первый запрос — cache miss (обращение к PostgreSQL), последующие — cache hit (данные из KeyDB). В логах user-service видны сообщения «пользователь взят из кэша» / «в кэше нет пользователя, читаем БД».

*[Рисунок 3 — Получение пользователя по ID GET /api/users/{id}]*

#### 4.1.4 Получение списка пользователей

Запрос `GET /api/users` возвращает массив всех пользователей.

*[Рисунок 4 — Получение списка GET /api/users]*

#### 4.1.5 Обновление и удаление пользователя

Запрос `PATCH /api/users/{id}` с телом `{"name": "Новое имя"}` возвращает 200 OK с обновлёнными данными. После обновления кеш инвалидируется.

Запрос `DELETE /api/users/{id}` возвращает 204 No Content.

*[Рисунок 5 — Обновление пользователя PATCH /api/users/{id}]*

### 4.2 Тестирование кеширования (лабораторная 2)

Скрипт `make init-student` создаёт в KeyDB набор ключей со структурами данных Redis. Скрипт `make scan` выполняет обход ключей через `SCAN` и выводит тип и содержимое каждого ключа.

*[Рисунок 6 — Результат SCAN по ключам student:BIVT-23-SP-2:17]*

### 4.3 Тестирование event-driven взаимодействия (лабораторная 3)

После создания пользователя через `POST /api/users` в логах `notification-service` появляется запись «уведомление: создан новый пользователь» с user_id, email и name.

В RabbitMQ Management UI (`http://localhost:15672`) видно, что сообщения публикуются в exchange `user.exchange` и потребляются из очереди `user.events`. После обработки очередь пуста (Ready = 0).

*[Рисунок 7 — RabbitMQ Management UI: очередь user.events]*

*[Рисунок 8 — Логи notification-service при создании пользователя]*

### 4.4 Тестирование observability (лабораторная 4)

#### 4.4.1 Prometheus

В интерфейсе Prometheus (`http://localhost:9090`) выполнен запрос `users_new_total` — метрика отображает количество созданных пользователей. Запрос `api_user_request_duration_seconds_count{method="get"}` показывает количество GET-запросов.

*[Рисунок 9 — Prometheus: метрика users_new_total]*

#### 4.4.2 Grafana

В Grafana (`http://localhost:3000`, admin/admin) доступны три дашборда:

- **CPU Load** — загрузка CPU процесса user-service;
- **User Registrations** — график `users_new_total` (скорость создания пользователей);
- **RabbitMQ Queue** — метрики очереди из RabbitMQ prometheus plugin.

*[Рисунок 10 — Дашборд Grafana: User Registrations]*

*[Рисунок 11 — Дашборд Grafana: RabbitMQ Queue]*

#### 4.4.3 Алерты и Telegram

Скрипт `make create-test-users` создаёт 12 пользователей. Через 1–2 минуты срабатывает алерт `HighUserCreationRate` (> 10 пользователей за 5 минут). Alertmanager отправляет webhook на `telegram-bot`, бот доставляет сообщение в Telegram.

*[Рисунок 12 — Алерт HighUserCreationRate в Alertmanager UI]*

*[Рисунок 13 — Уведомление об алерте в Telegram]*

#### 4.4.4 Targets Prometheus

На странице Status → Targets видны два job: `user-service` (UP) и `rabbitmq` (UP).

*[Рисунок 14 — Prometheus Targets: user-service и rabbitmq]*

---

## ЗАКЛЮЧЕНИЕ

В ходе выполнения курсовой работы спроектировано и реализовано распределённое приложение управления пользователями, объединяющее результаты четырёх лабораторных работ.

**Достигнутые результаты:**

1. Разработан REST API на Go (Gin) с полным CRUD для пользователей, контейнеризированный в Docker.
2. Реализовано кеширование в KeyDB по стратегии Cache-Aside с инвалидацией при изменении данных. Продемонстрирована работа со структурами данных Redis (string, hash, list, set, sorted set).
3. Организована событийно-ориентированная обработка через RabbitMQ: `user-service` публикует событие `UserCreated`, `notification-service` потребляет его асинхронно.
4. Настроена система наблюдаемости: Prometheus собирает кастомные метрики (`users_new_total`, `api_user_request_duration_seconds`) и метрики RabbitMQ; Grafana отображает три дашборда; Alertmanager доставляет три типа алертов в Telegram.
5. Все 9 компонентов системы запускаются одной командой `make up-lab4` через Docker Compose.

Система демонстрирует применение ключевых архитектурных паттернов: контейнеризация, кеширование, event-driven communication и observability. Возможные направления развития: внедрение паттерна Transactional Outbox для гарантированной доставки событий, горизонтальное масштабирование через Kubernetes, добавление распределённой трассировки (OpenTelemetry).

---

## СПИСОК ИСПОЛЬЗОВАННЫХ ИСТОЧНИКОВ

1. Docker Documentation. Containerization and Docker Compose. — URL: https://docs.docker.com (дата обращения: 17.06.2026).

2. Redis Documentation. Redis as a Cache. — URL: https://redis.io/docs/manual/client-side-caching (дата обращения: 17.06.2026).

3. KeyDB Documentation. — URL: https://docs.keydb.dev (дата обращения: 17.06.2026).

4. RabbitMQ Documentation. AMQP 0-9-1 Model Explained. — URL: https://www.rabbitmq.com/tutorials/amqp-concepts.html (дата обращения: 17.06.2026).

5. Kleppmann M. Designing Data-Intensive Applications. — O'Reilly Media, 2017. — 616 с.

6. Prometheus Documentation. Overview. — URL: https://prometheus.io/docs/introduction/overview (дата обращения: 17.06.2026).

7. Grafana Documentation. Grafana fundamentals. — URL: https://grafana.com/docs/grafana/latest (дата обращения: 17.06.2026).

8. Go Documentation. The Go Programming Language. — URL: https://go.dev/doc (дата обращения: 17.06.2026).

9. Gin Web Framework Documentation. — URL: https://gin-gonic.com/docs (дата обращения: 17.06.2026).

10. PostgreSQL Documentation. PostgreSQL 16. — URL: https://www.postgresql.org/docs/16 (дата обращения: 17.06.2026).

11. Richardson C. Microservices Patterns. — Manning Publications, 2018. — 520 с.

12. Alertmanager Documentation. — URL: https://prometheus.io/docs/alerting/latest/alertmanager (дата обращения: 17.06.2026).
