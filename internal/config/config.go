package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	JWTSecret   string
	ServerPort  string
}

// Load читает .env (если он есть — на проде переменные обычно и так
// заданы окружением, поэтому отсутствие файла не считаем ошибкой)
// и собирает конфиг. Если чего-то важного не хватает — падаем сразу
// при старте, а не посреди работы сервера.
func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		ServerPort:  os.Getenv("SERVER_PORT"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("не задана переменная DATABASE_URL")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("не задана переменная JWT_SECRET")
	}
	if cfg.ServerPort == "" {
		cfg.ServerPort = "8080"
	}

	return cfg, nil
}
