package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"gins-backend/internal/middleware"
	"gins-backend/internal/models"
	"gins-backend/internal/repository"
)

type TaskHandler struct {
	tasks *repository.TaskRepository
}

func NewTaskHandler(tasks *repository.TaskRepository) *TaskHandler {
	return &TaskHandler{tasks: tasks}
}

type createTaskRequest struct {
	Title        string `json:"title"`
	Description  string `json:"description"`
	Requirements string `json:"requirements"`
	// принимаем дедлайн строкой в формате RFC3339 (например,
	// "2026-08-01T00:00:00Z"), чтобы не тянуть в JSON свой формат времени
	Deadline string `json:"deadline"`
}

// Create — POST /api/tasks. Создавать задачи может только заказчик,
// это проверяется ещё до вызова хендлера в middleware/роутере.
func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	var req createTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "некорректный JSON в теле запроса")
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	req.Description = strings.TrimSpace(req.Description)

	if req.Title == "" || req.Description == "" {
		respondError(w, http.StatusBadRequest, "поля title и description обязательны")
		return
	}
	if len(req.Title) > 100 {
		respondError(w, http.StatusBadRequest, "title не должен превышать 100 символов")
		return
	}
	if len(req.Description) > 5000 {
		respondError(w, http.StatusBadRequest, "description не должен превышать 5000 символов")
		return
	}

	task := &models.Task{
		CustomerID:   userID,
		Title:        req.Title,
		Description:  req.Description,
		Requirements: req.Requirements,
		Status:       models.TaskStatusOpen,
	}

	if req.Deadline != "" {
		deadline, err := time.Parse(time.RFC3339, req.Deadline)
		if err != nil {
			respondError(w, http.StatusBadRequest, "deadline должен быть в формате RFC3339, например 2026-08-01T00:00:00Z")
			return
		}
		task.Deadline = &deadline
	}

	created, err := h.tasks.Create(r.Context(), task)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "не удалось создать задачу")
		return
	}

	respondJSON(w, http.StatusCreated, created)
}

// List — GET /api/tasks
func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.tasks.List(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "не удалось получить список задач")
		return
	}

	// отдаём пустой массив, а не null, если задач ещё нет — так удобнее
	// клиенту, не нужно отдельно проверять на null
	if tasks == nil {
		tasks = []models.Task{}
	}

	respondJSON(w, http.StatusOK, tasks)
}

// GetByID — GET /api/tasks/{id}
func (h *TaskHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "id задачи должен быть числом")
		return
	}

	task, err := h.tasks.GetByID(r.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		respondError(w, http.StatusNotFound, "задача не найдена")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "не удалось получить задачу")
		return
	}

	respondJSON(w, http.StatusOK, task)
}
