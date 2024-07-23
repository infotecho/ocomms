// Package i18n implements internationalization.
package i18n

import (
	"errors"
	"fmt"
	"log/slog"
	"regexp"

	"github.com/infotecho/ocomms/internal/config"
)

// MessageProvider provides localized messages.
type MessageProvider struct {
	messages map[string]Messages
	logger   *slog.Logger
	config   config.Config
}

// NewMessageProvider loads i18n messages and creates a MessageProvider instance to access them.
// Returns error if unable to load messages.
func NewMessageProvider(logger *slog.Logger, config config.Config) (*MessageProvider, error) {
	messages, err := loadMessages()
	if err != nil {
		return nil, fmt.Errorf("failed to load i18n messages: %w", err)
	}

	return &MessageProvider{
		messages: messages,
		logger:   logger,
		config:   config,
	}, nil
}

// Message returns a localized message given lang and getter,
// which identifies which property to retrieve on the Messages object.
func (mp MessageProvider) Message(lang string, getter func(Messages) string) (string, error) {
	messages, ok := mp.messages[lang]

	if !ok {
		defaultLang := mp.config.I18N.DefaultLang
		err := fmt.Errorf("no messages exist for lang '%s'. Defaulting to lang '%s'", lang, defaultLang) //nolint:err113
		messages = mp.messages[defaultLang]

		return getter(messages), err
	}

	return getter(messages), nil
}

// MessageReplace returns a localized message given lang, getter,
// and replacements for templated values in the i18n string.
func (mp MessageProvider) MessageReplace(
	lang string,
	getter func(Messages) string,
	replacements map[string]string,
) (string, error) {
	errs := []error{}

	msg, err := mp.Message(lang, getter)
	if err != nil {
		errs = append(errs, err)
	}

	re := regexp.MustCompile(`\{[^\}]*\}`)

	msg = re.ReplaceAllStringFunc(msg, func(sub string) string {
		key := sub[1 : len(sub)-1]

		val, ok := replacements[key]
		if !ok {
			errs = append(errs, fmt.Errorf("no replacement provided for '%s' in i18n message", key)) //nolint:err113

			return ""
		}

		return val
	})

	err = nil
	if len(errs) > 0 {
		err = errors.Join(errs...)
	}

	return msg, err
}
