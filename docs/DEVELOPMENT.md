# Development

## Local Setup

## VS Code Dev Container

This repository includes a Debian Bookworm-based Dev Container for the most reliable VS Code setup, especially on Windows.

Prerequisites:

- VS Code with the Dev Containers extension
- Docker Desktop with Linux containers enabled
- On Windows, WSL 2 is recommended for best filesystem and Docker performance

Open the repo in the container:

```bash
code .
```

Then run `Dev Containers: Reopen in Container` from the VS Code command palette. The container uses Go 1.25 on Debian Bookworm, installs GitHub CLI, and provisions Docker CLI access so the existing Docker-based smoke tests can run from inside VS Code.

If you prefer WSL on Windows, clone the repository inside your WSL distribution and open it from there with:

```bash
code .
```

That path also works with the included Dev Container and keeps Git operations synced with GitHub through the normal repository remote.

Prerequisites:

- Go 1.25+
- Docker Engine
- Optional: Python with MkDocs for docs preview

Clone and prepare:

```bash
git clone https://github.com/FugginOld/dockwatch.git
cd dockwatch
go mod download
```

Build:

```bash
go build ./...
```

Run locally against Docker socket:

```bash
go run . --schedule "@every 24h"
```

## Testing Scripts

Run all Go tests:

```bash
go test ./...
```

Smoke test script:

```bash
./scripts/test-dockwatch.sh
```

Bounded smoke test:

```bash
TEST_DURATION_SECONDS=25 ./scripts/test-dockwatch.sh
```

CI wrapper:

```bash
./scripts/test-dockwatch-ci.sh
```

Other helper scripts:

- scripts/lifecycle-tests.sh
- scripts/dependency-test.sh
- scripts/contnet-tests.sh

## Debugging tips

- Start with debug logs:

```bash
go run . --debug --schedule "@every 10m"
```

- Use trace only when needed (may expose sensitive data):

```bash
go run . --trace --run-once
```

- Validate API mode quickly:

```bash
go run . --http-api-update --http-api-periodic-polls --schedule "@every 24h"
```

- If behavior seems wrong, verify effective flags and env inputs first.
- Reproduce with --run-once or --force-update to isolate one execution path.
