package daemon

import (
	"testing"
	"time"

	"github.com/fugginold/dockwatch/pkg/filters"
	t "github.com/fugginold/dockwatch/pkg/types"
)

// blockingClient is a container.Client whose ListContainers blocks until
// released, letting a test hold a scan "in flight" and probe the guard.
type blockingClient struct {
	entered chan struct{}
	release chan struct{}
}

func (c *blockingClient) Ping() error { return nil }

func (c *blockingClient) ListContainers(t.Filter) ([]t.Container, error) {
	select {
	case c.entered <- struct{}{}:
	default:
	}
	<-c.release
	return nil, nil
}

func (c *blockingClient) GetContainer(t.ContainerID) (t.Container, error)   { return nil, nil }
func (c *blockingClient) StopContainer(t.Container, time.Duration) error    { return nil }
func (c *blockingClient) StartContainer(t.Container) (t.ContainerID, error) { return "", nil }
func (c *blockingClient) RenameContainer(t.Container, string) error         { return nil }
func (c *blockingClient) RemoveImageByID(t.ImageID) error                   { return nil }
func (c *blockingClient) WarnOnHeadPullFailed(t.Container) bool             { return false }
func (c *blockingClient) ExecuteCommand(t.ContainerID, string, int) (bool, error) {
	return false, nil
}
func (c *blockingClient) IsContainerStale(t.Container, t.UpdateParams) (bool, t.ImageID, error) {
	return false, "", nil
}

func TestRunnerSingleFlight(t *testing.T) {
	client := &blockingClient{
		entered: make(chan struct{}, 1),
		release: make(chan struct{}),
	}
	r := NewRunner(Config{Client: client})

	go r.Run(filters.NoFilter) // acquires the guard, then blocks in ListContainers
	<-client.entered           // the scan is now in flight

	if _, ran := r.TryRun(filters.NoFilter); ran {
		t.Fatal("TryRun ran while a scan was already in flight")
	}

	close(client.release) // let the in-flight scan (and any later one) complete
	r.Wait()              // block until the guard is free again

	if _, ran := r.TryRun(filters.NoFilter); !ran {
		t.Fatal("TryRun skipped even though the guard was free")
	}
}
