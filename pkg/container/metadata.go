package container

import "strconv"

const (
	dockwatchLabel         = "io.github.fugginold.dockwatch"
	signalLabel            = "io.github.fugginold.dockwatch.stop-signal"
	enableLabel            = "io.github.fugginold.dockwatch.enable"
	monitorOnlyLabel       = "io.github.fugginold.dockwatch.monitor-only"
	noPullLabel            = "io.github.fugginold.dockwatch.no-pull"
	dependsOnLabel         = "io.github.fugginold.dockwatch.depends-on"
	scope                  = "io.github.fugginold.dockwatch.scope"
	preCheckLabel          = "io.github.fugginold.dockwatch.lifecycle.pre-check"
	postCheckLabel         = "io.github.fugginold.dockwatch.lifecycle.post-check"
	preUpdateLabel         = "io.github.fugginold.dockwatch.lifecycle.pre-update"
	postUpdateLabel        = "io.github.fugginold.dockwatch.lifecycle.post-update"
	preUpdateTimeoutLabel  = "io.github.fugginold.dockwatch.lifecycle.pre-update-timeout"
	postUpdateTimeoutLabel = "io.github.fugginold.dockwatch.lifecycle.post-update-timeout"
)

// GetLifecyclePreCheckCommand returns the pre-check command set in the container metadata or an empty string
func (c Container) GetLifecyclePreCheckCommand() string {
	return c.getLabelValueOrEmpty(preCheckLabel)
}

// GetLifecyclePostCheckCommand returns the post-check command set in the container metadata or an empty string
func (c Container) GetLifecyclePostCheckCommand() string {
	return c.getLabelValueOrEmpty(postCheckLabel)
}

// GetLifecyclePreUpdateCommand returns the pre-update command set in the container metadata or an empty string
func (c Container) GetLifecyclePreUpdateCommand() string {
	return c.getLabelValueOrEmpty(preUpdateLabel)
}

// GetLifecyclePostUpdateCommand returns the post-update command set in the container metadata or an empty string
func (c Container) GetLifecyclePostUpdateCommand() string {
	return c.getLabelValueOrEmpty(postUpdateLabel)
}

// ContainsDockwatchLabel takes a map of labels and values and tells
// the consumer whether it contains a valid dockwatch instance label
func ContainsDockwatchLabel(labels map[string]string) bool {
	val, ok := labels[dockwatchLabel]
	return ok && val == "true"
}

// labels returns the container's labels, or nil if the daemon gave us no info or
// no config for it. Reads from a nil map are well defined, so putting the guard
// here fixes every label reader at once -- and a scan reads labels on every
// container, so a panic here takes the whole daemon down with it.
func (c Container) labels() map[string]string {
	if c.containerInfo == nil || c.containerInfo.Config == nil {
		return nil
	}
	return c.containerInfo.Config.Labels
}

func (c Container) getLabelValueOrEmpty(label string) string {
	if val, ok := c.labels()[label]; ok {
		return val
	}
	return ""
}

func (c Container) getLabelValue(label string) (string, bool) {
	val, ok := c.labels()[label]
	return val, ok
}

func (c Container) getBoolLabelValue(label string) (bool, error) {
	if strVal, ok := c.labels()[label]; ok {
		value, err := strconv.ParseBool(strVal)
		return value, err
	}
	return false, errorLabelNotFound
}
