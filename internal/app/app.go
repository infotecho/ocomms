// Package app is the top-level package that provides the main function with a server to run.
package app

import (
	"log/slog"
	"net/http"

	"github.com/infotecho/ocomms/internal/config"
	"github.com/infotecho/ocomms/internal/handler"
	"github.com/infotecho/ocomms/internal/i18n"
	"github.com/infotecho/ocomms/internal/mail"
	"github.com/infotecho/ocomms/internal/twigen"
	"github.com/infotecho/ocomms/internal/twilio"
	"github.com/sendgrid/sendgrid-go"
	twilioApi "github.com/twilio/twilio-go"
)

// Server returns the [http.Server] implementing the O-Comms API.
func Server(conf config.Config, logger *slog.Logger) http.Server {
	app := WireDependencies(conf, logger)

	return app.Server()
}

// WireDependencies handles dependency injection.
func WireDependencies(config config.Config, logger *slog.Logger) ServerFactory {
	i18n, err := i18n.NewMessageProvider(logger, config)
	if err != nil {
		logger.Error("Failed to load i18n messages", "err", err)
		panic(err)
	}

	twilioClient := twilioApi.NewRestClientWithParams(twilioApi.ClientParams{
		AccountSid: config.Twilio.Auth.AccountSID,
		Username:   config.Twilio.Auth.KeySID,
		Password:   config.Twilio.Auth.KeySecret,
	})

	return ServerFactory{
		Config: config,
		Logger: logger,
		MuxFactory: &handler.MuxFactory{
			Recordings: &handler.Recordings{
				Logger: logger,
			},
			Voice: &handler.Voice{
				Config: config,
				Emailer: &mail.SendGridMailer{
					Config:         config,
					I18n:           i18n,
					Logger:         logger,
					SendGridClient: sendgrid.NewSendClient(config.Mail.SendGrid.APIKey),
				},
				Logger: logger,
				Twigen: &twigen.Voice{
					Config: config,
					I18n:   i18n,
					Logger: logger,
				},
				Twilio: &twilio.API{
					Client: twilioClient.Api,
					Logger: logger,
				},
			},
		},
	}
}
