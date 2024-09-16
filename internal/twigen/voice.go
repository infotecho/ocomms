// Package twigen generates TwiML.
package twigen

import (
	"context"
	"fmt"
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

	voiceLang, ok := v.Config.Twilio.Languages[lang]
	if !ok {
		v.Logger.ErrorContext(ctx, fmt.Sprintf("No corresponding Twilio language found for language code '%s'", lang))
	}

	return &twiml.VoiceSay{
		Language: voiceLang,
		Message:  msg,
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

	voiceLang, ok := v.Config.Twilio.Languages[lang]
	if !ok {
		v.Logger.ErrorContext(ctx, fmt.Sprintf("No corresponding Twilio language found for language code '%s'", lang))
	}

	return &twiml.VoiceSay{
		Language: voiceLang,
		Message:  msg,
	}
}

// Noop generates an empty TwiML responds that instructs Twilio to do nothing.
func (v Voice) Noop(ctx context.Context) string {
	return v.voice(ctx, []twiml.Element{})
}

// GatherOutboundNumber generates TwiML gather a phone number to place an outbound call.
func (v Voice) GatherOutboundNumber(ctx context.Context, actionDialOut string) string {
	say := &twiml.VoiceSay{
		Language: "en-US",
		Message:  "Enter the number you wish to call, then press pound.",
	}
	gather := &twiml.VoiceGather{
		Action:        actionDialOut,
		InnerElements: []twiml.Element{say},
		Timeout:       strconv.Itoa(v.Config.Twilio.Timeouts.GatherOutboundNumber),
	}
	return v.voice(ctx, []twiml.Element{gather})
}

// DialOut generates TwiML to dial out as the company.
func (v Voice) DialOut(ctx context.Context, callbackRecordingStatus string, number string) string {
	dial := &twiml.VoiceDial{
		Number: number,
	}
	if v.Config.Twilio.RecordOutboundCalls {
		dial.Record = "record-from-answer"
		dial.RecordingStatusCallback = callbackRecordingStatus
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
func (v Voice) DialAgent(
	ctx context.Context,
	callbackRecordingStatus string,
	actionAcceptCall string,
	actionEndCall string,
	callerID string,
	lang string,
) string {
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
		CallerId:      callerID,
		InnerElements: numbers,
		Timeout:       strconv.Itoa(v.Config.Twilio.Timeouts.DialAgents),
	}
	if v.Config.Twilio.RecordInboundCalls {
		dialAgents.Record = "record-from-answer"
		dialAgents.RecordingStatusCallback = callbackRecordingStatus
	}

	return v.voice(ctx, []twiml.Element{sayHold, dialAgents})
}

// GatherAccept generates TwiML to have an agent confirm acceptance of a call.
func (v Voice) GatherAccept(ctx context.Context, actionConfirmConnected string, lang string) string {
	sayAccept := v.say(ctx, lang, func(m i18n.Messages) string { return m.Voice.AcceptCall })
	hangup := &twiml.VoiceHangup{}
	gather := &twiml.VoiceGather{
		Action:        actionConfirmConnected,
		NumDigits:     "1",
		Timeout:       strconv.Itoa(v.Config.Twilio.Timeouts.GatherAcceptCall),
		InnerElements: []twiml.Element{sayAccept},
	}
	return v.voice(ctx, []twiml.Element{gather, hangup})
}

// SayConnected generates TwiML to confirm to an agent that a call was connected.
func (v Voice) SayConnected(ctx context.Context, lang string) string {
	say := v.say(ctx, lang, func(m i18n.Messages) string { return m.Voice.ConfirmConnected })
	return v.voice(ctx, []twiml.Element{say})
}

// GatherVoicemailStart generates TwiML to instruct callers to leave a voicemail.
func (v Voice) GatherVoicemailStart(
	ctx context.Context,
	actionStartVoicemail string,
	recordKey string,
	lang string,
) string {
	say1 := v.sayTemplate(ctx, lang,
		func(m i18n.Messages) string { return m.Voice.Voicemail },
		map[string]string{"digit": recordKey},
	)
	gather1 := &twiml.VoiceGather{
		Action:        actionStartVoicemail + "?lang=" + lang,
		InnerElements: []twiml.Element{say1},
		NumDigits:     "1",
		Timeout:       strconv.Itoa(v.Config.Twilio.Timeouts.GatherStartVoicemail),
	}

	say2 := v.sayTemplate(ctx, lang,
		func(m i18n.Messages) string { return m.Voice.VoicemailRepeat },
		map[string]string{"digit": recordKey},
	)
	gather2 := &twiml.VoiceGather{
		Action:        actionStartVoicemail + "?lang=" + lang,
		InnerElements: []twiml.Element{say2},
		NumDigits:     "1",
		Timeout:       strconv.Itoa(v.Config.Twilio.Timeouts.GatherStartVoicemail),
	}

	return v.voice(ctx, []twiml.Element{gather1, gather2})
}

// RecordVoicemail generates TwiML instructing Twilio to record a caller's voicemail.
func (v Voice) RecordVoicemail(
	ctx context.Context,
	callbackRecordingStatus string,
	actionEndVoicemail string,
	recordKey string,
	lang string,
	rerecord bool,
) string {
	var say *twiml.VoiceSay
	if rerecord {
		say = v.say(ctx, lang, func(m i18n.Messages) string { return m.Voice.ReRecord })
	} else {
		say = v.say(ctx, lang, func(m i18n.Messages) string { return m.Voice.RecordAfterTone })
	}
	record := &twiml.VoiceRecord{
		Action:                  actionEndVoicemail + "?lang=" + lang,
		FinishOnKey:             recordKey,
		RecordingStatusCallback: callbackRecordingStatus,
		Timeout:                 "0",
	}
	return v.voice(ctx, []twiml.Element{say, record})
}
