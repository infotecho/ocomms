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

func (vh VoiceHandler) lang(r *http.Request) string {
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		vh.Logger.ErrorContext(r.Context(), "No lang query parameter provided. Defaulting to en.")
		lang = "en"
	}
	return lang
}

// Inbound handles inbound calls.
func (vh VoiceHandler) Inbound(actionDialOut string, actionConnectAgent string) http.HandlerFunc {
	return vh.handler(func(r *http.Request) string {
		vh.parseForm(r)

		if slices.Contains(vh.Config.Twilio.AgentDIDs, r.Form.Get("From")) {
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
func (vh VoiceHandler) ConnectAgent(
	actionConnectAgent string,
	actionAcceptCall string,
	actionEndCall string,
) http.HandlerFunc {
	return vh.handler(func(r *http.Request) string {
		vh.parseForm(r)

		switch r.Form.Get("Digits") {
		case "1":
			return vh.Twigen.DialAgent(r.Context(), actionAcceptCall, actionEndCall, "en")
		case "2":
			return vh.Twigen.DialAgent(r.Context(), actionAcceptCall, actionEndCall, "fr")
		default:
			return vh.Twigen.GatherLanguage(r.Context(), actionConnectAgent, false)
		}
	})
}

// AcceptCall prompts an agent to press a key to accept the call,
// to distinguish from their personal voicemail answering the call.
func (vh VoiceHandler) AcceptCall(actionConfirmConnected string) http.HandlerFunc {
	return vh.handler(func(r *http.Request) string {
		return vh.Twigen.GatherAccept(r.Context(), actionConfirmConnected, vh.lang(r))
	})
}

// ConfirmConnected confirms to the agent that they were connected to the call after accepting it.
func (vh VoiceHandler) ConfirmConnected() http.HandlerFunc {
	return vh.handler(func(r *http.Request) string {
		return vh.Twigen.SayConnected(r.Context(), vh.lang(r))
	})
}

// EndCall handles the end of an inbound call, whether successful (agent picks up)
// or unsuccessful (busy tone or call goes to agent voicemail).
func (vh VoiceHandler) EndCall(actionStartRecording string) http.HandlerFunc {
	return vh.handler(func(r *http.Request) string {
		vh.parseForm(r)

		callStatus := r.Form.Get("DialCallStatus")
		callDuration := r.Form.Get("DialCallDuration")
		switch {
		case callStatus == "busy",
			callStatus == "no-answer",
			// indicates call went agent's to voicemail - no key pressed to accept call
			callStatus == "completed" && callDuration == "":
			return vh.Twigen.GatherVoicemail(r.Context(), actionStartRecording, vh.lang(r))
		case callStatus == "completed":
			return ""
		default:
			vh.Logger.ErrorContext(r.Context(), "Unexpected DialCallStatus: "+callStatus)
			return ""
		}
	})
}
