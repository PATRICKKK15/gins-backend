package main

import (
	"context"
	"log"
	"net/http"

	"gins-backend/internal/config"
	"gins-backend/internal/repository"
	"gins-backend/internal/router"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("не удалось загрузить конфигурацию: %v", err)
	}

	ctx := context.Background()

	db, err := repository.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("не удалось подключиться к базе данных: %v", err)
	}
	defer db.Close()

	handler := router.New(db, cfg.JWTSecret)

	log.Printf("сервер запущен на порту %s", cfg.ServerPort)
	if err := http.ListenAndServe(":"+cfg.ServerPort, handler); err != nil {
		log.Fatalf("сервер остановился с ошибкой: %v", err)
	}
}
