package api

import (
	"database/sql"
	"encoding/json"
	openapi_types "github.com/oapi-codegen/runtime/types"
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

func (h *Handlers) ListManifests(
	w http.ResponseWriter,
	r *http.Request,
	params gen.ListManifestsParams,
) {
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

	out := make([]gen.ManifestMeta, len(repos))
	for i, m := range repos {
		out[i] = gen.ManifestMeta{
			Id:       &m.ID,
			Version:  &m.Version,
			Icon:     &m.Icon,
			Category: &m.Category,
			Tags:     &m.Tags,
			Author: gen.Author{
				Email: m.AuthorEmail,
				Name:  m.AuthorName,
			},
			CreatedAt:     &m.CreatedAt,
			MetaCreatedAt: &m.MetaCreatedAt,
			Title:         toStringPtr(m.Title), // sql.NullString
			Description:   toStringPtr(m.Title), // sql.NullString
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

func (h *Handlers) SearchManifests(
	w http.ResponseWriter,
	r *http.Request,
	params gen.SearchManifestsParams,
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

	out := make([]gen.ManifestMeta, len(repos))
	for i, m := range repos {
		out[i] = gen.ManifestMeta{
			Id:       &m.ID,
			Version:  &m.Version,
			Icon:     &m.Icon,
			Category: &m.Category,
			Tags:     &m.Tags,
			Author: gen.Author{
				Email: m.AuthorEmail,
				Name:  m.AuthorName,
			},
			CreatedAt:     &m.CreatedAt,
			MetaCreatedAt: &m.MetaCreatedAt,
			Title:         toStringPtr(m.Title),
			Description:   toStringPtr(m.Description),
		}
	}

	json.NewEncoder(w).Encode(out)
}

func (h *Handlers) GetManifestById(
	w http.ResponseWriter,
	r *http.Request,
	id openapi_types.UUID,
) {
	locale := utils.ParseAcceptLanguageHeader(r.Header.Get("Accept-Language"))

	repo, err := h.Svc.GetManifestById(r.Context(), id, locale)
	if err != nil {
		h.Logger.Error().Err(err).Msg("GetManifestById failed")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	meta := gen.ManifestMeta{
		Id:          &repo.ID,
		Title:       &repo.Title.String,
		Description: &repo.Description.String,
		Author: gen.Author{
			Email: repo.AuthorEmail,
			Name:  repo.AuthorName,
		},
		CreatedAt: &repo.CreatedAt,
		Version:   &repo.Version,
		Icon:      &repo.Icon,
		Category:  &repo.Category,
		Tags:      &repo.Tags,
	}

	var scriptRaw gen.ManifestScript
	if repo.Script.Valid {
		scriptRaw = gen.ManifestScript([]byte(repo.Script.String))
	} else {
		scriptRaw = nil
	}

	out := gen.Manifest{
		Meta:         meta,
		Localization: repo.Localization.RawMessage,
		Ui:           repo.UI.RawMessage,
		Script:       scriptRaw,
		Actions:      repo.Actions.RawMessage,
		Permissions:  repo.Permissions,
		Signature:    &repo.Signature,
	}

	JSON(w, http.StatusOK, out)
}

func (h *Handlers) CreateManifest(w http.ResponseWriter, r *http.Request) {

}
func (h *Handlers) UpdateManifest(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {}
