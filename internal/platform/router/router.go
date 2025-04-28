package router

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"pluto-backend/internal/platform/config"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/quic-go/quic-go/http3"
	"github.com/rs/zerolog"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// NewChiRouter создаёт chi-маршрутизатор с базовыми middleware.
func NewChiRouter(cfg config.Config, log *zerolog.Logger) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(requestLogger(log))
	r.Use(middleware.Recoverer)
	return r
}

// NewServer запускает два сервера на одном порту:
//  1. HTTP/1.1+HTTP/2 cleartext (h2c) через std http.Server
//  2. HTTP/3 через quic-go
//
// TLSConfig здесь нужен для HTTP/3; сертификаты можно положить в путь из конфига
// и nginx при желании пробросит их внутрь контейнера.
func NewServer(
	cfg config.Config,
	log *zerolog.Logger,
	handler http.Handler,
) *http.Server {
	addr := fmt.Sprintf(":%d", cfg.Server.Port)

	// 1) Оборачиваем хендлер в h2c для HTTP/2 cleartext:
	h2s := &http2.Server{}
	h2cHandler := h2c.NewHandler(handler, h2s)

	// 2) Готовим HTTP/3 сервер
	quicServer := &http3.Server{
		Addr:      addr,
		Handler:   handler,
		TLSConfig: &tls.Config{NextProtos: []string{"h3"}},
	}

	// Стартуем HTTP/3 в фоне
	go func() {
		log.Info().Msgf("⇒ HTTP/3 (QUIC) listening on %s", addr)
		// nginx всё равно может пробрасывать файлы сертификатов в /etc/ssl/...
		certFile := cfg.TLS.CertFile
		keyFile := cfg.TLS.KeyFile
		fmt.Println(certFile)
		fmt.Println(keyFile)
		if err := quicServer.ListenAndServeTLS(certFile, keyFile); err != nil {
			log.Fatal().Err(err).Msg("http3 ListenAndServeTLS failed")
		}
	}()

	// 3) Возвращаем обычный http.Server с h2c-handler, который поднимает HTTP/1 и HTTP/2
	srv := &http.Server{
		Addr:    addr,
		Handler: h2cHandler,
	}

	return srv
}

func requestLogger(logger *zerolog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Msg("incoming request")
			next.ServeHTTP(w, r)
		})
	}
}
