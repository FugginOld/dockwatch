# Container Selection

By default, dockwatch will watch all containers. However, sometimes only some containers should be updated.

There are two options:

- **Fully exclude**: You can choose to exclude containers entirely from being watched by dockwatch.
- **Monitor only**: In this mode, dockwatch checks for container updates and invokes the [pre-check/post-check hooks](lifecycle-hooks.md) on the containers but does **not** perform the update.

## Full Exclude

If you need to exclude some containers, set the _com.centurylinklabs.dockwatch.enable_ label to `false`.  For clarity this should be set **on the container(s)** you wish to be ignored, this is not set on dockwatch.

=== "dockerfile"

    ```docker
    LABEL com.centurylinklabs.dockwatch.enable="false"
    ```
=== "docker run"

    ```bash
    docker run -d --label=com.centurylinklabs.dockwatch.enable=false someimage
    ```

=== "docker-compose"

    ``` yaml
    services:
      someimage:
        container_name: someimage
        labels:
          - "com.centurylinklabs.dockwatch.enable=false"
    ```

If instead you want to [only include containers with the enable label](arguments.md#filter_by_enable_label), pass the `--label-enable` flag or the `DOCKWATCH_LABEL_ENABLE` environment variable on startup for dockwatch and set the _com.centurylinklabs.dockwatch.enable_ label with a value of `true` on the containers you want to watch.

=== "dockerfile"

    ```docker
    LABEL com.centurylinklabs.dockwatch.enable="true"
    ```
=== "docker run"

    ```bash
    docker run -d --label=com.centurylinklabs.dockwatch.enable=true someimage
    ```

=== "docker-compose"

    ``` yaml
    services:
      someimage:
        container_name: someimage
        labels:
          - "com.centurylinklabs.dockwatch.enable=true"
    ```

If you wish to create a monitoring scope, you will need to [run multiple instances and set a scope for each of them](running-multiple-instances.md).

Dockwatch filters running containers by testing them against each configured criteria. A container is monitored if all criteria are met. For example:

- If a container's name is on the monitoring name list (not empty `--name` argument) but it is not enabled (_centurylinklabs.dockwatch.enable=false_), it won't be monitored;
- If a container's name is not on the monitoring name list (not empty `--name` argument), even if it is enabled (_centurylinklabs.dockwatch.enable=true_ and `--label-enable` flag is set), it won't be monitored;

## Monitor Only

Individual containers can be marked to only be monitored (without being updated).

To do so, set the _com.centurylinklabs.dockwatch.monitor-only_ label to `true` on that container.

```docker
LABEL com.centurylinklabs.dockwatch.monitor-only="true"
```

Or, it can be specified as part of the `docker run` command line:

```bash
docker run -d --label=com.centurylinklabs.dockwatch.monitor-only=true someimage
```

When the label is specified on a container, dockwatch treats that container exactly as if [`DOCKWATCH_MONITOR_ONLY`](arguments.md#without_updating_containers) was set, but the effect is limited to the individual container.
