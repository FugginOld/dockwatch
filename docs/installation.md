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

Set `DOCKER_API_VERSION` from your daemon to avoid Docker client/server API mismatch errors:

```bash
docker rm -f dockwatch 2>/dev/null || true
DW_API_VERSION=$(docker version --format '{{.Server.APIVersion}}')

docker run -d \
  --name dockwatch \
  --restart unless-stopped \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -e DOCKER_API_VERSION="${DW_API_VERSION}" \
  fugginold/dockwatch:latest
```

### Run Dockwatch (Docker Compose)

```yaml
services:
  dockwatch:
    image: fugginold/dockwatch:latest
    container_name: dockwatch
    restart: unless-stopped
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
      DOCKER_API_VERSION: "${DW_API_VERSION}"
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
