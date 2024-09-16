package app //nolint:testpackage

import (
	"context"
	"encoding/xml"
	"flag"
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
)

const (
	agentDID = "+16138160938"
)

var update = flag.Bool("update", false, "rewrite testdata/*.xml files") //nolint:gochecknoglobals

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
		return err //nolint:wrapcheck
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

func rewriteTestFile(t *testing.T, path string, gotXML XMLElement) {
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

	indentedXML, err := xml.MarshalIndent(gotXML, "", "	")
	if err != nil {
		t.Fatalf("Failed to re-marshal response XML with indentation: %v", err)
	}
	indentedXML = append(indentedXML, byte('\n'))

	t.Logf("Rewriting %s", path)
	if err = os.WriteFile(path, indentedXML, 0600); err != nil { //nolint:gofumpt
		t.Fatalf("Failed to write XML to %s: %v", path, err)
	}
}

func runTableTest(t *testing.T, serverURL string, test TableTestInput) {
	t.Helper()

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		serverURL+test.hook,
		strings.NewReader(test.form.Encode()),
	)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Error making POST request: %v", err)
	}
	defer func() {
		if err = res.Body.Close(); err != nil {
			t.Fatalf("Failed to close request body: %v", err)
		}
	}()

	var gotXML XMLElement
	if err = xml.NewDecoder(res.Body).Decode(&gotXML); err != nil {
		t.Fatalf("Error decoding XML response: %v", err)
	}

	if *update {
		rewriteTestFile(t, test.want, gotXML)
		return
	}

	wantFile, err := os.Open(test.want)
	if err != nil {
		t.Fatalf("Error reading XML file: %v", err)
	}
	defer func() {
		if err = wantFile.Close(); err != nil {
			t.Fatalf("Failed to close XML file: %v", err)
		}
	}()

	var wantXML XMLElement
	if err = xml.NewDecoder(wantFile).Decode(&wantXML); err != nil {
		t.Fatalf("Error decoding XML file: %v", err)
	}

	if diff := cmp.Diff(wantXML, gotXML); diff != "" {
		t.Error(diff)
	}
}

type TableTestInput struct {
	hook string
	form url.Values
	want string
}

func TestGolden(t *testing.T) { //nolint:paralleltest, tparallel
	serverURL := setup(t)

	tests := []TableTestInput{
		{
			hook: "/voice/inbound",
			form: url.Values{},
			want: "testdata/voice/inbound-from-client.xml",
		},
		{
			hook: "/voice/inbound",
			form: url.Values{
				"From": []string{agentDID},
			},
			want: "testdata/voice/inbound-from-agent.xml",
		},
	}

	for _, test := range tests {
		t.Run(test.hook, func(t *testing.T) {
			t.Parallel()
			runTableTest(t, serverURL, test)
		})
	}
}
