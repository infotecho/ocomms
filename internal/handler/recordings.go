package handler

import (
	"log/slog"
	"net/http"
)

// Recordings handles routes under /recordings.
type Recordings struct {
	Logger *slog.Logger
}

func (rec Recordings) getRecording(_ http.ResponseWriter, _ *http.Request) {}
