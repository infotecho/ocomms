package i18n_test

import (
	"log/slog"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/infotecho/ocomms/internal/config"
	"github.com/infotecho/ocomms/internal/i18n"
)

func Test_Message_en(t *testing.T) {
	t.Parallel()

	mp, err := i18n.NewMessageProvider(slog.Default(), config.Config{})
	if err != nil {
		t.Errorf("Failed to load message provider: %s", err)
	}

	welcome, err := mp.Message("en", func(m i18n.Messages) string { return m.Voice.Welcome })
	if err != nil {
		t.Error(err)
	}

	if diff := cmp.Diff("Welcome to InfoTech Ottawa.", welcome); diff != "" {
		t.Error(diff)
	}
}

func Test_Message_fr(t *testing.T) {
	t.Parallel()

	mp, err := i18n.NewMessageProvider(slog.Default(), config.Config{})
	if err != nil {
		t.Errorf("Failed to load message provider: %s", err)
	}

	welcome, err := mp.Message("fr", func(m i18n.Messages) string { return m.Voice.Welcome })
	if err != nil {
		t.Error(err)
	}

	if diff := cmp.Diff("Vous avez rejoint l'infothèque d'Ottawa.", welcome); diff != "" {
		t.Error(diff)
	}
}

func Test_Message_invalidLang(t *testing.T) {
	t.Parallel()

	var config config.Config
	config.I18N.DefaultLang = "en"

	mp, err := i18n.NewMessageProvider(slog.Default(), config)
	if err != nil {
		t.Errorf("Failed to load message provider: %s", err)
	}

	welcome, err := mp.Message("foo", func(m i18n.Messages) string { return m.Voice.Welcome })

	if err == nil || !strings.Contains(err.Error(), "foo") {
		t.Errorf("Expected error invalid lang, got: %s", err)
	}

	if diff := cmp.Diff("Welcome to InfoTech Ottawa.", welcome); diff != "" {
		t.Error(diff)
	}
}

func Test_Message_replacementExpected(t *testing.T) {
	t.Parallel()

	mp, err := i18n.NewMessageProvider(slog.Default(), config.Config{})
	if err != nil {
		t.Errorf("Failed to load message provider: %s", err)
	}

	langSelect, err := mp.Message("en", func(m i18n.Messages) string { return m.Voice.LangSelect })

	if err == nil || !strings.Contains(err.Error(), "digit") {
		t.Errorf("Expected error missing 'digit' replacement, got: %s", err)
	}
	if diff := cmp.Diff("For service in English, press .", langSelect); diff != "" {
		t.Error(diff)
	}
}

func Test_MessageReplace(t *testing.T) {
	t.Parallel()

	mp, err := i18n.NewMessageProvider(slog.Default(), config.Config{})
	if err != nil {
		t.Errorf("Failed to load message provider: %s", err)
	}

	langSelect, err := mp.MessageReplace(
		"en",
		func(m i18n.Messages) string { return m.Voice.LangSelect },
		map[string]string{"digit": "1"},
	)
	if err != nil {
		t.Error(err)
	}

	if diff := cmp.Diff("For service in English, press 1.", langSelect); diff != "" {
		t.Error(diff)
	}
}

func Test_MessageReplace_InvalidReplacement(t *testing.T) {
	t.Parallel()

	mp, err := i18n.NewMessageProvider(slog.Default(), config.Config{})
	if err != nil {
		t.Errorf("Failed to load message provider: %s", err)
	}

	langSelect, err := mp.MessageReplace(
		"en",
		func(m i18n.Messages) string { return m.Voice.LangSelect },
		map[string]string{},
	)

	if err == nil || !strings.Contains(err.Error(), "digit") {
		t.Errorf("Expected error missing 'digit' replacement, got: %s", err)
	}
	if diff := cmp.Diff("For service in English, press .", langSelect); diff != "" {
		t.Error(diff)
	}
}
