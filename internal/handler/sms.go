package handler

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/infotecho/ocomms/internal/config"
	"github.com/infotecho/ocomms/internal/i18n"
	"github.com/infotecho/ocomms/internal/mail"
	"github.com/twilio/twilio-go/twiml"
)

// SMSHandler implements handlers for Twilio Programmable Messaging hooks.
type SMSHandler struct {
	Config         config.Config
	I18n           *i18n.MessageProvider
	HandlerFactory *TwimlHandlerFactory
	Logger         *slog.Logger
	Mailer         *mail.SendGridMailer
}

// inbound implements the Twilio incoming message webhook.
func (h SMSHandler) inbound() http.HandlerFunc {
	return h.HandlerFactory.handler(func(ctx context.Context, _ string, params map[string]string) string {
		from := params["From"]
		body := params["Body"]

		h.Mailer.TextMessage(ctx, h.Config.I18N.DefaultLang, from, body)

		replyBodyEn := h.I18n.Message(ctx, "en", func(m i18n.Messages) string { return m.Messaging.Response })
		replyBodyFr := h.I18n.Message(ctx, "fr", func(m i18n.Messages) string { return m.Messaging.Response })
		replyBody := replyBodyEn + "\n" + replyBodyFr

		twiml, err := twiml.Messages([]twiml.Element{
			&twiml.MessagingMessage{
				Body: replyBody,
			},
		})
		if err != nil {
			h.Logger.ErrorContext(ctx, "Error generating TWiML", "err", err)
			return ""
		}

		return twiml
	})
}
