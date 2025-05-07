package bootstrap

import (
	"encoding/base64"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/go-chi/chi/v5"
	middleware "github.com/oapi-codegen/nethttp-middleware"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/ed25519"
	"pluto-backend/internal/manifest/api"
	"pluto-backend/internal/manifest/api/gen"
	"pluto-backend/internal/manifest/service"
	"pluto-backend/internal/platform/config"
	"pluto-backend/internal/platform/db"
	"pluto-backend/internal/platform/errors"
	"pluto-backend/internal/platform/logger"
	routerpkg "pluto-backend/internal/platform/router"
	routermw "pluto-backend/internal/platform/router/middleware"
)

func RunManifestService() error {
	cfg := config.GetConfig()
	log := logger.New(cfg.Logging)

	sqlDB, err := db.NewDB(cfg.Database)
	if err != nil {
		return logFatalWrap(log, err, "failed to connect to database")
	}

	privKeyBytes, err := base64.StdEncoding.DecodeString(cfg.Signing.PrivateKeyB64)
	if err != nil {
		return logFatalWrap(log, err, "invalid signing.privateKey")
	}
	pubKeyBytes, err := base64.StdEncoding.DecodeString(cfg.Signing.PublicKeyB64)
	if err != nil {
		return logFatalWrap(log, err, "invalid signing.publicKey")
	}
	signer := service.NewEd25519Signer(
		ed25519.PrivateKey(privKeyBytes),
		ed25519.PublicKey(pubKeyBytes),
	)

	_, err = signer.Sign([]byte("test"))
	if err != nil {
		return logFatalWrap(log, err, "failed to sign test data")
	}

	svc := service.New(sqlDB, signer)

	impl := api.NewHandlers(svc, log)

	spec, err := gen.GetSwagger()
	if err != nil {
		return logFatalWrap(log, err, "failed to load swagger spec")
	}

	oapiOpts := middleware.Options{
		Options:              openapi3filter.Options{MultiError: true},
		ErrorHandlerWithOpts: errors.ErrorHandlerWithMultiError,
	}

	r := chi.NewRouter()
	r.Use(routerpkg.RequestLogger(log))
	r.Use(middleware.OapiRequestValidatorWithOptions(spec, &oapiOpts))
	r.Use(routermw.JSONContentType)

	serverOpts := gen.ChiServerOptions{
		BaseRouter:       r,
		ErrorHandlerFunc: errors.ChiErrorHandler,
	}
	r.Mount("/", gen.HandlerWithOptions(impl, serverOpts))

	srv := routerpkg.NewServer(*cfg, log, r)
	log.Info().Msgf("Starting manifest-service on %d", cfg.Server.Port)

	return srv.ListenAndServe()
}

func logFatalWrap(log *zerolog.Logger, err error, msg string) error {
	log.Fatal().Err(err).Msg(msg)
	return err
}
