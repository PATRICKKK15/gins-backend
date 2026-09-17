package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"gins-backend/internal/models"
)

type TaskRepository struct {
	db *pgxpool.Pool
}

func NewTaskRepository(db *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Create(ctx context.Context, t *models.Task) (*models.Task, error) {
	const q = `
		INSERT INTO tasks (customer_id, title, description, requirements, deadline, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`

	row := r.db.QueryRow(ctx, q, t.CustomerID, t.Title, t.Description, t.Requirements, t.Deadline, t.Status)
	if err := row.Scan(&t.ID, &t.CreatedAt); err != nil {
		return nil, fmt.Errorf("не удалось создать задачу: %w", err)
	}

	return t, nil
}

// List отдаёт все задачи, свежие сверху. Для реального проекта сюда
// напрашивается пагинация, но для объёма практики достаточно и так.
func (r *TaskRepository) List(ctx context.Context) ([]models.Task, error) {
	const q = `
		SELECT id, customer_id, title, description, requirements, deadline, status, created_at
		FROM tasks
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить список задач: %w", err)
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		if err := rows.Scan(&t.ID, &t.CustomerID, &t.Title, &t.Description, &t.Requirements, &t.Deadline, &t.Status, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("не удалось прочитать задачу из результата запроса: %w", err)
		}
		tasks = append(tasks, t)
	}

	return tasks, nil
}

func (r *TaskRepository) GetByID(ctx context.Context, id int) (*models.Task, error) {
	const q = `
		SELECT id, customer_id, title, description, requirements, deadline, status, created_at
		FROM tasks
		WHERE id = $1
	`

	var t models.Task
	err := r.db.QueryRow(ctx, q, id).Scan(
		&t.ID, &t.CustomerID, &t.Title, &t.Description, &t.Requirements, &t.Deadline, &t.Status, &t.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("не удалось получить задачу: %w", err)
	}

	return &t, nil
}
