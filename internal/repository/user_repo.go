package repository

import (
	"context"
	"errors"
	"time"

	"arch-oyu-lab3/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNotFound — общая ошибка «записи с таким id нет». Её проверяют в HTTP-слое через errors.Is.
var ErrNotFound = errors.New("user not found")

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// Migrate создаёт таблицу при старте приложения (для лабы достаточно простого DDL).
func (r *UserRepository) Migrate(ctx context.Context) error {
	_, err := r.db.Exec(ctx, `
CREATE TABLE IF NOT EXISTS users (
	id UUID PRIMARY KEY,
	name TEXT NOT NULL,
	email TEXT NOT NULL UNIQUE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
`)
	return err
}

func (r *UserRepository) List(ctx context.Context) ([]models.User, error) {
	const q = `SELECT id, name, email, created_at FROM users ORDER BY created_at`

	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (models.User, error) {
	const q = `SELECT id, name, email, created_at FROM users WHERE id = $1`

	var u models.User
	err := r.db.QueryRow(ctx, q, id).Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.User{}, ErrNotFound
	}
	return u, err
}

func (r *UserRepository) Create(ctx context.Context, data models.UserCreate) (models.User, error) {
	// NewV7 даёт UUID, отсортированные по времени; если недоступен — обычный v4.
	id, err := uuid.NewV7()
	if err != nil {
		id = uuid.New()
	}

	createdAt := time.Now().UTC()

	const q = `INSERT INTO users (id, name, email, created_at) VALUES ($1, $2, $3, $4)`
	_, err = r.db.Exec(ctx, q, id, data.Name, data.Email, createdAt)
	if err != nil {
		return models.User{}, err
	}

	return models.User{
		ID:        id,
		Name:      data.Name,
		Email:     data.Email,
		CreatedAt: createdAt,
	}, nil
}

func (r *UserRepository) Update(ctx context.Context, id uuid.UUID, data models.UserUpdate) (models.User, error) {
	current, err := r.GetByID(ctx, id)
	if err != nil {
		return models.User{}, err
	}

	newName := current.Name
	if data.Name != nil {
		newName = *data.Name
	}
	newEmail := current.Email
	if data.Email != nil {
		newEmail = *data.Email
	}

	const q = `UPDATE users SET name = $2, email = $3 WHERE id = $1`
	tag, err := r.db.Exec(ctx, q, id, newName, newEmail)
	if err != nil {
		return models.User{}, err
	}
	if tag.RowsAffected() == 0 {
		return models.User{}, ErrNotFound
	}

	return r.GetByID(ctx, id)
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const q = `DELETE FROM users WHERE id = $1`
	tag, err := r.db.Exec(ctx, q, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
