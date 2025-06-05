// File: internal/auth/api/handlers.go
package api

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog"
	"pluto-backend/internal/auth/api/gen"
	"pluto-backend/internal/auth/service"
)

// Handlers реализует интерфейс gen.ServerInterface.
type Handlers struct {
	Svc    *service.Service
	Logger *zerolog.Logger
}

// NewHandlers создаёт новый набор обработчиков.
func NewHandlers(svc *service.Service, log *zerolog.Logger) *Handlers {
	return &Handlers{Svc: svc, Logger: log}
}

// Health реализует GET /health.
func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	ResponseJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// AppLogin реализует POST /auth/app-login.
func (h *Handlers) AppLogin(w http.ResponseWriter, r *http.Request) {
	var req gen.AppLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.Logger.Error().Err(err).Msg("AppLogin: failed to decode request body")
		ErrorJSON(w, http.StatusBadRequest, "invalid_request", "failed to parse JSON body")
		return
	}

	resp, err := h.Svc.AppLogin(r.Context(), req)
	if err != nil {
		h.Logger.Error().Err(err).Msg("AppLogin: service error")
		ErrorJSON(w, http.StatusInternalServerError, "internal_error", "failed to create or retrieve session")
		return
	}

	ResponseJSON(w, http.StatusOK, resp)
}

//// AppLogout реализует POST /auth/app-logout.
//func (h *Handlers) AppLogout(w http.ResponseWriter, r *http.Request) {
//	var req gen.AppLogoutRequest
//	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
//		h.Logger.Error().Err(err).Msg("AppLogout: failed to decode request body")
//		ErrorJSON(w, http.StatusBadRequest, "invalid_request", "failed to parse JSON body")
//		return
//	}
//
//	if err := h.Svc.AppLogout(r.Context(), req); err != nil {
//		h.Logger.Error().Err(err).Msg("AppLogout: service error")
//		ErrorJSON(w, http.StatusInternalServerError, "internal_error", "failed to revoke session")
//		return
//	}
//
//	w.WriteHeader(http.StatusNoContent)
//}

// GetPublicKey реализует GET /auth/public-key.
func (h *Handlers) GetPublicKey(w http.ResponseWriter, r *http.Request) {
	resp, err := h.Svc.GetPublicKey(r.Context())
	if err != nil {
		h.Logger.Error().Err(err).Msg("GetPublicKey: service error")
		ErrorJSON(w, http.StatusInternalServerError, "internal_error", "failed to fetch public key")
		return
	}
	ResponseJSON(w, http.StatusOK, resp)
}

// ResponseJSON сериализует payload в JSON, ставит Content-Type и статус.
func ResponseJSON(w http.ResponseWriter, status int, payload interface{}) {
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(payload); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	buf.WriteTo(w)
}

// ErrorJSON формирует JSON-ошибку в едином формате.
func ErrorJSON(w http.ResponseWriter, status int, code, message string, details ...interface{}) {
	resp := ErrorResponse{
		Code:    code,
		Message: message,
	}
	if len(details) > 0 {
		resp.Details = details[0]
	}
	ResponseJSON(w, status, resp)
}

// ErrorResponse описывает формат JSON-ошибки.
type ErrorResponse struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}
