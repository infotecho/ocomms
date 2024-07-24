//go:build tools

// Genschema generates a JSON schema for config and i18n YAML files based on their unmarshalled structs.
// This allows for code completion in the file as well as build-time validation.
package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"reflect"

	"github.com/infotecho/ocomms/internal/config"
	"github.com/infotecho/ocomms/internal/i18n"
	"github.com/invopop/jsonschema"
)

func main() {
	var (
		config   config.Config
		messages i18n.Messages
	)
	genSchema(config)
	genSchema(messages)
}

func genSchema(goType any) {
	log.SetFlags(0)

	pkg := os.Getenv("GOPACKAGE")
	if pkg != filepath.Base(reflect.TypeOf(goType).PkgPath()) {
		return
	}

	schema := jsonschema.Reflect(goType)

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
		return
	}

	log.Printf("Generated %s schema", pkg)
}
