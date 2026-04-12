package actions

import (
	"fmt"
	"sort"
	"time"

	"github.com/fugginold/dockwatch/pkg/container"
	"github.com/fugginold/dockwatch/pkg/filters"
	"github.com/fugginold/dockwatch/pkg/sorter"
	"github.com/fugginold/dockwatch/pkg/types"

	log "github.com/sirupsen/logrus"
)

// CheckForSanity makes sure everything is sane before starting
func CheckForSanity(client container.Client, filter types.Filter, rollingRestarts bool) error {
	log.Debug("Making sure everything is sane before starting")

	if rollingRestarts {
		containers, err := client.ListContainers(filter)
		if err != nil {
			return err
		}
		for _, c := range containers {
			if len(c.Links()) > 0 {
				return fmt.Errorf(
					"%q is depending on at least one other container. This is not compatible with rolling restarts",
					c.Name(),
				)
			}
		}
	}
	return nil
}

// CheckForMultipleDockwatchInstances will ensure that there are not multiple instances of the
// dockwatch running simultaneously. If multiple dockwatch containers are detected, this function
// will stop and remove all but the most recently started container. This behaviour can be bypassed
// if a scope UID is defined.
func CheckForMultipleDockwatchInstances(client container.Client, cleanup bool, scope string) error {
	filter := filters.DockwatchContainersFilter
	if scope != "" {
		filter = filters.FilterByScope(scope, filter)
	}
	containers, err := client.ListContainers(filter)

	if err != nil {
		return err
	}

	if len(containers) <= 1 {
		log.Debug("There are no additional dockwatch containers")
		return nil
	}

	log.Info("Found multiple running dockwatch instances. Cleaning up.")
	return cleanupExcessDockwatchs(containers, client, cleanup)
}

func cleanupExcessDockwatchs(containers []types.Container, client container.Client, cleanup bool) error {
	var stopErrors int

	sort.Sort(sorter.ByCreated(containers))
	allContainersExceptLast := containers[0 : len(containers)-1]

	for _, c := range allContainersExceptLast {
		if err := client.StopContainer(c, 10*time.Minute); err != nil {
			// logging the original here as we're just returning a count
			log.WithError(err).Error("Could not stop a previous dockwatch instance.")
			stopErrors++
			continue
		}

		if cleanup {
			if err := client.RemoveImageByID(c.ImageID()); err != nil {
				log.WithError(err).Warning("Could not cleanup dockwatch images, possibly because of other dockwatchs instances in other scopes.")
			}
		}
	}

	if stopErrors > 0 {
		return fmt.Errorf("%d errors while stopping dockwatch containers", stopErrors)
	}

	return nil
}
