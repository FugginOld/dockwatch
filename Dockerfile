#
# Builder
#
FROM golang:1.26.5-alpine AS builder

ARG VERSION=dev

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-s -w -X github.com/fugginold/dockwatch/internal/meta.Version=${VERSION}" \
    -o /dockwatch .

#
# Runtime
#
FROM alpine:3.23

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /dockwatch /dockwatch

HEALTHCHECK --interval=10s --timeout=3s --retries=5 \
  CMD ["/dockwatch", "--health-check"]

ENTRYPOINT ["/dockwatch"]
