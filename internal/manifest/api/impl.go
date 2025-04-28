package api

import (
	"database/sql"
	"encoding/json"
	"github.com/rs/zerolog"
	"net/http"
	"pluto-backend/internal/manifest/api/gen"
	"pluto-backend/internal/manifest/service"
	"pluto-backend/internal/platform/utils"
)

type Handlers struct {
	Svc    *service.Service
	Logger *zerolog.Logger
}

func NewHandlers(svc *service.Service, log *zerolog.Logger) *Handlers {
	return &Handlers{Svc: svc, Logger: log}
}

// ListManifests соответствует операции GET /manifests
func (h *Handlers) ListManifests(
	w http.ResponseWriter,
	r *http.Request,
	params gen.ListManifestsParams,
) {
	// oapi-codegen уже проверил params.Limit и params.Offset по схеме
	limit := int32(100)
	offset := int32(0)
	if params.Limit != nil {
		limit = int32(*params.Limit)
	}
	if params.Offset != nil {
		offset = int32(*params.Offset)
	}

	locale := r.Header.Get("Accept-Language")
	if locale == "" {
		locale = "en"
	}

	repos, err := h.Svc.ListManifests(r.Context(), limit, offset, locale)
	if err != nil {
		h.Logger.Error().Err(err).Msg("ListManifests failed")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// преобразуем репозиторские модели в сгенерированный тип Manifest
	out := make([]gen.ManifestMeta, len(repos))
	for i, m := range repos {
		out[i] = gen.ManifestMeta{
			Id:            m.ID,
			Version:       m.Version,
			Icon:          m.Icon,
			Category:      m.Category,
			Tags:          m.Tags,
			AuthorName:    m.AuthorName,
			AuthorEmail:   m.AuthorEmail,
			CreatedAt:     m.CreatedAt,
			MetaCreatedAt: m.MetaCreatedAt,
			Title:         *toStringPtr(m.Title), // sql.NullString
			Description:   *toStringPtr(m.Title), // sql.NullString
		}
	}

	json.NewEncoder(w).Encode(out)
}

func toStringPtr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

func (h *Handlers) GetManifestsBySearch(
	w http.ResponseWriter,
	r *http.Request,
	params gen.GetManifestsBySearchParams,
) {

	locale := r.Header.Get("Accept-Language")
	var parsedLocale string

	if locale == "" {
		parsedLocale = "en"
	} else {
		var err error
		parsedLocale, err = utils.ParsePrimaryLanguage(locale)
		if err != nil {
			parsedLocale = "en"
		}
	}

	repos, err := h.Svc.SearchManifestsFTS(r.Context(), params.Query, parsedLocale, utils.GetDisplayName(parsedLocale))
	if err != nil {
		h.Logger.Error().Err(err).Msg("GetManifestsSearch failed")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// преобразуем репозиторские модели в сгенерированный тип Manifest
	out := make([]gen.ManifestMeta, len(repos))
	for i, m := range repos {
		out[i] = gen.ManifestMeta{
			Id:            m.ID,
			Version:       m.Version,
			Icon:          m.Icon,
			Category:      m.Category,
			Tags:          m.Tags,
			AuthorName:    m.AuthorName,
			AuthorEmail:   m.AuthorEmail,
			CreatedAt:     m.CreatedAt,
			MetaCreatedAt: m.MetaCreatedAt,
			Title:         *toStringPtr(m.Title),
			Description:   *toStringPtr(m.Description),
		}
	}

	json.NewEncoder(w).Encode(out)
}
