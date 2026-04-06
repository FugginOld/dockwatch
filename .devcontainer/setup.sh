#!/bin/bash
set -e

echo "🔧 Installing dependencies..."

# System tools
sudo apt-get update
sudo apt-get install -y curl git build-essential

# Install latest Go
GO_VERSION=1.22.5
curl -LO https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz

# Add Go to PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.profile
export PATH=$PATH:/usr/local/go/bin

# Verify
go version

# Go tools
go install golang.org/x/tools/gopls@latest
go install github.com/go-delve/delve/cmd/dlv@latest

echo "✅ Setup complete"