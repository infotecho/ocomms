// Package log implements the application's structured logger, wrapping [slog].
// Generates logs in a format that Google Cloud Run can ingest.
package log

import (
	"log/slog"

	"github.com/infotecho/ocomms/internal/config"
)

// New creates a new [slog.Logger] according to application config.
func New(conf config.Config) *slog.Logger {
	switch conf.Logging.Format {
	case config.LogFormatText:
		// use default logger
	default:
		// JSON is default to ensure that logs in live environments are always formatted correctly
		handler := newCloudLoggingHandler(conf)
		logger := slog.New(handler)
		slog.SetDefault(logger)
	}

	return slog.Default()
}
