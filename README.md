## ocomms
Twilio webhook handlers implementing InfoTech Ottawa's automated voice ([IVR](https://en.wikipedia.org/wiki/Interactive_voice_response)) system.

### Features
* Ability for callers to discard and re-record their voice messages before submitting (I always hated having to one-shot voicemails)
* Internationalized - fully English-French bilingual


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

### Apply Terraform changes
```
gcloud auth application-default login
```
```
cd terraform
```
```
terraform apply
```
