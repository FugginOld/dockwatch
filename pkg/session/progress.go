package session

import (
	"github.com/fugginold/dockwatch/pkg/types"
)

// Progress contains the current session container status
type Progress map[types.ContainerID]*ContainerStatus

// UpdateFromContainer sets various status fields from their corresponding container equivalents
func UpdateFromContainer(cont types.Container, newImage types.ImageID, state State) *ContainerStatus {
	return &ContainerStatus{
		containerID:   cont.ID(),
		containerName: cont.Name(),
		imageName:     cont.ImageName(),
		oldImage:      cont.SafeImageID(),
		newImage:      newImage,
		state:         state,
	}
}

// AddSkipped adds a container to the Progress with the state set as skipped
func (m Progress) AddSkipped(cont types.Container, err error) {
	update := UpdateFromContainer(cont, cont.SafeImageID(), SkippedState)
	update.error = err
	m.Add(update)
}

// AddScanned adds a container to the Progress with the state set as scanned
func (m Progress) AddScanned(cont types.Container, newImage types.ImageID) {
	m.Add(UpdateFromContainer(cont, newImage, ScannedState))
}

// UpdateFailed updates the containers passed, setting their state as failed with the supplied error
func (m Progress) UpdateFailed(failures map[types.ContainerID]error) {
	for id, err := range failures {
		update, ok := m[id]
		if !ok {
			continue
		}
		update.error = err
		update.state = FailedState
	}
}

// Add a container to the map using container ID as the key
func (m Progress) Add(update *ContainerStatus) {
	m[update.containerID] = update
}

// MarkForUpdate marks the container identified by containerID for update.
//
// A container already recorded as skipped keeps that state: actions.Update marks
// every container that is not monitor-only, including ones whose staleness check
// failed, and those were never established to be up to date.
func (m Progress) MarkForUpdate(containerID types.ContainerID) {
	update, ok := m[containerID]
	if !ok || update.state == SkippedState {
		return
	}
	update.state = UpdatedState
}

// MarkSkipped marks a container that was queued for update as intentionally skipped.
// Used when a pre-update lifecycle hook defers the update via EX_TEMPFAIL.
func (m Progress) MarkSkipped(id types.ContainerID, err error) {
	if update, ok := m[id]; ok {
		update.error = err
		update.state = SkippedState
	}
}

// Report creates a new Report from a Progress instance
func (m Progress) Report() types.Report {
	return NewReport(m)
}
