## ocomms
### Local Setup
To run `make`/`make check`:

1. Install [govulncheck](https://go.dev/doc/security/vuln/)
```sh
go install golang.org/x/vuln/cmd/govulncheck@latest
```

2. Install [golangci-lint](https://golangci-lint.run)
```sh
brew install golangci-lint
```

3. Install [Bun](https://bun.sh/docs/installation)
```sh
curl -fsSL https://bun.sh/install | bash
```

4. Install [ajv](https://github.com/ajv-validator/ajv-cli)
```sh
bun install -g ajv-cli
```
