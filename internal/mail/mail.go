// Package mail sends O-Comms notification emails
package mail

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/infotecho/ocomms/internal/config"
	"github.com/infotecho/ocomms/internal/i18n"
	"github.com/sendgrid/rest"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// SendGridClient is an interface for [github.com/sendgrid/sendgrid-go.Client].
type SendGridClient interface {
	SendWithContext(ctx context.Context, email *mail.SGMailV3) (*rest.Response, error)
}

// SendGridMailer sends emails via SendGrid API.
type SendGridMailer struct {
	Config         config.Config
	I18n           *i18n.MessageProvider
	Logger         *slog.Logger
	SendGridClient SendGridClient
}

// TextMessage notifies agents that a client send a text message.
func (m *SendGridMailer) TextMessage(ctx context.Context, lang string, fromDID string, messageBody string) {
	subject := m.I18n.MessageReplace(
		ctx,
		lang,
		func(m i18n.Messages) string { return m.Email.TextMessage.Subject },
		map[string]string{
			"phoneNumber": fromDID,
		},
	)
	content := m.I18n.MessageReplace(
		ctx,
		lang,
		func(m i18n.Messages) string { return m.Email.TextMessage.Content },
		map[string]string{
			"messageBody": messageBody,
		},
	)

	m.send(ctx, subject, content)
}

// Voicemail notifies agents by email that a client left a voicemail.
func (m *SendGridMailer) Voicemail(ctx context.Context, lang string, fromDID string, recordingSID string) {
	subject := m.I18n.MessageReplace(
		ctx,
		lang,
		func(m i18n.Messages) string { return m.Email.Voicemail.Subject },
		map[string]string{
			"phoneNumber": fromDID,
		},
	)
	content := m.I18n.MessageReplace(
		ctx,
		lang,
		func(m i18n.Messages) string { return m.Email.Voicemail.Content },
		map[string]string{
			"phoneNumber":  fromDID,
			"voicemailURL": "https://ocomms-539601029037.northamerica-northeast1.run.app/recordings/" + recordingSID,
		},
	)

	m.send(ctx, subject, content)
}

func (m *SendGridMailer) send(ctx context.Context, subject string, content string) {
	mailFrom := mail.NewEmail(m.Config.Mail.From.Name, m.Config.Mail.From.Address)
	mailTo := mail.NewEmail(m.Config.Mail.To.Name, m.Config.Mail.To.Address)

	email := mail.NewSingleEmailPlainText(mailFrom, subject, mailTo, content)

	res, err := m.SendGridClient.SendWithContext(ctx, email)
	if err != nil {
		m.Logger.ErrorContext(ctx, "Error sending email", "err", err)
	}
	if res.StatusCode >= http.StatusBadRequest {
		m.Logger.ErrorContext(
			ctx,
			"Error sending email: SendGrid responded with an error code.",
			"statusCode",
			res.StatusCode,
			"response",
			res.Body,
		)
	}
}
