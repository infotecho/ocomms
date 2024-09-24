package i18n_test

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/infotecho/ocomms/internal/config"
	"github.com/infotecho/ocomms/internal/i18n"
)

func Test_Message_en(t *testing.T) {
	t.Parallel()

	mp, err := i18n.NewMessageProvider(slog.Default(), config.Config{}) //nolint:exhaustruct
	if err != nil {
		t.Errorf("Failed to load message provider: %s", err)
	}

	welcome := mp.Message(context.Background(), "en", func(m i18n.Messages) string { return m.Voice.Welcome })
	if err != nil {
		t.Error(err)
	}

	if diff := cmp.Diff("Welcome to Infotech Ottawa.", welcome); diff != "" {
		t.Error(diff)
	}
}

func Test_Message_fr(t *testing.T) {
	t.Parallel()

	mp, err := i18n.NewMessageProvider(slog.Default(), config.Config{}) //nolint:exhaustruct
	if err != nil {
		t.Errorf("Failed to load message provider: %s", err)
	}

	welcome := mp.Message(context.Background(), "fr", func(m i18n.Messages) string { return m.Voice.Welcome })

	if diff := cmp.Diff("Vous avez rejoint l'infothèque d'Ottawa.", welcome); diff != "" {
		t.Error(diff)
	}
}

func Test_Message_invalidLang(t *testing.T) {
	t.Parallel()

	var config config.Config
	config.I18N.DefaultLang = "en"

	logBuf := bytes.Buffer{}
	logger := slog.New(slog.NewTextHandler(&logBuf, nil))

	mp, err := i18n.NewMessageProvider(logger, config)
	if err != nil {
		t.Errorf("Failed to load message provider: %s", err)
	}

	welcome := mp.Message(context.Background(), "foo", func(m i18n.Messages) string { return m.Voice.Welcome })

	logOut := logBuf.String()
	if !strings.Contains(logOut, "ERROR") && !strings.Contains(logOut, "foo") {
		t.Errorf("Expected error invalid lang, got: %s", logOut)
	}

	if diff := cmp.Diff("Welcome to Infotech Ottawa.", welcome); diff != "" {
		t.Error(diff)
	}
}

func Test_Message_replacementExpected(t *testing.T) {
	t.Parallel()

	logBuf := bytes.Buffer{}
	logger := slog.New(slog.NewTextHandler(&logBuf, nil))

	mp, err := i18n.NewMessageProvider(logger, config.Config{}) //nolint:exhaustruct
	if err != nil {
		t.Errorf("Failed to load message provider: %s", err)
	}

	langSelect := mp.Message(context.Background(), "en", func(m i18n.Messages) string { return m.Voice.LangSelect })

	logOut := logBuf.String()
	if !strings.Contains(logOut, "ERROR") && !strings.Contains(logOut, "digit") {
		t.Errorf("Expected error missing 'digit' replacement, got: %s", logOut)
	}

	if diff := cmp.Diff("For service in English, press .", langSelect); diff != "" {
		t.Error(diff)
	}
}

func Test_MessageReplace(t *testing.T) {
	t.Parallel()

	mp, err := i18n.NewMessageProvider(slog.Default(), config.Config{}) //nolint:exhaustruct
	if err != nil {
		t.Errorf("Failed to load message provider: %s", err)
	}

	langSelect := mp.MessageReplace(
		context.Background(),
		"en",
		func(m i18n.Messages) string { return m.Voice.LangSelect },
		map[string]string{"digit": "1"},
	)

	if diff := cmp.Diff("For service in English, press 1.", langSelect); diff != "" {
		t.Error(diff)
	}
}

func Test_MessageReplace_InvalidReplacement(t *testing.T) {
	t.Parallel()

	logBuf := bytes.Buffer{}
	logger := slog.New(slog.NewTextHandler(&logBuf, nil))

	mp, err := i18n.NewMessageProvider(logger, config.Config{}) //nolint:exhaustruct
	if err != nil {
		t.Errorf("Failed to load message provider: %s", err)
	}

	langSelect := mp.MessageReplace(
		context.Background(),
		"en",
		func(m i18n.Messages) string { return m.Voice.LangSelect },
		map[string]string{},
	)

	logOut := logBuf.String()
	if !strings.Contains(logOut, "ERROR") && !strings.Contains(logOut, "digit") {
		t.Errorf("Expected error missing 'digit' replacement, got: %s", err)
	}

	if diff := cmp.Diff("For service in English, press .", langSelect); diff != "" {
		t.Error(diff)
	}
}
