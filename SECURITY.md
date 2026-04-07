# Security Policy

## Supported Versions

Security updates will always only be applied to the latest version of Dockwatch.
As the software by default is set to auto-update if you use the `latest` tag, you will get these security updates automatically as soon as they are released.

## Reporting a Vulnerability

All security vulnerabilities, including critical ones, should be reported by opening an issue in the [GitHub repository issues section](https://github.com/FugginOld/dockwatch/issues).
We'll always try to get back to you as swiftly as possible, but keep in mind that since this is a community project, we can't really leave any guarantees about the speed.

## Current Dependency Alerts

Some dependency alerts may reference `github.com/docker/docker` or `github.com/moby/moby` vulnerabilities that affect Docker Engine daemon features such as AuthZ plugins or plugin installation flows.

Dockwatch uses Docker client and API packages, but does not implement or ship the affected daemon functionality. For these alerts, remediation is typically to update the Docker Engine version on the host running Docker rather than to change Dockwatch application code.
