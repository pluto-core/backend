package handlers

import (
	"encoding/json"
	"github.com/rs/zerolog"
	"net/http"
	"pluto-backend/internal/manifest/service"
	"strconv"
)

func listManifestsHandler(log *zerolog.Logger, svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const (
			defaultLimit  = 100
			defaultOffset = 0
		)

		q := r.URL.Query()
		limit := defaultLimit
		if l := q.Get("limit"); l != "" {
			if v, err := strconv.Atoi(l); err == nil && v > 0 {
				limit = v
			} else {
				http.Error(w, "invalid limit", http.StatusBadRequest)
				return
			}
		}
		offset := defaultOffset
		if o := q.Get("offset"); o != "" {
			if v, err := strconv.Atoi(o); err == nil && v >= 0 {
				offset = v
			} else {
				http.Error(w, "invalid offset", http.StatusBadRequest)
				return
			}
		}

		// Вызываем сервис
		manifests, err := svc.ListManifests(r.Context(), int32(limit), int32(offset))
		if err != nil {
			log.Error().Err(err).Msg("ListManifests failed")
			http.Error(w, "failed to list manifests", http.StatusInternalServerError)
			return
		}

		// Отправляем JSON
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(manifests); err != nil {
			log.Error().Err(err).Msg("encode response failed")
		}
	}
}

func searchManifestsHandler(log *zerolog.Logger, svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		locale := r.Header.Get("Accept-Language")
		if locale == "" {
			locale = "en"
		}
	}
}
