FROM mcr.microsoft.com/devcontainers/base:debian

# Install Go
ARG GO_VERSION=1.25.0
RUN curl -fsSL -o go${GO_VERSION}.linux-amd64.tar.gz https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz \
    && rm -rf /usr/local/go \
    && tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz \
    && rm go${GO_VERSION}.linux-amd64.tar.gz

ENV PATH="/usr/local/go/bin:${PATH}"

# Install dependencies
RUN apt-get update && apt-get install -y \
    git \
    curl \
    make \
    gcc \
    && rm -rf /var/lib/apt/lists/*