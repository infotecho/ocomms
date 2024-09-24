//go:build test

package fakes

import (
	"embed"
	"encoding/json"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	openapi "github.com/twilio/twilio-go/rest/api/v2010"
)

const (
	dirCallEvents     = "testdata/twilio/call-events"
	dirCallRecordings = "testdata/twilio/call-recordings"
)

const (
	// CallConnected is an SID of a call where the client was successfully connected to an agent.
	CallConnected = "CA6cbcb81f7e2c10588972e3988a6f7d6c"

	// CallHangup is an SID of a call where the client called but hung up before even selecting a language.
	CallHangup = "CA24dc06000329d013ce1d01aae4b5cfdf"

	// CallMissedEn is an SID of a call where the client selected English
	// and was redirected to company voicemail but did not leave a message.
	CallMissedEn = "CAf23601206eac4321d55fede98aaf9837"

	// CallMissedFr is an SID of a call where the client selected French
	// and was redirected to company voicemail but did not leave a message.
	CallMissedFr = "CAfbc31ed6ee2733fa8418131fa3056895"

	// CallWithVoicemailEn is an SID of a call where a client selected English
	// and was redirected to company voicemail after the call went to the agent's voicemail.
	CallWithVoicemailEn = "CA498c97dafc33fc3e62dad318e646d578"

	// CallWithVoicemailFr is an SID of a call where a client selected French
	// and was redirected to company voicemail after receiving no answer from the agents.
	CallWithVoicemailFr = "CAd7ee90dd36917a6e0966c8098fabb2db"
)

//go:embed testdata/*
var testData embed.FS

// TwilioClient is a test double of the Twilio API client, with samples of real API responses.
type TwilioClient struct {
	callEvents map[string]openapi.ListCallEventResponse
}

// NewTwilioClient creates an instance of TwilioClient populated with sample data.
func NewTwilioClient() (*TwilioClient, error) {
	client := &TwilioClient{
		callEvents: make(map[string]openapi.ListCallEventResponse),
	}

	if err := populateCallEvents(client.callEvents); err != nil {
		return client, err
	}

	return client, nil
}

// ListCallEvent fakes [github.com/twilio/twilio-go/rest/api/v2010.ApiService.ListCallEvent].
func (tc TwilioClient) ListCallEvent(
	callSID string, _ *openapi.ListCallEventParams,
) ([]openapi.ApiV2010CallEvent, error) {
	return tc.callEvents[callSID].Events, nil
}

func populateCallEvents(callEvents map[string]openapi.ListCallEventResponse) error {
	entries, err := testData.ReadDir(dirCallEvents)
	if err != nil {
		return fmt.Errorf("error reading Twilio call events sample data directory: %w", err)
	}

	for _, entry := range entries {
		bytes, err := testData.ReadFile(filepath.Join(dirCallEvents, entry.Name()))
		if err != nil {
			return fmt.Errorf("error reading Twilio call events sample data file: %w", err)
		}

		var callEvent openapi.ListCallEventResponse
		if err = json.Unmarshal(bytes, &callEvent); err != nil {
			return fmt.Errorf("error decoding Twilio call events sample data file: %w", err)
		}

		callSid := strings.TrimSuffix(entry.Name(), path.Ext(entry.Name()))
		callEvents[callSid] = callEvent
	}

	return nil
}
