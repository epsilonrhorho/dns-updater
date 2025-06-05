# syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM golang:1.23 AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /dns-updater .

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /dns-updater /dns-updater

ENTRYPOINT ["/dns-updater"]
