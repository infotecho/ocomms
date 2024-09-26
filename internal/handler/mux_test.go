package handler_test

import (
	"context"
	"encoding/xml"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/infotecho/ocomms/internal/config"
	"github.com/infotecho/ocomms/internal/fakes"
	"github.com/infotecho/ocomms/internal/handler"
	"github.com/infotecho/ocomms/internal/i18n"
	"github.com/infotecho/ocomms/internal/mail"
	"github.com/infotecho/ocomms/internal/twigen"
	"github.com/infotecho/ocomms/internal/twilio"
	"golang.org/x/tools/txtar"
)

const (
	clientDID  = "+17052223434" // An arbitrary DID"
	agentDID   = "+16138160938"
	companyDID = "+16137775650"
)

var update = flag.Bool("update", false, "rewrite testdata golden files")

type XMLElement struct {
	XMLName  xml.Name     `xml:""`
	Attrs    []xml.Attr   `xml:",any,attr"`
	Content  string       `xml:",innerxml"`
	Elements []XMLElement `xml:",any"`
}

func (e *XMLElement) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type innerElem XMLElement

	err := d.DecodeElement((*innerElem)(e), &start)
	if err != nil {
		return fmt.Errorf("failed to decode XML element: %w", err)
	}

	sort.Slice(e.Attrs, func(i, j int) bool {
		return e.Attrs[i].Name.Local < e.Attrs[j].Name.Local
	})

	if len(e.Elements) > 0 {
		e.Content = ""
	}
	return nil
}

func setupMux(t *testing.T, sgFake *fakes.SendGridClient) *http.ServeMux {
	t.Helper()

	logger := slog.Default()

	config, err := config.Load(true)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	config.Twilio.AgentDIDs = []string{agentDID}

	i18n, err := i18n.NewMessageProvider(logger, config)
	if err != nil {
		t.Fatalf("Error loading i18n dependency: %v", err)
	}

	twilioClient, err := fakes.NewTwilioClient()
	if err != nil {
		t.Fatalf("Failed to instantiate Twilio client test double: %v", err)
	}

	muxFactory := &handler.MuxFactory{
		Recordings: &handler.RecordingsHandler{
			Logger: logger,
		},
		Voice: &handler.VoiceHandler{
			Config: config,
			Emailer: &mail.SendGridMailer{
				Config:         config,
				I18n:           i18n,
				Logger:         logger,
				SendGridClient: sgFake,
			},
			Logger: logger,
			Twigen: &twigen.Voice{
				Config: config,
				I18n:   i18n,
				Logger: logger,
			},
			Twilio: &twilio.API{
				Client: twilioClient,
				Logger: logger,
			},
		},
	}

	return muxFactory.Mux()
}

func getLocalizedTwiml(t *testing.T, langs []string, path string, form url.Values) []byte {
	t.Helper()

	mux := setupMux(t, &fakes.SendGridClient{})

	var gotArchive txtar.Archive
	for _, lang := range langs {
		res := sendRequest(t, mux, path+"?lang="+lang, form.Encode())
		gotArchive.Files = append(gotArchive.Files, txtar.File{
			Name: lang,
			Data: res,
		})
	}

	got := txtar.Format(&gotArchive)
	return got
}

func sendRequest(t *testing.T, handler http.Handler, url string, form string) []byte {
	t.Helper()

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		url,
		strings.NewReader(form),
	)
	if err != nil {
		t.Errorf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	var twiml XMLElement
	if err = xml.NewDecoder(recorder.Body).Decode(&twiml); err != nil {
		t.Errorf("Error decoding XML response: %v", err)
	}

	twimlIndented, err := xml.MarshalIndent(twiml, "", "	")
	if err != nil {
		t.Errorf("Failed to re-marshal response XML with indentation: %v", err)
	}

	return twimlIndented
}

func updateGolden(t *testing.T, path string, got []byte) {
	t.Helper()

	file, err := os.Create(filepath.Clean(path))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err = file.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	t.Logf("Rewriting %s", path)
	if err = os.WriteFile(path, got, 0o600); err != nil {
		t.Fatal(err)
	}
}

