package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"arch-oyu-lab3/internal/config"
	"arch-oyu-lab3/internal/events"
	"arch-oyu-lab3/internal/mq"
	"arch-oyu-lab3/internal/service"
)

// notification-service — Consumer: читает user.events и «отправляет» уведомление (лог).
func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	cfg := config.Load()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	rabbitClient, err := mq.Connect(cfg.RabbitMQURL)
	if err != nil {
		slog.Error("подключение к RabbitMQ", "error", err)
		os.Exit(1)
	}
	defer rabbitClient.Close()

	if err := rabbitClient.SetupTopology(); err != nil {
		slog.Error("настройка RabbitMQ", "error", err)
		os.Exit(1)
	}

	notifier := service.NewNotificationService()
	slog.Info("notification-service слушает очередь", "queue", mq.QueueName)

	err = rabbitClient.ConsumeUserCreated(ctx, func(event events.UserCreated) error {
		notifier.NotifyUserCreated(event)
		return nil
	})
	if err != nil && ctx.Err() == nil {
		slog.Error("consumer остановлен", "error", err)
		os.Exit(1)
	}
	slog.Info("notification-service завершён")
}
