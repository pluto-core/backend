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

func (h *Handlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		h.Logger.Error().Err(err).Msg("HealthCheck: failed to encode response")
	}
}

func (h *Handlers) GetPublicKey(w http.ResponseWriter, r *http.Request) {
	pubKey, err := h.Svc.GetPublicKey(r.Context())
	if err != nil {
		h.Logger.Error().Err(err).Msg("GetPublicKey failed")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp := map[string]string{
		"public_key": pubKey,
	}
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.Logger.Error().Err(err).Msg("GetPublicKey: failed to encode response")
	}
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

	out := make([]gen.ManifestMetaLocalized, len(repos))
	for i, m := range repos {
		out[i] = gen.ManifestMetaLocalized{
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
			Localization:  m.Localization,
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

	out := make([]gen.ManifestMetaLocalized, len(repos))
	for i, m := range repos {
		out[i] = gen.ManifestMetaLocalized{
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
			Localization:  m.Localization,
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
		Id: &repo.ID,
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
		scriptJSON, err := json.Marshal(map[string]string{
			"code": repo.Script.String,
		})
		if err != nil {
			h.Logger.Error().Err(err).Msg("failed to marshal script object")
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		scriptRaw = scriptJSON
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
	var req gen.ManifestCreate

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.Logger.Error().Err(err).Msg("createManifest: failed to decode")
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	id, err := h.Svc.CreateManifest(r.Context(), req)
	if err != nil {
		h.Logger.Error().Err(err).Msg("createManifest: service error")
		http.Error(w, "failed to create manifest", http.StatusInternalServerError)
		return
	}

	resp := map[string]string{
		"id": id.String(),
	}
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.Logger.Error().Err(err).Msg("createManifest: failed to encode response")
	}
}
func (h *Handlers) UpdateManifest(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {}
