package main

import (
	"pluto-backend/internal/manifest/handlers"
	"pluto-backend/internal/manifest/repository"
	"pluto-backend/internal/manifest/service"
	"pluto-backend/internal/platform/config"
	"pluto-backend/internal/platform/db"
	"pluto-backend/internal/platform/logger"
	"pluto-backend/internal/platform/router"
)

func main() {
	// 1) Загрузка конфига (дополнили его полями TLS.CertFile и TLS.KeyFile)
	cfg := config.Load("configs/manifest.yaml")
	log := logger.New(cfg.Logging)

	sqlDB, err := db.NewDB(cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}

	repo := repository.New(sqlDB)
	svc := service.New(repo)

	// 2) Создаём chi-router и регистрируем эндпоинты
	r := router.NewChiRouter(cfg, log)
	handlers.RegisterRoutes(r, log, svc)

	// 3) Стартуем сразу HTTP/1.1+h2c и HTTP/3
	srv := router.NewServer(cfg, log, r)
	log.Info().Msgf("⇒ Starting manifest-service on %d", cfg.Server.Port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal().Err(err).Msg("server failed")
	}
}
