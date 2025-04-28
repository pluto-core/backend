package logger

import (
	"os"

	"github.com/rs/zerolog"
	"pluto-backend/internal/platform/config"
)

// New возвращает zerolog.Logger, настроенный согласно cfg.Level.
// При parse error – default INFO.
func New(cfg config.LoggingConfig) *zerolog.Logger {
	lvl, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		lvl = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(lvl)
	log := zerolog.
		New(os.Stdout).
		With().
		Timestamp().
		Logger()
	return &log
}
