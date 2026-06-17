package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"arch-oyu-lab3/internal/cache"
	"arch-oyu-lab3/internal/config"
	"arch-oyu-lab3/internal/handler"
	"arch-oyu-lab3/internal/mq"
	"arch-oyu-lab3/internal/repository"
	"arch-oyu-lab3/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	setupLogging()

	cfg := config.Load()
	ctx := context.Background()

	dbPool := mustOpenPostgres(ctx, cfg.DatabaseURL)
	defer dbPool.Close()

	userRepo := repository.NewUserRepository(dbPool)
	if err := userRepo.Migrate(ctx); err != nil {
		slog.Error("миграция БД", "error", err)
		os.Exit(1)
	}

	userCache := cache.NewUserCache(cfg.RedisAddr)
	defer userCache.Close()
	if err := userCache.Ping(ctx); err != nil {
		slog.Error("подключение к Redis/KeyDB", "error", err)
		os.Exit(1)
	}

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

	userService := service.NewUserService(userRepo, userCache, rabbitClient)
	userHandler := handler.NewUserHandler(userService)

	engine := newHTTPServerEngine(userHandler)
	httpServer := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           engine,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
	}

	go func() {
		slog.Info("HTTP-сервер слушает", "адрес", cfg.HTTPAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("сервер остановлен с ошибкой", "error", err)
			os.Exit(1)
		}
	}()

	waitForShutdownSignal()
	shutdownGracefully(httpServer)
}

func setupLogging() {
	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	logger := slog.New(jsonHandler)
	slog.SetDefault(logger)
}

// mustOpenPostgres подключается к БД или завершает процесс — так проще читать main сверху вниз.
func mustOpenPostgres(ctx context.Context, databaseURL string) *pgxpool.Pool {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		slog.Error("PostgreSQL: не удалось подключиться", "error", err)
		os.Exit(1)
	}
	return pool
}

func newHTTPServerEngine(userHandler *handler.UserHandler) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery(), gin.Logger())
	handler.RegisterRoutes(engine, userHandler)
	return engine
}

func waitForShutdownSignal() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	slog.Info("получен сигнал остановки, завершаем работу…")
}

func shutdownGracefully(server *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("ошибка при graceful shutdown", "error", err)
	}
}
