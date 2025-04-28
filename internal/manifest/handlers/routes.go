package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"pluto-backend/internal/manifest/service"
)

func RegisterRoutes(r chi.Router, logger *zerolog.Logger, svc *service.Service) {
	r.Route("/manifests", func(r chi.Router) {
		r.Get("/", listManifestsHandler(logger, svc))
	})
}
