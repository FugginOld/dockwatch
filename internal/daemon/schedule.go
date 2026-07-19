package daemon

import (
	"sync"
	"time"

	t "github.com/fugginold/dockwatch/pkg/types"
	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"
)

// Controller owns the cron schedule that drives periodic update scans. Each
// tick delegates to the Runner, which enforces the single-flight guard.
type Controller struct {
	mu        sync.Mutex
	scheduler *cron.Cron
	runner    *Runner
	filter    t.Filter
	spec      string
}

// NewController returns a Controller that will run scans for filter via runner.
func NewController(runner *Runner, filter t.Filter) *Controller {
	return &Controller{runner: runner, filter: filter}
}

// Set installs (or replaces) the schedule and returns the next run time.
func (c *Controller) Set(spec string) (time.Time, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	scheduler := cron.New()
	err := scheduler.AddFunc(spec, func() {
		c.runner.TryRun(c.filter)

		nextRuns := scheduler.Entries()
		if len(nextRuns) > 0 {
			log.Debug("Scheduled next run: " + nextRuns[0].Next.String())
		}
	})
	if err != nil {
		return time.Time{}, err
	}

	nextRun := scheduler.Entries()[0].Schedule.Next(time.Now())
	scheduler.Start()

	if c.scheduler != nil {
		c.scheduler.Stop()
	}

	c.scheduler = scheduler
	c.spec = spec

	return nextRun, nil
}

// Current returns the active cron spec.
func (c *Controller) Current() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.spec
}

// NextRun returns the next scheduled run time, or the zero time if unscheduled.
func (c *Controller) NextRun() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.scheduler == nil {
		return time.Time{}
	}
	entries := c.scheduler.Entries()
	if len(entries) == 0 {
		return time.Time{}
	}
	return entries[0].Next
}

// Stop halts the schedule.
func (c *Controller) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.scheduler != nil {
		c.scheduler.Stop()
	}
}
