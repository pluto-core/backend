// File: internal/auth/bootstrap.go
package bootstrap

import (
	"context"
	"encoding/base64"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	middleware "github.com/oapi-codegen/nethttp-middleware"
	"github.com/rs/zerolog"
	"pluto-backend/internal/auth/config"
	"pluto-backend/internal/auth/service"
	"pluto-backend/internal/platform/db"
	"pluto-backend/internal/platform/errors"
	"pluto-backend/internal/platform/logger"
	routerpkg "pluto-backend/internal/platform/router"
	routermw "pluto-backend/internal/platform/router/middleware"

	authapi "pluto-backend/internal/auth/api"
	"pluto-backend/internal/auth/api/gen"
)

func RunAuthService() error {
	cfg := config.GetConfig()

	log := logger.New(cfg.Logging.Level)

	sqlDB, err := db.NewDB(cfg.Database.DSN)
	if err != nil {
		return logFatalWrap(log, err, "failed to connect to database")
	}
	defer sqlDB.Close()

	_, err = base64.StdEncoding.DecodeString(cfg.Signing.PrivateKeyB64)
	if err != nil {
		return logFatalWrap(log, err, "invalid signing.privateKey")
	}
	_, err = base64.StdEncoding.DecodeString(cfg.Signing.PublicKeyB64)
	if err != nil {
		return logFatalWrap(log, err, "invalid signing.publicKey")
	}

	signer, err := service.NewRSASigner(
		cfg.Signing.PrivateKeyB64,
		cfg.Signing.PublicKeyB64,
	)
	if err != nil {
		return logFatalWrap(log, err, "failed to initialize signer")
	}

	svc := service.New(sqlDB, signer)

	handlers := authapi.NewHandlers(svc, log)

	swagger, err := gen.GetSwagger()
	if err != nil {
		return logFatalWrap(log, err, "failed to load OpenAPI spec")
	}

	oapiOpts := middleware.Options{
		Options:              openapi3filter.Options{MultiError: true},
		ErrorHandlerWithOpts: errors.ErrorHandlerWithMultiError,
	}

	r := chi.NewRouter()
	r.Use(routerpkg.RequestLogger(log))
	r.Use(middleware.OapiRequestValidatorWithOptions(swagger, &oapiOpts))
	r.Use(routermw.JSONContentType)
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Recoverer)

	// 11. Монтируем сгенерированные обработчики из OpenAPI
	serverOpts := gen.ChiServerOptions{
		BaseRouter:       r,
		ErrorHandlerFunc: errors.ChiErrorHandler,
	}
	r.Mount("/", gen.HandlerWithOptions(handlers, serverOpts))

	// 12. Запускаем HTTP-сервер
	addr := ":" + os.Getenv("AUTH_SERVER_PORT")
	if addr == ":" {
		addr = ":" + strconv.Itoa(cfg.Server.Port)
	}
	srv := routerpkg.NewServer(cfg.Server.Port, log, r)
	log.Info().Msgf("Starting auth-service on %d", cfg.Server.Port)

	// Graceful shutdown
	idleConnsClosed := make(chan struct{})
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Error().Err(err).Msg("Error during shutdown")
		}
		close(idleConnsClosed)
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("HTTP server failed")
	}

	<-idleConnsClosed
	log.Info().Msg("shutting down")
	return nil
}

func logFatalWrap(log *zerolog.Logger, err error, msg string) error {
	log.Fatal().Err(err).Msg(msg)
	return err
}
