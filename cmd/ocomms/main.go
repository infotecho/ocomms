// ocomms-server runs the O-Comms server
package main

import (
	"flag"
	"os"

	"github.com/infotecho/ocomms/internal/app"
	"github.com/infotecho/ocomms/internal/config"
	"github.com/infotecho/ocomms/internal/log"
)

func main() {
	flag.Parse()

	conf, err := config.Load(false)
	logger := log.New(conf)

	if err != nil {
		logger.Error("Failed to load app config", "err", err)
		os.Exit(1)
	}

	srv := app.Server(conf, logger)

	logger.Info("Listening and serving HTTP", "addr", srv.Addr)
	err = srv.ListenAndServe()
	logger.Error("Failed to listen and serve HTTP", "err", err)
	os.Exit(1)
}
