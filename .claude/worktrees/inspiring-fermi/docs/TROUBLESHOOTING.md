# Troubleshooting

## Docker Socket Issues

Symptoms:

- Cannot list containers
- Permission denied on /var/run/docker.sock
- Dockwatch starts but never scans

Checks:

```bash
ls -l /var/run/docker.sock
docker ps
```

Container run fix:

```bash
docker run -d \
  --name dockwatch \
  -v /var/run/docker.sock:/var/run/docker.sock \
  fugginold/dockwatch:latest
```

If using rootless Docker, confirm socket path and mount match your daemon setup.

## API Version mismatch

Symptoms:

- Client and server API mismatch errors

Fix:

```bash
DW_API_VERSION=$(docker version --format '{{.Server.APIVersion}}')
docker run -d \
  --name dockwatch \
  -e DOCKER_API_VERSION="${DW_API_VERSION}" \
  -v /var/run/docker.sock:/var/run/docker.sock \
  fugginold/dockwatch:latest
```

## Permission Errors

Symptoms:

- Docker command fails without sudo
- Access denied during local runs

Fix user group membership:

```bash
sudo usermod -aG docker "$USER"
newgrp docker
```

Container permission notes:

- Ensure dockwatch container has socket mount
- Confirm host Docker daemon is accessible from that mount

## Useful diagnostics

Logs:

```bash
docker logs --tail=200 dockwatch
```

Health check command:

```bash
docker exec dockwatch /dockwatch --health-check
```

One-shot update for debugging:

```bash
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock fugginold/dockwatch:latest --force-update
```

API auth check:

```bash
curl -H "Authorization: Bearer ${DW_TOKEN}" http://localhost:8080/v1/schedule
```
