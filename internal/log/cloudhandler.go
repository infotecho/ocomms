package log

import (
	"context"
	"log/slog"
	"os"

	"github.com/infotecho/ocomms/internal/config"
)

type cloudLoggingHandler struct {
	*slog.JSONHandler
}

func (h cloudLoggingHandler) Handle(ctx context.Context, rec slog.Record) error {
	if trace, ok := traceFromContext(ctx); ok {
		rec.AddAttrs(slog.String("logging.googleapis.com/trace", trace))
	}

	return h.JSONHandler.Handle(ctx, rec) //nolint:wrapcheck
}

func newCloudLoggingHandler(conf config.Config) cloudLoggingHandler {
	return cloudLoggingHandler{
		slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			AddSource: true,
			Level:     conf.Logging.Level,
			ReplaceAttr: func(_ []string, attr slog.Attr) slog.Attr {
				switch attr.Key {
				case slog.MessageKey:
					attr.Key = "message"
				case slog.LevelKey:
					attr.Key = "severity"
				case slog.SourceKey:
					attr.Key = "logging.googleapis.com/sourceLocation"
				}

				return attr
			},
		}),
	}
}
