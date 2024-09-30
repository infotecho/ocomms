package handler

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/twilio/twilio-go/client"
)

// TwimlHandlerFactory creates http.HandleFunc instances for Twilio webhook handlers.
type TwimlHandlerFactory struct {
	Logger           *slog.Logger
	RequestValidator *client.RequestValidator
}

func (f TwimlHandlerFactory) handler(
	twimlHandler func(ctx context.Context, lang string, params map[string]string) string,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			f.Logger.ErrorContext(r.Context(), "Error parsing Twilio hook HTML form", "err", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		params := make(map[string]string, len(r.PostForm))
		for k, v := range r.PostForm {
			params[k] = v[0]
		}

		url := "https://" + r.Host + r.URL.String()
		signature := r.Header.Get("X-Twilio-Signature")
		if !f.RequestValidator.Validate(url, params, signature) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/xml")

		lang := r.URL.Query().Get("lang")
		twiml := twimlHandler(r.Context(), lang, params)

		_, err = w.Write([]byte(twiml))
		if err != nil {
			f.Logger.ErrorContext(r.Context(), "Error writing response", "err", err)
		}
	}
}
