package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool открывает пул соединений к Postgres. Один пул на всё приложение,
// хендлеры и репозитории работают через него, отдельные коннекты не плодим.
func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать пул соединений: %w", err)
	}

	// сразу проверяем, что база реально отвечает, а не молча падать потом
	// на первом запросе где-нибудь в хендлере
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("база данных не отвечает: %w", err)
	}

	return pool, nil
}
