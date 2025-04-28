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

func (s *Service) ListManifests(ctx context.Context, limit, offset int32) ([]repository.Manifest, error) {
	params := repository.ListManifestsParams{Limit: int64(limit), Offset: int64(offset)}
	return s.repo.ListManifests(ctx, params)
}
