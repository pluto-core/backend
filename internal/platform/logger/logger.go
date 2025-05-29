package logger

import (
	"github.com/rs/zerolog"
	"os"
)

// New возвращает zerolog.Logger, настроенный согласно cfg.Level.
// При parse error – default INFO.
func New(level string) *zerolog.Logger {
	lvl, err := zerolog.ParseLevel(level)
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
