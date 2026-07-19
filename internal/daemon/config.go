package daemon

import (
	"errors"
	"time"

	"github.com/fugginold/dockwatch/pkg/container"
	t "github.com/fugginold/dockwatch/pkg/types"
)

// Config is the resolved runtime configuration for a dockwatch run. It is a
// plain value with no package-level mutable state: build it once, hand it to
// New, and its validation rules live in one testable place.
type Config struct {
	Client            container.Client
	ScheduleSpec      string
	Cleanup           bool
	NoRestart         bool
	NoPull            bool
	MonitorOnly       bool
	EnableLabel       bool
	LifecycleHooks    bool
	RollingRestart    bool
	LabelPrecedence   bool
	Scope             string
	DisableContainers []string
	Timeout           time.Duration
}

// Validate reports configuration combinations that make a run illegal.
func (c Config) Validate() error {
	if c.Timeout < 0 {
		return errors.New("please specify a positive value for timeout value")
	}
	if c.RollingRestart && c.MonitorOnly {
		return errors.New("rolling restarts is not compatible with the global monitor only flag")
	}
	return nil
}

// updateParams builds the per-scan parameters for the given filter.
func (c Config) updateParams(filter t.Filter) t.UpdateParams {
	return t.UpdateParams{
		Filter:          filter,
		Cleanup:         c.Cleanup,
		NoRestart:       c.NoRestart,
		Timeout:         c.Timeout,
		MonitorOnly:     c.MonitorOnly,
		LifecycleHooks:  c.LifecycleHooks,
		RollingRestart:  c.RollingRestart,
		LabelPrecedence: c.LabelPrecedence,
		NoPull:          c.NoPull,
	}
}
