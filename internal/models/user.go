package models

import (
	"time"

	"github.com/google/uuid"
)

// User — как пользователь хранится в БД и отдаётся в JSON-ответах.
type User struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// UserCreate — тело запроса на создание пользователя.
type UserCreate struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required"`
}

// UserUpdate — тело запроса на частичное обновление.
// Указатели (*string) нужны, чтобы отличать «поле не прислали» от «прислали пустую строку».
// nil = не менять поле, не-nil = записать новое значение (включая "").
type UserUpdate struct {
	Name  *string `json:"name"`
	Email *string `json:"email"`
}
