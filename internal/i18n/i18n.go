// Package i18n implements internationalization.
package i18n

import (
	"context"
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
func (mp MessageProvider) Message(ctx context.Context, lang string, getter func(Messages) string) string {
	return mp.MessageReplace(ctx, lang, getter, map[string]string{})
}

// MessageReplace returns a localized message given lang, getter,
// and replacements for templated values in the i18n string.
func (mp MessageProvider) MessageReplace(
	ctx context.Context,
	lang string,
	getter func(Messages) string,
	replacements map[string]string,
) string {
	messages, ok := mp.messages[lang]
	if !ok {
		defaultLang := mp.config.I18N.DefaultLang
		messages = mp.messages[defaultLang]
		mp.logger.ErrorContext(
			ctx,
			fmt.Sprintf("No messages exist for lang '%s'. Defaulting to lang '%s'", lang, defaultLang),
		)
	}

	msg := getter(messages)

	re := regexp.MustCompile(`\{[^\}]*\}`)
	msg = re.ReplaceAllStringFunc(msg, func(sub string) string {
		key := sub[1 : len(sub)-1]
		val, ok := replacements[key]
		if !ok {
			mp.logger.ErrorContext(ctx, fmt.Sprintf("No replacement provided for '%s' in i18n message", key))
			return ""
		}
		return val
	})

	return msg
}
