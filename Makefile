.DEFAULT_GOAL = check

fmt:
	go fmt ./...

lint:
	golangci-lint run

fix:
	golangci-lint run --fix

generate:
	go generate ./...

schemavalidate:
	ajv validate -s internal/config/schema.json -d internal/config/config.yaml --spec=draft2020
	ajv validate -s internal/i18n/schema.json -d internal/i18n/messages/en.yaml --spec=draft2020
	ajv validate -s internal/i18n/schema.json -d internal/i18n/messages/fr.yaml --spec=draft2020

vulncheck:
	govulncheck ./...

test:
	go test ./...

testupdate:
	go test ./internal/app -update

check: generate schemavalidate fmt lint vulncheck test

run:
	go run cmd/ocomms/main.go --logging.format=text
