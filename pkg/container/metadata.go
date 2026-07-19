package container

import "strconv"

const (
	dockwatchLabel         = "com.centurylinklabs.dockwatch"
	signalLabel            = "com.centurylinklabs.dockwatch.stop-signal"
	enableLabel            = "com.centurylinklabs.dockwatch.enable"
	monitorOnlyLabel       = "com.centurylinklabs.dockwatch.monitor-only"
	noPullLabel            = "com.centurylinklabs.dockwatch.no-pull"
	dependsOnLabel         = "com.centurylinklabs.dockwatch.depends-on"
	zodiacLabel            = "com.centurylinklabs.zodiac.original-image"
	scope                  = "com.centurylinklabs.dockwatch.scope"
	preCheckLabel          = "com.centurylinklabs.dockwatch.lifecycle.pre-check"
	postCheckLabel         = "com.centurylinklabs.dockwatch.lifecycle.post-check"
	preUpdateLabel         = "com.centurylinklabs.dockwatch.lifecycle.pre-update"
	postUpdateLabel        = "com.centurylinklabs.dockwatch.lifecycle.post-update"
	preUpdateTimeoutLabel  = "com.centurylinklabs.dockwatch.lifecycle.pre-update-timeout"
	postUpdateTimeoutLabel = "com.centurylinklabs.dockwatch.lifecycle.post-update-timeout"
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

func (c Container) getLabelValueOrEmpty(label string) string {
	if val, ok := c.containerInfo.Config.Labels[label]; ok {
		return val
	}
	return ""
}

func (c Container) getLabelValue(label string) (string, bool) {
	val, ok := c.containerInfo.Config.Labels[label]
	return val, ok
}

func (c Container) getBoolLabelValue(label string) (bool, error) {
	if strVal, ok := c.containerInfo.Config.Labels[label]; ok {
		value, err := strconv.ParseBool(strVal)
		return value, err
	}
	return false, errorLabelNotFound
}
