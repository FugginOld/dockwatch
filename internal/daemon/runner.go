package daemon

import (
	"github.com/fugginold/dockwatch/internal/actions"
	"github.com/fugginold/dockwatch/pkg/metrics"
	t "github.com/fugginold/dockwatch/pkg/types"
	log "github.com/sirupsen/logrus"
)

// maxQueuedScans bounds how many callers may hold a place in the scan queue --
// the one running scan holds a slot too, so it is 1 running plus 7 waiting.
// Each waiter is a parked goroutine holding an HTTP connection, and every one of
// them runs a full scan when its turn comes -- so an unbounded queue turns a burst
// of update requests into a backlog of scans that outlives the callers that asked
// for them. Queued work is redundant anyway: a scan that has not started yet will
// pick up whatever the earlier ones would have.
const maxQueuedScans = 8

// Runner executes container update scans and guarantees that only one scan
// runs at a time. It owns the single-flight guard that was previously
// hand-rolled as a `chan bool` in the scheduler, the shell and the HTTP API.
type Runner struct {
	cfg   Config
	sem   chan struct{} // capacity 1; a token present means a scan is in flight
	queue chan struct{} // capacity maxQueuedScans; bounds the waiters on sem
}

// NewRunner returns a Runner bound to cfg with no scan in flight.
func NewRunner(cfg Config) *Runner {
	return &Runner{
		cfg:   cfg,
		sem:   make(chan struct{}, 1),
		queue: make(chan struct{}, maxQueuedScans),
	}
}

// Run blocks until the single-flight guard is free, runs a scan for filter,
// registers the scan metric and returns it. It returns nil without scanning when
// maxQueuedScans callers are already waiting.
func (r *Runner) Run(filter t.Filter) *metrics.Metric {
	select {
	case r.queue <- struct{}{}:
		defer func() { <-r.queue }()
	default:
		metrics.RegisterScan(nil)
		log.Warn("Skipped update: too many scans already queued.")
		return nil
	}

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
