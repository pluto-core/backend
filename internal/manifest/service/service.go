package service

import (
	"context"

	"pluto-backend/internal/manifest/repository"
)

type Service struct {
	repo repository.Querier
}

func New(repo repository.Querier) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListManifests(ctx context.Context, limit, offset int32, locale string) ([]repository.ListManifestsRow, error) {
	params := repository.ListManifestsParams{Limit: int64(limit), Offset: int64(offset), Column3: locale}
	return s.repo.ListManifests(ctx, params)
}

func (s *Service) SearchManifests(ctx context.Context, search string, locale string) ([]repository.Manifest, error) {
	params := repository.SearchManifestsParams{Search: search, Locale: locale}
	return s.repo.SearchManifests(ctx, params)
}

func (s *Service) SearchManifestsFTS(ctx context.Context, query string, locale string, config string) ([]repository.SearchManifestsFTSRow, error) {
	params := repository.SearchManifestsFTSParams{Query: query, Locale: locale, Config: config}
	return s.repo.SearchManifestsFTS(ctx, params)
}
