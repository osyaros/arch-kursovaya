package events

import "github.com/google/uuid"

// UserCreated — событие «пользователь создан».
// Оба сервиса используют одну структуру, чтобы JSON в очереди был одинаковым.
type UserCreated struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
	Name  string    `json:"name"`
}
