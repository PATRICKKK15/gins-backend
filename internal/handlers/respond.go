package handlers

import (
	"encoding/json"
	"net/http"
)

// respondJSON — единая точка, через которую хендлеры отдают JSON.
// Чтобы формат ответа не расползался по проекту, все пишут через неё.
func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if payload == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		// сюда мы попадаем, только если данные вообще нельзя закодировать
		// в JSON — на практике такое означает баг в самом хендлере
		http.Error(w, `{"error": "внутренняя ошибка сервера"}`, http.StatusInternalServerError)
	}
}

// respondError — то же самое, но специально для ошибок, чтобы везде
// был один и тот же формат: {"error": "текст"}.
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}
