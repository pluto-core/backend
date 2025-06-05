// File: internal/auth/service.go
package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	_ "net/http"
	"time"

	"github.com/google/uuid"
	"pluto-backend/internal/auth/api/gen"
	"pluto-backend/internal/auth/repository"
)

// Предопределённые ошибки, чтобы handlers могли распознать
var (
	ErrSessionNotFound = errors.New("session not found or already revoked")
	ErrInternal        = errors.New("internal server error")
)

// Service реализует бизнес-логику auth-сервиса.
type Service struct {
	repo   repository.Querier
	db     *sql.DB
	signer Signer
}

// New конструирует новый Service.
func New(db *sql.DB, signer Signer) *Service {
	raw := repository.New(db)
	return &Service{
		repo:   raw,
		db:     db,
		signer: signer,
	}
}

// AppLogin создаёт или возвращает JWT по фингерпринту.
func (s *Service) AppLogin(ctx context.Context, request gen.AppLoginRequest) (gen.AppLoginResponse, error) {
	// Собираем JSON-фингерпринт
	fingerprintMap := map[string]interface{}{
		"device_id":   request.DeviceId,
		"os":          request.Os,
		"app_version": request.AppVersion,
	}
	if request.Additional != nil {
		fingerprintMap["additional"] = request.Additional
	}
	fpBytes, err := json.Marshal(fingerprintMap)
	if err != nil {
		return gen.AppLoginResponse{}, ErrInternal
	}

	// Ищем активную сессию по device_id
	var existingID uuid.UUID
	var existingJTI uuid.UUID
	var existingExp time.Time

	row, err := s.repo.GetActiveAppSessionByDeviceID(ctx, request.DeviceId)
	if err == nil {
		existingID = row.ID
		existingJTI = row.JwtID
		existingExp = row.ExpiresAt

		// Если осталось больше 5 минут — возвращаем старый токен
		if time.Until(existingExp) > 5*time.Minute {
			tokenString, err := s.signer.Sign(fpBytes, existingExp, existingJTI.String())
			if err != nil {
				return gen.AppLoginResponse{}, ErrInternal
			}
			return gen.AppLoginResponse{
				AccessToken: tokenString,
				ExpiresIn:   int32(time.Until(existingExp).Seconds()),
				SessionId:   existingID,
			}, nil
		}

		// Иначе отзываем старую сессию
		if err := s.repo.RevokeAppSession(ctx, repository.RevokeAppSessionParams{
			ID:    existingID,
			JwtID: existingJTI,
		}); err != nil {
			return gen.AppLoginResponse{}, ErrInternal
		}
	}

	// Создаём новую сессию
	newSessionID := uuid.New()
	newJTI := uuid.New()
	expiresAt := time.Now().Add(1 * time.Hour)

	if err := s.repo.CreateAppSession(ctx, repository.CreateAppSessionParams{
		ID:          newSessionID,
		Fingerprint: fpBytes,
		ExpiresAt:   expiresAt,
		JwtID:       newJTI,
	}); err != nil {
		return gen.AppLoginResponse{}, ErrInternal
	}

	tokenString, err := s.signer.Sign(fpBytes, expiresAt, newJTI.String())
	if err != nil {
		return gen.AppLoginResponse{}, ErrInternal
	}

	return gen.AppLoginResponse{
		AccessToken: tokenString,
		ExpiresIn:   3600,
		SessionId:   newSessionID,
	}, nil
}

//// AppLogout отзывает токен.
//func (s *Service) AppLogout(ctx context.Context, request gen.AppLogoutRequest) error {
//	// Преобразуем request.SessionID и request.Jti (строки) в uuid.UUID
//	sessionUUID, err := uuid.Parse(request.SessionId)
//	if err != nil {
//		return ErrSessionNotFound
//	}
//	jtiUUID, err := uuid.Parse(request.Jti)
//	if err != nil {
//		return ErrSessionNotFound
//	}
//
//	// Пытаемся пометить сессию revoked = true
//	if err := s.repo.RevokeAppSession(ctx, repository.RevokeAppSessionParams{
//		ID:    sessionUUID,
//		JwtID: request.SessionId,
//	}); err != nil {
//		return ErrInternal
//	}
//	return nil
//}

// GetPublicKey возвращает PEM-encoded public key.
func (s *Service) GetPublicKey(ctx context.Context) (gen.PublicKeyResponse, error) {
	pubKey, err := s.signer.GetPublicKey()
	if err != nil || pubKey == "" {
		return gen.PublicKeyResponse{}, ErrInternal
	}
	return gen.PublicKeyResponse{PublicKey: pubKey}, nil
}

// Health проверяет работоспособность сервиса.
func (s *Service) Health(ctx context.Context) (gen.HealthResponse, error) {
	return gen.HealthResponse{Status: "ok"}, nil
}
