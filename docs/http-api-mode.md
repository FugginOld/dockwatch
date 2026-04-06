Dockwatch provides an HTTP API mode that enables an HTTP endpoint that can be requested to trigger container updating. The current available endpoint list is:

-   `/v1/update` - triggers an update for all of the containers monitored by this Dockwatch instance.

---

To enable this mode, use the flag `--http-api-update`. For example, in a Docker Compose config file:

```yaml
version: '3'

services:
  app-monitored-by-dockwatch:
    image: myapps/monitored-by-dockwatch
    labels:
      - "com.centurylinklabs.dockwatch.enable=true"

  dockwatch:
    image: fugginold/dockwatch
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    command: --debug --http-api-update
    environment:
      - DOCKWATCH_HTTP_API_TOKEN=mytoken
    labels:
      - "com.centurylinklabs.dockwatch.enable=false"
    ports:
      - 8080:8080
```

By default, enabling this mode prevents periodic polls (i.e. what is specified using `--interval` or `--schedule`). To run periodic updates regardless, pass `--http-api-periodic-polls`.

Notice that there is an environment variable named DOCKWATCH_HTTP_API_TOKEN. To prevent external services from accidentally triggering image updates, all of the requests have to contain a "Token" field, valued as the token defined in DOCKWATCH_HTTP_API_TOKEN, in their headers. In this case, there is a port bind to the host machine, allowing to request localhost:8080 to reach Dockwatch. The following `curl` command would trigger an image update:

```bash
curl -H "Authorization: Bearer mytoken" localhost:8080/v1/update
```

---

In order to update only certain images, the image names can be provided as URL query parameters. The following `curl` command would trigger an update for the images `foo/bar` and `foo/baz`:

```bash
curl -H "Authorization: Bearer mytoken" localhost:8080/v1/update?image=foo/bar,foo/baz
```
