package middleware

import "net/http"

// RequireRole пропускает дальше только пользователей с указанной ролью.
// Ставится после RequireAuth, потому что читает роль из контекста,
// а туда её кладёт именно RequireAuth.
func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := RoleFromContext(r.Context())
			if !ok || userRole != role {
				writeForbidden(w, "недостаточно прав для этого действия")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func writeForbidden(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	w.Write([]byte(`{"error": "` + message + `"}`))
}
