.DEFAULT_GOAL = check

.PHONY: fmt lint fix test check run

fmt:
	go fmt ./...

lint:
	golangci-lint run

fix:
	golangci-lint run --fix

test:
	go test ./...

check: fmt lint test

run:
	go run src/main.go