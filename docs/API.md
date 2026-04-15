# API

Dockwatch HTTP API is optional and served on port 8080.

## Auth

All endpoints require:

- Authorization: Bearer <token>

Token source:

- DOCKWATCH_HTTP_API_TOKEN

If token is missing or invalid, response is 401 unauthorized.

## Endpoints

### POST /v1/update

Triggers an update run for all monitored containers.

Optional query:

- image: comma-separated image list

Example:

```bash
curl -X POST -H "Authorization: Bearer ${DW_TOKEN}" http://localhost:8080/v1/update
```

Target specific images:

```bash
curl -X POST -H "Authorization: Bearer ${DW_TOKEN}" "http://localhost:8080/v1/update?image=nginx:latest,redis:latest"
```

### GET /v1/schedule

Returns active schedule and next planned run.

Example:

```bash
curl -H "Authorization: Bearer ${DW_TOKEN}" http://localhost:8080/v1/schedule
```

Typical response:

```json
{"schedule":"@every 24h","nextRun":"2026-04-07T01:23:45Z"}
```

### POST or PUT /v1/schedule

Updates the active runtime schedule without restarting Dockwatch.

Query parameters:

- schedule
- cron (alias)

Examples:

```bash
curl -X POST -H "Authorization: Bearer ${DW_TOKEN}" "http://localhost:8080/v1/schedule?schedule=@every%2030m"
```

```bash
curl -X PUT -H "Authorization: Bearer ${DW_TOKEN}" "http://localhost:8080/v1/schedule?cron=@daily"
```

Invalid schedule returns 400 bad request.

## Enable API Mode

```bash
docker run -d \
  --name dockwatch \
  --restart unless-stopped \
  -p 8080:8080 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -e DOCKER_API_VERSION="${DW_API_VERSION}" \
  -e DOCKWATCH_HTTP_API_TOKEN="${DW_TOKEN}" \
  fugginold/dockwatch:latest \
  --http-api-update \
  --http-api-periodic-polls \
  --schedule "@every 24h"
```

## Notes

- Schedule endpoint is registered only when periodic polling is enabled.
- The update lock prevents concurrent update sessions between scheduler and API triggers.
