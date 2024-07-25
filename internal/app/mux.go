package app

import (
	"log/slog"
	"net/http"

	"github.com/infotecho/ocomms/internal/config"
	"github.com/infotecho/ocomms/internal/twihooks"
)

const (
	voiceInbound          = "/voice/inbound"
	voiceDialOut          = "/voice/dial-out"
	voiceConnectAgent     = "/voice/connect-agent"
	voiceAcceptCall       = "/voice/accept-call"
	voiceConfirmConnected = "/voice/confirm-connected"
	voiceEndCall          = "/voice/end-call"
	voiceStartRecording   = "/voice/start-recording"
)

type muxFactory struct {
	Config       config.Config
	Logger       *slog.Logger
	VoiceHandler *twihooks.VoiceHandler
}

func (mf muxFactory) Mux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc(voiceInbound, mf.VoiceHandler.Inbound(voiceDialOut, voiceConnectAgent))
	mux.HandleFunc(voiceDialOut, mf.VoiceHandler.DialOut())
	mux.HandleFunc(voiceConnectAgent, mf.VoiceHandler.ConnectAgent(voiceConnectAgent, voiceAcceptCall, voiceEndCall))
	mux.HandleFunc(voiceAcceptCall, mf.VoiceHandler.AcceptCall(voiceConfirmConnected))
	mux.HandleFunc(voiceConfirmConnected, mf.VoiceHandler.ConfirmConnected())
	mux.HandleFunc(voiceEndCall, mf.VoiceHandler.EndCall(voiceStartRecording))

	return mux
}
