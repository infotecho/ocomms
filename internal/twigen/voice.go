package twigen

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/infotecho/ocomms/internal/config"
	"github.com/twilio/twilio-go/twiml"
)

// Voice generates TwiML for Programmable Voice.
type Voice struct {
	Config config.Config
	Logger *slog.Logger
}

func (v Voice) voice(ctx context.Context, verbs []twiml.Element) string {
	res, err := twiml.Voice(verbs)
	if err != nil {
		v.Logger.ErrorContext(ctx, "Error generating TwiML response", "err", err)
	}

	return res
}

func (v *Voice) say(lang string, msg string) *twiml.VoiceSay {
	return &twiml.VoiceSay{
		Message: msg,
		Voice:   v.Config.Twilio.Voice[lang],
	}
}

// GatherOutboundNumber generates TwiML gather a phone number to place an outbound call.
func (v Voice) GatherOutboundNumber(ctx context.Context, actionDialOut string) string {
	say := v.say("en", "Enter the number you wish to call, then press pound.")
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
	sayWelcome := v.say("en", "Welcome to InfoTech Ottawa.")
	sayEn := v.say("en", "For service in English, press 1.")
	sayFr := v.say("fr", "Pour le service en français, appuyer sur le 2.")

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

// DialAgent genreates TwiML to connect a caller to an agent
func (v Voice) DialAgent(ctx context.Context, lang string) string {
	return ""
}
