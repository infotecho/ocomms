package test_test

import (
	"context"
	"encoding/xml"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/infotecho/ocomms/internal/app"
	"github.com/infotecho/ocomms/internal/config"
	"golang.org/x/tools/txtar"
)

const (
	agentDID = "+16138160938"
)

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

func setupServer(t *testing.T) string {
	t.Helper()

	config, err := config.Load(true)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	config.Twilio.AgentDIDs = []string{agentDID}

	muxFactory := app.WireDependencies(config, slog.Default()).MuxFactory
	mux := muxFactory.Mux()
	server := httptest.NewServer(mux)
	t.Cleanup(func() {
		server.Close()
	})
	return server.URL
}

func getLocalizedTwiml(t *testing.T, langs []string, url string, form url.Values) []byte {
	t.Helper()

	var gotArchive txtar.Archive
	for _, lang := range langs {
		res := postForTwiml(t, url+"?lang="+lang, form.Encode())
		gotArchive.Files = append(gotArchive.Files, txtar.File{
			Name: lang,
			Data: res,
		})
	}

	got := txtar.Format(&gotArchive)
	return got
}

func postForTwiml(t *testing.T, url string, form string) []byte {
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
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Errorf("Error making POST request: %v", err)
	}
	t.Cleanup(func() {
		if err = res.Body.Close(); err != nil {
			t.Errorf("Failed to close request body: %v", err)
		}
	})

	var twiml XMLElement
	if err = xml.NewDecoder(res.Body).Decode(&twiml); err != nil {
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
			"Digits": []string{"1234567890"},
		},
		lang: "all",
	},

	{
		name: "connect-agent-en",
		path: "/voice/connect-agent",
		form: url.Values{
			"To":     []string{"1234567890"},
			"Digits": []string{"1"},
		},
		lang: "en",
	},
	{
		name: "connect-agent-fr",
		path: "/voice/connect-agent",
		form: url.Values{
			"To":     []string{"1234567890"},
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

	serverURL := setupServer(t)

	langs := []string{"en", "fr"}

	for _, test := range goldenTwimlTests {
		pathRoot := strings.Split(test.path, "/")[1]
		name := path.Join(pathRoot, test.name)
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			testLangs := langs
			if test.lang != "" {
				testLangs = []string{test.lang}
			}

			got := getLocalizedTwiml(t, testLangs, serverURL+test.path, test.form)

			goldenName := test.golden
			if test.golden == "" {
				goldenName = test.name
			}
			goldenFilePath := filepath.Join("testdata", "twiml", pathRoot, goldenName+".golden.xml")

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
