package cache

import (
	"context"
	"encoding/json"
	"time"

	"arch-oyu-lab3/internal/models"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Как мы кладём пользователя в Redis: префикс + версия схемы (если формат изменится — можно сделать v2).
const redisKeyPrefix = "user:v1:"

// Сколько живёт запись в кэше без обращений (совпадает с типичным TTL из методички).
const userCacheTTL = 60 * time.Second

// UserCache оборачивает go-redis: только строковые ключи и JSON в значениях.
type UserCache struct {
	client *redis.Client
}

func NewUserCache(serverAddr string) *UserCache {
	return &UserCache{
		client: redis.NewClient(&redis.Options{Addr: serverAddr}),
	}
}

func (c *UserCache) redisKey(userID uuid.UUID) string {
	return redisKeyPrefix + userID.String()
}

// Get читает пользователя из Redis.
//
// Возвращает:
//   - cacheHit == true  — запись нашли, user заполнен;
//   - cacheHit == false, err == nil — ключа не было (нужно идти в БД);
//   - err != nil — сеть, битый JSON и т.п.
func (c *UserCache) Get(ctx context.Context, userID uuid.UUID) (user models.User, cacheHit bool, err error) {
	jsonBytes, err := c.client.Get(ctx, c.redisKey(userID)).Bytes()
	if err == redis.Nil {
		// Ключа нет — это нормальный «промах», не ошибка.
		return models.User{}, false, nil
	}
	if err != nil {
		return models.User{}, false, err
	}

	if err := json.Unmarshal(jsonBytes, &user); err != nil {
		return models.User{}, false, err
	}
	return user, true, nil
}

// Set сохраняет пользователя в Redis с TTL.
func (c *UserCache) Set(ctx context.Context, user models.User) error {
	jsonBytes, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, c.redisKey(user.ID), jsonBytes, userCacheTTL).Err()
}

// Delete убирает пользователя из кэша (после изменения или удаления в БД — чтобы не отдавать старые данные).
func (c *UserCache) Delete(ctx context.Context, userID uuid.UUID) error {
	return c.client.Del(ctx, c.redisKey(userID)).Err()
}

func (c *UserCache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

func (c *UserCache) Close() error {
	return c.client.Close()
}
