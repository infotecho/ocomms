package config

import (
	_ "embed"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"reflect"

	_ "github.com/joho/godotenv/autoload" // load .env
	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v3"
)

//go:embed config.yaml
var configFile []byte

// Load reads app config from config.yaml and unmarshals it to a Config struct.
func Load(allowUndefinedEnvVariables bool) (Config, error) {
	var errs []error
	expanded := os.Expand(string(configFile), func(s string) string {
		envVar := os.Getenv(s)
		if envVar == "" && !allowUndefinedEnvVariables {
			errs = append(errs, fmt.Errorf("undefined environment variable %s", s))
		}
		return envVar
	})
	if len(errs) > 0 {
		return Config{}, errors.Join(errs...)
	}

	var configMap map[string]any
	err := yaml.Unmarshal([]byte(expanded), &configMap)
	if err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal config.yaml: %w", err)
	}

	var config Config
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			stringToLogLevelHookFunc(),
		),
		ErrorUnset: true,
		Result:     &config,
	})
	if err != nil {
		return Config{}, fmt.Errorf("failed to create config decoder: %w", err)
	}

	err = decoder.Decode(configMap)
	if err != nil {
		return Config{}, fmt.Errorf("failed to decode config.yaml: %w", err)
	}

	applyCommandLineFlags(&config)

	return config, nil
}

//nolint:ireturn
func stringToLogLevelHookFunc() mapstructure.DecodeHookFunc {
	return func(
		fromType reflect.Type,
		toType reflect.Type,
		data interface{},
	) (interface{}, error) {
		if fromType.Kind() != reflect.String {
			return data, nil
		}

		var level slog.Level
		if toType != reflect.TypeOf(level) {
			return data, nil
		}

		err := level.UnmarshalText([]byte(data.(string))) //nolint:forcetypeassert

		return level, err //nolint:wrapcheck
	}
}
