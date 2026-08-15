package mocks

import (
	"errors"
	"fmt"
	"time"

	t "github.com/fugginold/dockwatch/pkg/types"
)

// MockClient is a mock that passes as a dockwatch Client
type MockClient struct {
	TestData      *TestData
	pullImages    bool
	removeVolumes bool
}

// TestData is the data used to perform the test
type TestData struct {
	TriedToRemoveImageCount int
	NameOfContainerToKeep   string
	Containers              []t.Container
	Staleness               map[string]bool
	// StalenessError maps a container name to the error its staleness check should
	// return, for exercising the paths where dockwatch cannot determine staleness.
	StalenessError map[string]error
	// StopErrors maps a container name to the error its stop should return, for
	// exercising how the update flow reacts to a specific stop failure.
	StopErrors map[string]error
	// StartedContainers records the name of every container StartContainer was
	// called for, so a test can assert that a container dockwatch failed to stop
	// was not started again.
	StartedContainers []string
}

// TriedToRemoveImage is a test helper function to check whether RemoveImageByID has been called
func (testdata *TestData) TriedToRemoveImage() bool {
	return testdata.TriedToRemoveImageCount > 0
}

// CreateMockClient creates a mock dockwatch Client for usage in tests
func CreateMockClient(data *TestData, pullImages bool, removeVolumes bool) MockClient {
	return MockClient{
		data,
		pullImages,
		removeVolumes,
	}
}

// Ping is a mock method that always succeeds
func (client MockClient) Ping() error {
	return nil
}

// ListContainers is a mock method returning the provided container testdata
func (client MockClient) ListContainers(_ t.Filter) ([]t.Container, error) {
	return client.TestData.Containers, nil
}

// StopContainer is a mock method
func (client MockClient) StopContainer(c t.Container, _ time.Duration) error {
	if err, ok := client.TestData.StopErrors[c.Name()]; ok {
		// Wrapped on purpose: production wraps the sentinel with the container name,
		// so a consumer comparing with == instead of errors.Is would pass here while
		// failing against the real client.
		return fmt.Errorf("mock stop failed: %w", err)
	}
	if c.Name() == client.TestData.NameOfContainerToKeep {
		return errors.New("tried to stop the instance we want to keep")
	}
	return nil
}

// StartContainer is a mock method that records which containers were started
func (client MockClient) StartContainer(c t.Container) (t.ContainerID, error) {
	client.TestData.StartedContainers = append(client.TestData.StartedContainers, c.Name())
	return "", nil
}

// RenameContainer is a mock method
func (client MockClient) RenameContainer(_ t.Container, _ string) error {
	return nil
}

// RemoveImageByID increments the TriedToRemoveImageCount on being called
func (client MockClient) RemoveImageByID(_ t.ImageID) error {
	client.TestData.TriedToRemoveImageCount++
	return nil
}

// GetContainer is a mock method
func (client MockClient) GetContainer(_ t.ContainerID) (t.Container, error) {
	return client.TestData.Containers[0], nil
}

// ExecuteCommand is a mock method
func (client MockClient) ExecuteCommand(_ t.ContainerID, command string, _ int) (SkipUpdate bool, err error) {
	switch command {
	case "/PreUpdateReturn0.sh":
		return false, nil
	case "/PreUpdateReturn1.sh":
		return false, fmt.Errorf("command exited with code 1")
	case "/PreUpdateReturn75.sh":
		return true, nil
	default:
		return false, nil
	}
}

// IsContainerStale is true if not explicitly stated in TestData for the mock client.
// A container named in StalenessError fails its check instead, as it would when the
// registry is unreachable or the credentials have expired.
func (client MockClient) IsContainerStale(cont t.Container, params t.UpdateParams) (bool, t.ImageID, error) {
	if err, found := client.TestData.StalenessError[cont.Name()]; found {
		return false, "", err
	}

	stale, found := client.TestData.Staleness[cont.Name()]
	if !found {
		stale = true
	}
	return stale, "", nil
}

// WarnOnHeadPullFailed is always true for the mock client
func (client MockClient) WarnOnHeadPullFailed(_ t.Container) bool {
	return true
}
