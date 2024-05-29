package config

//go:generate go run ../../cmd/genschema/genschema.go

import "time"

// Config is the unmarshalled representation of config.yaml.
type Config struct {
	Server struct {
		Addr string `json:"addr"`

		// [http.Server] timeout values
		Timeouts struct {
			ReadHeaderTimeout time.Duration `jsonschema:"type=string"`
			ReadTimeout       time.Duration `jsonschema:"type=string"`
			WriteTimeout      time.Duration `jsonschema:"type=string"`
			IdleTimeout       time.Duration `jsonschema:"type=string"`
		} `json:"timeouts"`
	} `json:"server"`
}
