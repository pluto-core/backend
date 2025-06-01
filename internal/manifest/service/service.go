package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"pluto-backend/internal/manifest/api/gen"

	"pluto-backend/internal/manifest/repository"
)

type Service struct {
	repo    repository.Querier
	rawRepo *repository.Queries
	db      *sql.DB
	signer  Signer
}

func New(db *sql.DB, signer Signer) *Service {
	raw := repository.New(db)
	return &Service{
		repo:    raw,
		rawRepo: raw,
		db:      db,
		signer:  signer,
	}
}

func (s *Service) GetPublicKey(ctx context.Context) (string, error) {
	pubKey, _ := s.signer.GetPublicKey()
	if pubKey == "" {
		return "", errors.New("public key is empty")
	}
	return pubKey, nil

}

func (s *Service) ListManifests(ctx context.Context, limit, offset int32, locale string) ([]repository.ListManifestsRow, error) {
	params := repository.ListManifestsParams{Limit: int64(limit), Offset: int64(offset), Column3: locale}
	return s.repo.ListManifests(ctx, params)
}

func (s *Service) SearchManifests(ctx context.Context, search string, locale string) ([]repository.SearchManifestsRow, error) {
	params := repository.SearchManifestsParams{Search: search, Locale: locale}
	return s.repo.SearchManifests(ctx, params)
}

func (s *Service) SearchManifestsFTS(ctx context.Context, query string, locale string, config string) ([]repository.SearchManifestsFTSRow, error) {
	params := repository.SearchManifestsFTSParams{Query: query, Locale: locale, Config: config}
	return s.repo.SearchManifestsFTS(ctx, params)
}

func (s *Service) GetManifestById(ctx context.Context, id uuid.UUID, locale string) (repository.GetManifestRow, error) {
	params := repository.GetManifestParams{ManifestID: id, Locale: locale}
	return s.repo.GetManifest(ctx, params)
}

func (s *Service) CreateManifest(ctx context.Context, req gen.ManifestCreate) (uuid.UUID, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.Nil, err
	}

	enLocale, ok := req.Localization["en"]
	if !ok {
		return uuid.Nil, errors.New("localization must include 'en'")
	}
	if _, ok := enLocale["title"]; !ok {
		return uuid.Nil, errors.New("localization 'en' must include 'title'")
	}
	if _, ok := enLocale["description"]; !ok {
		return uuid.Nil, errors.New("localization 'en' must include 'description'")
	}

	var locales, keys, values []string
	for locale, entries := range req.Localization {
		for key, value := range entries {
			locales = append(locales, locale)
			keys = append(keys, key)
			values = append(values, value)
		}
	}

	scriptCode, err := extractScriptCode(req.Script)
	if err != nil {
		return uuid.Nil, err
	}

	// формируем каноническое представление
	payload, err := buildCanonicalPayload(id, "1.0.0", req, scriptCode)
	if err != nil {
		return uuid.Nil, err
	}

	// подписываем
	signature, err := s.signer.Sign(payload)
	if err != nil {
		return uuid.Nil, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return uuid.Nil, err
	}
	defer tx.Rollback()

	q := s.rawRepo.WithTx(tx)

	if _, err := q.CreateManifest(ctx, repository.CreateManifestParams{
		ID:          id,
		Version:     "1.0.0",
		Icon:        req.Icon,
		Category:    req.Category,
		Tags:        req.Tags,
		AuthorName:  req.Author.Name,
		AuthorEmail: req.Author.Email,
		Signature:   signature,
	}); err != nil {
		return uuid.Nil, err
	}

	actionsJSON, err := json.Marshal(req.Actions)
	if err != nil {
		return uuid.Nil, err
	}
	// — content (ui, script, actions, permissions)
	if err := q.CreateManifestContent(ctx, repository.CreateManifestContentParams{
		ManifestID:  id,
		Ui:          req.Ui,
		Script:      scriptCode,
		Actions:     actionsJSON,
		Permissions: req.Permissions,
	}); err != nil {
		return uuid.Nil, err
	}

	// — localizations (batch insert)
	if err := q.CreateLocalizations(ctx, repository.CreateLocalizationsParams{
		ManifestID: id,
		Locales:    locales,
		Keys:       keys,
		Values:     values,
	}); err != nil {
		return uuid.Nil, err
	}

	if err := tx.Commit(); err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

func extractScriptCode(raw json.RawMessage) (string, error) {
	var payload struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return "", err
	}
	return payload.Code, nil
}
