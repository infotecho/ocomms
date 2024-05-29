// ocomms-server runs the O-Comms server
package main

import (
	"log"

	"github.com/infotecho/ocomms/internal/app"
)

func main() {
	srv, err := app.Server()
	if err != nil {
		log.Fatal(err)
	}

	err = srv.ListenAndServe()
	log.Fatal(err)
}
