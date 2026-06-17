package service

import (
	"context"
	"log/slog"

	"arch-oyu-lab3/internal/cache"
	"arch-oyu-lab3/internal/events"
	"arch-oyu-lab3/internal/metrics"
	"arch-oyu-lab3/internal/models"
	"arch-oyu-lab3/internal/mq"
	"arch-oyu-lab3/internal/repository"

	"github.com/google/uuid"
)

// UserService связывает БД, кэш и публикацию событий в RabbitMQ.
type UserService struct {
	userRepo   *repository.UserRepository
	userCache  *cache.UserCache
	rabbit     *mq.Client
}

func NewUserService(userRepo *repository.UserRepository, userCache *cache.UserCache, rabbit *mq.Client) *UserService {
	return &UserService{userRepo: userRepo, userCache: userCache, rabbit: rabbit}
}

func (s *UserService) List(ctx context.Context) ([]models.User, error) {
	return s.userRepo.List(ctx)
}

// GetByID — cache-aside:
//  1. смотрим Redis;
//  2. если пусто — читаем PostgreSQL и кладём ответ в Redis;
//  3. если в Redis было — сразу отдаём (быстрее).
func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (models.User, error) {
	fromCache, found, err := s.userCache.Get(ctx, id)
	if err != nil {
		return models.User{}, err
	}
	if found {
		slog.Info("пользователь взят из кэша (cache hit)")
		return fromCache, nil
	}

	slog.Info("в кэше нет пользователя, читаем БД (cache miss)")
	fromDB, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return models.User{}, err
	}

	if err := s.userCache.Set(ctx, fromDB); err != nil {
		return models.User{}, err
	}
	return fromDB, nil
}

func (s *UserService) Create(ctx context.Context, data models.UserCreate) (models.User, error) {
	user, err := s.userRepo.Create(ctx, data)
	if err != nil {
		return models.User{}, err
	}

	// Асинхронная связь: после сохранения в БД шлём событие в exchange.
	// HTTP-ответ клиенту не ждёт notification-service.
	event := events.UserCreated{ID: user.ID, Email: user.Email, Name: user.Name}
	if err := s.rabbit.PublishUserCreated(ctx, event); err != nil {
		return models.User{}, err
	}
	slog.Info("событие user.created отправлено в RabbitMQ", "user_id", user.ID.String())
	metrics.UsersCreated.Inc()

	return user, nil
}

// Update сначала меняет строку в БД, потом удаляет запись из кэша,
// чтобы следующий GetByID не прочитал устаревшее имя/email.
func (s *UserService) Update(ctx context.Context, id uuid.UUID, data models.UserUpdate) (models.User, error) {
	updated, err := s.userRepo.Update(ctx, id, data)
	if err != nil {
		return models.User{}, err
	}
	if err := s.userCache.Delete(ctx, id); err != nil {
		return models.User{}, err
	}
	slog.Info("кэш сброшен после обновления пользователя", "user_id", id.String())
	return updated, nil
}

// Delete удаляет из БД и из кэша (если бы удалили только из БД, в Redis мог бы остаться «воскресший» пользователь до TTL).
func (s *UserService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.userRepo.Delete(ctx, id); err != nil {
		return err
	}
	if err := s.userCache.Delete(ctx, id); err != nil {
		return err
	}
	slog.Info("кэш сброшен после удаления пользователя", "user_id", id.String())
	return nil
}
