// Package twihooks handles Twilio webhooks for responding to communication events, such as receiving a phone call.
package twihooks

import (
	"log/slog"
	"net/http"
	"slices"

	"github.com/infotecho/ocomms/internal/config"
	"github.com/infotecho/ocomms/internal/twigen"
)

const (
	keyRecordVoicemail = "9"
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
func (vh VoiceHandler) DialOut(callbackRecordingStatus string) http.HandlerFunc {
	return vh.handler(func(r *http.Request) string {
		vh.parseForm(r)

		digits := r.Form.Get("Digits")

		return vh.Twigen.DialOut(r.Context(), callbackRecordingStatus, digits)
	})
}

// ConnectAgent connects an incoming caller to an agent.
func (vh VoiceHandler) ConnectAgent(
	callbackRecordingStatus string,
	actionConnectAgent string,
	actionAcceptCall string,
	actionEndCall string,
) http.HandlerFunc {
	return vh.handler(func(r *http.Request) string {
		vh.parseForm(r)

		callerID := r.Form.Get("To")
		digits := r.Form.Get("Digits")

		switch digits {
		case "1":
			return vh.Twigen.DialAgent(r.Context(), callbackRecordingStatus, actionAcceptCall, actionEndCall, callerID, "en")
		case "2":
			return vh.Twigen.DialAgent(r.Context(), callbackRecordingStatus, actionAcceptCall, actionEndCall, callerID, "fr")
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
			// indicates call went to agent's voicemail - no key pressed to accept call
			callStatus == "completed" && callDuration == "":
			return vh.Twigen.GatherVoicemailStart(r.Context(), actionStartRecording, keyRecordVoicemail, vh.lang(r))
		case callStatus == "completed":
			return vh.Twigen.Noop(r.Context())
		default:
			vh.Logger.ErrorContext(r.Context(), "Unexpected DialCallStatus: "+callStatus)
			return vh.Twigen.Noop(r.Context())
		}
	})
}

// StartVoicemail handles a key press after a caller was invited to press 9 to leave a message.
func (vh VoiceHandler) StartVoicemail(
	callbackRecordingStatus string,
	actionStartVoicemail string,
	actionEndVoicemail string,
) http.HandlerFunc {
	return vh.handler(func(r *http.Request) string {
		vh.parseForm(r)

		digits := r.Form.Get("Digits")

		if digits != keyRecordVoicemail {
			return vh.Twigen.GatherVoicemailStart(r.Context(), actionStartVoicemail, keyRecordVoicemail, vh.lang(r))
		}

		return vh.Twigen.RecordVoicemail(
			r.Context(),
			callbackRecordingStatus,
			actionEndVoicemail,
			keyRecordVoicemail,
			vh.lang(r),
			false,
		)
	})
}

// EndVoicemail handles the end of a voicemail recording
// either due to a keypress (rerecord) or caller hangup (end recording).
func (vh VoiceHandler) EndVoicemail(callbackRecordingStatus string, actionEndVoicemail string) http.HandlerFunc {
	return vh.handler(func(r *http.Request) string {
		vh.parseForm(r)

		digits := r.Form.Get("Digits")

		if digits == "hangup" {
			return vh.Twigen.Noop(r.Context())
		}

		return vh.Twigen.RecordVoicemail(
			r.Context(),
			callbackRecordingStatus,
			actionEndVoicemail,
			keyRecordVoicemail,
			vh.lang(r),
			true,
		)
	})
}
