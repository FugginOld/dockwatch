## Installation

Dockwatch is typically installed as a Docker container that monitors the local Docker daemon.

### Debian: Install Docker Engine

```bash
sudo apt-get update
sudo apt-get install -y ca-certificates curl gnupg
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/debian/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
sudo chmod a+r /etc/apt/keyrings/docker.gpg

echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/debian $(. /etc/os-release; echo $VERSION_CODENAME) stable" | sudo tee /etc/apt/sources.list.d/docker.list >/dev/null

sudo apt-get update
sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
sudo systemctl enable --now docker
```

### Run Dockwatch (docker run)

You can optionally use the install helper script to enter a cron schedule at install time:

```bash
./scripts/install-dockwatch.sh
```

The script prompts for a schedule and runs Dockwatch with `--schedule <value>`.

If you prefer manual installation, run Dockwatch with API-version auto-detection to avoid Docker client/server API mismatch errors:

```bash
docker rm -f dockwatch 2>/dev/null || true
DW_API_VERSION=$(docker version --format '{{.Server.APIVersion}}')

docker run -d \
  --name dockwatch \
  --restart unless-stopped \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v dockwatch-config:/config \
  -e DOCKER_API_VERSION="${DW_API_VERSION}" \
  fugginold/dockwatch:latest \
  --schedule "@every 24h"
```

Dockwatch performs one update check immediately at startup, then continues on the configured schedule. Runtime schedule changes are saved in `/config/dockwatch.json`; mount `/config` if you want those changes to survive container recreation.

### Run Dockwatch (Docker Compose)

```yaml
services:
  dockwatch:
    image: fugginold/dockwatch:latest
    container_name: dockwatch
    restart: unless-stopped
    command: --schedule "@every 24h"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - dockwatch-config:/config
    environment:
      DOCKER_API_VERSION: "${DW_API_VERSION}"

volumes:
  dockwatch-config:
```

Then run:

```bash
DW_API_VERSION=$(docker version --format '{{.Server.APIVersion}}')
export DW_API_VERSION
docker compose up -d
```

### Verify installation

```bash
docker ps --filter name=dockwatch
docker logs --tail=100 dockwatch
```
