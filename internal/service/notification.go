package service

import (
	"log/slog"

	"arch-oyu-lab3/internal/events"
)

// NotificationService — «отправка уведомления» для лабы = запись в лог.
type NotificationService struct{}

func NewNotificationService() *NotificationService {
	return &NotificationService{}
}

func (s *NotificationService) NotifyUserCreated(event events.UserCreated) {
	slog.Info(
		"уведомление: создан новый пользователь",
		"user_id", event.ID.String(),
		"email", event.Email,
		"name", event.Name,
	)
}
