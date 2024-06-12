package log

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type traceKey struct{}

func contextWithTrace(ctx context.Context, trace string) context.Context {
	return context.WithValue(ctx, traceKey{}, trace)
}

func traceFromContext(ctx context.Context) (string, bool) {
	t, ok := ctx.Value(traceKey{}).(string)

	return t, ok
}

// Middleware applies logging middleware to an [http.Handler] which adds GCP trace information to request context
// so that log entries can be associated to their requests.
func Middleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")

		if projectID != "" {
			traceHeader := req.Header.Get("X-Cloud-Trace-Context")
			traceParts := strings.Split(traceHeader, "/")

			if len(traceParts) > 0 && len(traceParts[0]) > 0 {
				trace := fmt.Sprintf("projects/%s/traces/%s", projectID, traceParts[0])
				ctx := contextWithTrace(req.Context(), trace)
				req = req.WithContext(ctx)
			}
		}

		handler.ServeHTTP(writer, req)
	})
}
