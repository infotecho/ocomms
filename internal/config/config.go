// Package config is responsible for reading application configuration.
package config

import (
	"bytes"
	_ "embed"
	"fmt"

	"github.com/spf13/viper"
)

//go:embed config.yaml
var configFile []byte

// Load reads app config from config.yaml and unmarshals it to a Config struct.
func Load() (Config, error) {
	viper.SetConfigType("yaml")

	err := viper.ReadConfig(bytes.NewBuffer(configFile))
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config.yaml: %w", err)
	}

	var config Config

	err = viper.Unmarshal(&config)
	if err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal config.yaml: %w", err)
	}

	return config, nil
}
