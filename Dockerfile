#
# Builder
#
FROM golang:1.26.6-alpine AS builder

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

COPY --from=builder /dockwatch /usr/local/bin/dockwatch

# Dockwatch recognises its own container by this label. Without it a self-update
# stops this container like any other watched one, never starts the replacement,
# and leaves any renamed old instance unreaped. Dropped once in e9cc67b; there is
# a test in pkg/container that fails if it goes missing or moves to the builder.
# Keep the key in sync with dockwatchLabel in pkg/container/metadata.go.
LABEL io.github.fugginold.dockwatch="true"

HEALTHCHECK --interval=10s --timeout=3s --retries=5 \
  CMD ["dockwatch", "--health-check"]

ENTRYPOINT ["dockwatch"]
