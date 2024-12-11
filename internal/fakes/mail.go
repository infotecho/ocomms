//go:build test

package fakes

import (
	"bytes"
	"context"
	"fmt"

	"github.com/sendgrid/rest"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// SendGridClient is a fake [github.com/sendgrid/sendgrid-go.Client].
type SendGridClient struct {
	sent [][]byte
}

// SendWithContext fakes [github.com/sendgrid/sendgrid-go.Client.SendWithContext].
func (sgc *SendGridClient) SendWithContext(_ context.Context, email *mail.SGMailV3) (*rest.Response, error) {
	var message bytes.Buffer
	fmt.Fprintf(&message, "From: %s <%s>\r\n", email.From.Name, email.From.Address)
	fmt.Fprintf(&message, "To: ")
	for i, to := range email.Personalizations[0].To {
		if i != 0 {
			message.WriteString(", ")
		}
		fmt.Fprintf(&message, "%s <%s>", to.Name, to.Address)
	}
	fmt.Fprint(&message, "\r\n")
	fmt.Fprintf(&message, "Subject: %s \r\n", email.Subject)
	fmt.Fprintf(&message, "\r\n")
	fmt.Fprint(&message, email.Content[0].Value)

	sgc.sent = append(sgc.sent, message.Bytes())

	return &rest.Response{}, nil
}

// SentEmails returns the list of emails that were requested to be "sent" by the fake, in RFC5322 format.
func (sgc *SendGridClient) SentEmails() [][]byte {
	return sgc.sent
}
