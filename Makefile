.DEFAULT_GOAL = check

.PHONY: fmt lint fix test check run

fmt:
	go fmt ./...

lint:
	golangci-lint run

fix:
	golangci-lint run --fix

generate:
	go generate ./...

check-config:
	ajv validate -s src/internal/config/schema.json -d src/internal/config/config.yaml --spec=draft2020

vulncheck:
	govulncheck ./...

test:
	go test ./...

check: generate check-config fmt lint vulncheck test

run:
	go run src/main.go
