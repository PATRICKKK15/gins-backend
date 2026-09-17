package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"gins-backend/internal/models"
)

var ErrNotFound = errors.New("запись не найдена")
var ErrEmailTaken = errors.New("пользователь с таким email уже существует")

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, u *models.User) (*models.User, error) {
	const q = `
		INSERT INTO users (full_name, email, password_hash, role, competencies)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`

	row := r.db.QueryRow(ctx, q, u.FullName, u.Email, u.PasswordHash, u.Role, u.Competencies)
	if err := row.Scan(&u.ID, &u.CreatedAt); err != nil {
		// 23505 — это код unique_violation в постгресе, но проверять его
		// текстом сообщения не очень надёжно, поэтому проще заранее
		// проверять email на существование в хендлере. Здесь оставляем
		// как есть и просто оборачиваем ошибку.
		return nil, fmt.Errorf("не удалось создать пользователя: %w", err)
	}

	return u, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	const q = `
		SELECT id, full_name, email, password_hash, role, competencies, created_at
		FROM users
		WHERE email = $1
	`

	u := &models.User{}
	err := r.db.QueryRow(ctx, q, email).Scan(
		&u.ID, &u.FullName, &u.Email, &u.PasswordHash, &u.Role, &u.Competencies, &u.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("не удалось получить пользователя по email: %w", err)
	}

	return u, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int) (*models.User, error) {
	const q = `
		SELECT id, full_name, email, password_hash, role, competencies, created_at
		FROM users
		WHERE id = $1
	`

	u := &models.User{}
	err := r.db.QueryRow(ctx, q, id).Scan(
		&u.ID, &u.FullName, &u.Email, &u.PasswordHash, &u.Role, &u.Competencies, &u.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("не удалось получить пользователя по id: %w", err)
	}

	return u, nil
}
