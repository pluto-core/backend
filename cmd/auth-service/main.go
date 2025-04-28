package main

import (
	stdhttp "net/http"

	"pluto-backend/internal/platform/config"
	"pluto-backend/internal/platform/logger"
	"pluto-backend/internal/platform/router"
)

func main() {
	// 1) Загрузка конфига (дополнили его полями TLS.CertFile и TLS.KeyFile)
	cfg := config.Load("configs/manifest.yaml")
	log := logger.New(cfg.Logging)

	// 2) Создаём chi-router и регистрируем эндпоинты
	r := router.NewChiRouter(cfg, log)
	r.Get("/health", func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
		w.Write([]byte("auth OK"))
	})

	// 3) Стартуем сразу HTTP/1.1+h2c и HTTP/3
	srv := router.NewServer(cfg, log, r)
	log.Info().Msgf("⇒ Starting manifest-service on %d", cfg.Server.Port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal().Err(err).Msg("server failed")
	}
}
