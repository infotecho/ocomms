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

func (v *Voice) sayRepl(
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
	say := v.say(ctx, "en", func(m i18n.Messages) string { return m.Voice.GatherOutbound })
	gather := &twiml.VoiceGather{
		Action:        actionDialOut,
		InnerElements: []twiml.Element{say},
		Timeout:       strconv.Itoa(v.Config.Twilio.Timeouts.GatherOutboundNumber),
	}
	verbs := []twiml.Element{gather}

	return v.voice(ctx, verbs)
}

// DialOut generates TwiML to dial out as the company.
func (v Voice) DialOut(ctx context.Context, number string) string {
	dial := &twiml.VoiceDial{
		Number: number,
	}
	verbs := []twiml.Element{dial}

	return v.voice(ctx, verbs)
}

// GatherLanguage generates TwiML to gather a caller's language preference.
func (v Voice) GatherLanguage(ctx context.Context, actionConnectAgent string, intro bool) string {
	sayWelcome := v.say(ctx, "en", func(m i18n.Messages) string { return m.Voice.Welcome })
	sayEn := v.sayRepl(ctx, "en",
		func(m i18n.Messages) string { return m.Voice.LangSelect },
		map[string]string{"digit": "1"},
	)
	sayFr := v.sayRepl(ctx, "fr",
		func(m i18n.Messages) string { return m.Voice.LangSelect },
		map[string]string{"digit": "2"},
	)

	gather := &twiml.VoiceGather{
		Action: actionConnectAgent,
	}
	if intro {
		gather.InnerElements = []twiml.Element{sayWelcome, sayEn, sayFr}
	} else {
		gather.InnerElements = []twiml.Element{sayEn, sayFr}
	}

	verbs := []twiml.Element{gather}

	return v.voice(ctx, verbs)
}

// DialAgent generates TwiML to connect a caller to an agent.
func (v Voice) DialAgent(ctx context.Context, _ string) string {
	return v.voice(ctx, []twiml.Element{})
}
