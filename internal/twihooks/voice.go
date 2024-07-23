// Package twihooks handles Twilio webhooks for responding to communication events, such as receiving a phone call.
package twihooks

import (
	"log/slog"
	"net/http"
	"slices"

	"github.com/infotecho/ocomms/internal/config"
	"github.com/infotecho/ocomms/internal/twigen"
)

// VoiceHandler implements Twilio Programmable VoiceHandler hooks.
type VoiceHandler struct {
	Config config.Config
	Logger *slog.Logger
	Twigen *twigen.Voice
}

func (vh VoiceHandler) handler(hookHandler func(*http.Request) string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")

		twiml := hookHandler(r)

		_, err := w.Write([]byte(twiml))
		if err != nil {
			vh.Logger.ErrorContext(r.Context(), "Failed to write response", "err", err)
		}
	}
}

func (vh VoiceHandler) parseForm(r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		vh.Logger.ErrorContext(r.Context(), "Failed to parse Twilio hook HTML form", "err", err)
	}
}

// Inbound handles inbound calls.
func (vh VoiceHandler) Inbound(actionDialOut string, actionConnectAgent string) http.HandlerFunc {
	return vh.handler(func(r *http.Request) string {
		agentDIDs := vh.Config.Twilio.AgentDIDs
		from := r.Header.Get("From")

		if slices.Contains(agentDIDs, from) {
			return vh.Twigen.GatherOutboundNumber(r.Context(), actionDialOut)
		}

		return vh.Twigen.GatherLanguage(r.Context(), actionConnectAgent, true)
	})
}

// DialOut dials out from the company to a gathered phone number.
func (vh VoiceHandler) DialOut() http.HandlerFunc {
	return vh.handler(func(r *http.Request) string {
		vh.parseForm(r)

		digits := r.Form.Get("Digits")

		return vh.Twigen.DialOut(r.Context(), digits)
	})
}

// ConnectAgent connects an incoming caller to an agent.
func (vh VoiceHandler) ConnectAgent(actionConnectAgent string) http.HandlerFunc {
	return vh.handler(func(r *http.Request) string {
		vh.parseForm(r)

		switch r.Form.Get("Digits") {
		case "1":
			return vh.Twigen.DialAgent(r.Context(), "en")
		case "2":
			return vh.Twigen.DialAgent(r.Context(), "fr")
		default:
			return vh.Twigen.GatherLanguage(r.Context(), actionConnectAgent, false)
		}
	})
}
