package lifecycle_test

import (
	"errors"
	"testing"
	"time"

	"github.com/fugginold/dockwatch/internal/actions/mocks"
	"github.com/fugginold/dockwatch/pkg/container"
	"github.com/fugginold/dockwatch/pkg/lifecycle"
	"github.com/fugginold/dockwatch/pkg/types"
	"github.com/stretchr/testify/assert"

	dockerContainer "github.com/docker/docker/api/types/container"
)

const (
	preCheckLabel   = "com.centurylinklabs.dockwatch.lifecycle.pre-check"
	postCheckLabel  = "com.centurylinklabs.dockwatch.lifecycle.post-check"
	preUpdateLabel  = "com.centurylinklabs.dockwatch.lifecycle.pre-update"
	postUpdateLabel = "com.centurylinklabs.dockwatch.lifecycle.post-update"
)

// errorClient is a minimal container.Client that always fails ListContainers.
type errorClient struct {
	mocks.MockClient
}

func (e errorClient) ListContainers(_ types.Filter) ([]types.Container, error) {
	return nil, errors.New("docker down")
}

// Ensure errorClient satisfies container.Client at compile time.
var _ container.Client = errorClient{}

func containerWithLabels(labels map[string]string) types.Container {
	config := &dockerContainer.Config{
		Image:  "image:latest",
		Labels: labels,
	}
	return mocks.CreateMockContainerWithConfig("cid1", "c1", "image:latest", true, false, time.Now(), config)
}

// ExecutePreCheckCommand

func TestExecutePreCheckCommand_NoCommand(t *testing.T) {
	c := containerWithLabels(map[string]string{})
	client := mocks.CreateMockClient(&mocks.TestData{}, false, false)

	// Should not panic and not call ExecuteCommand (no command label set)
	assert.NotPanics(t, func() {
		lifecycle.ExecutePreCheckCommand(client, c)
	})
}

func TestExecutePreCheckCommand_WithCommand(t *testing.T) {
	c := containerWithLabels(map[string]string{
		preCheckLabel: "/PreUpdateReturn0.sh",
	})
	client := mocks.CreateMockClient(&mocks.TestData{Containers: []types.Container{c}}, false, false)

	assert.NotPanics(t, func() {
		lifecycle.ExecutePreCheckCommand(client, c)
	})
}

func TestExecutePreCheckCommand_WithCommandError(t *testing.T) {
	c := containerWithLabels(map[string]string{
		preCheckLabel: "/PreUpdateReturn1.sh",
	})
	client := mocks.CreateMockClient(&mocks.TestData{Containers: []types.Container{c}}, false, false)

	// Should not panic even if ExecuteCommand returns an error
	assert.NotPanics(t, func() {
		lifecycle.ExecutePreCheckCommand(client, c)
	})
}

// ExecutePostCheckCommand

func TestExecutePostCheckCommand_NoCommand(t *testing.T) {
	c := containerWithLabels(map[string]string{})
	client := mocks.CreateMockClient(&mocks.TestData{}, false, false)

	assert.NotPanics(t, func() {
		lifecycle.ExecutePostCheckCommand(client, c)
	})
}

func TestExecutePostCheckCommand_WithCommand(t *testing.T) {
	c := containerWithLabels(map[string]string{
		postCheckLabel: "/PreUpdateReturn0.sh",
	})
	client := mocks.CreateMockClient(&mocks.TestData{Containers: []types.Container{c}}, false, false)

	assert.NotPanics(t, func() {
		lifecycle.ExecutePostCheckCommand(client, c)
	})
}

// ExecutePreChecks

func TestExecutePreChecks_ListContainersError(t *testing.T) {
	client := errorClient{}

	// Should not panic when ListContainers fails
	assert.NotPanics(t, func() {
		lifecycle.ExecutePreChecks(client, types.UpdateParams{})
	})
}

func TestExecutePreChecks_EmptyContainers(t *testing.T) {
	client := mocks.CreateMockClient(&mocks.TestData{Containers: []types.Container{}}, false, false)

	assert.NotPanics(t, func() {
		lifecycle.ExecutePreChecks(client, types.UpdateParams{})
	})
}

func TestExecutePreChecks_WithContainers(t *testing.T) {
	c := containerWithLabels(map[string]string{
		preCheckLabel: "/PreUpdateReturn0.sh",
	})
	client := mocks.CreateMockClient(&mocks.TestData{Containers: []types.Container{c}}, false, false)

	assert.NotPanics(t, func() {
		lifecycle.ExecutePreChecks(client, types.UpdateParams{})
	})
}

// ExecutePostChecks

func TestExecutePostChecks_EmptyContainers(t *testing.T) {
	client := mocks.CreateMockClient(&mocks.TestData{Containers: []types.Container{}}, false, false)

	assert.NotPanics(t, func() {
		lifecycle.ExecutePostChecks(client, types.UpdateParams{})
	})
}

func TestExecutePostChecks_ListContainersError(t *testing.T) {
	client := errorClient{}

	assert.NotPanics(t, func() {
		lifecycle.ExecutePostChecks(client, types.UpdateParams{})
	})
}

// ExecutePreUpdateCommand

