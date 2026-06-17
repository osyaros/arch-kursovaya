.PHONY: up down build run run-notification init-student scan up-lab4 build-lab4 down-lab4 telegram-chat-id create-test-users

up:
	docker compose up -d postgres keydb rabbitmq

down:
	docker compose down --remove-orphans

build:
	go build -o bin/user-service ./cmd/server
	go build -o bin/notification-service ./cmd/notification
	go build -o bin/telegram-bot ./cmd/telegram-bot

up-lab4:
	docker compose up -d --build

build-lab4:
	docker compose build

down-lab4:
	docker compose down --remove-orphans

telegram-chat-id:
	./scripts/get-telegram-chat-id.sh

create-test-users:
	./scripts/create-test-users.sh

run:
	HTTP_ADDR=:8080 \
	DATABASE_URL=postgres://app:app@localhost:5433/app?sslmode=disable \
	REDIS_ADDR=localhost:6379 \
	RABBITMQ_URL=amqp://guest:guest@localhost:5672/ \
	go run ./cmd/server

run-notification:
	RABBITMQ_URL=amqp://guest:guest@localhost:5672/ \
	go run ./cmd/notification

run-telegram-bot:
	set -a && . ./.env && set +a && \
	HTTP_ADDR=:8081 go run ./cmd/telegram-bot

init-student:
	STUDENT_GROUP=BIVT-23-SP-2 STUDENT_NO=17 ./scripts/keydb-student-init.sh

scan:
	./scripts/redis-scan-dump.sh
