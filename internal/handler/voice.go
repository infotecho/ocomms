package handler

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

// Voice implements handlers for Twilio Programmable Voice hooks.
type Voice struct {
	Config config.Config
	Logger *slog.Logger
	Twigen *twigen.Voice
}

func (v Voice) handler(hookHandler func(*http.Request) string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")

		twiml := hookHandler(r)

		_, err := w.Write([]byte(twiml))
		if err != nil {
			v.Logger.ErrorContext(r.Context(), "Failed to write response", "err", err)
		}
	}
}

func (v Voice) parseForm(r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		v.Logger.ErrorContext(r.Context(), "Failed to parse Twilio hook HTML form", "err", err)
	}
}

func (v Voice) lang(r *http.Request) string {
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		v.Logger.ErrorContext(r.Context(), "No lang query parameter provided. Defaulting to en.")
		lang = "en"
	}
	return lang
}

// Inbound handles inbound calls.
func (v Voice) Inbound(actionDialOut string, actionConnectAgent string) http.HandlerFunc {
	return v.handler(func(r *http.Request) string {
		v.parseForm(r)

		if slices.Contains(v.Config.Twilio.AgentDIDs, r.Form.Get("From")) {
			return v.Twigen.GatherOutboundNumber(r.Context(), actionDialOut)
		}

		return v.Twigen.GatherLanguage(r.Context(), actionConnectAgent, true)
	})
}

// DialOut dials out from the company to a gathered phone number.
func (v Voice) DialOut(callbackRecordingStatus string) http.HandlerFunc {
	return v.handler(func(r *http.Request) string {
		v.parseForm(r)

		digits := r.Form.Get("Digits")

		return v.Twigen.DialOut(r.Context(), callbackRecordingStatus, digits)
	})
}

// ConnectAgent connects an incoming caller to an agent.
func (v Voice) ConnectAgent(
	callbackRecordingStatus string,
	actionConnectAgent string,
	actionAcceptCall string,
	actionEndCall string,
) http.HandlerFunc {
	return v.handler(func(r *http.Request) string {
		v.parseForm(r)

		callerID := r.Form.Get("To")
		digits := r.Form.Get("Digits")

		switch digits {
		case "1":
			return v.Twigen.DialAgent(r.Context(), callbackRecordingStatus, actionAcceptCall, actionEndCall, callerID, "en")
		case "2":
			return v.Twigen.DialAgent(r.Context(), callbackRecordingStatus, actionAcceptCall, actionEndCall, callerID, "fr")
		default:
			return v.Twigen.GatherLanguage(r.Context(), actionConnectAgent, false)
		}
	})
}

// AcceptCall prompts an agent to press a key to accept the call,
// to distinguish from their personal voicemail answering the call.
func (v Voice) AcceptCall(actionConfirmConnected string) http.HandlerFunc {
	return v.handler(func(r *http.Request) string {
		return v.Twigen.GatherAccept(r.Context(), actionConfirmConnected, v.lang(r))
	})
}

// ConfirmConnected confirms to the agent that they were connected to the call after accepting it.
func (v Voice) ConfirmConnected() http.HandlerFunc {
	return v.handler(func(r *http.Request) string {
		return v.Twigen.SayConnected(r.Context(), v.lang(r))
	})
}

// EndCall handles the end of an inbound call, whether successful (agent picks up)
// or unsuccessful (busy tone or call goes to agent voicemail).
func (v Voice) EndCall(actionStartRecording string) http.HandlerFunc {
	return v.handler(func(r *http.Request) string {
		v.parseForm(r)

		callStatus := r.Form.Get("DialCallStatus")
		callDuration := r.Form.Get("DialCallDuration")

		switch {
		case callStatus == "busy",
			callStatus == "no-answer",
			// indicates call went to agent's voicemail - no key pressed to accept call
			callStatus == "completed" && callDuration == "":
			return v.Twigen.GatherVoicemailStart(r.Context(), actionStartRecording, keyRecordVoicemail, v.lang(r))
		case callStatus == "completed":
			return v.Twigen.Noop(r.Context())
		default:
			v.Logger.ErrorContext(r.Context(), "Unexpected DialCallStatus: "+callStatus)
			return v.Twigen.Noop(r.Context())
		}
	})
}

// StartVoicemail handles a key press after a caller was invited to press 9 to leave a message.
func (v Voice) StartVoicemail(
	callbackRecordingStatus string,
	actionStartVoicemail string,
	actionEndVoicemail string,
) http.HandlerFunc {
	return v.handler(func(r *http.Request) string {
		v.parseForm(r)

		digits := r.Form.Get("Digits")

		if digits != keyRecordVoicemail {
			return v.Twigen.GatherVoicemailStart(r.Context(), actionStartVoicemail, keyRecordVoicemail, v.lang(r))
		}

		return v.Twigen.RecordVoicemail(
			r.Context(),
			callbackRecordingStatus,
			actionEndVoicemail,
			keyRecordVoicemail,
			v.lang(r),
			false,
		)
	})
}

// EndVoicemail handles the end of a voicemail recording
// either due to a keypress (rerecord) or caller hangup (end recording).
func (v Voice) EndVoicemail(callbackRecordingStatus string, actionEndVoicemail string) http.HandlerFunc {
	return v.handler(func(r *http.Request) string {
		v.parseForm(r)

		digits := r.Form.Get("Digits")

		if digits == "hangup" {
			return v.Twigen.Noop(r.Context())
		}

		return v.Twigen.RecordVoicemail(
			r.Context(),
			callbackRecordingStatus,
			actionEndVoicemail,
			keyRecordVoicemail,
			v.lang(r),
			true,
		)
	})
}
