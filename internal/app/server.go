package app

import (
	"log/slog"
	"net/http"

	"github.com/infotecho/ocomms/internal/config"
	"github.com/infotecho/ocomms/internal/log"
)

type serverFactory struct {
	Config     config.Config
	Logger     *slog.Logger
	MuxFactory *muxFactory
}

func (sf serverFactory) Server() http.Server {
	mux := sf.MuxFactory.Mux()
	handler := appyMilddleware(mux)

	return http.Server{
		Addr:              ":" + sf.Config.Server.Port,
		Handler:           handler,
		ReadHeaderTimeout: sf.Config.Server.Timeouts.ReadHeaderTimeout,
		ReadTimeout:       sf.Config.Server.Timeouts.ReadTimeout,
		WriteTimeout:      sf.Config.Server.Timeouts.WriteTimeout,
		IdleTimeout:       sf.Config.Server.Timeouts.IdleTimeout,
	}
}

func appyMilddleware(h http.Handler) http.Handler {
	return log.Middleware(h)
}
