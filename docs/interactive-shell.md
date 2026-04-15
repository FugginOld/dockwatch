# Interactive Shell

Dockwatch ships natively with an interactive command-line shell environment. This environment replaces traditional daemon-backgrounding and gives you a real-time `dockwatch>` prompt to control the update lifecycle exactly when you want to.

## Automatic Activation

There are no special flags or configurations required to use the shell. By default, Dockwatch checks whether standard input is attached to a real interactive TTY.

If it detects an interactive terminal (e.g., you ran the binary directly from Bash/PowerShell or ran `docker run -it`), it will automatically start the interactive shell:

```text
dockwatch>
```

## Available Commands

The following commands can currently be executed from within the interactive shell:

- **`schedule <cron spec>`**  
  Modifies the container update schedule at runtime and saves it to the runtime config file. If no cron specification is provided, it acts as a getter and returns your currently configured schedule.  
  _Example:_ `schedule @every 5m`

- **`update`**  
  Triggers a forceful dockwatch scan immediately, performing a container update iteration without waiting for the scheduler or having to use the HTTP API.

- **`help`**  
  Returns a helpful list of available terminal commands.

- **`exit`** (or **`quit`**)  
  Sends an interactive termination signal to the application. It acts identically to sending a `SIGTERM` / `SIGINT`; the scheduler is gracefully halted, any running container updates are drained/completed first, and the process safely exits.
