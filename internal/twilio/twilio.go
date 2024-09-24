// Package twilio interfaces with the Twilio REST API
package twilio

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"

	openapi "github.com/twilio/twilio-go/rest/api/v2010"
)

// CallMetadata contains information about an O-Comms call.
type CallMetadata struct {
	CallConnected        bool
	VoicemailRecordingID string
	Lang                 string
}

// Client is an interface corresponding the twilio.RestClient struct.
type Client interface {
	ListCallEvent(CallSid string, params *openapi.ListCallEventParams) ([]openapi.ApiV2010CallEvent, error)
}

// API interfaces with the Twilio REST API.
type API struct {
	Client Client
	Logger *slog.Logger
}

// GetCallMetadata returns metadata extracted from a Twilio [Call Event Resource]
// [Call Event Resource]: https://www.twilio.com/docs/voice/api/call-event-resource
func (a API) GetCallMetadata(ctx context.Context, callSID string) CallMetadata {
	logger := a.Logger.With("callSID", callSID)

	callMetadata := CallMetadata{
		CallConnected:        false,
		VoicemailRecordingID: "",
		Lang:                 "",
	}

	callEventRequests, err := a.listCallEventRequests(callSID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to fetch call event resources from Twilio API", "err", err)
		return callMetadata
	}

	callMetadata.CallConnected = a.callConnected(callEventRequests)
	callMetadata.Lang = a.lang(ctx, callEventRequests)
	callMetadata.VoicemailRecordingID = a.voicemailRecordingID(callEventRequests)

	return callMetadata
}

func (a API) listCallEventRequests(callSID string) ([]CallEventRequest, error) {
	callEvents, err := a.Client.ListCallEvent(callSID, &openapi.ListCallEventParams{})
	if err != nil {
		return nil, fmt.Errorf("error fetching call event metadata from Twilio API: %w", err)
	}

	callEventRequests := make([]CallEventRequest, len(callEvents))
	for i := range callEventRequests {
		if err = callEventRequests[i].decode(callEvents[i]); err != nil {
			return nil, fmt.Errorf("failed to decode CallEvent resources from Twilio API: %w", err)
		}
	}
	return callEventRequests, nil
}

func (a API) callConnected(callEventRequests []CallEventRequest) bool {
	for _, req := range callEventRequests {
		if req.Parameters.DialCallStatus == "completed" && req.Parameters.DialCallDuration != "" {
			return true
		}
	}
	return false
}

func (a API) voicemailRecordingID(callEventRequests []CallEventRequest) string {
	for _, req := range callEventRequests {
		if req.Parameters.Digits == "hangup" && req.Parameters.RecordingSID != "" {
			return req.Parameters.RecordingSID
		}
	}
	return ""
}

func (a API) lang(ctx context.Context, callEventRequests []CallEventRequest) string {
	lang := ""
	for _, req := range callEventRequests {
		url, err := url.Parse(req.URL)
		if err != nil {
			a.Logger.ErrorContext(ctx, "Unable to parse URL", "url", req.URL, "err", err)
			continue
		}

		urlLang := url.Query().Get("lang")
		if urlLang != "" {
			lang = urlLang
			break
		}
	}

	return lang
}
