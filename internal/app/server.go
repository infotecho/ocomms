package app

import (
	"log/slog"
	"net/http"

	"github.com/infotecho/ocomms/internal/config"
	"github.com/infotecho/ocomms/internal/handler"
	"github.com/infotecho/ocomms/internal/log"
)

// ServerFactory creates the O-Comms [http.Server] instance.
type ServerFactory struct {
	Config     config.Config
	Logger     *slog.Logger
	MuxFactory *handler.MuxFactory
}

// Server returns an [http.Server] instance for O-Comms.
func (sf ServerFactory) Server() http.Server {
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
