package api

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// JSON — общая функция для обычного JSON-ответа
func JSON(w http.ResponseWriter, status int, payload interface{}) {
	// 1. Сериализуем в буфер
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(payload); err != nil {
		// Если упало — сразу отдаём обычный http.Error и выходим
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// 2. Заголовок и тело уже после успешной сериализации
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	buf.WriteTo(w)
}

// Error — JSON-ошибка в едином формате
func Error(w http.ResponseWriter, status int, code, message string, details ...interface{}) {
	resp := ErrorResponse{
		Code:    code,
		Message: message,
	}
	if len(details) > 0 {
		resp.Details = details[0]
	}
	JSON(w, status, resp)
}
