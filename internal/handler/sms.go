package handler

import (
	"log/slog"
	"net/http"

	"github.com/infotecho/ocomms/internal/config"
	"github.com/infotecho/ocomms/internal/i18n"
	"github.com/infotecho/ocomms/internal/mail"
	"github.com/twilio/twilio-go/twiml"
)

// SMSHandler implements handlers for Twilio Programmable Messaging hooks.
type SMSHandler struct {
	Config config.Config
	I18n   *i18n.MessageProvider
	Logger *slog.Logger
	Mailer *mail.SendGridMailer
}

// Inbound implements the Twilio incoming message webhook.
func (h SMSHandler) Inbound() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			h.Logger.ErrorContext(r.Context(), "Failed to parse Twilio hook HTML form", "err", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		from := r.Form.Get("From")
		body := r.Form.Get("Body")

		h.Mailer.TextMessage(r.Context(), h.Config.I18N.DefaultLang, from, body)

		replyBodyEn := h.I18n.Message(r.Context(), "en", func(m i18n.Messages) string { return m.Messaging.Response })
		replyBodyFr := h.I18n.Message(r.Context(), "fr", func(m i18n.Messages) string { return m.Messaging.Response })
		replyBody := replyBodyEn + "\n" + replyBodyFr

		twiml, err := twiml.Messages([]twiml.Element{
			&twiml.MessagingMessage{
				Body: replyBody,
			},
		})
		if err != nil {
			h.Logger.ErrorContext(r.Context(), "Error generating TWiML", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/xml")
		_, err = w.Write([]byte(twiml))
		if err != nil {
			h.Logger.ErrorContext(r.Context(), "Failed to write response", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}
