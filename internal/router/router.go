package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"gins-backend/internal/auth"
	"gins-backend/internal/handlers"
	"gins-backend/internal/middleware"
	"gins-backend/internal/models"
	"gins-backend/internal/repository"
)

// New собирает роутер: репозитории -> хендлеры -> маршруты. Всё в одном
// месте, чтобы было видно всю карту API сразу, а не бегать по файлам.
func New(db *pgxpool.Pool, jwtSecret string) http.Handler {
	tokenManager := auth.NewTokenManager(jwtSecret)

	userRepo := repository.NewUserRepository(db)
	taskRepo := repository.NewTaskRepository(db)
	responseRepo := repository.NewResponseRepository(db)

	authHandler := handlers.NewAuthHandler(userRepo, tokenManager)
	taskHandler := handlers.NewTaskHandler(taskRepo)
	responseHandler := handlers.NewResponseHandler(responseRepo, taskRepo)

	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ГИНС-платформа: backend работает"))
	})

	r.Route("/api", func(r chi.Router) {
		// открытые эндпоинты — токен ещё не нужен
		r.Post("/auth/register", authHandler.Register)
		r.Post("/auth/login", authHandler.Login)

		// всё, что ниже, требует валидный JWT-токен
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAuth(tokenManager))

			r.Get("/users/me", authHandler.Me)

			r.Get("/tasks", taskHandler.List)
			r.Get("/tasks/{id}", taskHandler.GetByID)

			// создавать задачи может только заказчик
			r.With(middleware.RequireRole(models.RoleCustomer)).
				Post("/tasks", taskHandler.Create)

			// откликаться на задачу может только исполнитель
			r.With(middleware.RequireRole(models.RoleExecutor)).
				Post("/tasks/{id}/responses", responseHandler.Create)

			// принять/отклонить отклик может только заказчик — владельца
			// задачи хендлер дополнительно проверяет сам, роли тут мало
			r.With(middleware.RequireRole(models.RoleCustomer)).
				Patch("/responses/{id}", responseHandler.Decide)
		})
	})

	return r
}
