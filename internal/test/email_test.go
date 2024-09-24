package test_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/infotecho/ocomms/internal/config"
	"github.com/infotecho/ocomms/internal/fakes"
	"github.com/infotecho/ocomms/internal/handler"
	"github.com/infotecho/ocomms/internal/i18n"
	"github.com/infotecho/ocomms/internal/mail"
	"github.com/infotecho/ocomms/internal/twigen"
	"github.com/infotecho/ocomms/internal/twilio"
)

const (
	phoneNumber = "+17052223434" // An arbitrary DID
)

func setup(t *testing.T, sgc *fakes.SendGridClient) *handler.Voice {
	t.Helper()

	logger := slog.Default()

	config, err := config.Load(true)
	if err != nil {
		t.Fatalf("Failed to load app config: %v", err)
	}

	i18n, err := i18n.NewMessageProvider(logger, config)
	if err != nil {
		t.Fatalf("Failed to initialize SUT dependency: %v", err)
	}

	twilioClient, err := fakes.NewTwilioClient()
	if err != nil {
		t.Fatalf("Failed to initialize SUT dependency: %v", err)
	}

	return &handler.Voice{
		Config: config,
		Emailer: &mail.SendGridMailer{
			Config:         config,
			I18n:           i18n,
			Logger:         logger,
			SendGridClient: sgc,
		},
		Logger: slog.Default(),
		Twigen: &twigen.Voice{}, //nolint:exhaustruct
		Twilio: &twilio.API{
			Client: twilioClient,
			Logger: slog.Default(),
		},
	}
}

func updateGoldenEmail(t *testing.T, filePath string, got []byte) {
	t.Helper()

	file, err := os.Create(filepath.Clean(filePath))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err = file.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	t.Logf("Rewriting %s", filePath)
	if err = os.WriteFile(filePath, got, 0o600); err != nil {
		t.Fatal(err)
	}
}

var goldenEmailTests = []struct {
	name      string
	callSID   string
	emailSent bool
}{
	{
		name:      "call-connected",
		callSID:   fakes.CallConnected,
		emailSent: false,
	},
	{
		name:      "missed-call-en",
		callSID:   fakes.CallMissedEn,
		emailSent: true,
	},
	{
		name:      "missed-call-fr",
		callSID:   fakes.CallMissedFr,
		emailSent: true,
	},
	{
		name:      "no-lang-select",
		callSID:   fakes.CallHangup,
		emailSent: false,
	},
	{
		name:      "voicemail-en",
		callSID:   fakes.CallWithVoicemailEn,
		emailSent: true,
	},
	{
		name:      "voicemail-fr",
		callSID:   fakes.CallWithVoicemailFr,
		emailSent: true,
	},
}

func TestGoldenEmails(t *testing.T) {
	t.Parallel()

	for _, test := range goldenEmailTests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			sgFake := &fakes.SendGridClient{}
			sut := setup(t, sgFake)

			w := httptest.NewRecorder()
			r := &http.Request{}
			r.Form = url.Values{
				"CallSid":    []string{test.callSID},
				"CallStatus": []string{"completed"},
				"Direction":  []string{"inbound"},
				"From":       []string{phoneNumber},
			}
			sut.StatusCallback().ServeHTTP(w, r)

			sentEmails := sgFake.SentEmails()
			if test.emailSent && len(sentEmails) != 1 {
				t.Fatalf("Expected 1 sent email but got: %d", len(sentEmails))
			}
			if !test.emailSent && len(sentEmails) != 0 {
				t.Fatalf("Expected 0 sent emails but got: %d", len(sentEmails))
			}
			if len(sentEmails) == 0 {
				return
			}

			filePath := filepath.Join("testdata", "email", test.name+".golden.eml")
			got := sentEmails[0]

			if *update {
				updateGoldenEmail(t, filePath, got)
				return
			}

			want, err := os.ReadFile(filepath.Clean(filePath))
			if err != nil {
				t.Errorf("Error reading golden file: %v", err)
			}

			if diff := cmp.Diff(want, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}
