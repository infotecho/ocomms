.DEFAULT_GOAL = check

fmt:
	go fmt ./...

lint:
	golangci-lint run --build-tags test,tools

fix:
	golangci-lint run --build-tags test,tools --fix

generate:
	go generate ./...

schemavalidate:
	ajv validate -s internal/config/schema.json -d internal/config/config.yaml --spec=draft2020
	ajv validate -s internal/i18n/schema.json -d internal/i18n/messages/en.yaml --spec=draft2020
	ajv validate -s internal/i18n/schema.json -d internal/i18n/messages/fr.yaml --spec=draft2020

vulncheck:
	govulncheck ./...

test:
	go test -tags=test ./...

testupdate:
	find . -name "*.golden.*" -exec rm -f {} +
	go test ./internal/handler/ -tags=test -update

cover:
	go test ./... -cover -coverprofile=coverage.out -tags=test
	go tool cover -html=coverage.out

check: generate schemavalidate fmt lint vulncheck test

run:
	go run cmd/ocomms/main.go --logging.format=text
