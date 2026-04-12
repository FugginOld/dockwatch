By default, dockwatch is set-up to monitor the local Docker daemon (the same daemon running the dockwatch container itself). However, it is possible to configure dockwatch to monitor a remote Docker endpoint. When starting the dockwatch container you can specify a remote Docker endpoint with either the `--host` flag or the `DOCKER_HOST` environment variable:

```bash
docker run -d \
  --name dockwatch \
  fugginold/dockwatch --host "tcp://10.0.1.2:2375"
```

or

```bash
docker run -d \
  --name dockwatch \
  -e DOCKER_HOST="tcp://10.0.1.2:2375" \
  fugginold/dockwatch
```

Note in both of the examples above that it is unnecessary to mount the _/var/run/docker.sock_ into the dockwatch container.
