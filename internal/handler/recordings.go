package handler

import (
	"fmt"
	"log/slog"
	"net/http"
)

// RecordingsHandler handles routes under /recordings.
type RecordingsHandler struct {
	Logger *slog.Logger
}

func (h RecordingsHandler) getRecording(w http.ResponseWriter, r *http.Request) {
	recordingSID := r.PathValue("id")
	if recordingSID == "" {
		h.Logger.ErrorContext(r.Context(), "No {id} value in path")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	recordingURL := fmt.Sprintf(
		"https://www.twilio.com/console/voice/api/recordings/recording-logs/%s/download",
		recordingSID,
	)

	http.Redirect(w, r, recordingURL, http.StatusSeeOther)
}
