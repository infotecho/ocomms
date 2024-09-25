package handler

import (
	"log/slog"
	"net/http"
)

// RecordingsHandler handles routes under /recordings.
type RecordingsHandler struct {
	Logger *slog.Logger
}

func (h RecordingsHandler) getRecording(_ http.ResponseWriter, _ *http.Request) {}
