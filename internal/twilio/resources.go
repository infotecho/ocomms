package twilio

import (
	"errors"

	"github.com/go-viper/mapstructure/v2"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
)

// CallEventRequest represents hook requests made by Twilio to O-Comms,
// corresponding to the [request] field of the Call Event Resource.
// [request]: https://www.twilio.com/docs/voice/api/call-event-resource#request
type CallEventRequest struct {
	URL        string `mapstructure:"url"`
	Parameters struct {
		DialCallStatus   string `mapstructure:"dial_call_status"`
		DialCallDuration string `mapstructure:"dial_call_duration"`
		Digits           string `mapstructure:"digits"`
		RecordingSID     string `mapstructure:"recording_sid"`
	} `mapstructure:"parameters"`
}

func (cer *CallEventRequest) decode(callEvent openapi.ApiV2010CallEvent) error {
	requestMap, ok := (*callEvent.Request).(map[string]any)
	if !ok {
		return errors.New("failed to decode Twilio CallEvent API resource")
	}

	if err := mapstructure.Decode(requestMap, cer); err != nil {
		return errors.New("failed to decode Twilio CallEvent API resource")
	}

	return nil
}
