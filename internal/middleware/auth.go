package middleware

import (
	"context"
	"net/http"
	"strings"

	"gins-backend/internal/auth"
)

// через эти ключи хендлеры достают данные пользователя из контекста
// после того, как middleware их туда положил
type contextKey string

const (
	userIDKey contextKey = "userID"
	roleKey   contextKey = "role"
)

// RequireAuth проверяет заголовок Authorization: Bearer <токен>,
// достаёт из него пользователя и роль, и кладёт их в контекст запроса.
// Если токена нет или он невалиден — дальше запрос не пускаем.
func RequireAuth(tokenManager *auth.TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				writeUnauthorized(w, "заголовок Authorization отсутствует")
				return
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				writeUnauthorized(w, "ожидался формат: Authorization: Bearer <токен>")
				return
			}

			claims, err := tokenManager.Parse(parts[1])
			if err != nil {
				writeUnauthorized(w, "токен недействителен или просрочен")
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
			ctx = context.WithValue(ctx, roleKey, claims.Role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func writeUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"error": "` + message + `"}`))
}

// UserIDFromContext и RoleFromContext — небольшие хелперы, чтобы в
// хендлерах не тянуть напрямую contextKey и не рисковать опечаткой в имени ключа.

func UserIDFromContext(ctx context.Context) (int, bool) {
	id, ok := ctx.Value(userIDKey).(int)
	return id, ok
}

func RoleFromContext(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(roleKey).(string)
	return role, ok
}
