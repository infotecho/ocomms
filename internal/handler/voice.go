package handler

import (
	"context"
	"log/slog"
	"net/http"
	"slices"

	"github.com/infotecho/ocomms/internal/config"
	"github.com/infotecho/ocomms/internal/mail"
	"github.com/infotecho/ocomms/internal/twigen"
)

const (
	callStatusCompleted = "completed"
	keyRecordVoicemail  = "9"
)

// VoiceHandler implements handlers for Twilio Programmable Voice hooks.
type VoiceHandler struct {
	Config         config.Config
	Emailer        *mail.SendGridMailer
	HandlerFactory *TwimlHandlerFactory
	Logger         *slog.Logger
	Twigen         *twigen.Voice
}

func (h VoiceHandler) inbound(actionDialOut string, actionConnectAgent string) http.HandlerFunc {
	return h.HandlerFactory.handler(func(ctx context.Context, _ string, params map[string]string) string {
		if slices.Contains(h.Config.Twilio.AgentDIDs, params["From"]) {
			return h.Twigen.GatherOutboundNumber(ctx, actionDialOut)
		}

		return h.Twigen.GatherLanguage(ctx, actionConnectAgent, true)
	})
}

// dialOut dials out from the company to a gathered phone number.
func (h VoiceHandler) dialOut() http.HandlerFunc {
	return h.HandlerFactory.handler(func(ctx context.Context, _ string, params map[string]string) string {
		digits := params["Digits"]

		return h.Twigen.DialOut(ctx, digits)
	})
}

// connectAgent connects an incoming caller to an agent.
func (h VoiceHandler) connectAgent(
	actionConnectAgent string,
	actionAcceptCall string,
	actionEndCall string,
) http.HandlerFunc {
	return h.HandlerFactory.handler(func(ctx context.Context, _ string, params map[string]string) string {
		callerID := params["To"]
		digits := params["Digits"]

		switch digits {
		case "1":
			return h.Twigen.DialAgent(ctx, actionAcceptCall, actionEndCall, callerID, "en")
		case "2":
			return h.Twigen.DialAgent(ctx, actionAcceptCall, actionEndCall, callerID, "fr")
		default:
			return h.Twigen.GatherLanguage(ctx, actionConnectAgent, false)
		}
	})
}

// acceptCall prompts an agent to press a key to accept the call,
// to distinguish from their personal voicemail answering the call.
func (h VoiceHandler) acceptCall(actionConfirmConnected string) http.HandlerFunc {
	return h.HandlerFactory.handler(func(ctx context.Context, lang string, _ map[string]string) string {
		return h.Twigen.GatherAccept(ctx, actionConfirmConnected, lang)
	})
}

// confirmConnected confirms to the agent that they were connected to the call after accepting it.
func (h VoiceHandler) confirmConnected() http.HandlerFunc {
	return h.HandlerFactory.handler(func(ctx context.Context, lang string, _ map[string]string) string {
		return h.Twigen.SayConnected(ctx, lang)
	})
}

// endCall handles the end of an inbound call, whether successful (agent picks up)
// or unsuccessful (busy tone or call goes to agent voicemail).
func (h VoiceHandler) endCall(actionStartRecording string) http.HandlerFunc {
	return h.HandlerFactory.handler(func(ctx context.Context, lang string, params map[string]string) string {
		callStatus := params["DialCallStatus"]
		callDuration := params["DialCallDuration"]

		switch {
		case callStatus == "busy",
			callStatus == "no-answer",
			// indicates call went to agent's voicemail - no key pressed to accept call
			callStatus == callStatusCompleted && callDuration == "":
			return h.Twigen.GatherVoicemailStart(ctx, actionStartRecording, keyRecordVoicemail, lang)
		case callStatus == callStatusCompleted:
			return h.Twigen.Noop(ctx)
		default:
			h.Logger.ErrorContext(ctx, "Unexpected DialCallStatus: "+callStatus)
			return h.Twigen.Noop(ctx)
		}
	})
}

// startVoicemail handles a key press after a caller was invited to press 9 to leave a message.
func (h VoiceHandler) startVoicemail(
	actionStartVoicemail string,
	actionEndVoicemail string,
) http.HandlerFunc {
	return h.HandlerFactory.handler(func(ctx context.Context, lang string, params map[string]string) string {
		digits := params["Digits"]

		if digits != keyRecordVoicemail {
			return h.Twigen.GatherVoicemailStart(ctx, actionStartVoicemail, keyRecordVoicemail, lang)
		}

		return h.Twigen.RecordVoicemail(
			ctx,
			actionEndVoicemail,
			keyRecordVoicemail,
			lang,
			false,
		)
	})
}

// endVoicemail handles the end of a voicemail recording
// either due to a keypress (rerecord) or caller hangup (end recording).
func (h VoiceHandler) endVoicemail(actionEndVoicemail string) http.HandlerFunc {
	return h.HandlerFactory.handler(func(ctx context.Context, lang string, params map[string]string) string {
		digits := params["Digits"]

		if digits == "hangup" {
			from := params["From"]
			recordingSID := params["RecordingSid"]
			h.Emailer.Voicemail(ctx, lang, from, recordingSID)
			return h.Twigen.Noop(ctx)
		}

		return h.Twigen.RecordVoicemail(
			ctx,
			actionEndVoicemail,
			keyRecordVoicemail,
			lang,
			true,
		)
	})
}
