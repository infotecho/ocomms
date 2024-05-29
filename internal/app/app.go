// Package app is the top-level package that provides the main function with a server to run.
package app

import (
	"fmt"
	"net/http"

	"github.com/infotecho/ocomms/internal/config"
)

// Server returns the [http.Server] implementing the O-Comms API.
func Server() (http.Server, error) {
	appConfig, err := config.Load()
	if err != nil {
		//nolint:gosec //gosec doesn't allow zero value of http.Server due to unset timeouts (Slowloris check)
		return http.Server{}, fmt.Errorf("failed to initialize app: %w", err)
	}

	app := wireDependencies(appConfig)

	return app.Server(), nil
}
