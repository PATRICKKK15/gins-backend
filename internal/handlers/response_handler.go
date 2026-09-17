package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"gins-backend/internal/middleware"
	"gins-backend/internal/models"
	"gins-backend/internal/repository"
)

type ResponseHandler struct {
	responses *repository.ResponseRepository
	tasks     *repository.TaskRepository
}

func NewResponseHandler(responses *repository.ResponseRepository, tasks *repository.TaskRepository) *ResponseHandler {
	return &ResponseHandler{responses: responses, tasks: tasks}
}

type createResponseRequest struct {
	Motivation string `json:"motivation"`
}

// Create — POST /api/tasks/{id}/responses. Откликаться может только
// исполнитель, это уже проверено в роутере на уровне middleware по роли.
func (h *ResponseHandler) Create(w http.ResponseWriter, r *http.Request) {
	executorID, _ := middleware.UserIDFromContext(r.Context())

	taskID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "id задачи должен быть числом")
		return
	}

	// проверяем, что задача вообще существует и ещё открыта — глупо
	// принимать отклики на уже закрытую задачу
	task, err := h.tasks.GetByID(r.Context(), taskID)
	if errors.Is(err, repository.ErrNotFound) {
		respondError(w, http.StatusNotFound, "задача не найдена")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "не удалось проверить задачу")
		return
	}
	if task.Status != models.TaskStatusOpen {
		respondError(w, http.StatusConflict, "задача уже не принимает отклики")
		return
	}

	var req createResponseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "некорректный JSON в теле запроса")
		return
	}

	req.Motivation = strings.TrimSpace(req.Motivation)
	if len(req.Motivation) > 2000 {
		respondError(w, http.StatusBadRequest, "motivation не должен превышать 2000 символов")
		return
	}

	resp := &models.Response{
		TaskID:     taskID,
		ExecutorID: executorID,
		Motivation: req.Motivation,
		Status:     models.ResponseStatusPending,
	}

	created, err := h.responses.Create(r.Context(), resp)
	if errors.Is(err, repository.ErrDuplicateResponse) {
		respondError(w, http.StatusConflict, "вы уже откликались на эту задачу")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "не удалось создать отклик")
		return
	}

	respondJSON(w, http.StatusCreated, created)
}

type decideResponseRequest struct {
	// ожидаем "accepted" или "rejected" — решение заказчика по отклику
	Status string `json:"status"`
}

// Decide — PATCH /api/responses/{id}. Принять или отклонить отклик
// может только заказчик, которому принадлежит задача из этого отклика.
func (h *ResponseHandler) Decide(w http.ResponseWriter, r *http.Request) {
	customerID, _ := middleware.UserIDFromContext(r.Context())

	responseID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "id отклика должен быть числом")
		return
	}

	var req decideResponseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "некорректный JSON в теле запроса")
		return
	}
	if req.Status != models.ResponseStatusAccepted && req.Status != models.ResponseStatusRejected {
		respondError(w, http.StatusBadRequest, "status должен быть accepted или rejected")
		return
	}

	resp, err := h.responses.GetByID(r.Context(), responseID)
	if errors.Is(err, repository.ErrNotFound) {
		respondError(w, http.StatusNotFound, "отклик не найден")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "не удалось получить отклик")
		return
	}

	// главная проверка прав: решение по отклику может принять только
	// тот заказчик, чья это задача, а не любой залогиненный заказчик
	task, err := h.tasks.GetByID(r.Context(), resp.TaskID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "не удалось проверить задачу отклика")
		return
	}
	if task.CustomerID != customerID {
		respondError(w, http.StatusForbidden, "решение по отклику может принять только владелец задачи")
		return
	}

	if err := h.responses.UpdateStatus(r.Context(), responseID, req.Status); err != nil {
		respondError(w, http.StatusInternalServerError, "не удалось обновить статус отклика")
		return
	}

	resp.Status = req.Status
	respondJSON(w, http.StatusOK, resp)
}
