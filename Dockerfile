##
## Stage 1 - Build image
##
FROM golang:1.22 AS build

WORKDIR /app

COPY go.mod .
COPY src/* .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /ocomms

##
## Stage 2 - Deploy
##
FROM gcr.io/distroless/static

WORKDIR /

COPY --from=build --chown=nonroot:nonroot /ocomms /ocomms

# Run binary as nonroot user added by Distroless
USER nonroot:nonroot

ENTRYPOINT ["/ocomms"]
