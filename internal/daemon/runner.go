package daemon

import (
	"github.com/fugginold/dockwatch/internal/actions"
	"github.com/fugginold/dockwatch/pkg/metrics"
	t "github.com/fugginold/dockwatch/pkg/types"
	log "github.com/sirupsen/logrus"
)

// Runner executes container update scans and guarantees that only one scan
// runs at a time. It owns the single-flight guard that was previously
// hand-rolled as a `chan bool` in the scheduler, the shell and the HTTP API.
type Runner struct {
	cfg Config
	sem chan struct{} // capacity 1; a token present means a scan is in flight
}

// NewRunner returns a Runner bound to cfg with no scan in flight.
func NewRunner(cfg Config) *Runner {
	return &Runner{cfg: cfg, sem: make(chan struct{}, 1)}
}

// Run blocks until the single-flight guard is free, runs a scan for filter,
// registers the scan metric and returns it.
func (r *Runner) Run(filter t.Filter) *metrics.Metric {
	r.sem <- struct{}{}
	defer func() { <-r.sem }()
	return r.scan(filter)
}

// TryRun runs a scan only when no other scan is in flight. It returns the
// scan metric and whether it ran; a skipped run registers an empty scan.
func (r *Runner) TryRun(filter t.Filter) (*metrics.Metric, bool) {
	select {
	case r.sem <- struct{}{}:
		defer func() { <-r.sem }()
		return r.scan(filter), true
	default:
		metrics.RegisterScan(nil)
		log.Debug("Skipped another update already running.")
		return nil, false
	}
}

// Wait blocks until any in-flight scan has finished. Used at shutdown.
func (r *Runner) Wait() {
	r.sem <- struct{}{}
	<-r.sem
}

func (r *Runner) scan(filter t.Filter) *metrics.Metric {
	result, err := actions.Update(r.cfg.Client, r.cfg.updateParams(filter))
	if err != nil {
		log.Error(err)
	}
	metric := metrics.NewMetric(result)
	log.WithFields(log.Fields{
		"Scanned": metric.Scanned,
		"Updated": metric.Updated,
		"Failed":  metric.Failed,
	}).Info("Session done")
	metrics.RegisterScan(metric)
	return metric
}
