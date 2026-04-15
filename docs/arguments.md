# Arguments

By default, dockwatch will monitor all containers running within the Docker daemon to which it is pointed (in most cases this
will be the local Docker daemon, but you can override it with the `--host` option described in the next section). However, you
can restrict dockwatch to monitoring a subset of the running containers by specifying the container names as arguments when
launching dockwatch.

```bash

$ docker run -d \
    --name dockwatch \
    -v /var/run/docker.sock:/var/run/docker.sock \
    fugginold/dockwatch \
    nginx redis

```

In the example above, dockwatch will only monitor the containers named "nginx" and "redis" for updates -- all of the other
running containers will be ignored. If you do not want dockwatch to run as a daemon you can pass the `--run-once` flag and remove
the dockwatch container after its execution.

```bash

$ docker run --rm \
    -v /var/run/docker.sock:/var/run/docker.sock \
    fugginold/dockwatch \
    --run-once \
    nginx redis

```

In the example above, dockwatch will execute an upgrade attempt on the containers named "nginx" and "redis". Using this mode will enable debugging output showing all actions performed, as usage is intended for interactive users. Once the attempt is completed, the container will exit and remove itself due to the `--rm` flag.

If you want a clearer command name for one-shot runs, use `--force-update` (alias for `--run-once`).

When no arguments are specified, dockwatch will monitor all running containers.

## Secrets/Files

Some arguments can also reference a file, in which case the contents of the file are used as the value.
This can be used to avoid putting secrets in the configuration file or command line.

The following arguments are currently supported (including their corresponding `DOCKWATCH_` environment variables):

- `http-api-token`

### Example docker-compose usage

```yaml

secrets:
  access_token:
    file: access_token

services:
  dockwatch:
    secrets:
      - access_token
    environment:
      - DOCKWATCH_HTTP_API_TOKEN=/run/secrets/access_token

```

## Help

Shows documentation about the supported flags.

```text

            Argument: --help
Environment Variable: N/A
                Type: N/A
             Default: N/A

```

## Time Zone

Sets the time zone to be used by Dockwatch's logs and the optional Cron scheduling argument (--schedule). If this environment variable is not set, Dockwatch will use the default time zone: UTC.
To find out the right value, see [this list](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones), find your location and use the value in _TZ Database Name_, e.g _Europe/Rome_. The timezone can alternatively be set by volume mounting your hosts /etc/localtime file. `-v /etc/localtime:/etc/localtime:ro`

```text

            Argument: N/A
Environment Variable: TZ
                Type: String
             Default: "UTC"

```

## Cleanup

Removes old images after updating. When this flag is specified, dockwatch will remove the old image after restarting a container with a new image. Use this option to prevent the accumulation of orphaned images on your system as containers are updated.

```text

            Argument: --cleanup
Environment Variable: DOCKWATCH_CLEANUP
                Type: Boolean
             Default: false

```

## Remove anonymous volumes

Removes anonymous volumes after updating. When this flag is specified, dockwatch will remove all anonymous volumes from the container before restarting with a new image. Named volumes will not be removed!

```text

            Argument: --remove-volumes
Environment Variable: DOCKWATCH_REMOVE_VOLUMES
                Type: Boolean
             Default: false

```

## Debug

Enable debug mode with verbose logging.

