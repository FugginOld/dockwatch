# Configuration

## Env vars

Docker client and connectivity:

- DOCKER_HOST
- DOCKER_TLS_VERIFY
- DOCKER_API_VERSION
- DOCKWATCH_REGISTRY_TLS_SKIP_VERIFY

Scheduling and update behavior:

- DOCKWATCH_POLL_INTERVAL
- DOCKWATCH_SCHEDULE
- DOCKWATCH_TIMEOUT
- DOCKWATCH_NO_PULL
- DOCKWATCH_NO_RESTART
- DOCKWATCH_RUN_ONCE
- DOCKWATCH_MONITOR_ONLY
- DOCKWATCH_CLEANUP
- DOCKWATCH_REMOVE_VOLUMES
- DOCKWATCH_INCLUDE_RESTARTING
- DOCKWATCH_INCLUDE_STOPPED
- DOCKWATCH_REVIVE_STOPPED
- DOCKWATCH_ROLLING_RESTART
- DOCKWATCH_LIFECYCLE_HOOKS

Targeting and filtering:

- DOCKWATCH_LABEL_ENABLE
- DOCKWATCH_DISABLE_CONTAINERS
- DOCKWATCH_SCOPE
- DOCKWATCH_LABEL_TAKE_PRECEDENCE

HTTP API and metrics:

- DOCKWATCH_HTTP_API_UPDATE
- DOCKWATCH_HTTP_API_TOKEN
- DOCKWATCH_HTTP_API_PERIODIC_POLLS
- DOCKWATCH_HTTP_API_METRICS

Logging and output:

- DOCKWATCH_LOG_LEVEL
- DOCKWATCH_LOG_FORMAT
- DOCKWATCH_DEBUG
- DOCKWATCH_TRACE
- NO_COLOR
- DOCKWATCH_PORCELAIN

## Flags

Core:

- --schedule, -s
- --cron
- --interval, -i
- --run-once, -R
- --force-update
- --stop-timeout, -t

Update controls:

- --no-pull
- --no-restart
- --cleanup, -c
- --remove-volumes
- --monitor-only, -m
- --rolling-restart

Targeting:

- --label-enable, -e
- --disable-containers, -x
- --scope
- --label-take-precedence

Docker client:

- --host, -H
- --tlsverify, -v
- --api-version, -a
- --registry-tls-skip-verify

API and observability:

- --http-api-update
- --http-api-token
- --http-api-periodic-polls
- --http-api-metrics
- --health-check

Logging:

- --log-level
- --log-format, -l
- --debug, -d
- --trace
- --no-color

## Scheduling Examples

Every 24 hours:

```bash
--schedule "@every 24h"
```

Every 10 minutes:

```bash
--schedule "@every 10m"
```

Daily at 03:15:

```bash
--schedule "15 3 * * *"
```

Weekdays at 06:00:

```bash
--schedule "0 6 * * 1-5"
```

Alias form:

```bash
--cron "@every 30m"
```

Runtime update through API:

```bash
curl -X POST -H "Authorization: Bearer ${DW_TOKEN}" "http://localhost:8080/v1/schedule?schedule=@every%2030m"
```
