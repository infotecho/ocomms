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
	Voice      *VoiceHandler
}

// Mux creates the app's HTTP request multiplexer.
func (mf MuxFactory) Mux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/voice/inbound", mf.Voice.Inbound(voiceDialOut, voiceConnectAgent))
	mux.HandleFunc(voiceDialOut, mf.Voice.DialOut())
	mux.HandleFunc(voiceConnectAgent, mf.Voice.ConnectAgent(voiceConnectAgent, voiceAcceptCall, voiceEndCall))
	mux.HandleFunc(voiceAcceptCall, mf.Voice.AcceptCall(voiceConfirmConnected))
	mux.HandleFunc(voiceConfirmConnected, mf.Voice.ConfirmConnected())
	mux.HandleFunc(voiceEndCall, mf.Voice.EndCall(voicemailStart))
	mux.HandleFunc(voicemailStart, mf.Voice.StartVoicemail(voicemailStart, voicemailEnd))
	mux.HandleFunc(voicemailEnd, mf.Voice.EndVoicemail(voicemailEnd))
	mux.HandleFunc("/voice/status-callback", mf.Voice.StatusCallback())

	mux.HandleFunc("/recordings/{id}", mf.Recordings.getRecording)

	return mux
}
