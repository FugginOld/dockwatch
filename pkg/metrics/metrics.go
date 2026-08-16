package metrics

import (
	"sync"

	"github.com/fugginold/dockwatch/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	metrics     *Metrics
	defaultOnce sync.Once
	// registerer is the collector registry the singleton builds into. It is a
	// variable so a test can swap in a fresh registry and rebuild from scratch --
	// without that, the duplicate-registration panic this Once exists to prevent
	// cannot be reproduced in a test at all.
	registerer prometheus.Registerer = prometheus.DefaultRegisterer
)

// Metric is the data points of a single scan
type Metric struct {
	Scanned int
	Updated int
	Failed  int
}

// Metrics is the handler processing all individual scan metrics
type Metrics struct {
	channel chan *Metric
	scanned prometheus.Gauge
	updated prometheus.Gauge
	failed  prometheus.Gauge
	total   prometheus.Counter
	skipped prometheus.Counter
}

// NewMetric returns a Metric with the counts taken from the appropriate types.Report fields
func NewMetric(report types.Report) *Metric {
	return &Metric{
		Scanned: len(report.Scanned()),
		// Note: This is for backwards compatibility. ideally, stale containers should be counted separately
		Updated: len(report.Updated()) + len(report.Stale()),
		Failed:  len(report.Failed()),
	}
}

// QueueIsEmpty checks whether any messages are enqueued in the channel
func (metrics *Metrics) QueueIsEmpty() bool {
	return len(metrics.channel) == 0
}

// Register registers metrics for an executed scan
func (metrics *Metrics) Register(metric *Metric) {
	metrics.channel <- metric
}

// Default creates a new metrics handler if none exists, otherwise returns the existing one.
//
// The cron goroutine and a finishing scan's goroutine both reach this, so a plain
// nil check let both build the singleton: the second promauto call then panics the
// daemon with "duplicate metrics collector registration attempted", and two
// HandleUpdate goroutines end up racing for the same channel.
func Default() *Metrics {
	defaultOnce.Do(buildDefault)
	return metrics
}

func buildDefault() {
	metrics = &Metrics{
		scanned: promauto.With(registerer).NewGauge(prometheus.GaugeOpts{
			Name: "dockwatch_containers_scanned",
			Help: "Number of containers scanned for changes by dockwatch during the last scan",
		}),
		updated: promauto.With(registerer).NewGauge(prometheus.GaugeOpts{
			Name: "dockwatch_containers_updated",
			Help: "Number of containers updated by dockwatch during the last scan",
		}),
		failed: promauto.With(registerer).NewGauge(prometheus.GaugeOpts{
			Name: "dockwatch_containers_failed",
			Help: "Number of containers where update failed during the last scan",
		}),
		total: promauto.With(registerer).NewCounter(prometheus.CounterOpts{
			Name: "dockwatch_scans_total",
			Help: "Number of scans since the dockwatch started",
		}),
		skipped: promauto.With(registerer).NewCounter(prometheus.CounterOpts{
			Name: "dockwatch_scans_skipped",
			Help: "Number of skipped scans since dockwatch started",
		}),
		channel: make(chan *Metric, 10),
	}

	go metrics.HandleUpdate(metrics.channel)
}

// RegisterScan fetches a metric handler and enqueues a metric
func RegisterScan(metric *Metric) {
	metrics := Default()
	metrics.Register(metric)
}

// HandleUpdate dequeue the metric channel and processes it
func (metrics *Metrics) HandleUpdate(channel <-chan *Metric) {
	for change := range channel {
		if change == nil {
			// Update was skipped and rescheduled
			metrics.total.Inc()
			metrics.skipped.Inc()
			metrics.scanned.Set(0)
			metrics.updated.Set(0)
			metrics.failed.Set(0)
			continue
		}
		// Update metrics with the new values
		metrics.total.Inc()
		metrics.scanned.Set(float64(change.Scanned))
		metrics.updated.Set(float64(change.Updated))
		metrics.failed.Set(float64(change.Failed))
	}
}
