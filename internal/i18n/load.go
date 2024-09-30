package i18n

import (
	"embed"
	"fmt"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"gopkg.in/yaml.v3"
)

//go:embed messages
var messagesDir embed.FS

const messagesDirName = "messages"

func loadMessages() (map[string]Messages, error) {
	dirEntries, err := messagesDir.ReadDir(messagesDirName)
	if err != nil {
		return nil, fmt.Errorf("failed to load i18n messages: %w", err)
	}

	messages := map[string]Messages{}

	for _, dirEntry := range dirEntries {
		if dirEntry.IsDir() {
			continue
		}

		filename := dirEntry.Name()

		langMessages, err := loadMessagesFromFile(filename)
		if err != nil {
			return nil, err
		}

		lang := strings.Split(filename, ".")[0]
		messages[lang] = langMessages
	}

	return messages, nil
}

func loadMessagesFromFile(filename string) (Messages, error) {
	file, err := messagesDir.ReadFile(messagesDirName + "/" + filename)
	if err != nil {
		return Messages{}, fmt.Errorf("failed to load i18n messages from %s: %w", filename, err)
	}

	var rawMap map[string]any

	err = yaml.Unmarshal(file, &rawMap)
	if err != nil {
		return Messages{}, fmt.Errorf("failed to unmarshal %s: %w", filename, err)
	}

	var messages Messages

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		ErrorUnset: true,
		Result:     &messages,
	})
	if err != nil {
		return Messages{}, fmt.Errorf("failed to create i18n message decoder: %w", err)
	}

	err = decoder.Decode(rawMap)
	if err != nil {
		return Messages{}, fmt.Errorf("failed to decode %s: %w", filename, err)
	}

	return messages, nil
}
