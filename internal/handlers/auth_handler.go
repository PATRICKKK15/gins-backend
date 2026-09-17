package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"gins-backend/internal/auth"
	"gins-backend/internal/middleware"
	"gins-backend/internal/models"
	"gins-backend/internal/repository"
)

type AuthHandler struct {
	users        *repository.UserRepository
	tokenManager *auth.TokenManager
}

func NewAuthHandler(users *repository.UserRepository, tokenManager *auth.TokenManager) *AuthHandler {
	return &AuthHandler{users: users, tokenManager: tokenManager}
}

type registerRequest struct {
	FullName     string `json:"full_name"`
	Email        string `json:"email"`
	Password     string `json:"password"`
	Role         string `json:"role"`
	Competencies string `json:"competencies"`
}

// Register — POST /api/auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "некорректный JSON в теле запроса")
		return
	}

	// минимальная валидация — на реальном проекте стоит вынести в
	// отдельный слой, но для объёма практики достаточно проверок тут
	req.FullName = strings.TrimSpace(req.FullName)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	if req.FullName == "" || req.Email == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "поля full_name, email и password обязательны")
		return
	}
	if len(req.Password) < 8 {
		respondError(w, http.StatusBadRequest, "пароль должен быть не короче 8 символов")
		return
	}
	if req.Role != models.RoleCustomer && req.Role != models.RoleExecutor {
		respondError(w, http.StatusBadRequest, "role должна быть customer или executor")
		return
	}

	// проверяем заранее, чтобы вернуть понятную ошибку, а не разбирать
	// текст ошибки уникального индекса из постгреса
	if _, err := h.users.GetByEmail(r.Context(), req.Email); err == nil {
		respondError(w, http.StatusConflict, "пользователь с таким email уже зарегистрирован")
		return
	} else if !errors.Is(err, repository.ErrNotFound) {
		respondError(w, http.StatusInternalServerError, "не удалось проверить email")
		return
	}

	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "не удалось обработать пароль")
		return
	}

	user := &models.User{
		FullName:     req.FullName,
		Email:        req.Email,
		PasswordHash: passwordHash,
		Role:         req.Role,
		Competencies: req.Competencies,
	}

	created, err := h.users.Create(r.Context(), user)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "не удалось создать пользователя")
		return
	}

	respondJSON(w, http.StatusCreated, created)
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

// Login — POST /api/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "некорректный JSON в теле запроса")
		return
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))

	user, err := h.users.GetByEmail(r.Context(), email)
	if errors.Is(err, repository.ErrNotFound) {
		// намеренно не уточняем, что именно неверно — email или пароль,
		// чтобы не помогать перебору email-адресов
		respondError(w, http.StatusUnauthorized, "неверный email или пароль")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "не удалось выполнить вход")
		return
	}

	if !auth.CheckPassword(user.PasswordHash, req.Password) {
		respondError(w, http.StatusUnauthorized, "неверный email или пароль")
		return
	}

	token, err := h.tokenManager.Generate(user.ID, user.Role)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "не удалось сгенерировать токен")
		return
	}

	respondJSON(w, http.StatusOK, loginResponse{Token: token, User: *user})
}

// Me — GET /api/users/me, отдаёт профиль пользователя по токену из запроса
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "не удалось определить пользователя по токену")
		return
	}

	user, err := h.users.GetByID(r.Context(), userID)
	if errors.Is(err, repository.ErrNotFound) {
		respondError(w, http.StatusNotFound, "пользователь не найден")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "не удалось получить профиль")
		return
	}

	respondJSON(w, http.StatusOK, user)
}
