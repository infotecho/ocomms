// Package log implements the application's structured logger, wrapping [slog].
// Generates logs in a format that Google Cloud Run can ingest.
package log

import (
	"log/slog"
	"os"

	"github.com/infotecho/ocomms/internal/config"
	"github.com/lmittmann/tint"
)

// New creates a new [slog.Logger] according to application config.
func New(conf config.Config) *slog.Logger {
	var handler slog.Handler

	switch conf.Logging.Format {
	case config.LogFormatText:
		handler = tint.NewHandler(os.Stderr, &tint.Options{
			AddSource:   false,
			Level:       conf.Logging.Level,
			NoColor:     false,
			ReplaceAttr: nil,
			TimeFormat:  "15:04:05.000000",
		})
	// JSON is default to ensure that logs in live environments are always formatted correctly
	default:
		handler = slog.NewJSONHandler(os.Stderr, nil)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return logger
}
