By default, Dockwatch will clean up other instances and won't allow multiple instances running on the same Docker host or swarm. It is possible to override this behavior by defining a [scope](arguments.md#filter_by_scope) to each running instance.

!!! note
    - Multiple instances can't run with the same scope;
    - An instance without a scope will clean up other running instances, even if they have a defined scope;
    - Supplying `none` as the scope will treat `com.centurylinklabs.dockwatch.scope=none`, `com.centurylinklabs.dockwatch.scope=` and the lack of a `com.centurylinklabs.dockwatch.scope` label as the scope `none`. This effectively enables you to run both scoped and unscoped dockwatch instances on the same machine.

To define an instance monitoring scope, use the `--scope` argument or the `DOCKWATCH_SCOPE` environment variable on startup and set the `com.centurylinklabs.dockwatch.scope` label with the same value for the containers you want to include in this instance's scope (including the instance itself).

For example, in a Docker Compose config file:

```yaml
services:
  app-with-scope:
    image: myapps/monitored-by-dockwatch
    labels: [ "com.centurylinklabs.dockwatch.scope=myscope" ]

  scoped-dockwatch:
    image: fugginold/dockwatch
    volumes: [ "/var/run/docker.sock:/var/run/docker.sock" ]
    command: --interval 30 --scope myscope
    labels: [ "com.centurylinklabs.dockwatch.scope=myscope" ] 

  unscoped-app-a:
    image: myapps/app-a

  unscoped-app-b:
    image: myapps/app-b
    labels: [ "com.centurylinklabs.dockwatch.scope=none" ]
    
  unscoped-app-c:
    image: myapps/app-b
    labels: [ "com.centurylinklabs.dockwatch.scope=" ]
    
  unscoped-dockwatch:
    image: fugginold/dockwatch
    volumes: [ "/var/run/docker.sock:/var/run/docker.sock" ]
    command: --interval 30 --scope none
```
