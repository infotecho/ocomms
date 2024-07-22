//go:build tools

// Genschema generates a JSON schema for config.yaml from the Config struct it is unmarshalled to.
// This allows for code completion in the file as well as compile-time validation.
package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/infotecho/ocomms/internal/config"
	"github.com/invopop/jsonschema"
)

func main() {
	var config config.Config
	schema := jsonschema.Reflect(&config)

	jsonStr, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	file, err := os.Create("schema.json")
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		err = file.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	_, err = file.Write(jsonStr)
	if err != nil {
		log.Println(err)
	}
}
