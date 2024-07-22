package app

import (
	"log/slog"
	"net/http"

	"github.com/infotecho/ocomms/internal/config"
)

const (
	voiceInbound      = "/voice/inbound"
	voiceDialOut      = "/voice/dial-out"
	voiceConnectAgent = "/voice/connect-agent"
)

type VoiceHandler interface {
	Inbound(actionDialOut string, actionDialAgent string) http.HandlerFunc
	DialOut() http.HandlerFunc
	ConnectAgent(actionConnectAgent string) http.HandlerFunc
}

type muxFactory struct {
	Config       config.Config
	Logger       *slog.Logger
	VoiceHandler VoiceHandler
}

func (mf muxFactory) Mux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc(voiceInbound, mf.VoiceHandler.Inbound(voiceDialOut, voiceConnectAgent))
	mux.HandleFunc(voiceDialOut, mf.VoiceHandler.DialOut())
	mux.HandleFunc(voiceConnectAgent, mf.VoiceHandler.ConnectAgent(voiceConnectAgent))

	return mux
}
