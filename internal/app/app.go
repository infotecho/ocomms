// Package app is the top-level package that provides the main function with a server to run.
package app

import (
	"log/slog"
	"net/http"

	"github.com/infotecho/ocomms/internal/config"
	"github.com/infotecho/ocomms/internal/twigen"
	"github.com/infotecho/ocomms/internal/twihooks"
)

// Server returns the [http.Server] implementing the O-Comms API.
func Server(conf config.Config, logger *slog.Logger) http.Server {
	app := wireDependencies(conf, logger)

	return app.Server()
}

func wireDependencies(config config.Config, logger *slog.Logger) serverFactory {
	return serverFactory{
		Config: config,
		Logger: logger,
		MuxFactory: &muxFactory{
			Config: config,
			Logger: logger,
			VoiceHandler: &twihooks.VoiceHandler{
				Config: config,
				Logger: logger,
				Twigen: &twigen.Voice{
					Config: config,
					Logger: logger,
				},
			},
		},
	}
}
