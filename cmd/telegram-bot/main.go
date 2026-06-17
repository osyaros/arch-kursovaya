package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"arch-oyu-lab3/internal/telegram"
)

func main() {
	setupLogging()

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")
	if token == "" || chatID == "" {
		slog.Error("нужны TELEGRAM_BOT_TOKEN и TELEGRAM_CHAT_ID")
		os.Exit(1)
	}

	addr := envOr("HTTP_ADDR", ":8081")
	client := telegram.NewClient(token, chatID)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go telegram.RunCommandListener(ctx, token, client)

	mux := http.NewServeMux()
	mux.Handle("/alerts", telegram.NewAlertHandler(client))
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		slog.Info("telegram-bot слушает", "addr", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("сервер остановлен", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("останавливаем telegram-bot…")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)
}

func setupLogging() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))
}

func envOr(name, defaultValue string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return defaultValue
}
