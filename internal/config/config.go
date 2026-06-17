package config

import "os"

// Config — всё, что приложение читает из переменных окружения.
type Config struct {
	HTTPAddr    string
	DatabaseURL string
	RedisAddr   string
	// RabbitMQURL — AMQP, например amqp://guest:guest@localhost:5672/
	RabbitMQURL string
}

func Load() Config {
	return Config{
		HTTPAddr:    envOr("HTTP_ADDR", ":8080"),
		DatabaseURL: envOr("DATABASE_URL", "postgres://app:app@localhost:5433/app?sslmode=disable"),
		RedisAddr:   envOr("REDIS_ADDR", "localhost:6379"),
		RabbitMQURL: envOr("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
	}
}

func envOr(name, defaultValue string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return defaultValue
}
