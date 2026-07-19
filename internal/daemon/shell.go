package daemon

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"time"

	t "github.com/fugginold/dockwatch/pkg/types"
	log "github.com/sirupsen/logrus"
)

// RunShell provides the interactive prompt when dockwatch is attached to a TTY.
// Input and output are injected so the shell can be driven in tests.
func RunShell(sched *Controller, runner *Runner, filter t.Filter, in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	fmt.Fprint(out, "dockwatch> ")

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "exit" || line == "quit" {
			break
		}

		parts := strings.Fields(line)
		if len(parts) > 0 {
			switch strings.ToLower(parts[0]) {
			case "schedule":
				handleScheduleCommand(sched, parts)
			case "update":
				log.Info("Starting manual update...")
				if _, ran := runner.TryRun(filter); ran {
					log.Info("Manual update complete.")
				} else {
					log.Warn("An update is already running.")
				}
			case "help":
				fmt.Fprintln(out, "Available commands:")
				fmt.Fprintln(out, "  schedule <cron> - Set or view the container update schedule")
				fmt.Fprintln(out, "  update          - Trigger an immediate container update")
				fmt.Fprintln(out, "  exit, quit      - Exit the shell and Dockwatch")
			default:
				log.Errorf("Unknown command: %s. Type 'help' for a list of commands.", parts[0])
			}
		}

		fmt.Fprint(out, "dockwatch> ")
	}

	if err := scanner.Err(); err != nil {
		log.Errorf("Error reading standard input: %v", err)
	}

	log.Info("Exiting interactive shell...")
}

func handleScheduleCommand(sched *Controller, parts []string) {
	if len(parts) > 1 {
		spec := strings.Join(parts[1:], " ")
		nextRun, err := sched.Set(spec)
		if err != nil {
			log.Errorf("Error setting schedule: %v", err)
			return
		}
		log.Infof("Schedule set to %s", spec)
		if !nextRun.IsZero() {
			log.Infof("Next run scheduled for: %s", nextRun.Format(time.RFC1123))
		}
		return
	}

	log.Infof("Current schedule: %s", sched.Current())
	if next := sched.NextRun(); !next.IsZero() {
		log.Infof("Next run at: %s", next.Format(time.RFC1123))
	}
}
