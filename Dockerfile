##
## Stage 1 - Build image
##
FROM golang:1.23 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /ocomms ./cmd/ocomms

##
## Stage 2 - Deploy
##
FROM gcr.io/distroless/static

WORKDIR /

COPY --from=build --chown=nonroot:nonroot /ocomms /ocomms

# Run binary as nonroot user added by Distroless
USER nonroot:nonroot

ENTRYPOINT ["/ocomms"]
