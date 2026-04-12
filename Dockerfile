# syntax=docker/dockerfile:1

# -------- Build stage --------
FROM --platform=$BUILDPLATFORM golang:1.22-alpine AS builder

WORKDIR /app

# Cache deps
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build for target platform
ARG TARGETOS
ARG TARGETARCH

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o dockwatch

# -------- Runtime stage --------
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /app/dockwatch .

# Basic runtime health check (process must be alive)
HEALTHCHECK --interval=10s --timeout=3s --retries=5 \
  CMD pgrep dockwatch || exit 1

ENTRYPOINT ["./dockwatch"]
