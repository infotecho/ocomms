package app //nolint:testpackage

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
	"golang.org/x/tools/txtar"
)

const (
	agentDID = "+16138160938"
)

var update = flag.Bool("update", false, "rewrite testdata/*.xml files") //nolint:gochecknoglobals

type TableTestInput struct {
	hook string
	form url.Values
	want string
	lang string
}

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

func setup(t *testing.T) string {
	t.Helper()

	t.Setenv("TWILIO_API_KEY_SID", "fake")
	t.Setenv("TWILIO_API_KEY_SECRET", "fake")

	config, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	config.Twilio.AgentDIDs = []string{agentDID}

	muxFactory := wireDependencies(config, slog.Default()).MuxFactory
	mux := muxFactory.Mux()
	server := httptest.NewServer(mux)
	t.Cleanup(func() {
		server.Close()
	})
	return server.URL
}

func postHook(t *testing.T, url string, form string) []byte {
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

func rewriteGoldenFile(t *testing.T, path string, got []byte) {
	t.Helper()

	path = filepath.Clean(path)
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("Failed to create %v", err)
	}
	defer func() {
		if err = file.Close(); err != nil {
			t.Fatalf("Failed to close %s: %v", path, err)
		}
	}()

	t.Logf("Rewriting %s", path)
	if err = os.WriteFile(path, got, 0600); err != nil { //nolint:gofumpt
		t.Fatalf("Failed to write XML to %s: %v", path, err)
	}
}

func runTableTest(t *testing.T, serverURL string, test TableTestInput) {
	t.Helper()

	langs := []string{"en", "fr"}
	if test.lang != "" {
		langs = []string{test.lang}
	}

	var gotArchive txtar.Archive
	for _, lang := range langs {
		url := serverURL + test.hook + "?lang=" + lang
		res := postHook(t, url, test.form.Encode())
		gotArchive.Files = append(gotArchive.Files, txtar.File{
			Name: lang,
			Data: res,
		})
	}

	got := txtar.Format(&gotArchive)

	if *update {
		rewriteGoldenFile(t, test.want, got)
		return
	}

	want, err := os.ReadFile(test.want)
	if err != nil {
		t.Errorf("Error reading golden file: %v", err)
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Error(diff)
	}
}

func TestGolden(t *testing.T) { //nolint:paralleltest, tparallel
	serverURL := setup(t)

	tests := []TableTestInput{
		{
			hook: "/voice/inbound",
			form: url.Values{},
			want: "testdata/voice/inbound-client.xml",
			lang: "en",
		},
		{
			hook: "/voice/inbound",
			form: url.Values{
				"From": []string{agentDID},
			},
			want: "testdata/voice/inbound-agent.xml",
			lang: "en",
		},
		{
			hook: "/voice/accept-call",
			form: url.Values{},
			want: "testdata/voice/accept-call.xml",
		},
	}

	for _, test := range tests {
		t.Run(test.hook, func(t *testing.T) {
			t.Parallel()
			runTableTest(t, serverURL, test)
		})
	}
}
