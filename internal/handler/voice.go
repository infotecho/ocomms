package handler

import (
	"context"
	"log/slog"
	"net/http"
	"slices"

	"github.com/infotecho/ocomms/internal/config"
	"github.com/infotecho/ocomms/internal/twigen"
	"github.com/infotecho/ocomms/internal/twilio"
)

const (
	callStatusCompleted = "completed"
	keyRecordVoicemail  = "9"
)

type emailer interface {
	MissedCall(ctx context.Context, lang string, from string)
	Voicemail(ctx context.Context, lang string, from string, recordingSID string)
}

// VoiceHandler implements handlers for Twilio Programmable Voice hooks.
type VoiceHandler struct {
	Config  config.Config
	Emailer emailer
	Logger  *slog.Logger
	Twigen  *twigen.Voice
	Twilio  *twilio.API
}

func (h VoiceHandler) handler(hookHandler func(*http.Request) string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			h.Logger.ErrorContext(r.Context(), "Failed to parse Twilio hook HTML form", "err", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/xml")

		twiml := hookHandler(r)

		_, err = w.Write([]byte(twiml))
		if err != nil {
			h.Logger.ErrorContext(r.Context(), "Failed to write response", "err", err)
		}
	}
}

func (h VoiceHandler) lang(r *http.Request) string {
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		h.Logger.ErrorContext(r.Context(), "No lang query parameter provided. Defaulting to en.")
		lang = "en"
	}
	return lang
}

// Inbound handles inbound calls.
func (h VoiceHandler) Inbound(actionDialOut string, actionConnectAgent string) http.HandlerFunc {
	return h.handler(func(r *http.Request) string {
		if slices.Contains(h.Config.Twilio.AgentDIDs, r.Form.Get("From")) {
			return h.Twigen.GatherOutboundNumber(r.Context(), actionDialOut)
		}

		return h.Twigen.GatherLanguage(r.Context(), actionConnectAgent, true)
	})
}

// DialOut dials out from the company to a gathered phone number.
func (h VoiceHandler) DialOut() http.HandlerFunc {
	return h.handler(func(r *http.Request) string {
		digits := r.Form.Get("Digits")

		return h.Twigen.DialOut(r.Context(), digits)
	})
}

// ConnectAgent connects an incoming caller to an agent.
func (h VoiceHandler) ConnectAgent(
	actionConnectAgent string,
	actionAcceptCall string,
	actionEndCall string,
) http.HandlerFunc {
	return h.handler(func(r *http.Request) string {
		callerID := r.Form.Get("To")
		digits := r.Form.Get("Digits")

		switch digits {
		case "1":
			return h.Twigen.DialAgent(r.Context(), actionAcceptCall, actionEndCall, callerID, "en")
		case "2":
			return h.Twigen.DialAgent(r.Context(), actionAcceptCall, actionEndCall, callerID, "fr")
		default:
			return h.Twigen.GatherLanguage(r.Context(), actionConnectAgent, false)
		}
	})
}

// AcceptCall prompts an agent to press a key to accept the call,
// to distinguish from their personal voicemail answering the call.
func (h VoiceHandler) AcceptCall(actionConfirmConnected string) http.HandlerFunc {
	return h.handler(func(r *http.Request) string {
		return h.Twigen.GatherAccept(r.Context(), actionConfirmConnected, h.lang(r))
	})
}

// ConfirmConnected confirms to the agent that they were connected to the call after accepting it.
func (h VoiceHandler) ConfirmConnected() http.HandlerFunc {
	return h.handler(func(r *http.Request) string {
		return h.Twigen.SayConnected(r.Context(), h.lang(r))
	})
}

// EndCall handles the end of an inbound call, whether successful (agent picks up)
// or unsuccessful (busy tone or call goes to agent voicemail).
func (h VoiceHandler) EndCall(actionStartRecording string) http.HandlerFunc {
	return h.handler(func(r *http.Request) string {
		callStatus := r.Form.Get("DialCallStatus")
		callDuration := r.Form.Get("DialCallDuration")

		switch {
		case callStatus == "busy",
			callStatus == "no-answer",
			// indicates call went to agent's voicemail - no key pressed to accept call
			callStatus == callStatusCompleted && callDuration == "":
			return h.Twigen.GatherVoicemailStart(r.Context(), actionStartRecording, keyRecordVoicemail, h.lang(r))
		case callStatus == callStatusCompleted:
			return h.Twigen.Noop(r.Context())
		default:
			h.Logger.ErrorContext(r.Context(), "Unexpected DialCallStatus: "+callStatus)
			return h.Twigen.Noop(r.Context())
		}
	})
}

// StartVoicemail handles a key press after a caller was invited to press 9 to leave a message.
func (h VoiceHandler) StartVoicemail(
	actionStartVoicemail string,
	actionEndVoicemail string,
) http.HandlerFunc {
	return h.handler(func(r *http.Request) string {
		digits := r.Form.Get("Digits")

		if digits != keyRecordVoicemail {
			return h.Twigen.GatherVoicemailStart(r.Context(), actionStartVoicemail, keyRecordVoicemail, h.lang(r))
		}

		return h.Twigen.RecordVoicemail(
			r.Context(),
			actionEndVoicemail,
			keyRecordVoicemail,
			h.lang(r),
			false,
		)
	})
}

// EndVoicemail handles the end of a voicemail recording
// either due to a keypress (rerecord) or caller hangup (end recording).
func (h VoiceHandler) EndVoicemail(actionEndVoicemail string) http.HandlerFunc {
	return h.handler(func(r *http.Request) string {
		digits := r.Form.Get("Digits")

		if digits == "hangup" {
			return h.Twigen.Noop(r.Context())
		}

		return h.Twigen.RecordVoicemail(
			r.Context(),
			actionEndVoicemail,
			keyRecordVoicemail,
			h.lang(r),
			true,
		)
	})
}

// StatusCallback handles call status changes.
func (h VoiceHandler) StatusCallback() http.HandlerFunc {
	return h.handler(func(r *http.Request) string {
		direction := r.Form.Get("Direction")
		from := r.Form.Get("From")
		callSID := r.Form.Get("CallSid")
		callStatus := r.Form.Get("CallStatus")

		if direction != "inbound" || callStatus != callStatusCompleted {
			return h.Twigen.Noop(r.Context())
		}

		metadata := h.Twilio.GetCallMetadata(r.Context(), callSID)

		switch {
		case metadata.VoicemailRecordingID != "":
			h.Emailer.Voicemail(r.Context(), metadata.Lang, from, metadata.VoicemailRecordingID)
		case !metadata.CallConnected && metadata.Lang != "":
			h.Emailer.MissedCall(r.Context(), metadata.Lang, from)
		}

		return h.Twigen.Noop(r.Context())
	})
}
