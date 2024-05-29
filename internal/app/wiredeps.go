package app

import (
	"github.com/infotecho/ocomms/internal/config"
)

func wireDependencies(config config.Config) serverFactory {
	return serverFactory{Config: config}
}
