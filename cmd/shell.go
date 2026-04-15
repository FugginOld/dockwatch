package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	t "github.com/fugginold/dockwatch/pkg/types"
	log "github.com/sirupsen/logrus"
)

// runShell provides an interactive prompt when the system is attached to a TTY.
func runShell(scheduleCtrl *scheduleController, updateLock chan bool, filter t.Filter, configFile string) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("dockwatch> ")

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "exit" || line == "quit" {
			break
		}

		parts := strings.Fields(line)
		if len(parts) > 0 {
			switch strings.ToLower(parts[0]) {
			case "schedule":
				if len(parts) > 1 {
					spec := strings.Join(parts[1:], " ")
					nextRun, err := setScheduleAndPersist(scheduleCtrl, spec, configFile)
					if err != nil {
						log.Errorf("Error setting schedule: %v", err)
					} else {
						log.Infof("Schedule set to %s", spec)
						if !nextRun.IsZero() {
							log.Infof("Next run scheduled for: %s", nextRun.Format(time.RFC1123))
						}
					}
				} else {
					log.Infof("Current schedule: %s", scheduleCtrl.Current())
					next := scheduleCtrl.NextRun()
					if !next.IsZero() {
						log.Infof("Next run at: %s", next.Format(time.RFC1123))
					}
				}
			case "update":
				select {
				case v := <-updateLock:
					log.Info("Starting manual update...")
					runUpdates(filter)
					log.Info("Manual update complete.")
					updateLock <- v
				default:
					log.Warn("An update is already running.")
				}
			case "help":
				fmt.Println("Available commands:")
				fmt.Println("  schedule <cron> - Set or view the container update schedule")
				fmt.Println("  update          - Trigger an immediate container update")
				fmt.Println("  exit, quit      - Exit the shell and Dockwatch")
			default:
				log.Errorf("Unknown command: %s. Type 'help' for a list of commands.", parts[0])
			}
		}

		fmt.Print("dockwatch> ")
	}

	if err := scanner.Err(); err != nil {
		log.Errorf("Error reading standard input: %v", err)
	}

	log.Info("Exiting interactive shell...")
}
