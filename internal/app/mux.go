package app

import (
	"log/slog"
	"net/http"

	"github.com/infotecho/ocomms/internal/config"
	"github.com/infotecho/ocomms/internal/twihooks"
)

const (
	voiceAcceptCall       = "/voice/accept-call"
	voiceConfirmConnected = "/voice/confirm-connected"
	voiceConnectAgent     = "/voice/connect-agent"
	voiceDialOut          = "/voice/dial-out"
	voiceEndCall          = "/voice/end-call"
	voiceInbound          = "/voice/inbound"
	voiceRecordingStatus  = "/voice/recording-status-callback"
	voicemailStart        = "/voice/start-voicemail"
	voicemailEnd          = "/voice/end-voicemail"
)

type muxFactory struct {
	Config       config.Config
	Logger       *slog.Logger
	VoiceHandler *twihooks.VoiceHandler
}

func (mf muxFactory) Mux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc(voiceInbound, mf.VoiceHandler.Inbound(voiceDialOut, voiceConnectAgent))
	mux.HandleFunc(voiceDialOut, mf.VoiceHandler.DialOut(voiceRecordingStatus))
	mux.HandleFunc(voiceConnectAgent,
		mf.VoiceHandler.ConnectAgent(voiceRecordingStatus, voiceConnectAgent, voiceAcceptCall, voiceEndCall),
	)
	mux.HandleFunc(voiceAcceptCall, mf.VoiceHandler.AcceptCall(voiceConfirmConnected))
	mux.HandleFunc(voiceConfirmConnected, mf.VoiceHandler.ConfirmConnected())
	mux.HandleFunc(voiceEndCall, mf.VoiceHandler.EndCall(voicemailStart))
	mux.HandleFunc(voicemailStart,
		mf.VoiceHandler.StartVoicemail(voiceRecordingStatus, voicemailStart, voicemailEnd),
	)
	mux.HandleFunc(voicemailEnd, mf.VoiceHandler.EndVoicemail(voiceRecordingStatus, voicemailEnd))

	return mux
}
