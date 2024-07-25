// Package twigen generates TwiML.
package twigen

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/infotecho/ocomms/internal/config"
	"github.com/infotecho/ocomms/internal/i18n"
	"github.com/twilio/twilio-go/twiml"
)

// Voice generates TwiML for Programmable Voice.
type Voice struct {
	Config config.Config
	Logger *slog.Logger
	I18n   *i18n.MessageProvider
}

func (v Voice) voice(ctx context.Context, verbs []twiml.Element) string {
	res, err := twiml.Voice(verbs)
	if err != nil {
		v.Logger.ErrorContext(ctx, "Error generating TwiML response", "err", err)
	}

	return res
}

func (v *Voice) say(ctx context.Context, lang string, getter func(m i18n.Messages) string) *twiml.VoiceSay {
	msg, err := v.I18n.Message(lang, getter)
	if err != nil {
		v.Logger.ErrorContext(ctx, "Error loading i18n message", "err", err)
	}

	return &twiml.VoiceSay{
		Message: msg,
		Voice:   v.Config.Twilio.Voice[lang],
	}
}

func (v *Voice) sayTemplate(
	ctx context.Context,
	lang string,
	getter func(m i18n.Messages) string,
	replaments map[string]string,
) *twiml.VoiceSay {
	msg, err := v.I18n.MessageReplace(lang, getter, replaments)
	if err != nil {
		v.Logger.ErrorContext(ctx, "Error loading i18n message", "err", err)
	}

	return &twiml.VoiceSay{
		Message: msg,
		Voice:   v.Config.Twilio.Voice[lang],
	}
}

// GatherOutboundNumber generates TwiML gather a phone number to place an outbound call.
func (v Voice) GatherOutboundNumber(ctx context.Context, actionDialOut string) string {
	say := &twiml.VoiceSay{
		Message: "Enter the number you wish to call, then press pound.",
		Voice:   v.Config.Twilio.Voice["en"],
	}
	gather := &twiml.VoiceGather{
		Action:        actionDialOut,
		InnerElements: []twiml.Element{say},
		Timeout:       strconv.Itoa(v.Config.Twilio.Timeouts.GatherOutboundNumber),
	}
	return v.voice(ctx, []twiml.Element{gather})
}

// DialOut generates TwiML to dial out as the company.
func (v Voice) DialOut(ctx context.Context, number string) string {
	dial := &twiml.VoiceDial{
		Number: number,
	}
	return v.voice(ctx, []twiml.Element{dial})
}

// GatherLanguage generates TwiML to gather a caller's language preference.
func (v Voice) GatherLanguage(ctx context.Context, actionConnectAgent string, intro bool) string {
	sayWelcome := v.say(ctx, "en", func(m i18n.Messages) string { return m.Voice.Welcome })
	sayEn := v.sayTemplate(ctx, "en",
		func(m i18n.Messages) string { return m.Voice.LangSelect },
		map[string]string{"digit": "1"},
	)
	sayFr := v.sayTemplate(ctx, "fr",
		func(m i18n.Messages) string { return m.Voice.LangSelect },
		map[string]string{"digit": "2"},
	)

	gatherWelcome := &twiml.VoiceGather{
		Action:        actionConnectAgent,
		NumDigits:     "1",
		Timeout:       strconv.Itoa(v.Config.Twilio.Timeouts.GatherLanguage),
		InnerElements: []twiml.Element{sayWelcome, sayEn, sayFr},
	}
	gather := &twiml.VoiceGather{
		Action:        actionConnectAgent,
		NumDigits:     "1",
		Timeout:       strconv.Itoa(v.Config.Twilio.Timeouts.GatherLanguage),
		InnerElements: []twiml.Element{sayEn, sayFr},
	}

	if intro {
		return v.voice(ctx, []twiml.Element{gatherWelcome, gather})
	}
	return v.voice(ctx, []twiml.Element{gather, gather})
}

// DialAgent generates TwiML to connect a caller to an agent.
func (v Voice) DialAgent(ctx context.Context, actionAcceptCall string, actionEndCall string, lang string) string {
	sayHold := v.say(ctx, lang, func(m i18n.Messages) string { return m.Voice.PleaseHold })

	agentDIDs := v.Config.Twilio.AgentDIDs
	numbers := make([]twiml.Element, len(agentDIDs))
	for i, agentDID := range agentDIDs {
		numbers[i] = &twiml.VoiceNumber{
			PhoneNumber: agentDID,
			Url:         actionAcceptCall + "?lang=" + lang,
		}
	}

	dialAgents := &twiml.VoiceDial{
		Action:        actionEndCall + "?lang=" + lang,
		InnerElements: numbers,
		Timeout:       strconv.Itoa(v.Config.Twilio.Timeouts.DialAgents),
	}

	return v.voice(ctx, []twiml.Element{sayHold, dialAgents})
}

// GatherAccept generates TwiML to have an agent confirm acceptance of a call.
func (v Voice) GatherAccept(ctx context.Context, actionConfirmConnected string, lang string) string {
	sayAccept := v.say(ctx, lang, func(m i18n.Messages) string { return m.Voice.AcceptCall })
	gather := &twiml.VoiceGather{
		Action:        actionConfirmConnected,
		NumDigits:     "1",
		Timeout:       strconv.Itoa(v.Config.Twilio.Timeouts.GatherAcceptCall),
		InnerElements: []twiml.Element{sayAccept},
	}
	return v.voice(ctx, []twiml.Element{gather})
}

// SayConnected generates TwiML to confirm to an agent that a call was connected.
func (v Voice) SayConnected(ctx context.Context, lang string) string {
	say := v.say(ctx, lang, func(m i18n.Messages) string { return m.Voice.ConfirmConnected })
	return v.voice(ctx, []twiml.Element{say})
}

// GatherVoicemail generates TwiML to instruct callers to leave a voicemail.
func (v Voice) GatherVoicemail(ctx context.Context, _ string, lang string) string {
	say := v.sayTemplate(ctx, lang,
		func(m i18n.Messages) string { return m.Voice.Voicemail },
		map[string]string{"digit": "9"},
	)
	return v.voice(ctx, []twiml.Element{say})
}
