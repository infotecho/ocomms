package config

import (
	"flag"
)

func applyCommandLineFlags(config *Config) {
	flag.StringVar(&config.Logging.Format, "logging.format", config.Logging.Format, "Logging format (json or text)")
	flag.Parse()
}
