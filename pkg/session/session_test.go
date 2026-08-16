package session_test

import (
	"errors"
	"testing"
	"time"

	"github.com/fugginold/dockwatch/internal/actions/mocks"
	"github.com/fugginold/dockwatch/pkg/session"
	"github.com/fugginold/dockwatch/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// helpers

func newContainer(id, name, image string) types.Container {
	return mocks.CreateMockContainer(id, name, image, time.Now())
}

// ContainerStatus via UpdateFromContainer

func TestUpdateFromContainer(t *testing.T) {
	c := newContainer("cid1", "test-container", "myimage:latest")
	newImg := types.ImageID("newimgid")

	status := session.UpdateFromContainer(c, newImg, session.ScannedState)

	assert.Equal(t, c.ID(), status.ID())
	assert.Equal(t, "test-container", status.Name())
	assert.Equal(t, "myimage:latest", status.ImageName())
	assert.Equal(t, newImg, status.LatestImageID())
	assert.Equal(t, "", status.Error())
}

func TestContainerStatus_State_AllValues(t *testing.T) {
	c := newContainer("cid1", "c1", "image:latest")

	tests := []struct {
		state    session.State
		expected string
	}{
		{session.SkippedState, "Skipped"},
		{session.ScannedState, "Scanned"},
		{session.UpdatedState, "Updated"},
		{session.FailedState, "Failed"},
		{session.FreshState, "Fresh"},
		{session.StaleState, "Stale"},
		{session.UnknownState, "Unknown"},
	}

	for _, tt := range tests {
		status := session.UpdateFromContainer(c, c.SafeImageID(), tt.state)
		assert.Equal(t, tt.expected, status.State(), "state=%v", tt.state)
	}
}

// Progress tests

func TestProgress_AddScanned(t *testing.T) {
	p := session.Progress{}
	c := newContainer("cid1", "c1", "image:latest")
	newImg := types.ImageID("newimg")

	p.AddScanned(c, newImg)

	require.Len(t, p, 1)
	assert.Equal(t, "Scanned", p[c.ID()].State())
}

func TestProgress_AddSkipped(t *testing.T) {
	p := session.Progress{}
	c := newContainer("cid1", "c1", "image:latest")
	err := errors.New("skipped because of something")

	p.AddSkipped(c, err)

	require.Len(t, p, 1)
	assert.Equal(t, "Skipped", p[c.ID()].State())
	assert.Equal(t, "skipped because of something", p[c.ID()].Error())
}

func TestProgress_AddSkipped_NilError(t *testing.T) {
	p := session.Progress{}
	c := newContainer("cid1", "c1", "image:latest")

	p.AddSkipped(c, nil)

	assert.Equal(t, "Skipped", p[c.ID()].State())
	assert.Equal(t, "", p[c.ID()].Error())
}

func TestProgress_MarkForUpdate(t *testing.T) {
	p := session.Progress{}
	c := newContainer("cid1", "c1", "image:latest")
	p.AddScanned(c, types.ImageID("newimg"))

	p.MarkForUpdate(c.ID())

	assert.Equal(t, "Updated", p[c.ID()].State())
}

// actions.Update calls MarkForUpdate for every container that is not monitor-only,
// including ones AddSkipped already recorded because their staleness check failed.
// Overwriting the skip there loses the only record that the check never ran.
func TestProgress_MarkForUpdateDoesNotClobberSkipped(t *testing.T) {
	p := session.Progress{}
	c := newContainer("cid1", "c1", "image:latest")
	p.AddSkipped(c, errors.New("pull failed: unauthorized"))

	p.MarkForUpdate(c.ID())

	assert.Equal(t, "Skipped", p[c.ID()].State())

	report := session.NewReport(p)
	assert.Len(t, report.Skipped(), 1, "a skipped container must stay in Skipped")
	assert.Empty(t, report.Fresh(), "a container whose staleness check failed is not known to be fresh")
}

// A container stopped and removed as part of another container's update, which then
// fails to start again, keeps its original image -- but it is not up to date, it is
// deleted and not running.
func TestReport_FailedIsNotReportedAsFresh(t *testing.T) {
	p := session.Progress{}
	c := newContainer("cid1", "linked-db", "image:latest")
	p.AddScanned(c, c.SafeImageID()) // image never changed

	p.UpdateFailed(map[types.ContainerID]error{
		c.ID(): errors.New(`Conflict. The container name "/linked-db" is already in use`),
	})

	report := session.NewReport(p)

	assert.Len(t, report.Failed(), 1, "a failed container must be reported as failed")
	assert.Empty(t, report.Fresh())
}

func TestProgress_UpdateFailed(t *testing.T) {
	p := session.Progress{}
	c := newContainer("cid1", "c1", "image:latest")
	p.AddScanned(c, types.ImageID("newimg"))

	failures := map[types.ContainerID]error{
		c.ID(): errors.New("update failed"),
	}
	p.UpdateFailed(failures)

	assert.Equal(t, "Failed", p[c.ID()].State())
	assert.Equal(t, "update failed", p[c.ID()].Error())
}

func TestProgress_Report(t *testing.T) {
	p := session.Progress{}
	c := newContainer("cid1", "c1", "image:latest")
	p.AddScanned(c, types.ImageID("newimg"))

	report := p.Report()
	assert.NotNil(t, report)
}

// NewReport tests

func TestNewReport_Empty(t *testing.T) {
	p := session.Progress{}
	r := session.NewReport(p)

	assert.Empty(t, r.Scanned())
	assert.Empty(t, r.Updated())
	assert.Empty(t, r.Failed())
	assert.Empty(t, r.Skipped())
	assert.Empty(t, r.Stale())
	assert.Empty(t, r.Fresh())
	assert.Empty(t, r.All())
}

func TestNewReport_FreshContainer(t *testing.T) {
	// same old and new image => Fresh
	p := session.Progress{}
	c := newContainer("cid1", "c1", "image:latest")
	p.AddScanned(c, c.SafeImageID())

	r := session.NewReport(p)

	assert.Len(t, r.Scanned(), 1)
	assert.Len(t, r.Fresh(), 1)
	assert.Empty(t, r.Stale())
	assert.Empty(t, r.Updated())
}

func TestNewReport_StaleContainer(t *testing.T) {
	// different old and new image, not marked updated => Stale
	p := session.Progress{}
	c := newContainer("cid1", "c1", "image:latest")
	p.AddScanned(c, types.ImageID("different-image-id"))

	r := session.NewReport(p)

	assert.Len(t, r.Scanned(), 1)
	assert.Len(t, r.Stale(), 1)
	assert.Empty(t, r.Fresh())
	assert.Empty(t, r.Updated())
}

func TestNewReport_UpdatedContainer(t *testing.T) {
	p := session.Progress{}
	c := newContainer("cid1", "c1", "image:latest")
	p.AddScanned(c, types.ImageID("different-image-id"))
	p.MarkForUpdate(c.ID())

	r := session.NewReport(p)

	assert.Len(t, r.Updated(), 1)
	assert.Empty(t, r.Stale())
}

func TestNewReport_FailedContainer(t *testing.T) {
	p := session.Progress{}
	c := newContainer("cid1", "c1", "image:latest")
	p.AddScanned(c, types.ImageID("different-image-id"))
	p.UpdateFailed(map[types.ContainerID]error{c.ID(): errors.New("boom")})

	r := session.NewReport(p)

	assert.Len(t, r.Failed(), 1)
	assert.Empty(t, r.Updated())
	assert.Equal(t, "boom", r.Failed()[0].Error())
}

func TestNewReport_SkippedContainer(t *testing.T) {
	p := session.Progress{}
	c := newContainer("cid1", "c1", "image:latest")
	p.AddSkipped(c, nil)

	r := session.NewReport(p)

	assert.Len(t, r.Skipped(), 1)
	assert.Empty(t, r.Scanned())
}

func TestNewReport_All_Deduplication(t *testing.T) {
	p := session.Progress{}
	c1 := newContainer("cid1", "c1", "image:latest")
	c2 := newContainer("cid2", "c2", "image:latest")
	c3 := newContainer("cid3", "c3", "image:latest")

	// c1 fresh, c2 stale, c3 skipped
	p.AddScanned(c1, c1.SafeImageID())
	p.AddScanned(c2, types.ImageID("different"))
	p.AddSkipped(c3, nil)

	r := session.NewReport(p)
	all := r.All()

	// All should contain each container exactly once
	assert.Len(t, all, 3)

	seen := make(map[types.ContainerID]int)
	for _, cr := range all {
		seen[cr.ID()]++
	}
	for id, count := range seen {
		assert.Equal(t, 1, count, "container %s appeared more than once in All()", id)
	}
}

func TestNewReport_ContainerReport_Fields(t *testing.T) {
	p := session.Progress{}
	c := newContainer("cid1", "c1", "image:latest")
	p.AddScanned(c, types.ImageID("different-image-id"))
	p.MarkForUpdate(c.ID())

	r := session.NewReport(p)
	require.Len(t, r.Updated(), 1)

	cr := r.Updated()[0]
	assert.Equal(t, c.ID(), cr.ID())
	assert.Equal(t, "c1", cr.Name())
	assert.NotEmpty(t, cr.ImageName())
}
