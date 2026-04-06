package sorter_test

import (
	"sort"
	"testing"
	"time"

	dockerTypes "github.com/docker/docker/api/types"
	dockerContainer "github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/fugginold/dockwatch/internal/actions/mocks"
	"github.com/fugginold/dockwatch/pkg/container"
	"github.com/fugginold/dockwatch/pkg/sorter"
	"github.com/fugginold/dockwatch/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createContainerWithRFC3339Date creates a container whose Created field is formatted
// as RFC3339Nano so that ByCreated.Less can parse it correctly.
func createContainerWithRFC3339Date(id, name, image string, created time.Time) types.Container {
	content := &dockerTypes.ContainerJSON{
		ContainerJSONBase: &dockerTypes.ContainerJSONBase{
			ID:      id,
			Image:   image,
			Name:    name,
			Created: created.UTC().Format(time.RFC3339Nano),
			HostConfig: &dockerContainer.HostConfig{
				PortBindings: nat.PortMap{},
			},
		},
		Config: &dockerContainer.Config{
			Image:        image,
			Labels:       map[string]string{},
			ExposedPorts: nat.PortSet{},
		},
	}
	imageInfo := &dockerTypes.ImageInspect{
		ID:          image,
		RepoDigests: []string{image},
	}
	return container.NewContainer(content, imageInfo)
}

func TestByCreated_Len(t *testing.T) {
	containers := sorter.ByCreated{
		mocks.CreateMockContainer("id1", "c1", "image:latest", time.Now()),
		mocks.CreateMockContainer("id2", "c2", "image:latest", time.Now()),
	}
	assert.Equal(t, 2, containers.Len())
}

func TestByCreated_Len_Empty(t *testing.T) {
	containers := sorter.ByCreated{}
	assert.Equal(t, 0, containers.Len())
}

func TestByCreated_Swap(t *testing.T) {
	c1 := mocks.CreateMockContainer("id1", "c1", "image:latest", time.Now())
	c2 := mocks.CreateMockContainer("id2", "c2", "image:latest", time.Now())
	containers := sorter.ByCreated{c1, c2}
	containers.Swap(0, 1)
	assert.Equal(t, c2.ID(), containers[0].ID())
	assert.Equal(t, c1.ID(), containers[1].ID())
}

func TestByCreated_Less(t *testing.T) {
	now := time.Now()
	older := now.AddDate(0, 0, -1)

	c1 := createContainerWithRFC3339Date("id1", "c1", "image:latest", older)
	c2 := createContainerWithRFC3339Date("id2", "c2", "image:latest", now)
	containers := sorter.ByCreated{c1, c2}

	// older container (index 0) should be less than newer container (index 1)
	assert.True(t, containers.Less(0, 1))
	assert.False(t, containers.Less(1, 0))
}

func TestByCreated_Sort(t *testing.T) {
	now := time.Now()
	containers := sorter.ByCreated{
		createContainerWithRFC3339Date("id3", "c3", "image:latest", now),
		createContainerWithRFC3339Date("id1", "c1", "image:latest", now.AddDate(0, 0, -2)),
		createContainerWithRFC3339Date("id2", "c2", "image:latest", now.AddDate(0, 0, -1)),
	}

	sort.Sort(containers)

	assert.Equal(t, types.ContainerID("id1"), containers[0].ID())
	assert.Equal(t, types.ContainerID("id2"), containers[1].ID())
	assert.Equal(t, types.ContainerID("id3"), containers[2].ID())
}

func TestSortByDependencies_Empty(t *testing.T) {
	result, err := sorter.SortByDependencies([]types.Container{})
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestSortByDependencies_SingleContainer(t *testing.T) {
	c := mocks.CreateMockContainer("id1", "c1", "image:latest", time.Now())
	result, err := sorter.SortByDependencies([]types.Container{c})
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, c.ID(), result[0].ID())
}

func TestSortByDependencies_NoDependencies(t *testing.T) {
	containers := []types.Container{
		mocks.CreateMockContainer("id1", "c1", "image:latest", time.Now()),
		mocks.CreateMockContainer("id2", "c2", "image:latest", time.Now()),
		mocks.CreateMockContainer("id3", "c3", "image:latest", time.Now()),
	}

	result, err := sorter.SortByDependencies(containers)
	require.NoError(t, err)
	assert.Len(t, result, 3)
}

func TestSortByDependencies_WithLinks(t *testing.T) {
	// c2 depends on c1, so c1 should appear before c2 in the sorted result
	c1 := mocks.CreateMockContainer("id1", "c1", "image:latest", time.Now())
	c2 := mocks.CreateMockContainerWithLinks("id2", "c2", "image:latest", time.Now(), []string{"c1:c2"}, mocks.CreateMockImageInfo("image:latest"))

	// Present c2 before c1 - sort should reorder
	result, err := sorter.SortByDependencies([]types.Container{c2, c1})
	require.NoError(t, err)
	require.Len(t, result, 2)

	// c1 must come before c2
	var c1Pos, c2Pos int
	for i, c := range result {
		if c.ID() == "id1" {
			c1Pos = i
		}
		if c.ID() == "id2" {
			c2Pos = i
		}
	}
	assert.Less(t, c1Pos, c2Pos, "c1 should appear before c2 in sorted result")
}

func TestSortByDependencies_CircularReference(t *testing.T) {
	// c1 -> c2 -> c1 is circular
	c1 := mocks.CreateMockContainerWithLinks("id1", "c1", "image:latest", time.Now(), []string{"c2:c1"}, mocks.CreateMockImageInfo("image:latest"))
	c2 := mocks.CreateMockContainerWithLinks("id2", "c2", "image:latest", time.Now(), []string{"c1:c2"}, mocks.CreateMockImageInfo("image:latest"))

	_, err := sorter.SortByDependencies([]types.Container{c1, c2})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "circular reference")
}

func TestSortByDependencies_LinkedToMissing(t *testing.T) {
	// c1 links to a container not in the list - should not error, just skip
	c1 := mocks.CreateMockContainerWithLinks("id1", "c1", "image:latest", time.Now(), []string{"missing:c1"}, mocks.CreateMockImageInfo("image:latest"))

	result, err := sorter.SortByDependencies([]types.Container{c1})
	require.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestSortByDependencies_ChainedDependencies(t *testing.T) {
	// c3 -> c2 -> c1; expected order: c1, c2, c3
	c1 := mocks.CreateMockContainer("id1", "c1", "image:latest", time.Now())
	c2 := mocks.CreateMockContainerWithLinks("id2", "c2", "image:latest", time.Now(), []string{"c1:c2"}, mocks.CreateMockImageInfo("image:latest"))
	c3 := mocks.CreateMockContainerWithLinks("id3", "c3", "image:latest", time.Now(), []string{"c2:c3"}, mocks.CreateMockImageInfo("image:latest"))

	result, err := sorter.SortByDependencies([]types.Container{c3, c2, c1})
	require.NoError(t, err)
	require.Len(t, result, 3)

	positions := make(map[types.ContainerID]int, 3)
	for i, c := range result {
		positions[c.ID()] = i
	}
	assert.Less(t, positions[types.ContainerID("id1")], positions[types.ContainerID("id2")], "c1 should come before c2")
	assert.Less(t, positions[types.ContainerID("id2")], positions[types.ContainerID("id3")], "c2 should come before c3")
}