func TestExecutePreUpdateCommand_NoCommand(t *testing.T) {
	c := containerWithLabels(map[string]string{})
	client := mocks.CreateMockClient(&mocks.TestData{Containers: []types.Container{c}}, false, false)

	skip, err := lifecycle.ExecutePreUpdateCommand(client, c)

	assert.NoError(t, err)
	assert.False(t, skip)
}

func TestExecutePreUpdateCommand_NotRunning(t *testing.T) {
	config := &dockerContainer.Config{
		Image:  "image:latest",
		Labels: map[string]string{preUpdateLabel: "/PreUpdateReturn0.sh"},
	}
	c := mocks.CreateMockContainerWithConfig("cid1", "c1", "image:latest", false, false, time.Now(), config)
	client := mocks.CreateMockClient(&mocks.TestData{Containers: []types.Container{c}}, false, false)

	skip, err := lifecycle.ExecutePreUpdateCommand(client, c)

	assert.NoError(t, err)
	assert.False(t, skip)
}

func TestExecutePreUpdateCommand_IsRestarting(t *testing.T) {
	config := &dockerContainer.Config{
		Image:  "image:latest",
		Labels: map[string]string{preUpdateLabel: "/PreUpdateReturn0.sh"},
	}
	c := mocks.CreateMockContainerWithConfig("cid1", "c1", "image:latest", true, true, time.Now(), config)
	client := mocks.CreateMockClient(&mocks.TestData{Containers: []types.Container{c}}, false, false)

	skip, err := lifecycle.ExecutePreUpdateCommand(client, c)

	assert.NoError(t, err)
	assert.False(t, skip)
}

func TestExecutePreUpdateCommand_RunningWithCommand_NoSkip(t *testing.T) {
	config := &dockerContainer.Config{
		Image:  "image:latest",
		Labels: map[string]string{preUpdateLabel: "/PreUpdateReturn0.sh"},
	}
	c := mocks.CreateMockContainerWithConfig("cid1", "c1", "image:latest", true, false, time.Now(), config)
	client := mocks.CreateMockClient(&mocks.TestData{Containers: []types.Container{c}}, false, false)

	skip, err := lifecycle.ExecutePreUpdateCommand(client, c)

	assert.NoError(t, err)
	assert.False(t, skip)
}

func TestExecutePreUpdateCommand_RunningWithCommand_Skip(t *testing.T) {
	// /PreUpdateReturn75.sh causes ExecuteCommand to return skip=true
	config := &dockerContainer.Config{
		Image:  "image:latest",
		Labels: map[string]string{preUpdateLabel: "/PreUpdateReturn75.sh"},
	}
	c := mocks.CreateMockContainerWithConfig("cid1", "c1", "image:latest", true, false, time.Now(), config)
	client := mocks.CreateMockClient(&mocks.TestData{Containers: []types.Container{c}}, false, false)

	skip, err := lifecycle.ExecutePreUpdateCommand(client, c)

	assert.NoError(t, err)
	assert.True(t, skip)
}

func TestExecutePreUpdateCommand_RunningWithCommand_Error(t *testing.T) {
	// /PreUpdateReturn1.sh causes ExecuteCommand to return an error
	config := &dockerContainer.Config{
		Image:  "image:latest",
		Labels: map[string]string{preUpdateLabel: "/PreUpdateReturn1.sh"},
	}
	c := mocks.CreateMockContainerWithConfig("cid1", "c1", "image:latest", true, false, time.Now(), config)
	client := mocks.CreateMockClient(&mocks.TestData{Containers: []types.Container{c}}, false, false)

	skip, err := lifecycle.ExecutePreUpdateCommand(client, c)

	assert.Error(t, err)
	assert.False(t, skip)
}

// ExecutePostUpdateCommand

func TestExecutePostUpdateCommand_NoCommand(t *testing.T) {
	c := containerWithLabels(map[string]string{})
	client := mocks.CreateMockClient(&mocks.TestData{Containers: []types.Container{c}}, false, false)

	assert.NotPanics(t, func() {
		lifecycle.ExecutePostUpdateCommand(client, c.ID())
	})
}

func TestExecutePostUpdateCommand_WithCommand(t *testing.T) {
	config := &dockerContainer.Config{
		Image:  "image:latest",
		Labels: map[string]string{postUpdateLabel: "/PreUpdateReturn0.sh"},
	}
	c := mocks.CreateMockContainerWithConfig("cid1", "c1", "image:latest", true, false, time.Now(), config)
	client := mocks.CreateMockClient(&mocks.TestData{Containers: []types.Container{c}}, false, false)

	assert.NotPanics(t, func() {
		lifecycle.ExecutePostUpdateCommand(client, c.ID())
	})
}

func TestExecutePostUpdateCommand_WithCommandError(t *testing.T) {
	config := &dockerContainer.Config{
		Image:  "image:latest",
		Labels: map[string]string{postUpdateLabel: "/PreUpdateReturn1.sh"},
	}
	c := mocks.CreateMockContainerWithConfig("cid1", "c1", "image:latest", true, false, time.Now(), config)
	client := mocks.CreateMockClient(&mocks.TestData{Containers: []types.Container{c}}, false, false)

	assert.NotPanics(t, func() {
		lifecycle.ExecutePostUpdateCommand(client, c.ID())
	})
}
