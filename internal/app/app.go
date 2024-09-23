// Package app is the top-level package that provides the main function with a server to run.
package app

import (
	"log/slog"
	"net/http"

	"github.com/infotecho/ocomms/internal/config"
	"github.com/infotecho/ocomms/internal/handler"
	"github.com/infotecho/ocomms/internal/i18n"
	"github.com/infotecho/ocomms/internal/twigen"
)

// Server returns the [http.Server] implementing the O-Comms API.
func Server(conf config.Config, logger *slog.Logger) http.Server {
	app := WireDependencies(conf, logger)

	return app.Server()
}

// WireDependencies handles dependency injection.
func WireDependencies(config config.Config, logger *slog.Logger) ServerFactory {
	i18n, err := i18n.NewMessageProvider(logger, config)
	if err != nil {
		logger.Error("Failed to load i18n messages", "err", err)
		panic(err)
	}

	return ServerFactory{
		Config: config,
		Logger: logger,
		MuxFactory: &handler.MuxFactory{
			Config: config,
			Logger: logger,
			VoiceHandler: &handler.Voice{
				Config: config,
				Logger: logger,
				Twigen: &twigen.Voice{
					Config: config,
					I18n:   i18n,
					Logger: logger,
				},
			},
		},
	}
}
