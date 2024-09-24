package config

import (
	"flag"
)

//nolint:gochecknoglobals
var loggingFormat = flag.String("logging.format", "", "Logging format (json or text)")

func applyCommandLineFlags(config *Config) {
	if *loggingFormat != "" {
		config.Logging.Format = *loggingFormat
	}
}
