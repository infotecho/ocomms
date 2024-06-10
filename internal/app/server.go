package app

import (
	"log/slog"
	"net/http"

	"github.com/infotecho/ocomms/internal/config"
)

type serverFactory struct {
	config config.Config
	logger *slog.Logger
}

func (f serverFactory) Server() http.Server {
	config := f.config.Server
	logger := f.logger
	//nolint:exhaustruct
	return http.Server{
		Addr: ":" + config.Port,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, err := w.Write([]byte("Hello world 2"))
			if err != nil {
				logger.Error("ResponseWriter failed", "err", err)
			}
		}),
		ReadHeaderTimeout: config.Timeouts.ReadHeaderTimeout,
		ReadTimeout:       config.Timeouts.ReadTimeout,
		WriteTimeout:      config.Timeouts.WriteTimeout,
		IdleTimeout:       config.Timeouts.IdleTimeout,
	}
}
