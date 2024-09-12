// Package config is responsible for reading application configuration.
package config

import (
	"log/slog"
	"time"
)

//go:generate go run ../../cmd/genschema/genschema.go

// Config is the unmarshalled representation of config.yaml.
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

	I18N struct {
		DefaultLang string `json:"defaultLang"`
	} `json:"i18n"`

	Twilio struct {
		AgentDIDs []string `json:"agentDIDs"`
		Auth      struct {
			KeySID    string `json:"keySID"`
			KeySecret string `json:"keySecret"`
		} `json:"auth"`
		Languages           map[string]string `json:"languages"`
		RecordInboundCalls  bool              `json:"recordInboundCalls"`
		RecordOutboundCalls bool              `json:"recordOutboundCalls"`
		Timeouts            struct {          // time in seconds
			DialAgents           int `json:"dialAgents"`
			GatherLanguage       int `json:"gatherLanguage"`
			GatherOutboundNumber int `json:"gatherOutboundNumber"`
			GatherAcceptCall     int `json:"gatherAcceptCall"`
			GatherStartVoicemail int `json:"gatherStartVoicemail"`
		} `json:"timeouts"`
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
