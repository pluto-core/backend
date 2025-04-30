package bootstrap

import (
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/go-chi/chi/v5"
	middleware "github.com/oapi-codegen/nethttp-middleware"
	"github.com/rs/zerolog"
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

func RunManifestService(configPath string) error {
	cfg := config.Load(configPath)
	log := logger.New(cfg.Logging)

	sqlDB, err := db.NewDB(cfg.Database)
	if err != nil {
		return logFatalWrap(log, err, "failed to connect to database")
	}

	svc := service.New(sqlDB)

	impl := api.NewHandlers(svc, log)

	spec, err := gen.GetSwagger()
	if err != nil {
		return logFatalWrap(log, err, "failed to load swagger spec")
	}

	r := chi.NewRouter()
	r.Use(routerpkg.RequestLogger(log))

	oapiOpts := middleware.Options{
		Options:              openapi3filter.Options{MultiError: true},
		ErrorHandlerWithOpts: errors.ErrorHandlerWithMultiError,
	}
	r.Use(middleware.OapiRequestValidatorWithOptions(spec, &oapiOpts))

	serverOpts := gen.ChiServerOptions{
		BaseRouter: r,
		Middlewares: []gen.MiddlewareFunc{
			routermw.JSONContentType,
		},
		ErrorHandlerFunc: errors.ChiErrorHandler,
	}
	r.Mount("/", gen.HandlerWithOptions(impl, serverOpts))

	srv := routerpkg.NewServer(cfg, log, r)
	log.Info().Msgf("⇒ Starting manifest-service on %d", cfg.Server.Port)
	return srv.ListenAndServe()
}

func logFatalWrap(log *zerolog.Logger, err error, msg string) error {
	log.Fatal().Err(err).Msg(msg)
	return err
}