var goldenTwimlTests = []struct {
	name   string
	path   string
	form   url.Values
	lang   string `exhaustruct:"optional"`
	golden string `exhaustruct:"optional"`
}{
	{
		name: "inbound-client",
		path: "/voice/inbound",
		form: url.Values{},
		lang: "all",
	},
	{
		name: "inbound-agent",
		path: "/voice/inbound",
		form: url.Values{
			"From": []string{agentDID},
		},
		lang: "en",
	},

	{
		name: "dial-out",
		path: "/voice/dial-out",
		form: url.Values{
			"Digits": []string{clientDID},
		},
		lang: "all",
	},

	{
		name: "connect-agent-en",
		path: "/voice/connect-agent",
		form: url.Values{
			"To":     []string{companyDID},
			"Digits": []string{"1"},
		},
		lang: "en",
	},
	{
		name: "connect-agent-fr",
		path: "/voice/connect-agent",
		form: url.Values{
			"To":     []string{companyDID},
			"Digits": []string{"2"},
		},
		lang: "fr",
	},
	{
		name: "invalid-lang-select",
		path: "/voice/connect-agent",
		form: url.Values{
			"Digits": []string{"3"},
		},
		lang: "all",
	},

	{
		name: "accept-call",
		path: "/voice/accept-call",
		form: url.Values{},
	},

	{
		name: "confirm-connected",
		path: "/voice/confirm-connected",
		form: url.Values{},
	},

	{
		name: "dial-agent-busy",
		path: "/voice/end-call",
		form: url.Values{
			"DialCallStatus": []string{"busy"},
		},
		golden: "go-to-voicemail",
	},
	{
		name: "dial-agent-no-answer",
		path: "/voice/end-call",
		form: url.Values{
			"DialCallStatus": []string{"no-answer"},
		},
		golden: "go-to-voicemail",
	},
	{
		name: "dial-agent-voicemail", // Dial connects to agent's voicemail
		path: "/voice/end-call",
		form: url.Values{
			"DialCallStatus":   []string{"completed"},
			"DialCallDuration": nil,
		},
		golden: "go-to-voicemail",
	},
	{
		name: "dial-agent-connected",
		path: "/voice/end-call",
		form: url.Values{
			"DialCallStatus":   []string{"completed"},
			"DialCallDuration": []string{"10"},
		},
		golden: "noop",
	},
	{
		name: "dial-agent-misc-status",
		path: "/voice/end-call",
		form: url.Values{
			"DialCallStatus": []string{"someotherstatus"},
		},
		golden: "noop",
	},

	{
		name: "start-voicemail-invalid-key",
		path: "/voice/start-voicemail",
		form: url.Values{
			"Digits": []string{"8"},
		},
		golden: "go-to-voicemail",
	},
	{
		name: "record-voicemail",
		path: "/voice/start-voicemail",
		form: url.Values{
			"Digits": []string{"9"},
		},
	},

	{
		name: "rerecord-voicemail",
		path: "/voice/end-voicemail",
		form: url.Values{
			"Digits": []string{"9"},
		},
	},
	{
		name: "end-voicemail",
		path: "/voice/end-voicemail",
		form: url.Values{
			"Digits": []string{"hangup"},
		},
		golden: "noop",
	},

	{
		name:   "status-callback",
		path:   "/voice/status-callback",
		form:   url.Values{},
		golden: "noop",
	},
}

func TestGoldenTwiml(t *testing.T) {
	t.Parallel()

	langs := []string{"en", "fr"}

	for _, test := range goldenTwimlTests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			testLangs := langs
			if test.lang != "" {
				testLangs = []string{test.lang}
			}

			got := getLocalizedTwiml(t, testLangs, test.path, test.form)

			goldenName := test.golden
			if test.golden == "" {
				goldenName = test.name
			}
			goldenFilePath := filepath.Join("testdata", "twiml", goldenName+".golden.xml")

			if *update {
				updateGolden(t, goldenFilePath, got)
				return
			}

			want, err := os.ReadFile(filepath.Clean(goldenFilePath))
			if err != nil {
				t.Errorf("Error reading golden file: %v", err)
			}

			if diff := cmp.Diff(want, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}

var goldenEmailTests = []struct {
	name      string
	path      string
	form      url.Values
	emailSent bool
}{
	{
		name: "call-connected",
		path: "/voice/status-callback",
		form: url.Values{
			"CallSid":    []string{fakes.CallConnected},
			"CallStatus": []string{"completed"},
			"Direction":  []string{"inbound"},
			"From":       []string{clientDID},
		},
		emailSent: false,
	},
	{
		name: "missed-call-en",
		path: "/voice/status-callback",
		form: url.Values{
			"CallSid":    []string{fakes.CallMissedEn},
			"CallStatus": []string{"completed"},
			"Direction":  []string{"inbound"},
			"From":       []string{clientDID},
		},
		emailSent: true,
	},
	{
		name: "missed-call-fr",
		path: "/voice/status-callback",
		form: url.Values{
			"CallSid":    []string{fakes.CallMissedFr},
			"CallStatus": []string{"completed"},
			"Direction":  []string{"inbound"},
			"From":       []string{clientDID},
		},
		emailSent: true,
	},
	{
		name: "no-lang-select",
		path: "/voice/status-callback",
		form: url.Values{
			"CallSid":    []string{fakes.CallHangup},
			"CallStatus": []string{"completed"},
			"Direction":  []string{"inbound"},
			"From":       []string{clientDID},
		},
		emailSent: false,
	},
	{
		name: "voicemail-en",
		path: "/voice/status-callback",
		form: url.Values{
			"CallSid":    []string{fakes.CallWithVoicemailEn},
			"CallStatus": []string{"completed"},
			"Direction":  []string{"inbound"},
			"From":       []string{clientDID},
		},
		emailSent: true,
	},
	{
		name: "voicemail-fr",
		path: "/voice/status-callback",
		form: url.Values{
			"CallSid":    []string{fakes.CallWithVoicemailFr},
			"CallStatus": []string{"completed"},
			"Direction":  []string{"inbound"},
			"From":       []string{clientDID},
		},
		emailSent: true,
	},
}

func TestGoldenEmails(t *testing.T) {
	t.Parallel()

	for _, test := range goldenEmailTests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			sgFake := &fakes.SendGridClient{}
			mux := setupMux(t, sgFake)

			sendRequest(t, mux, test.path, test.form.Encode())

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
				updateGolden(t, filePath, got)
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
