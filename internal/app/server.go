package app

import (
	"log/slog"
	"net/http"

	"github.com/infotecho/ocomms/internal/config"
	"github.com/infotecho/ocomms/internal/log"
)

type serverFactory struct {
	config config.Config
	logger *slog.Logger
}

func (f serverFactory) Server() http.Server {
	config := f.config.Server
	logger := f.logger

	var handler http.Handler
	handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		_, err := w.Write([]byte("Hello world 2"))
		if err != nil {
			logger.Error("ResponseWriter failed", "err", err)
		}

		logger.InfoContext(req.Context(), "Hello world request")
	})
	handler = appyMilddleware(handler)

	return http.Server{
		Addr:              ":" + config.Port,
		Handler:           handler,
		ReadHeaderTimeout: config.Timeouts.ReadHeaderTimeout,
		ReadTimeout:       config.Timeouts.ReadTimeout,
		WriteTimeout:      config.Timeouts.WriteTimeout,
		IdleTimeout:       config.Timeouts.IdleTimeout,
	}
}

func appyMilddleware(h http.Handler) http.Handler {
	return log.Middleware(h)
}
