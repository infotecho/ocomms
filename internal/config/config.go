// Package config is responsible for reading application configuration.
package config

import (
	"log/slog"
	"time"
)

//go:generate go run ../../cmd/genschema/genschema.go

// Config is the unmarshalled representation of config.yaml.
// [internal/log] configuration.
type Config struct {
	Server struct {
		Port     string `json:"port"`
		Timeouts struct {
			ReadHeaderTimeout time.Duration `jsonschema:"type=string"`
			ReadTimeout       time.Duration `jsonschema:"type=string"`
			WriteTimeout      time.Duration `jsonschema:"type=string"`
			IdleTimeout       time.Duration `jsonschema:"type=string"`
		} `json:"timeouts"`
	} `json:"server"`

	Logging struct {
		Format LogFormat  `json:"format" jsonschema:"type=string,enum=text,enum=json"`
		Level  slog.Level `json:"level"  jsonschema:"type=string,enum=debug,enum=info,enum=warn,enum=error"`
	} `json:"logging"`

	Twilio struct {
		AgentDIDs []string `json:"agentDIDs"`
		Timeouts  struct { // time in seconds
			GatherOutboundNumber int `json:"gatherOutboundNumber"`
		} `json:"timeouts"`
		Voice map[string]string `json:"voice"`
	} `json:"twilio"`
}

// LogFormat determines the output format of logs: JSON or plain text.
type LogFormat = string

const (
	// LogFormatText represents the plain text logging format for local development.
	LogFormatText LogFormat = "text"

	// LogFormatJSON represents the JSON logging format for live environments in Cloud Run.
	LogFormatJSON LogFormat = "json"
)
