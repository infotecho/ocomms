// Package handler contains the app's HTTP request handling functions
package handler

import (
	"net/http"
)

const (
	voiceAcceptCall       = "/voice/accept-call"
	voiceConfirmConnected = "/voice/confirm-connected"
	voiceConnectAgent     = "/voice/connect-agent"
	voiceDialOut          = "/voice/dial-out"
	voiceEndCall          = "/voice/end-call"
	voicemailStart        = "/voice/start-voicemail"
	voicemailEnd          = "/voice/end-voicemail"
)

// MuxFactory is responsible for creating the app's HTTP request multiplexer.
type MuxFactory struct {
	Recordings *RecordingsHandler
	SMS        *SMSHandler
	Voice      *VoiceHandler
}

// Mux creates the app's HTTP request multiplexer.
func (mf MuxFactory) Mux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("/sms/inbound", mf.SMS.inbound())

	mux.HandleFunc("/voice/inbound", mf.Voice.inbound(voiceDialOut, voiceConnectAgent))
	mux.HandleFunc(voiceDialOut, mf.Voice.dialOut())
	mux.HandleFunc(voiceConnectAgent, mf.Voice.connectAgent(voiceConnectAgent, voiceAcceptCall, voiceEndCall))
	mux.HandleFunc(voiceAcceptCall, mf.Voice.acceptCall(voiceConfirmConnected))
	mux.HandleFunc(voiceConfirmConnected, mf.Voice.confirmConnected())
	mux.HandleFunc(voiceEndCall, mf.Voice.endCall(voicemailStart))
	mux.HandleFunc(voicemailStart, mf.Voice.startVoicemail(voicemailStart, voicemailEnd))
	mux.HandleFunc(voicemailEnd, mf.Voice.endVoicemail(voicemailEnd))

	mux.HandleFunc("/recordings/{id}", mf.Recordings.getRecording)

	return mux
}
