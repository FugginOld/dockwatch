package updater

import (
	"time"

	"github.com/docker/docker/client"
	"github.com/fugginold/dockwatch/internal/actions"
	"github.com/fugginold/dockwatch/pkg/container"
	"github.com/fugginold/dockwatch/pkg/filters"
	t "github.com/fugginold/dockwatch/pkg/types"
	log "github.com/sirupsen/logrus"
)

const defaultTimeout = 10 * time.Second

// CheckForUpdates runs a one-shot scan using dockwatch's native update pipeline.
// The Docker SDK client argument is preserved for backward compatibility.
func CheckForUpdates(_ *client.Client) {
	dockerClient, err := container.NewClient(container.ClientOptions{})
	if err != nil {
		log.WithError(err).Error("failed to initialize Docker client")
		return
	}

	_, err = actions.Update(dockerClient, t.UpdateParams{
		Filter:  filters.NoFilter,
		Timeout: defaultTimeout,
	})
	if err != nil {
		log.WithError(err).Error("update run failed")
	}
}