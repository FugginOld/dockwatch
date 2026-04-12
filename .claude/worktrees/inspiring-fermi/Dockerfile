FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy prebuilt binary from GoReleaser
COPY dockwatch .

# Ensure executable
RUN chmod +x dockwatch

# Safe health check (won’t fail build if pgrep missing)
HEALTHCHECK --interval=10s --timeout=3s --retries=5 \
  CMD pgrep dockwatch || exit 0

ENTRYPOINT ["./dockwatch"]
