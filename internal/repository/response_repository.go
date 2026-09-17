package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"gins-backend/internal/models"
)

var ErrDuplicateResponse = errors.New("исполнитель уже откликался на эту задачу")

type ResponseRepository struct {
	db *pgxpool.Pool
}

func NewResponseRepository(db *pgxpool.Pool) *ResponseRepository {
	return &ResponseRepository{db: db}
}

func (r *ResponseRepository) Create(ctx context.Context, resp *models.Response) (*models.Response, error) {
	const q = `
		INSERT INTO responses (task_id, executor_id, motivation, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	row := r.db.QueryRow(ctx, q, resp.TaskID, resp.ExecutorID, resp.Motivation, resp.Status)
	err := row.Scan(&resp.ID, &resp.CreatedAt)
	if err != nil {
		// уникальный индекс (task_id, executor_id) не даст откликнуться дважды —
		// ловим это здесь по тексту ошибки постгреса, чтобы вернуть
		// понятный ответ, а не голый 500-й
		if isUniqueViolation(err) {
			return nil, ErrDuplicateResponse
		}
		return nil, fmt.Errorf("не удалось создать отклик: %w", err)
	}

	return resp, nil
}

func (r *ResponseRepository) GetByID(ctx context.Context, id int) (*models.Response, error) {
	const q = `
		SELECT id, task_id, executor_id, motivation, status, created_at
		FROM responses
		WHERE id = $1
	`

	var resp models.Response
	err := r.db.QueryRow(ctx, q, id).Scan(
		&resp.ID, &resp.TaskID, &resp.ExecutorID, &resp.Motivation, &resp.Status, &resp.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("не удалось получить отклик: %w", err)
	}

	return &resp, nil
}

func (r *ResponseRepository) UpdateStatus(ctx context.Context, id int, status string) error {
	const q = `UPDATE responses SET status = $1 WHERE id = $2`

	tag, err := r.db.Exec(ctx, q, status, id)
	if err != nil {
		return fmt.Errorf("не удалось обновить статус отклика: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// isUniqueViolation грубо проверяет, что ошибка — нарушение уникального
// индекса. pgx оборачивает *pgconn.PgError, но чтобы не тащить лишний
// импорт ради одной проверки, смотрим на код прямо через errors.As было бы
// правильнее в более крупном проекте — здесь оставлено просто для наглядности.
func isUniqueViolation(err error) bool {
	var pgErr interface{ SQLState() string }
	if errors.As(err, &pgErr) {
		return pgErr.SQLState() == "23505"
	}
	return false
}
