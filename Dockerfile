FROM --platform=$BUILDPLATFORM golang:1.25-bookworm AS builder

ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG VERSION=dev

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags "-s -w -X github.com/fugginold/dockwatch/internal/meta.Version=${VERSION}" -o /out/dockwatch ./

FROM alpine:3.23.3 AS certs

RUN apk add --no-cache \
    ca-certificates \
    tzdata

FROM scratch

LABEL "com.centurylinklabs.dockwatch"="true"

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=certs /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /out/dockwatch /dockwatch

EXPOSE 8080

HEALTHCHECK CMD ["/dockwatch", "--health-check"]

ENTRYPOINT ["/dockwatch"]