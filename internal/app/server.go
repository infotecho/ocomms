package app

import (
	"log"
	"net/http"

	"github.com/infotecho/ocomms/internal/config"
)

type serverFactory struct {
	Config config.Config
}

func (f serverFactory) Server() http.Server {
	config := f.Config.Server
	//nolint:exhaustruct
	return http.Server{
		Addr: ":" + config.Addr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, err := w.Write([]byte("Hello world 2"))
			if err != nil {
				log.Printf("ResponseWriter failed: %v", err)
			}
		}),
		ReadHeaderTimeout: config.Timeouts.ReadHeaderTimeout,
		ReadTimeout:       config.Timeouts.ReadTimeout,
		WriteTimeout:      config.Timeouts.WriteTimeout,
		IdleTimeout:       config.Timeouts.IdleTimeout,
	}
}
