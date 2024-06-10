// Package app is the top-level package that provides the main function with a server to run.
package app

import (
	"log/slog"
	"net/http"

	"github.com/infotecho/ocomms/internal/config"
)

// Server returns the [http.Server] implementing the O-Comms API.
func Server(conf config.Config, logger *slog.Logger) http.Server {
	app := wireDependencies(conf, logger)

	return app.Server()
}

func wireDependencies(config config.Config, logger *slog.Logger) serverFactory {
	return serverFactory{
		config: config,
		logger: logger,
	}
}