!!! note "Notes"  
   Alias for `--log-level debug`. See [Maximum log level](#maximum-log-level).  
    Does _not_ take an argument when used as an argument. Using `--debug true` will **not** work.

```text

            Argument: --debug, -d
Environment Variable: DOCKWATCH_DEBUG
                Type: Boolean
             Default: false

```

## Trace

Enable trace mode with very verbose logging. Caution: exposes credentials!

!!! note "Notes"  
   Alias for `--log-level trace`. See [Maximum log level](#maximum-log-level).  
    Does _not_ take an argument when used as an argument. Using `--trace true` will **not** work.

```text

            Argument: --trace
Environment Variable: DOCKWATCH_TRACE
                Type: Boolean
             Default: false

```

## Maximum log level

The maximum log level that will be written to STDERR (shown in `docker log` when used in a container).

```text

            Argument: --log-level
Environment Variable: DOCKWATCH_LOG_LEVEL
     Possible values: panic, fatal, error, warn, info, debug or trace
             Default: info

```

## Logging format

Sets what logging format to use for console output.

```text

            Argument: --log-format, -l
Environment Variable: DOCKWATCH_LOG_FORMAT
     Possible values: Auto, LogFmt, Pretty or JSON
             Default: Auto

```

## ANSI colors

Disable ANSI color escape codes in log output.

```text

            Argument: --no-color
Environment Variable: NO_COLOR
                Type: Boolean
             Default: false

```

## Docker host

Docker daemon socket to connect to. Can be pointed at a remote Docker host by specifying a TCP endpoint as `tcp://hostname:port`.

```text

            Argument: --host, -H
Environment Variable: DOCKER_HOST
                Type: String
             Default: "unix:///var/run/docker.sock"

```

## Docker API version

The API version to use by the Docker client for connecting to the Docker daemon. The minimum supported version is 1.25.

```text

            Argument: --api-version, -a
Environment Variable: DOCKER_API_VERSION
                Type: String
             Default: "1.25"

```

## Registry TLS skip verify

Skip TLS certificate verification for registry HEAD requests. This should only be used for testing or trusted private networks.

```text

            Argument: --registry-tls-skip-verify
Environment Variable: DOCKWATCH_REGISTRY_TLS_SKIP_VERIFY
                Type: Boolean
             Default: false

```

## Include restarting

Will also include restarting containers.

```text

            Argument: --include-restarting
Environment Variable: DOCKWATCH_INCLUDE_RESTARTING
                Type: Boolean
             Default: false

```

## Include stopped

Will also include created and exited containers.

```text

            Argument: --include-stopped, -S
Environment Variable: DOCKWATCH_INCLUDE_STOPPED
                Type: Boolean
             Default: false

```

## Revive stopped

Start any stopped containers that have had their image updated. This argument is only usable with the `--include-stopped` argument.

```text

            Argument: --revive-stopped
Environment Variable: DOCKWATCH_REVIVE_STOPPED
                Type: Boolean
             Default: false

```

## Poll interval

Poll interval (in seconds). This value controls how frequently dockwatch will poll for new images. Either `--schedule` or a poll interval can be defined, but not both.

```text

            Argument: --interval, -i
Environment Variable: DOCKWATCH_POLL_INTERVAL
                Type: Integer
             Default: 86400 (24 hours)

```

## Filter by enable label

Monitor and update containers that have a `com.centurylinklabs.dockwatch.enable` label set to true.

```text

            Argument: --label-enable
Environment Variable: DOCKWATCH_LABEL_ENABLE
                Type: Boolean
             Default: false

```

## Filter by disable label

**Do not** Monitor and update containers that have `com.centurylinklabs.dockwatch.enable` label set to false and
no `--label-enable` argument is passed. Note that only one or the other (targeting by enable label) can be
used at the same time to target containers.

## Filter by disabling specific container names

Monitor and update containers whose names are not in a given set of names.

This can be used to exclude specific containers, when setting labels is not an option.
The listed containers will be excluded even if they have the enable filter set to true.

```text

            Argument: --disable-containers, -x
Environment Variable: DOCKWATCH_DISABLE_CONTAINERS
                Type: Comma- or space-separated string list
             Default: ""

```

## Without updating containers

Will only monitor for new images and invoke
the [pre-check/post-check hooks](lifecycle-hooks.md), but will **not** update the
containers.

!!! note
    Due to Docker API limitations the latest image will still be pulled from the registry.
    The HEAD digest checks allows dockwatch to skip pulling when there are no changes, but to know _what_ has changed it
    will still do a pull whenever the repository digest doesn't match the local image digest.

```text

            Argument: --monitor-only
Environment Variable: DOCKWATCH_MONITOR_ONLY
                Type: Boolean
             Default: false

```

Note that monitor-only can also be specified on a per-container basis with the `com.centurylinklabs.dockwatch.monitor-only` label set on those containers.

See [With label taking precedence over arguments](#with-label-taking-precedence-over-arguments) for behavior when both argument and label are set

## With label taking precedence over arguments

By default, arguments will take precedence over labels. This means that if you set `DOCKWATCH_MONITOR_ONLY` to true or use `--monitor-only`, a container with `com.centurylinklabs.dockwatch.monitor-only` set to false will not be updated. If you set `DOCKWATCH_LABEL_TAKE_PRECEDENCE` to true or use `--label-take-precedence`, then the container will also be updated. This also apply to the no pull option. if you set `DOCKWATCH_NO_PULL` to true or use `--no-pull`, a container with `com.centurylinklabs.dockwatch.no-pull` set to false will not pull the new image. If you set `DOCKWATCH_LABEL_TAKE_PRECEDENCE` to true or use `--label-take-precedence`, then the container will pull image

```text

            Argument: --label-take-precedence
Environment Variable: DOCKWATCH_LABEL_TAKE_PRECEDENCE
                Type: Boolean
             Default: false

```

## Without restarting containers

Do not restart containers after updating. This option can be useful when the start of the containers
is managed by an external system such as systemd.

```text

            Argument: --no-restart
Environment Variable: DOCKWATCH_NO_RESTART
                Type: Boolean
             Default: false

```

## Without pulling new images

Do not pull new images. When this flag is specified, dockwatch will not attempt to pull
new images from the registry. Instead it will only monitor the local image cache for changes.
Use this option if you are building new images directly on the Docker host without pushing
them to a registry.

```text

            Argument: --no-pull
Environment Variable: DOCKWATCH_NO_PULL
                Type: Boolean
             Default: false

```

Note that no-pull can also be specified on a per-container basis with the
`com.centurylinklabs.dockwatch.no-pull` label set on those containers.

See [With label taking precedence over arguments](#with-label-taking-precedence-over-arguments) for behavior when both argument and label are set

## Without sending a startup message

Do not send a startup message after dockwatch started.

```text

            Argument: --no-startup-message
Environment Variable: DOCKWATCH_NO_STARTUP_MESSAGE
                Type: Boolean
             Default: false

```

## Run once

Run a single update check immediately and exit.

```text

            Argument: --run-once, -R
Environment Variable: DOCKWATCH_RUN_ONCE
                Type: Boolean
             Default: false

```

## Force update

Alias for `--run-once`. Triggers an immediate update check and exits instead of waiting for the next scheduled poll.

```text

            Argument: --force-update
Environment Variable: N/A
                Type: Boolean
             Default: false

```

## Cron

Alias for `--schedule`.

```text

            Argument: --cron
Environment Variable: N/A
                Type: String
             Default: ""

```

## HTTP API Mode

Runs Dockwatch in HTTP API mode, only allowing image updates to be triggered by an HTTP request.
For details see [HTTP API](http-api-mode.md).

```text

            Argument: --http-api-update
Environment Variable: DOCKWATCH_HTTP_API_UPDATE
                Type: Boolean
             Default: false

```

## HTTP API Token

Sets an authentication token to HTTP API requests.
Can also reference a file, in which case the contents of the file are used.

```text

            Argument: --http-api-token
Environment Variable: DOCKWATCH_HTTP_API_TOKEN
                Type: String
             Default: -

```

## HTTP API periodic polls

Keep running periodic updates if the HTTP API mode is enabled, otherwise the HTTP API would prevent periodic polls.  

```text

            Argument: --http-api-periodic-polls
Environment Variable: DOCKWATCH_HTTP_API_PERIODIC_POLLS
                Type: Boolean
             Default: false

```

## Filter by scope

Update containers that have a `com.centurylinklabs.dockwatch.scope` label set with the same value as the given argument.
This enables [running multiple instances](running-multiple-instances.md).

!!! note "Filter by lack of scope"
    If you want other instances of dockwatch to ignore the scoped containers, set this argument to `none`.
    When omitted, dockwatch will update all containers regardless of scope.

```text

            Argument: --scope
Environment Variable: DOCKWATCH_SCOPE
                Type: String
             Default: -

```

## Scheduling

[Cron expression](https://pkg.go.dev/github.com/robfig/cron@v1.2.0?tab=doc#hdr-CRON_Expression_Format) in 6 fields (rather than the traditional 5) which defines when and how often to check for new images. Either `--interval` or the schedule expression
can be defined, but not both. An example: `--schedule "0 0 4 * * *"`

```text

            Argument: --schedule, -s
Environment Variable: DOCKWATCH_SCHEDULE
                Type: String
             Default: -

```

## Runtime config file

Path to the runtime config file used to persist interactive schedule changes.

```text

            Argument: --config-file
Environment Variable: DOCKWATCH_CONFIG_FILE
                Type: String
             Default: /config/dockwatch.json

```

## Rolling restart

Restart one image at time instead of stopping and starting all at once.  Useful in conjunction with lifecycle hooks
to implement zero-downtime deploy.

```text

            Argument: --rolling-restart
Environment Variable: DOCKWATCH_ROLLING_RESTART
                Type: Boolean
             Default: false

```

## Wait until timeout

Timeout before the container is forcefully stopped. When set, this option will change the default (`10s`) wait time to the given value. An example: `--stop-timeout 30s` will set the timeout to 30 seconds.

```text

            Argument: --stop-timeout
Environment Variable: DOCKWATCH_TIMEOUT
                Type: Duration
             Default: 10s

```

## TLS Verification

Use TLS when connecting to the Docker socket and verify the server's certificate.

```text

            Argument: --tlsverify
Environment Variable: DOCKER_TLS_VERIFY
                Type: Boolean
             Default: false

```

## HEAD failure warnings

When to warn about HEAD pull requests failing. Auto means that it will warn when the registry is known to handle the
requests and may rate limit pull requests (mainly docker.io).

```text

            Argument: --warn-on-head-failure
Environment Variable: DOCKWATCH_WARN_ON_HEAD_FAILURE
     Possible values: always, auto, never
             Default: auto

```

## Health check

Returns a success exit code to enable usage with docker `HEALTHCHECK`. This check is naive and only checks whether there is another process running inside the container, as it is the only known form of failure state for a dockwatch container.

!!! note "Only for HEALTHCHECK use"
    Never put this on the main container executable command line as it is only meant to be run from docker HEALTHCHECK.

```text

            Argument: --health-check

```

## Programatic Output (porcelain)

Writes the session results to STDOUT using a stable, machine-readable format (indicated by the argument VERSION).

```text
            Argument: --porcelain, -P
Environment Variable: DOCKWATCH_PORCELAIN
     Possible values: v1
             Default: -
```
